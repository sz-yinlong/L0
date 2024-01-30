package main

import (
	"L0/cache"
	"L0/jsonImporter"
	natsTestMessage "L0/nats-test-message"
	"L0/resources/config"
	model "L0/resources/dbmodels"
	"L0/server"
	"L0/utility"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"

	_ "github.com/lib/pq"
	"github.com/nats-io/stan.go"
)

var (
	db         *sql.DB
	sc         stan.Conn
	cfg        *config.Config
	orderCache *cache.OrderCache
	port       string
)

func init() {
	cfg = config.NewConfig()
	orderCache = cache.NewOrderCache()

	port = cfg.Port
	if port == "" {
		log.Fatalf("PORT must be set")
	}

	var err error
	db, err = setupDatabase(cfg)
	if err != nil {
		log.Fatalf("Error setting up database: %v", err)
	}

	sc, err = setupNatsStreaming(cfg)
	if err != nil {
		log.Fatalf("Error setting up NATS Streaming: %v", err)
	}

	jsonImporter.ImportJson(db, "resources/json/model.json", orderCache)
	if err := orderCache.LoadOrdersIntoCache(db); err != nil {
		log.Fatalf("Failed to load orders into cache: %v", err)
	}
}

func main() {

	readyChan := make(chan struct{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		go server.StartServer(port, orderCache, db)
		log.Println("HTTP server is running on port", port)

		orderUID := "b563feb7b2b84b6test"
		order, found := orderCache.Get(orderUID)
		if found {
			fmt.Printf("Order found in cache: %+v\n", order)
		} else {
			fmt.Printf("Order not found in cache, loading from DB")
		}
		go handleMessages(ctx, sc, db, orderCache)
		close(readyChan)
	}()
	<-readyChan

	natsTestMessage.MessageTest(sc)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	cancel()
}

func handleMessages(ctx context.Context, sc stan.Conn, db *sql.DB, cache *cache.OrderCache) {
	sub, err := sc.Subscribe("my_channel", func(msg *stan.Msg) {
		log.Printf("Received message: %s", string(msg.Data))

		var order model.Order
		err := json.Unmarshal(msg.Data, &order)
		if err != nil {
			log.Printf("Error unmarshalling message: %v\n", err)
			return
		}
		err = utility.SaveOrder(db, cache, &order)
		if err != nil {
			log.Printf("Error saving order: %v", err)
			return
		}
	})
	if err != nil {
		fmt.Println("Error creating subscriber", err)
		return
	}
	defer sub.Unsubscribe()
	go func() {
		<-ctx.Done()
		sub.Unsubscribe()
	}()
}

func setupNatsStreaming(cfg *config.Config) (stan.Conn, error) {
	clientID := cfg.NATSClientID
	clusterID := cfg.NATSClusterID
	natsURL := cfg.NATSURL
	return stan.Connect(clusterID, clientID, stan.NatsURL(natsURL))
}

func setupDatabase(cfg *config.Config) (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	return db, nil
}

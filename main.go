package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/nats-io/stan.go"

	"github.com/sz-yinlong/L0/cache"
	"github.com/sz-yinlong/L0/jsonImporter"
	model "github.com/sz-yinlong/L0/models"
	"github.com/sz-yinlong/L0/server"
	"github.com/sz-yinlong/L0/utility"
)

var orderCache *cache.OrderCache

var (
	db  *sql.DB
	sc  stan.Conn
	err error
)

func init() {
	db, err = setupDatabase()
	if err != nil {
		log.Fatalf("Error setting up database: %v", err)
	}
	sc, err = setupNatsStreaming()
	if err != nil {
		log.Fatalf("Error setting up NATS Streaming: %v", err)
	}
}

func main() {

	orderCache = cache.NewOrderCache()

	jsonImporter.ImportJson(db, "json/model.json", orderCache)

	if err = db.Ping(); err != nil {
		log.Fatalf("Error pinging database: %v", err)
	}

	if err := orderCache.LoadOrdersIntoCache(db); err != nil {
		log.Fatalf("Failed to load orders into cache: %v", err)
	}
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatalf("PORT must be set")
	}
	go server.StartServer(port, orderCache, db)
	log.Println("HTTP server is running on port", port)

	orderUID := "b563feb7b2b84b6test"
	order, found := orderCache.Get(orderUID)
	if found {
		fmt.Printf("Order found in cache: %+v\n", order)
	} else {
		fmt.Printf("Order not found in cache, loading from DB")
	}
	handleMessages(sc, db, orderCache)

}

func handleMessages(sc stan.Conn, db *sql.DB, cache *cache.OrderCache) {
	sub, err := sc.Subscribe("your_channel", func(msg *stan.Msg) {
		var order model.Order
		err := json.Unmarshal(msg.Data, &order)
		if err != nil {
			log.Printf("Error unmarshalling message: %v\n", err)
			return
		}
		err = utility.SaveOrder(db, cache, &order)
		if err != nil {
			log.Printf("Error subscribing to channel: %v", err)
			return
		}

	})
	if err != nil {
		fmt.Println("Error creating subscriber", err)
		return
	}
	defer sub.Unsubscribe()
	select {}
}

func setupNatsStreaming() (stan.Conn, error) {

	clientID := os.Getenv("NATS_CLIENT_ID")
	clusterID := os.Getenv("NATS_CLUSTER_ID")
	natsURL := os.Getenv("NATS_URL")

	return stan.Connect(clusterID, clientID, stan.NatsURL(natsURL))
}

func setupDatabase() (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	return db, nil
}

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
)

func main() {
	err := godotenv.Load("/Users/artem/Development/go/L0/enivroment.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
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
	defer db.Close()

	importJson(db, "/Users/artem/Development/go/L0/json/model.json")

	if err = db.Ping(); err != nil {
		log.Fatalf("Error pinging database: %v", err)
	}

	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		fmt.Println("Connection Error", err)
		return
	}
	defer nc.Close()

	fmt.Println("Connected to NATS Streaming Server")

	orderCache := NewOrderCache()
	if err := orderCache.LoadOrdersIntoCache(db); err != nil {
		log.Fatalf("Failed to load orders into cache: %v", err)
	}
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatalf("PORT must be set")
	}
	go StartServer(port, orderCache, db)
	log.Println("HTTP server is running on port", port)

	orderUID := "b563feb7b2b84b6test"
	order, found := orderCache.Get(orderUID)
	if found {
		fmt.Printf("Order found in cache: %+v\n", order)
	} else {
		fmt.Printf("Order not found in cache, loading from DB")
	}
	handleMessages(nc, db, orderCache)

}

func handleMessages(nc *nats.Conn, db *sql.DB, cache *OrderCache) {
	sub, err := nc.Subscribe("your_channel", func(msg *nats.Msg) {
		var order Order
		err := json.Unmarshal(msg.Data, &order)
		if err != nil {
			log.Printf("Error unmarshalling message: %v\n", err)
			return
		}
		err = saveOrder(db, cache, &order)
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

func saveOrder(db *sql.DB, cache *OrderCache, order *Order) error {
	orderUID := order.OrderUid

	if _, found := cache.Get(orderUID); found {
		log.Printf("Order %s already exists in cache, skipping save to DB", orderUID)
		return nil
	}

	jsonData, err := json.Marshal(order)
	if err != nil {
		return err
	}
	sqlStatement := `INSERT INTO orders (order_uid, order_data) VALUES ($1, $2) ON CONFLICT (order_uid) DO NOTHING`
	log.Println("Executing SQL statement:", sqlStatement)
	log.Printf("Order UID: %s, Data: %s\n", order.OrderUid, jsonData)
	_, err = db.Exec(sqlStatement, order.OrderUid, jsonData)
	if err != nil {
		return err
	}
	return nil

}

package jsonImporter

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"

	cache "github.com/sz-yinlong/L0/cache"
	model "github.com/sz-yinlong/L0/models"
)

func ImportJson(db *sql.DB, jsonFilePath string) {
	jsonData, err := ioutil.ReadFile(jsonFilePath)
	if err != nil {
		log.Fatalf("Error reading JSON file: %v", err)
	}

	var order model.Order

	if err := json.Unmarshal(jsonData, &order); err != nil {
		log.Fatalf("Error saving order to database: %v", err)
	}
	orderCache := cache.NewOrderCache()
	orderCache.Set(order.OrderUid, order)

	log.Println("Order imported and saved to cache successfully.")
}

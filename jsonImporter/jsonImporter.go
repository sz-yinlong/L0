package jsonImporter

import (
	"L0/cache"

	model "L0/resources/dbmodels"
	"L0/utility"
	"database/sql"
	"encoding/json"
	"log"
	"os"
)

func ImportJson(db *sql.DB, jsonFilePath string, cache *cache.OrderCache) {
	jsonData, err := os.ReadFile(jsonFilePath)
	if err != nil {
		log.Fatalf("Error reading JSON file: %v", err)
	}

	var order model.Order
	if err := json.Unmarshal(jsonData, &order); err != nil {
		log.Fatalf("Error unmarshalling JSON: %v", err)
	}

	cache.Set(order.OrderUid, order)

	if err := utility.SaveOrder(db, cache, &order); err != nil {
		log.Fatalf("Error saving order to database: %v", err)
	}
	log.Println("Order imported and saved to cache successfully.")
}

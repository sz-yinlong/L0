package main

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
)

func importJson(db *sql.DB, jsonFilePath string) {
	jsonData, err := ioutil.ReadFile(jsonFilePath)
	if err != nil {
		log.Fatalf("Error reading JSON file: %v", err)
	}

	var order Order
	orderCache := NewOrderCache()
	if err := json.Unmarshal(jsonData, &order); err != nil {
		log.Fatalf("Error saving order to database: %v", err)
	}
	if err := saveOrder(db, orderCache, &order); err != nil {
		log.Fatalf("Error saving order to database: %v", err)
	}
	log.Println("Order imported and saved successfully.")
}

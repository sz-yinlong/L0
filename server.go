package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func StartServer(port string, cache *OrderCache, db *sql.DB) {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})
	http.HandleFunc("/getOrder/", getOrderHandler(cache, db))

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Error starting HTTP server:", err)
	}

}

func getOrderHandler(cache *OrderCache, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		orderUID := strings.TrimPrefix(r.URL.Path, "/getOrder/")
		log.Printf("Requested orderUID: %s", orderUID)
		if orderUID == "" {
			http.Error(w, "Order ID is required", http.StatusBadRequest)
			return
		}
		if order, found := cache.Get(orderUID); found {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(order)
			return
		}

		var orderData []byte
		query := `SELECT order_data FROM orders WHERE order_uid = $1`
		row := db.QueryRow(query, orderUID)
		err := row.Scan(&orderData)

		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Order not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Server error", http.StatusInternalServerError)
			log.Printf("Error querying order from DB: %v", err)
			return
		}

		var order Order
		if err := json.Unmarshal(orderData, &order); err != nil {
			http.Error(w, "Error decoding order data", http.StatusInternalServerError)
			log.Printf("Error unmarshalling order data: %v", err)
			return
		}

		cache.Set(orderUID, order)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(order)
	}
}

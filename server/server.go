package server

import (
	"L0/cache"
	"L0/utility"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func StartServer(port string, cache *cache.OrderCache, db *sql.DB) {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})
	http.HandleFunc("/getOrder/", getOrderHandler(cache, db))

	fs := http.FileServer(http.Dir("static"))

	http.Handle("/", fs)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Error starting HTTP server:", err)
	}
}

func getOrderHandler(cache *cache.OrderCache, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		orderUID := strings.TrimPrefix(r.URL.Path, "/getOrder/")
		log.Printf("Requested orderUID: %s", orderUID)

		if orderUID == "" {
			http.Error(w, "Order ID is required", http.StatusBadRequest)
			return
		}
		if Order, found := cache.Get(orderUID); found {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(Order)
			return
		}
		order, err := utility.GetOrderFromDB(db, orderUID)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		if order == nil {
			http.Error(w, "Order not found", http.StatusNotFound)
			return
		}
		cache.Set(orderUID, *order)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(order)
	}
}

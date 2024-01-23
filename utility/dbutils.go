package utility

import (
	"database/sql"
	"encoding/json"
	"log"

	_ "github.com/lib/pq"

	"github.com/sz-yinlong/L0/cache"
	model "github.com/sz-yinlong/L0/models"
)

func SaveOrder(db *sql.DB, cache *cache.OrderCache, order *model.Order) error {
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

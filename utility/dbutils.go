package utility

import (
	"L0/cache"
	model "L0/resources/dbmodels"
	"database/sql"
	"encoding/json"
	"log"

	_ "github.com/lib/pq"
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
	_, err = db.Exec(sqlStatement, order.OrderUid, jsonData)
	if err != nil {
		return err
	}
	return nil
}

func GetOrderFromDB(db *sql.DB, orderUID string) (*model.Order, error) {
	var orderData []byte
	querry := `SELECT order_data FROM orders WHERE order_uid = $1`
	row := db.QueryRow(querry, orderUID)
	err := row.Scan(&orderData)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	var order model.Order
	if err := json.Unmarshal(orderData, &order); err != nil {
		return nil, err
	}
	return &order, nil
}

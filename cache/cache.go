package cache

import (
	model "L0/resources/dbmodels"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
)

type Order = model.Order

type OrderCache struct {
	mu sync.RWMutex

	cache map[string]Order
}

func NewOrderCache() *OrderCache {
	return &OrderCache{
		cache: make(map[string]Order),
	}
}
func (oc *OrderCache) Set(orderUID string, order Order) {
	oc.mu.Lock()
	defer oc.mu.Unlock()
	oc.cache[orderUID] = order
}
func (oc *OrderCache) Get(orderUID string) (Order, bool) {
	order, found := oc.cache[orderUID]
	return order, found
}
func (oc *OrderCache) Delete(orderUID string) {
	oc.mu.Lock()
	defer oc.mu.Unlock()
	delete(oc.cache, orderUID)
}
func (oc *OrderCache) LoadOrdersIntoCache(db *sql.DB) error {
	query := "SELECT order_uid, order_data FROM orders"
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("querry execution error: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var orderUID string
		var orderData []byte
		if err := rows.Scan(&orderUID, &orderData); err != nil {
			return fmt.Errorf("error scanning row: %v", err)
		}

		var order Order
		if err := json.Unmarshal(orderData, &order); err != nil {
			return fmt.Errorf("error unmarshaling order data: %v", err)
		}
		oc.Set(orderUID, order)
	}
	if err = rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %v", err)
	}
	return nil
}

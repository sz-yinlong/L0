package cache

import (
	model "L0/resources/dbmodels"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
)

type OrderRecord struct {
	mu    sync.RWMutex
	Order model.Order
}

type OrderCache struct {
	mu sync.RWMutex

	cache map[string]*OrderRecord
}

func NewOrderCache() *OrderCache {
	return &OrderCache{
		cache: make(map[string]*OrderRecord),
	}
}
func (oc *OrderCache) Set(orderUID string, order model.Order) {
	oc.mu.Lock()
	if _, exists := oc.cache[orderUID]; !exists {
		oc.cache[orderUID] = &OrderRecord{}
	}
	oc.cache[orderUID].mu.Lock()
	oc.mu.Unlock()

	oc.cache[orderUID].Order = order
	oc.cache[orderUID].mu.Unlock()
}

func (oc *OrderCache) Get(orderUID string) (model.Order, bool) {
	oc.mu.RLock()
	record, exists := oc.cache[orderUID]
	oc.mu.RUnlock()

	if !exists {
		return model.Order{}, false
	}

	record.mu.RLock()
	defer record.mu.RUnlock()
	return record.Order, true
}

func (oc *OrderCache) LoadOrdersIntoCache(db *sql.DB) error {
	query := "SELECT order_uid, order_data FROM orders"
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("query execution error: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var orderUID string
		var orderData []byte
		if err := rows.Scan(&orderUID, &orderData); err != nil {
			return fmt.Errorf("error scanning row: %v", err)
		}

		var order model.Order
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

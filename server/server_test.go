package server

import (
	"L0/cache"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetOrderHandler_DataInDB(t *testing.T) {
	mockCache := cache.NewOrderCache()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"order_data"}).AddRow(`{"order_uid":"testorderUID", "other_fields": "values"}`)
	mock.ExpectQuery("SELECT order_data FROM orders WHERE order_uid = \\$1").WithArgs("testorderUID").WillReturnRows(rows)

	req, err := http.NewRequest("GET", "/getOrder/testorderUID", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := getOrderHandler(mockCache, db)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

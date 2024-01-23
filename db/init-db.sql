CREATE TABLE IF NOT EXISTS orders (
    order_uid VARCHAR(255) PRIMARY KEY,
    order_data JSONB NOT NULL
);

INSERT INTO orders (order_uid, order_data) VALUES ('order1', '{"item": "product1", "quantity": 1, "price": 10.99}'::JSONB);
INSERT INTO orders (order_uid, order_data) VALUES ('order2', '{"item": "product2", "quantity": 2, "price": 21.98}'::JSONB);

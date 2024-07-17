-- Table for cart
CREATE TABLE cart (
    id TEXT PRIMARY KEY,
    created_at DATETIME,
    last_updated_at DATETIME,
    total_value REAL
);

-- Table for cart_item
CREATE TABLE cart_item (
    id TEXT PRIMARY KEY,
    cart_id TEXT,
    name TEXT,
    sale_price REAL,
    qty INTEGER,
    created_at DATETIME,
    FOREIGN KEY (cart_id) REFERENCES cart(id)
);

-- Table for cart_item_component
CREATE TABLE cart_item_component (
    id TEXT PRIMARY KEY,
    cart_item_id TEXT,
    cart_id TEXT,
    product_id INTEGER,
    qty INTEGER,
    name TEXT,
    created_at DATETIME,
    FOREIGN KEY (cart_item_id) REFERENCES cart_item(id),
    FOREIGN KEY (cart_id) REFERENCES cart(id)
);

CREATE TABLE cart (
    id              TEXT     PRIMARY KEY,    -- stored in cart table
    created_at      DATETIME NOT NULL,       -- stored in cart table
    last_updated_at DATETIME NOT NULL        -- stored in cart table
);
CREATE TABLE cart_item (
    id          TEXT     NOT NULL,        -- stored in cart_item table
    cart_id     TEXT     NOT NULL,           -- stored in cart_item table
    qty         INTEGER  NOT NULL,           -- stored in cart_item table
    created_at  DATETIME NOT NULL,           -- stored in cart_item table
    FOREIGN KEY (cart_id) REFERENCES cart(id) ON DELETE CASCADE
);
CREATE TABLE cart_item_component (
    cart_item_id TEXT    NOT NULL,           -- stored in cart_item_component table
    cart_id      TEXT    NOT NULL,           -- stored in cart_item_component table
    product_id   TEXT    NOT NULL,           -- from Product: only product_id stored
    qty          INTEGER NOT NULL,           -- from Product: only qty stored
    created_at   DATETIME NOT NULL,          -- stored in cart_item_component table
    PRIMARY KEY (cart_item_id, product_id),
    FOREIGN KEY (cart_item_id) REFERENCES cart_item(id) ON DELETE CASCADE,
    FOREIGN KEY (cart_id)      REFERENCES cart(id) ON DELETE CASCADE
);


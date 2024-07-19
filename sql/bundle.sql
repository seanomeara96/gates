CREATE TABLE IF NOT EXISTS bundle_gates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    gate_id INTEGER NOT NULL,
    bundle_id INTEGER NOT NULL,
    qty INTEGER NOT NULL,
    FOREIGN KEY (gate_id) REFERENCES products(id),
    FOREIGN KEY (bundle_id) REFERENCES products(id)
);

 CREATE TABLE IF NOT EXISTS bundle_extensions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    extension_id INTEGER NOT NULL,
    bundle_id INTEGER NOT NULL,
    qty INTEGER NOT NULL,
    FOREIGN KEY (extension_id) REFERENCES products(id),
    FOREIGN KEY (bundle_id) REFERENCES products(id)
);
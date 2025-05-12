package repos

import (
	"database/sql"
	"errors"

	"github.com/seanomeara96/gates/models"
)

/*
SQL statements to create the required tables:

CREATE TABLE orders (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				cart_id INTEGER NOT NULL,
				status TEXT DEFAULT 'pending',
				customer_name TEXT,
				customer_email TEXT,
				customer_phone TEXT,
				shipping_address TEXT,
				billing_address TEXT,
				payment_method TEXT,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE order_items (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				order_id INTEGER NOT NULL,
				item_name TEXT NOT NULL,
				item_quantity INTEGER NOT NULL,
				FOREIGN KEY (order_id) REFERENCES orders(id)
);

CREATE TABLE order_item_components (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				order_id INTEGER NOT NULL,
				order_item_id INTEGER NOT NULL,
				product_id INTEGER NOT NULL,
				product_name TEXT NOT NULL,
				product_price REAL NOT NULL,
				product_qty INTEGER NOT NULL,
				FOREIGN KEY (order_id) REFERENCES orders(id),
				FOREIGN KEY (order_item_id) REFERENCES order_items(id)
);
*/

type OrderRepo struct {
	db *sql.DB
}

type CustomerDetails struct {
	Name            string
	Email           string
	Phone           string
	ShippingAddress string
	BillingAddress  string
	PaymentMethod   string
}

func NewOrderRepo(db *sql.DB) *OrderRepo {
	return &OrderRepo{db}
}

func (r *OrderRepo) New(cart *models.Cart) (int, error) {
	if cart == nil {
		return 0, errors.New("cart cannot be nil")
	}

	tx, err := r.db.Begin()
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	res, err := tx.Exec(`INSERT INTO orders(cart_id) VALUES(?)`, cart.ID)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	_id, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	id := int(_id)

	for _, item := range cart.Items {
		if err := r.InsertItem(tx, id, item); err != nil {
			tx.Rollback()
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return 0, err
	}

	return id, nil
}

func (r *OrderRepo) UpdateStatus(orderID int, status string) error {
	_, err := r.db.Exec("UPDATE orders SET status = ? WHERE id = ?", status, orderID)
	if err != nil {
		return err
	}
	return nil
}

func (r *OrderRepo) UpdateCustomerDetails(orderID int, details CustomerDetails) error {
	_, err := r.db.Exec(`
		UPDATE orders
		SET customer_name = ?,
			customer_email = ?,
			customer_phone = ?,
			shipping_address = ?,
			billing_address = ?,
			payment_method = ?
		WHERE id = ?`,
		details.Name,
		details.Email,
		details.Phone,
		details.ShippingAddress,
		details.BillingAddress,
		details.PaymentMethod,
		orderID)
	if err != nil {
		return err
	}
	return nil
}

func (r *OrderRepo) InsertItem(tx *sql.Tx, orderID int, item models.CartItem) error {
	res, err := tx.Exec(`INSERT INTO order_items(order_id, item_name, item_quantity) VALUES (?,?,?)`, orderID, item.Name, item.Qty)
	if err != nil {
		return err
	}

	_id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	id := int(_id)

	for _, component := range item.Components {
		if err := r.InsertComponent(tx, orderID, id, component); err != nil {
			return err
		}
	}
	return nil
}

func (r *OrderRepo) InsertComponent(tx *sql.Tx, orderID int, orderItemID int, component models.CartItemComponent) error {
	_, err := tx.Exec(
		`INSERT INTO order_item_components(order_id, order_item_id, product_id, product_name, product_price, product_qty) VALUES (?,?,?,?,?,?)`,
		orderID, orderItemID, component.Product.Id, component.Product.Name, component.Product.Price, component.Qty,
	)
	if err != nil {
		return err
	}
	return nil
}

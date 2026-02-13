package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/seanomeara96/gates/models"
	"github.com/seanomeara96/gates/repos"
)

/*
SQL statements to create the required tables:

CREATE TABLE orders (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				cart_id INTEGER NOT NULL,
				session_id TEXT,
				status TEXT DEFAULT 'pending_payment',
				customer_name TEXT,
				customer_email TEXT,
				customer_phone TEXT,
				shipping_address TEXT,
				billing_address TEXT,
				payment_method TEXT,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				stripe_ref STRING
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

// Create operations
func (r *OrderRepo) New(cart models.Cart) (int, error) {
	if cart.ID == "" {
		return 0, errors.New("new order: cart ID cannot be empty")
	}

	tx, err := r.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("new order: begin transaction (cart_id=%s): %w", cart.ID, err)
	}

	// Default status to pending payment
	defaultStatus := models.OrderStatusPendingPayment

	res, err := tx.Exec(`INSERT INTO orders(cart_id, status) VALUES(?, ?)`, cart.ID, defaultStatus)
	if err != nil {
		_ = tx.Rollback()
		return 0, fmt.Errorf("new order: insert into orders (cart_id=%s, status=%s): %w", cart.ID, defaultStatus, err)
	}

	_id, err := res.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		return 0, fmt.Errorf("new order: get last insert id (cart_id=%s): %w", cart.ID, err)
	}

	id := int(_id)

	for idx, item := range cart.Items {
		if err := r.InsertItem(tx, id, item); err != nil {
			_ = tx.Rollback()
			return 0, fmt.Errorf("new order: insert item (order_id=%d, item_index=%d, item_name=%q): %w", id, idx, item.Name, err)
		}
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return 0, fmt.Errorf("new order: commit transaction (order_id=%d, cart_id=%s): %w", id, cart.ID, err)
	}

	return id, nil
}

func (r *OrderRepo) InsertItem(tx *sql.Tx, orderID int, item models.CartItem) error {
	if tx == nil {
		return errors.New("insert item: transaction cannot be nil")
	}

	res, err := tx.Exec(`INSERT INTO order_items(order_id, item_name, item_quantity) VALUES (?,?,?)`, orderID, item.Name, item.Qty)
	if err != nil {
		return fmt.Errorf("insert item: insert into order_items (order_id=%d, item_name=%q, item_qty=%d): %w", orderID, item.Name, item.Qty, err)
	}

	_id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("insert item: get last insert id (order_id=%d, item_name=%q): %w", orderID, item.Name, err)
	}

	id := int(_id)

	for idx, component := range item.Components {
		if err := r.InsertComponent(tx, orderID, id, component); err != nil {
			return fmt.Errorf(
				"insert item: insert component (order_id=%d, order_item_id=%d, component_index=%d, product_id=%d): %w",
				orderID, id, idx, component.Product.Id, err,
			)
		}
	}
	return nil
}

func (r *OrderRepo) InsertComponent(tx *sql.Tx, orderID int, orderItemID int, component models.CartItemComponent) error {
	if tx == nil {
		return errors.New("insert component: transaction cannot be nil")
	}

	_, err := tx.Exec(
		`INSERT INTO order_item_components(order_id, order_item_id, product_id, product_name, product_price, product_qty) VALUES (?,?,?,?,?,?)`,
		orderID, orderItemID, component.Product.Id, component.Product.Name, component.Product.Price, component.Qty,
	)
	if err != nil {
		return fmt.Errorf(
			"insert component: insert into order_item_components (order_id=%d, order_item_id=%d, product_id=%d, product_name=%q, product_qty=%d): %w",
			orderID, orderItemID, component.Product.Id, component.Product.Name, component.Qty, err,
		)
	}
	return nil
}

// Read operations
func (r *OrderRepo) GetOrders(params repos.GetOrdersParams) ([]models.Order, error) {
	query := `SELECT id, cart_id, session_id, status, customer_name, customer_email,
											customer_phone, shipping_address, billing_address, payment_method,
											created_at, stripe_ref FROM orders LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, params.Limit, params.Offset)
	if err != nil {
		return nil, fmt.Errorf("get orders: query orders (limit=%d, offset=%d): %w", params.Limit, params.Offset, err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		err := rows.Scan(
			&o.ID, &o.CartID, &o.SessionID, &o.Status, &o.CustomerName, &o.CustomerEmail,
			&o.CustomerPhone, &o.ShippingAddress, &o.BillingAddress, &o.PaymentMethod,
			&o.CreatedAt, &o.StripeRef,
		)
		if err != nil {
			return nil, fmt.Errorf("get orders: scan order row: %w", err)
		}
		orders = append(orders, o)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("get orders: iterate order rows: %w", err)
	}

	return orders, nil
}

func (r *OrderRepo) GetOrderByID(id int) (*models.Order, error) {
	query := `SELECT id, cart_id, session_id, status, customer_name, customer_email,
											customer_phone, shipping_address, billing_address, payment_method,
											created_at, stripe_ref FROM orders WHERE id = ?`

	row := r.db.QueryRow(query, id)

	var o models.Order
	err := row.Scan(
		&o.ID, &o.CartID, &o.SessionID, &o.Status, &o.CustomerName, &o.CustomerEmail,
		&o.CustomerPhone, &o.ShippingAddress, &o.BillingAddress, &o.PaymentMethod,
		&o.CreatedAt, &o.StripeRef,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("get order by id: order not found (id=%d)", id)
		}
		return nil, fmt.Errorf("get order by id: scan order row (id=%d): %w", id, err)
	}

	return &o, nil
}

func (r *OrderRepo) GetOrderItems(orderID int) ([]models.CartItem, error) {
	query := `SELECT id, item_name, item_quantity FROM order_items WHERE order_id = ?`

	rows, err := r.db.Query(query, orderID)
	if err != nil {
		return nil, fmt.Errorf("get order items: query order_items (order_id=%d): %w", orderID, err)
	}
	defer rows.Close()

	var items []models.CartItem
	for rows.Next() {
		var item models.CartItem
		var itemID int

		if err := rows.Scan(&itemID, &item.Name, &item.Qty); err != nil {
			return nil, fmt.Errorf("get order items: scan order item row (order_id=%d): %w", orderID, err)
		}

		// Get components for this item
		components, err := r.GetOrderItemComponents(orderID, itemID)
		if err != nil {
			return nil, fmt.Errorf("get order items: get components (order_id=%d, order_item_id=%d): %w", orderID, itemID, err)
		}

		item.Components = components
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("get order items: iterate order item rows (order_id=%d): %w", orderID, err)
	}

	return items, nil
}

func (r *OrderRepo) GetOrderItemComponents(orderID, itemID int) ([]models.CartItemComponent, error) {
	query := `SELECT product_id, product_name, product_price, product_qty
											FROM order_item_components
											WHERE order_id = ? AND order_item_id = ?`

	rows, err := r.db.Query(query, orderID, itemID)
	if err != nil {
		return nil, fmt.Errorf("get order item components: query order_item_components (order_id=%d, order_item_id=%d): %w", orderID, itemID, err)
	}
	defer rows.Close()

	var components []models.CartItemComponent
	for rows.Next() {
		var component models.CartItemComponent
		var productID int
		var productName string
		var productPrice float32

		if err := rows.Scan(&productID, &productName, &productPrice, &component.Qty); err != nil {
			return nil, fmt.Errorf("get order item components: scan component row (order_id=%d, order_item_id=%d): %w", orderID, itemID, err)
		}

		// Create product for this component
		component.Product = models.Product{
			Id:    productID,
			Name:  productName,
			Price: productPrice,
		}

		components = append(components, component)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("get order item components: iterate component rows (order_id=%d, order_item_id=%d): %w", orderID, itemID, err)
	}

	return components, nil
}

// Update operations
func (r *OrderRepo) UpdateStatus(orderID int, status models.OrderStatus) error {
	// Validate the status before updating
	if err := status.Validate(); err != nil {
		return fmt.Errorf("update order status: invalid status (order_id=%d, status=%q): %w", orderID, status, err)
	}

	_, err := r.db.Exec("UPDATE orders SET status = ? WHERE id = ?", status, orderID)
	if err != nil {
		return fmt.Errorf("update order status: exec update (order_id=%d, status=%s): %w", orderID, status, err)
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
		return fmt.Errorf("update customer details: exec update (order_id=%d, customer_email=%q): %w", orderID, details.Email, err)
	}
	return nil
}

func (r *OrderRepo) UpdateOrder(order *models.Order) error {
	if order == nil {
		return errors.New("update order: order cannot be nil")
	}

	// Validate the status before updating
	if err := order.Status.Validate(); err != nil {
		return fmt.Errorf("update order: invalid status (order_id=%d, status=%q): %w", order.ID, order.Status, err)
	}

	_, err := r.db.Exec(`
		UPDATE orders
		SET cart_id = ?,
						session_id = ?,
						status = ?,
						customer_name = ?,
						customer_email = ?,
						customer_phone = ?,
						shipping_address = ?,
						billing_address = ?,
						payment_method = ?,
						stripe_ref = ?
		WHERE id = ?`,
		order.CartID,
		order.SessionID,
		order.Status,
		order.CustomerName,
		order.CustomerEmail,
		order.CustomerPhone,
		order.ShippingAddress,
		order.BillingAddress,
		order.PaymentMethod,
		order.StripeRef,
		order.ID)

	if err != nil {
		return fmt.Errorf("update order: exec update (order_id=%d, cart_id=%s): %w", order.ID, order.CartID, err)
	}

	return nil
}

func (r *OrderRepo) UpdateStripeRef(orderID int, stripeRef string) error {
	_, err := r.db.Exec("UPDATE orders SET stripe_ref = ? WHERE id = ?", stripeRef, orderID)
	if err != nil {
		return fmt.Errorf("update stripe reference: exec update (order_id=%d, stripe_ref=%q): %w", orderID, stripeRef, err)
	}
	return nil
}

func (r *OrderRepo) UpdateSessionID(orderID int, sessionID string) error {
	_, err := r.db.Exec("UPDATE orders SET session_id = ? WHERE id = ?", sessionID, orderID)
	if err != nil {
		return fmt.Errorf("update session ID: exec update (order_id=%d, session_id=%q): %w", orderID, sessionID, err)
	}
	return nil
}

// Delete operations
func (r *OrderRepo) DeleteOrder(orderID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("delete order: begin transaction (order_id=%d): %w", orderID, err)
	}

	// First delete from components (child)
	_, err = tx.Exec("DELETE FROM order_item_components WHERE order_id = ?", orderID)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("delete order: delete order_item_components (order_id=%d): %w", orderID, err)
	}

	// Then delete from items (middle)
	_, err = tx.Exec("DELETE FROM order_items WHERE order_id = ?", orderID)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("delete order: delete order_items (order_id=%d): %w", orderID, err)
	}

	// Finally delete the order itself (parent)
	_, err = tx.Exec("DELETE FROM orders WHERE id = ?", orderID)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("delete order: delete orders row (order_id=%d): %w", orderID, err)
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("delete order: commit transaction (order_id=%d): %w", orderID, err)
	}

	return nil
}

func (r *OrderRepo) DeleteOrderItem(orderID, itemID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("delete order item: begin transaction (order_id=%d, item_id=%d): %w", orderID, itemID, err)
	}

	// First delete from components (child)
	_, err = tx.Exec("DELETE FROM order_item_components WHERE order_id = ? AND order_item_id = ?", orderID, itemID)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("delete order item: delete order_item_components (order_id=%d, item_id=%d): %w", orderID, itemID, err)
	}

	// Then delete the item itself (parent)
	_, err = tx.Exec("DELETE FROM order_items WHERE order_id = ? AND id = ?", orderID, itemID)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("delete order item: delete order_items row (order_id=%d, item_id=%d): %w", orderID, itemID, err)
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("delete order item: commit transaction (order_id=%d, item_id=%d): %w", orderID, itemID, err)
	}

	return nil
}

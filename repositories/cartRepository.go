package repositories

import (
	"database/sql"

	"github.com/seanomeara96/gates/models"
)

type CartRepository struct {
	db *sql.DB
}

func NewCartRepository(db *sql.DB) *CartRepository {
	return &CartRepository{
		db,
	}
}

func (r *CartRepository) CreateTables() (sql.Result, error) {
	res, err := r.db.Exec(`CREATE TABLE IF NOT EXISTS carts(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		created_at DATETIME,
		last_updated_at DATETIME
	)`)

	if err != nil {
		return res, err
	}

	res, err = r.db.Exec(`CREATE TABLE IF NOT EXISTS cart_items(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		cart_id INTEGER NOT NULL,
		product_id INTEGER NOT NULL,
		quantity INTEGER DEFAULT 1,
		created_at DATETIME,
		FOREIGN KEY (cart_id) REFERENCES carts(id),
		FOREIGN KEY(product_id) REFERENCES products(id)
	)`)

	return res, err
}

func (r *CartRepository) SaveCart(cart models.Cart) (sql.Result, error) {
	return r.db.Exec(`INSERT INTO 
		carts(
			user_id, 
			created_at, 
			last_updated_at) 
		VALUES 
			(?, ?, ?)`,
		cart.UserID,
		cart.CreatedAt,
		cart.LastUpdatedAt,
	)
}

func (r *CartRepository) SaveCartItem(cartItem models.CartItem) (sql.Result, error) {
	return r.db.Exec(`INSERT INTO
		cart_items(
			cart_id,
			product_id,
			quantity,
			created_at
		)
		VALUES
			(?, ?, ?, ?)`,
		cartItem.CartID,
		cartItem.ProductID,
		cartItem.Quantity,
		cartItem.CreatedAt,
	)

}

func (r *CartRepository) GetCartByUserID(userID int) (*models.Cart, error) {
	var cart models.Cart
	err := r.db.QueryRow(`SELECT 
			id, user_id, created_at, last_updated_at 
		FROM carts 
		WHERE user_id = ?`,
		userID,
	).Scan(
		&cart.ID,
		&cart.UserID,
		&cart.CreatedAt,
		&cart.LastUpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &cart, nil
}

func (r *CartRepository) GetCartItemsByCartID(cartID int) ([]*models.CartItem, error) {
	var items []*models.CartItem
	rows, err := r.db.Query(`SELECT id, cart_id, product_id, quantity, created_at FROM cart_items WHERE cart_id = ?`, cartID)
	if err != nil {
		return items, err
	}
	defer rows.Close()
	for rows.Next() {
		cartItem := models.CartItem{}
		err := rows.Scan(
			&cartItem.ID,
			&cartItem.CartID,
			&cartItem.ProductID,
			&cartItem.Quantity,
			&cartItem.CreatedAt,
		)
		if err != nil {
			return items, err
		}
		items = append(items, &cartItem)
	}
	return items, nil
}

func (r *CartRepository) GetCartItemByID(cartItemID int) (*models.CartItem, error) {
	var cartItem models.CartItem
	err := r.db.QueryRow(`SELECT 
			id, cart_id, product_id, quantity, created_at 
		FROM cart_items 
		WHERE id = ?`,
		cartItemID,
	).Scan(
		&cartItem.ID,
		&cartItem.CartID,
		&cartItem.ProductID,
		&cartItem.Quantity,
		&cartItem.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &cartItem, nil
}

func (r *CartRepository) GetCartItemByProductID(cartID, productID int) (*models.CartItem, error) {
	var cartItem models.CartItem
	err := r.db.QueryRow(`SELECT
			id, cart_id, product_id, quantity, created_at
		FROM cart_items
		WHERE cart_id = ? AND product_id = ?`,
		cartID, productID,
	).Scan(
		&cartItem.ID,
		&cartItem.CartID,
		&cartItem.ProductID,
		&cartItem.Quantity,
		&cartItem.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &cartItem, nil
}

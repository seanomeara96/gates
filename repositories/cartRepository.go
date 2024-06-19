package repositories

import (
	"database/sql"
	"errors"
	"fmt"

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
		id STRING PRIMARY KEY,
		created_at DATETIME,
		last_updated_at DATETIME
	)`)

	if err != nil {
		return res, err
	}

	res, err = r.db.Exec(`CREATE TABLE IF NOT EXISTS cart_items(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		cart_id STRING NOT NULL,
		product_id INTEGER NOT NULL,
		quantity INTEGER DEFAULT 1,
		created_at DATETIME,
		FOREIGN KEY (cart_id) REFERENCES carts(id),
		FOREIGN KEY(product_id) REFERENCES products(id)
	)`)

	return res, err
}

func (r *CartRepository) SaveCart(cart models.Cart) (*sql.Result, error) {
	res, err := r.db.Exec(`INSERT INTO 
		carts(
			id, 
			created_at, 
			last_updated_at) 
		VALUES 
			(?, ?, ?)`,
		cart.ID,
		cart.CreatedAt,
		cart.LastUpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to save cart. %v", err)
	}
	return &res, nil
}

func (r *CartRepository) SaveCartItem(cartItem models.CartItem) error {
	return errors.New("Save cart item not implemented")
}

func (r *CartRepository) GetCartByUserID(userID int) (*models.Cart, error) {
	var cart models.Cart
	err := r.db.QueryRow(`SELECT 
			id, created_at, last_updated_at 
		FROM carts 
		WHERE user_id = ?`,
		userID,
	).Scan(
		&cart.ID,
		&cart.CreatedAt,
		&cart.LastUpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &cart, nil
}

func (r *CartRepository) GetCartItemsByCartID(cartID string) ([]*models.CartItem, error) {
	return nil, errors.New("Getting cart items not implemented")
}

func (r *CartRepository) GetCartItemByID(cartItemID int) (*models.CartItem, error) {
	return nil, errors.New("Getting cart item not implemented")
}
func (r *CartRepository) UpdateCartItem(item models.CartItem) error {

	return errors.New("Updating cart item not implemented")
}

func (r *CartRepository) SaveCartItemComponents(cartID string, components []models.CartItemComponent) error {

	return errors.New("Save Cart Item Components not yet implemented")

}

func (r *CartRepository) RemoveCartItem(cartID, itemID string) error {
	return errors.New("Remove cart item not implemented")
}

func (r *CartRepository) RemoveCartItemComponents(itemID string) error {
	return errors.New("Remove Cart Item Components not yet implemented")
}

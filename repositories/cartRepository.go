package repositories

import (
	"database/sql"
	"fmt"

	"github.com/seanomeara96/gates/models"
)

var carts []models.Cart = []models.Cart{}
var cartItems []models.CartItem = []models.CartItem{}

type CartRepository struct {
	db *sql.DB
}

func NewCartRepository(db *sql.DB) *CartRepository {
	return &CartRepository{
		db,
	}
}

func (r *CartRepository) SaveCart(cart models.Cart) (models.Cart, error) {
	cart.ID = len(carts)
	carts = append(carts, cart)
	return cart, nil
}

func (r *CartRepository) SaveCartItem(cartItem models.CartItem) (models.CartItem, error) {
	cartItems = append(cartItems, cartItem)
	return cartItem, nil
}

func (r *CartRepository) GetCartByUserId(userID int) (*models.Cart, error) {
	for _, cart := range carts {
		if cart.UserID == userID {
			return &cart, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (r *CartRepository) GetCartItemsByCartId(cartID int) ([]*models.CartItem, error) {
	var items []*models.CartItem
	for _, item := range cartItems {
		if item.CartID == cartID {
			items = append(items, &item)
		}
	}
	return items, nil
}

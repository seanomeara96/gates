package models

import (
	"time"

	"github.com/google/uuid"
)

type Cart struct {
	ID            string    `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	LastUpdatedAt time.Time `json:"last_updated_at"`
}

func NewCart() Cart {
	cartID := uuid.New().String()
	return Cart{
		ID:            cartID,
		CreatedAt:     time.Now(),
		LastUpdatedAt: time.Now(),
	}
}

type CartItem struct {
	ID        int       `json:"id"`
	CartID    string    `json:"cart_id"`
	ProductID int       `json:"product_id"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
}

func NewCartItem(cartID string, productID, quantity int) CartItem {
	return CartItem{
		CartID:    cartID,
		ProductID: productID,
		Quantity:  quantity,
		CreatedAt: time.Now(),
	}
}

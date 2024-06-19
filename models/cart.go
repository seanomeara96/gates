package models

import (
	"time"

	"github.com/google/uuid"
)

type Cart struct {
	ID            string    `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	LastUpdatedAt time.Time `json:"last_updated_at"`
	Items         []CartItem
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
	ID         string `json:"id"`
	CartID     string `json:"cart_id"`
	Components []CartItemComponent
	Qty        int       `json:"qty"`
	CreatedAt  time.Time `json:"created_at"`
}

func NewCartItem(cartID string) CartItem {
	id := uuid.New().String()
	return CartItem{
		ID:        id,
		CartID:    cartID,
		CreatedAt: time.Now(),
		Qty:       1,
	}
}

type CartItemComponent struct {
	CartItemID string
	ProductID  int `json:"id"`
	Qty        int `json:"qty"`
}

package models

import (
	"time"

	"github.com/google/uuid"
)

type Cart struct {
	ID            string     `json:"id"`
	CreatedAt     time.Time  `json:"created_at"`
	LastUpdatedAt time.Time  `json:"last_updated_at"`
	Items         []CartItem `json:"items"`
	TotalValue    float64    `json:"total_value"`
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
	ID         string              `json:"id"`
	CartID     string              `json:"cart_id"`
	Name       string              `json:"name"`
	SalePrice  string              `json:"sale_price"`
	Components []CartItemComponent `json:"components"`
	Qty        int                 `json:"qty"`
	CreatedAt  time.Time           `json:"created_at"`
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
	CartItemID string `json:"id"`
	ProductID  int    `json:"product_id"`
	Qty        int    `json:"qty"`
	Name       string `json:"name"`
}

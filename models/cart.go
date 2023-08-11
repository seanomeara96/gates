package models

import "time"

type Cart struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	CreatedAt     time.Time `json:"created_at"`
	LastUpdatedAt time.Time `json:"last_updated_at"`
}

func NewCart(userID int) Cart {
	return Cart{
		UserID:        userID,
		CreatedAt:     time.Now(),
		LastUpdatedAt: time.Now(),
	}
}

type CartItem struct {
	ID        int       `json:"id"`
	CartID    int       `json:"cart_id"`
	ProductID int       `json:"product_id"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
}

func NewCartItem(cartID, productID, quantity int) CartItem {
	return CartItem{
		CartID:    cartID,
		ProductID: productID,
		Quantity:  quantity,
		CreatedAt: time.Now(),
	}
}

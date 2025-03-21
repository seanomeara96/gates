package models

import (
	"strconv"
	"strings"
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

type CartItem struct {
	ID         string              `json:"id"`
	CartID     string              `json:"cart_id"`
	Name       string              `json:"name"`
	SalePrice  float64             `json:"sale_price"`
	Components []CartItemComponent `json:"components"`
	Qty        int                 `json:"qty"`
	CreatedAt  time.Time           `json:"created_at"`
}

type CartItemComponent struct {
	CartItemID string    `json:"cart_item_id"`
	CartID     string    `json:"cart_id"`
	ProductID  int       `json:"product_id"`
	Qty        int       `json:"qty"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"created_at"`
}

func NewCart() Cart {
	cartID := uuid.New().String()
	return Cart{
		ID:            cartID,
		CreatedAt:     time.Now(),
		LastUpdatedAt: time.Now(),
	}
}

func NewCartItem(cartID string, components []CartItemComponent) CartItem {
	idParts := []string{}
	for _, c := range components {
		idParts = append(idParts, strconv.Itoa(c.ProductID)+"-"+strconv.Itoa(c.Qty))
	}
	id := strings.Join(idParts, "_")

	for i := range components {
		components[i].CartID = cartID
		components[i].CartItemID = id
	}

	name := components[0].Name

	if len(components) > 1 {
		name += " and " + strconv.Itoa(len(components) - 1) + " components"
	}

	return CartItem{
		ID:         id,
		CartID:     cartID,
		Name:
		Components: components,
		CreatedAt:  time.Now(),
		Qty:        1,
	}
}

func NewCartItemComponent(cartID string) CartItemComponent {
	return CartItemComponent{
		CartID:    cartID,
		CreatedAt: time.Now(),
	}
}

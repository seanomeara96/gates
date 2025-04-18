package models

import (
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Cart struct {
	ID            string     `json:"id"`              // stored in cart table
	CreatedAt     time.Time  `json:"created_at"`      // stored in cart table
	LastUpdatedAt time.Time  `json:"last_updated_at"` // stored in cart table
	Items         []CartItem `json:"items"`
	TotalValue    float32    `json:"total_value"`
}

type CartItem struct {
	ID         string              `json:"id"`      // stored in cart_item table
	CartID     string              `json:"cart_id"` // stored in cart_item table
	Name       string              `json:"name"`
	SalePrice  float32             `json:"sale_price"`
	Components []CartItemComponent `json:"components"`
	Qty        int                 `json:"qty"`        // stored in cart_item table
	CreatedAt  time.Time           `json:"created_at"` // stored in cart_item table
}

type CartItemComponent struct {
	CartItemID string    `json:"cart_item_id"` // stored in cart_item_component table
	CartID     string    `json:"cart_id"`      // stored in cart_item_component table
	CreatedAt  time.Time `json:"created_at"`   // stored in cart_item_component table
	Product              // only product_id and qty stored in cart_item_component table
}

func NewCart() Cart {
	cartID := uuid.New().String()
	return Cart{
		ID:            cartID,
		CreatedAt:     time.Now(),
		LastUpdatedAt: time.Now(),
	}
}

func (i *CartItem) SetName() {
	i.Name = i.Components[0].Name
	if len(i.Components) > 1 {
		i.Name += " and " + strconv.Itoa(len(i.Components)-1) + " components"
	}
}

func (i *CartItem) SetPrice() {

	for c := range i.Components {
		component := i.Components[c]
		i.SalePrice += (component.Price * float32(component.Qty))
	}
	i.SalePrice *= float32(i.Qty)

}

func NewCartItem(cartID string, components []CartItemComponent) CartItem {

	idParts := []string{}
	for _, c := range components {
		idParts = append(idParts, strconv.Itoa(c.Product.Id)+"-"+strconv.Itoa(c.Qty))
	}
	id := strings.Join(idParts, "_")

	item := CartItem{
		ID:         id,
		CartID:     cartID,
		Components: components,
		CreatedAt:  time.Now(),
		Qty:        1,
	}

	for i := range item.Components {
		item.Components[i].CartID = cartID
		item.Components[i].CartItemID = id
	}

	item.SetName()

	return item
}

func NewCartItemComponent(cartID string) CartItemComponent {
	return CartItemComponent{
		CartID:    cartID,
		CreatedAt: time.Now(),
	}
}

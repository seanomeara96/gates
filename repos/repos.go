package repos

import (
	"github.com/seanomeara96/gates/models"
)

// ProductFilterParams defines the parameters for filtering product lists.
type ProductFilterParams struct {
	MaxWidth       float32
	Limit          int
	Color          string
	InventoryLevel int     // Assumed filter: inventory_level >= ? (if > 0)
	Price          float32 // Assumed filter: price <= ? (if > 0)
	Type           models.ProductType
}

// CustomerDetails holds optional customer-provided data related to an order.
type CustomerDetails struct {
	Name            string
	Email           string
	Phone           string
	ShippingAddress string
	BillingAddress  string
	PaymentMethod   string
}

type GetOrdersParams struct {
	Limit, Offset int
}

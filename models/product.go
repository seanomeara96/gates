package models

// Define a custom type for the product type stored in the database.
type ProductType string

// Define constants representing the product type values.
const (
	ProductTypeGate      ProductType = "gate"
	ProductTypeExtension ProductType = "extension"
	ProductTypeBundle    ProductType = "bundle"
)

type Product struct {
	Id             int         `json:"product_id"`
	Type           ProductType `json:"type"`
	Name           string      `json:"name"`
	Width          float32     `json:"width"`
	Price          float32     `json:"price"`
	Img            string      `json:"img"`
	Color          string      `json:"color"`
	Tolerance      float32     `json:"tolerance"`
	Qty            int         `json:"qty"`
	InventoryLevel int         `json:"inventory_level"`
}

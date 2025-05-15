package repos

import (
	"database/sql"

	"github.com/seanomeara96/gates/models"
)

// CartRepository defines the interface for interacting with the cart storage
type CartRepository interface {
	SaveCart(cart models.Cart) (*sql.Result, error)
	CartExists(id string) (bool, error)
	GetCartByID(id string) (*models.Cart, error)
	GetCartByUserID(userID int) (*models.Cart, error)
	InsertCartItem(cartItem models.CartItem) error
	SelectCartItem(cartID, itemID string) (*models.CartItem, error)
	GetCartItemsByCartID(cartID string) ([]*models.CartItem, error)
	DoesCartItemExist(cartID, cartItemID string) (bool, error)
	SaveCartItemComponents(components []models.CartItemComponent) error
	SetLastUpdated(cartID string) error
	IncrementCartItem(cartID, itemID string) error
	DecrementCartItem(cartID, itemID string) error
	RemoveCartItem(cartID, itemID string) error
	RemoveCartItemComponents(itemID string) error
}

// ProductFilterParams defines the parameters for filtering product lists.
type ProductFilterParams struct {
	MaxWidth       float32
	Limit          int
	Color          string
	InventoryLevel int     // Assumed filter: inventory_level >= ? (if > 0)
	Price          float32 // Assumed filter: price <= ? (if > 0)
}

type ProductRepository interface {
	InsertProduct(product *models.Product) (int, error)
	GetProductPrice(id int) (float32, error)
	GetProductByName(name string) (*models.Product, error)
	GetProductByID(id int) (*models.Product, error)
	UpdateProductByID(id int, product *models.Product) error
	DeleteProductByID(id int) error
	CountProductByID(id int) (int, error)

	// Filtering
	GetProducts(productType models.ProductType, params ProductFilterParams) ([]*models.Product, error)
	CountProducts(productType models.ProductType, params ProductFilterParams) (int, error)

	// Type-specific retrieval
	GetGates(params ProductFilterParams) ([]*models.Product, error)
	GetExtensions(params ProductFilterParams) ([]*models.Product, error)
	GetBundles(params ProductFilterParams) ([]*models.Product, error)

	// Relationships
	GetCompatibleExtensionsByGateID(gateID int) ([]*models.Product, error)
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

// OrderRepository defines the contract for order-related persistence operations.
type OrderRepository interface {
	// Creates a new order from a validated cart, returning the new order ID.
	New(cart *models.Cart) (int, error)

	// Updates the status of an existing order.
	UpdateStatus(orderID int, status models.OrderStatus) error

	// Updates customer-related fields on an order.
	UpdateCustomerDetails(orderID int, details CustomerDetails) error

	// These are internal utilities, typically only used during transaction within `New()`,
	// so if you don't plan to call them outside, you can omit them from the public interface.
	InsertItem(tx *sql.Tx, orderID int, item models.CartItem) error
	InsertComponent(tx *sql.Tx, orderID int, orderItemID int, component models.CartItemComponent) error
}

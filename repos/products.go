package repos

import (
	"database/sql"
	"errors"
	"fmt"
	"strings" // Import strings package for query building

	"github.com/seanomeara96/gates/models"
)

type ProductRepo struct {
	db *sql.DB
}

func NewProductRepo(db *sql.DB) *ProductRepo {
	return &ProductRepo{db}
}

// Define a custom type for the product
type ProductType string

// Define constants representing the product values
const (
	Gate      ProductType = "gate"
	Extension ProductType = "extension"
	Bundle    ProductType = "bundle"
)

type scannable interface {
	Scan(dest ...any) error
}

// scanProductFromRow now includes inventory_level
func scanProductFromRow(row scannable) (*models.Product, error) {
	var product models.Product
	err := row.Scan(
		&product.Id,
		&product.Type,
		&product.Name,
		&product.Width,
		&product.Price,
		&product.Img,
		&product.Color,
		&product.Tolerance,
		&product.InventoryLevel, // Added inventory_level scanning
	)
	if err != nil {
		// Check specifically for sql.ErrNoRows which might be handled differently upstream
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("could not scan product from row(s): %w", err)
	}
	return &product, nil
}

// InsertProduct now includes inventory_level
func (r *ProductRepo) InsertProduct(product *models.Product) (sql.Result, error) {
	if r.db == nil {
		return nil, errors.New("insert product requires a non nil db pointer")
	}

	res, err := r.db.Exec(
		`INSERT INTO
			products (
				type,
				name,
				width,
				price,
				img,
				color,
				tolerance,
				inventory_level -- Added column
			)
		VALUES
			(?, ?, ?, ?, ?, ?, ?, ?)`, // Added placeholder
		product.Type,
		product.Name,
		product.Width,
		product.Price,
		product.Img,
		product.Color,
		product.Tolerance,
		product.InventoryLevel, // Added value
	)
	if err != nil {
		return nil, fmt.Errorf("error inserting product into database: %w", err)
	}

	return res, nil

}

func (r *ProductRepo) GetProductPrice(id int) (float32, error) {
	if r.db == nil {
		return 0, errors.New("get product price requires a non nil db pointer")
	}

	var price float32
	if err := r.db.QueryRow("SELECT price FROM products WHERE id = ?", id).Scan(&price); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("product with id %d not found: %w", id, err)
		}
		return 0, fmt.Errorf("error getting price for product id %d: %w", id, err)
	}
	return price, nil
}

// GetProductByName now selects inventory_level
func (r *ProductRepo) GetProductByName(name string) (*models.Product, error) {
	if r.db == nil {
		return nil, errors.New("need a non nil db to get products by name")
	}
	product, err := scanProductFromRow(
		r.db.QueryRow("SELECT id, type, name, width, price, img, color, tolerance, inventory_level FROM products WHERE name = ?", name), // Added inventory_level
	)
	if err != nil {
		// Let scanProductFromRow handle ErrNoRows if needed, or return the wrapped error
		return nil, err
	}
	return product, nil
}

// Updated ProductFilterParams struct definition (as provided by user)
type ProductFilterParams struct {
	MaxWidth       float32
	Limit          int // Limit is ignored by CountProducts but kept for consistency with GetProducts
	Color          string
	InventoryLevel int     // Filter: inventory_level >= ? (if > 0)
	Price          float32 // Filter: price <= ? (if > 0)
}

// GetProducts updated to use new filter params
func (r *ProductRepo) GetProducts(productType ProductType, params ProductFilterParams) ([]*models.Product, error) {
	if r.db == nil {
		return nil, errors.New("get products requires a non nil db pointer")
	}

	args := []any{productType}
	conditions := []string{"type = ?"} // Start with the mandatory type condition

	// Dynamically build WHERE clauses based on provided params
	if params.MaxWidth > 0 {
		conditions = append(conditions, "width < ?")
		args = append(args, params.MaxWidth)
	}
	if params.Color != "" {
		conditions = append(conditions, "color = ?")
		args = append(args, params.Color)
	}
	// Filter for inventory level >= value, only if value > 0
	if params.InventoryLevel > 0 {
		conditions = append(conditions, "inventory_level >= ?")
		args = append(args, params.InventoryLevel)
	}
	// Filter for price <= value, only if value > 0
	if params.Price > 0 {
		conditions = append(conditions, "price <= ?")
		args = append(args, params.Price)
	}

	// Construct the base query
	baseQuery := "SELECT id, type, name, width, price, img, color, tolerance, inventory_level FROM products WHERE " + strings.Join(conditions, " AND ")

	// Add LIMIT clause if provided
	if params.Limit > 0 {
		baseQuery += " LIMIT ?"
		args = append(args, params.Limit)
	}

	// Execute the query
	var products []*models.Product
	rows, err := r.db.Query(baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying products with filters (%s): %w", baseQuery, err)
	}
	defer rows.Close()

	// Scan rows
	for rows.Next() {
		product, err := scanProductFromRow(rows)
		if err != nil {
			return nil, err // Error already wrapped in scanProductFromRow
		}
		products = append(products, product)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating product rows: %w", err)
	}
	return products, nil
}

// CountProducts updated to use new filter params (ignoring Limit)
func (r *ProductRepo) CountProducts(productType ProductType, params ProductFilterParams) (int, error) {
	if r.db == nil {
		return 0, errors.New("count products requires a non nil db pointer")
	}

	args := []any{productType}
	conditions := []string{"type = ?"} // Start with the mandatory type condition

	// Dynamically build WHERE clauses based on provided params (same logic as GetProducts, excluding Limit)
	if params.MaxWidth > 0 {
		conditions = append(conditions, "width < ?")
		args = append(args, params.MaxWidth)
	}
	if params.Color != "" {
		conditions = append(conditions, "color = ?")
		args = append(args, params.Color)
	}
	// Filter for inventory level >= value, only if value > 0
	if params.InventoryLevel > 0 {
		conditions = append(conditions, "inventory_level >= ?")
		args = append(args, params.InventoryLevel)
	}
	// Filter for price <= value, only if value > 0
	if params.Price > 0 {
		conditions = append(conditions, "price <= ?")
		args = append(args, params.Price)
	}

	// Construct the count query
	baseQuery := "SELECT COUNT(*) FROM products WHERE " + strings.Join(conditions, " AND ")

	// Execute the count query
	var count int
	row := r.db.QueryRow(baseQuery, args...)
	err := row.Scan(&count)
	if err != nil {
		// sql.ErrNoRows is not expected for COUNT(*), but handle other potential DB errors
		return 0, fmt.Errorf("error counting products with filters (%s): %w", baseQuery, err)
	}

	return count, nil
}

// GetCompatibleExtensionsByGateID now selects inventory_level
func (r *ProductRepo) GetCompatibleExtensionsByGateID(gateID int) ([]*models.Product, error) {
	if r.db == nil {
		return nil, errors.New("get compatible extensions requires a non nil db pointer")
	}

	var extensions []*models.Product
	rows, err := r.db.Query(
		`SELECT
			p.id,
			p.type,
			p.name,
			p.width,
			p.price,
			p.img,
			p.color,
			p.tolerance,
			p.inventory_level -- Added inventory_level
		FROM
			products p
		INNER JOIN
			compatibles c
		ON
			p.id = c.extension_id
		WHERE
			c.gate_id = ? -- Qualify gate_id with table alias c
        AND p.type = ?`, // Also ensure we only get extensions
		gateID,
		Extension, // Explicitly filter for extensions
	)
	if err != nil {
		return nil, fmt.Errorf("error querying compatible extensions for gate ID %d: %w", gateID, err)
	}
	defer rows.Close()
	for rows.Next() {
		extension, err := scanProductFromRow(rows)
		if err != nil {
			return nil, err
		}
		extensions = append(extensions, extension)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating compatible extension rows: %w", err)
	}
	return extensions, nil
}

// UpdateProductByID now updates inventory_level
func (r *ProductRepo) UpdateProductByID(productID int, product *models.Product) error {
	if r.db == nil {
		return errors.New("update product requires a non nil db pointer")
	}

	res, err := r.db.Exec(
		`UPDATE products
		SET
			type = ?,
			name = ?,
			width = ?,
			price = ?,
			img = ?,
			color = ?,
			tolerance = ?,
			inventory_level = ? -- Added inventory_level
		WHERE id = ?`,
		product.Type,
		product.Name,
		product.Width,
		product.Price,
		product.Img,
		product.Color,
		product.Tolerance,
		product.InventoryLevel, // Added product.InventoryLevel
		productID,
	)

	if err != nil {
		return fmt.Errorf("error updating product with ID %d: %w", productID, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		// Log this error but don't necessarily fail the update
		fmt.Printf("Warning: could not get rows affected after update for product ID %d: %v\n", productID, err)
	} else if rowsAffected == 0 {
		// Consider returning a specific error like sql.ErrNoRows or a custom not found error
		return fmt.Errorf("no product found with ID %d to update", productID)
	}

	return nil
}

func (r *ProductRepo) DeleteProductByID(productID int) error {
	if r.db == nil {
		return errors.New("delete product requires a non nil db pointer")
	}

	res, err := r.db.Exec("DELETE FROM products WHERE id = ?", productID)
	if err != nil {
		return fmt.Errorf("error deleting product with ID %d: %w", productID, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		fmt.Printf("Warning: could not get rows affected after delete for product ID %d: %v\n", productID, err)
	} else if rowsAffected == 0 {
		// Consider returning a specific error like sql.ErrNoRows or a custom not found error
		return fmt.Errorf("no product found with ID %d to delete", productID)
	}

	return nil
}

// CreateProductParams now includes InventoryLevel
type CreateProductParams struct {
	Type           string
	Name           string
	Width          float32
	Price          float32
	Img            string
	Tolerance      float32
	Color          string
	InventoryLevel int // Added InventoryLevel
}

var validProductTypes = map[ProductType]bool{ // Use slice for easier iteration
	Gate:      true,
	Extension: true,
	// Consider if Bundle should be creatable this way.
}

func (r *ProductRepo) CreateProduct(params CreateProductParams) (int64, error) {

	// Validate input parameters
	if params.Name == "" || params.Type == "" || params.Color == "" {
		return 0, errors.New("name, type, and color are required")
	}

	hasValidType := validProductTypes[ProductType(params.Type)]

	if !hasValidType {
		return 0, fmt.Errorf("invalid product type '%s'. Valid types are: Gate or Extension", params.Type)
	}

	if params.Price <= 0.0 || params.Width <= 0.0 { // Check <= 0
		return 0, errors.New("price and width must be greater than 0")
	}
	if params.InventoryLevel < 0 {
		return 0, errors.New("inventory level cannot be negative")
	}

	// Check for existing product by name using QueryRow for efficiency
	var existingID int
	err := r.db.QueryRow("SELECT id FROM products WHERE name = ?", params.Name).Scan(&existingID)
	if err == nil {
		// No error means a product was found
		return 0, fmt.Errorf("product with name '%s' already exists (ID: %d)", params.Name, existingID)
	} else if !errors.Is(err, sql.ErrNoRows) {
		// An actual DB error occurred (other than not found)
		return 0, fmt.Errorf("error checking for existing product by name '%s': %w", params.Name, err)
	}
	// If err is sql.ErrNoRows, proceed with creation

	product := &models.Product{
		Id:             0, // ID will be auto-generated
		Type:           params.Type,
		Name:           params.Name,
		Width:          params.Width,
		Price:          params.Price,
		Img:            params.Img,
		Color:          params.Color,
		Tolerance:      params.Tolerance,
		InventoryLevel: params.InventoryLevel, // Added InventoryLevel
	}

	res, err := r.InsertProduct(product)
	if err != nil {
		return 0, err // Error already wrapped in InsertProduct
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error getting last insert ID after creating product '%s': %w", params.Name, err)
	}

	return lastID, nil
}

// GetProductByID now selects inventory_level
func (r *ProductRepo) GetProductByID(productID int) (*models.Product, error) {
	if r.db == nil {
		return nil, errors.New("get product by ID requires a non nil db pointer")
	}
	product, err := scanProductFromRow(
		r.db.QueryRow("SELECT id, type, name, width, price, img, color, tolerance, inventory_level FROM products WHERE id = ?", productID), // Added inventory_level
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("product with id %d not found", productID) // More specific error for not found
		}
		// Keep the wrapped error from scanProductFromRow for other errors
		return nil, err
	}
	return product, nil
}

// --- Wrapper functions using GetProducts ---

func (r *ProductRepo) GetGates(params ProductFilterParams) ([]*models.Product, error) {
	gates, err := r.GetProducts(Gate, params)
	if err != nil {
		return nil, err // Error already wrapped in GetProducts
	}
	for i := range gates {
		gates[i].Qty = 1 // Consider if this logic belongs elsewhere
	}
	return gates, nil
}

func (r *ProductRepo) GetExtensions(params ProductFilterParams) ([]*models.Product, error) {
	extensions, err := r.GetProducts(Extension, params)
	if err != nil {
		return nil, err // Error already wrapped in GetProducts
	}
	for i := range extensions {
		extensions[i].Qty = 1 // Consider if this logic belongs elsewhere
	}
	return extensions, nil
}

func (r *ProductRepo) GetBundles(params ProductFilterParams) ([]*models.Product, error) {
	bundles, err := r.GetProducts(Bundle, params)
	if err != nil {
		return nil, err // Error already wrapped in GetProducts
	}
	// Bundles likely don't need Qty=1 set here.
	return bundles, nil
}

// Add this function to your ProductRepo

// CountProductByID counts products matching a specific ID.
// Returns 1 if the product exists, 0 otherwise.
func (r *ProductRepo) CountProductByID(productID int) (int, error) {
	if r.db == nil {
		return 0, errors.New("count product by ID requires a non nil db pointer")
	}

	var count int
	query := "SELECT COUNT(*) FROM products WHERE id = ?"
	err := r.db.QueryRow(query, productID).Scan(&count)
	if err != nil {
		// sql.ErrNoRows should NOT occur for COUNT(*), but handle other errors
		return 0, fmt.Errorf("error counting product with ID %d: %w", productID, err)
	}

	// count will be 0 or 1 because id is a primary key
	return count, nil
}

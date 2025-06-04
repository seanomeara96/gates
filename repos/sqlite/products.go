package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/seanomeara96/gates/models" // Assuming your models package path
	"github.com/seanomeara96/gates/repos"
)

// ProductRepo handles database operations for products.
// It should focus solely on data access (CRUD) and not contain business logic validation.
type ProductRepo struct {
	db *sql.DB
}

// NewProductRepo creates a new instance of ProductRepo.
func NewProductRepo(db *sql.DB) *ProductRepo {
	if db == nil {
		// Consider panic or returning an error if a nil db is critical
		panic("database connection is nil for ProductRepo")
	}
	return &ProductRepo{db}
}

// scannable interface decouples scanning logic from specific sql types (*sql.Row, *sql.Rows).
type scannable interface {
	Scan(dest ...any) error
}

// scanProductFromRow scans a single product row into a models.Product struct.
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
		&product.InventoryLevel,
	)
	if err != nil {
		// Specifically check for ErrNoRows and return it so callers can distinguish
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		// Wrap other errors for context
		return nil, fmt.Errorf("could not scan product from row: %w", err)
	}
	return &product, nil
}

// InsertProduct inserts a new product record into the database.
// Assumes the input product object has been validated by the service layer.
func (r *ProductRepo) InsertProduct(product *models.Product) (int, error) {
	// Basic check for nil db pointer remains relevant
	if r.db == nil {
		return 0, errors.New("database connection is nil")
	}

	// The product object is assumed to be valid at this point.
	// The repository's job is just to execute the INSERT statement.
	res, err := r.db.Exec(
		`INSERT INTO products (
			type, name, width, price, img, color, tolerance, inventory_level
		 ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		product.Type,
		product.Name,
		product.Width,
		product.Price,
		product.Img,
		product.Color,
		product.Tolerance,
		product.InventoryLevel,
	)
	if err != nil {
		// Handle potential DB constraint errors if needed, or just wrap
		// Example: Could check for SQLite UNIQUE constraint error code here
		return 0, fmt.Errorf("database error inserting product: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// GetProductPrice retrieves only the price for a given product ID.
func (r *ProductRepo) GetProductPrice(id int) (float32, error) {
	if r.db == nil {
		return 0, errors.New("database connection is nil")
	}
	var price float32
	err := r.db.QueryRow("SELECT price FROM products WHERE id = ?", id).Scan(&price)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Return a clear "not found" error
			return 0, fmt.Errorf("product with id %d not found", id)
		}
		return 0, fmt.Errorf("error getting price for product id %d: %w", id, err)
	}
	return price, nil
}

// GetProductByName retrieves a product by its unique name.
// Returns sql.ErrNoRows if no product with that name exists.
func (r *ProductRepo) GetProductByName(name string) (*models.Product, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}
	product, err := scanProductFromRow(
		r.db.QueryRow("SELECT id, type, name, width, price, img, color, tolerance, inventory_level FROM products WHERE name = ?", name),
	)
	// scanProductFromRow handles wrapping and sql.ErrNoRows detection
	return product, err
}

// GetProducts retrieves a list of products based on type and filter parameters.
func (r *ProductRepo) GetProducts(params repos.ProductFilterParams) ([]*models.Product, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}

	baseSelect := "SELECT id, type, name, width, price, img, color, tolerance, inventory_level FROM products"
	args := []any{}
	conditions := []string{}

	if params.Type != "" {
		conditions = append(conditions, "type = ?")
		args = append(args, params.Type)
	}

	if params.MaxWidth > 0 {
		conditions = append(conditions, "width < ?")
		args = append(args, params.MaxWidth)
	}
	if params.Color != "" {
		conditions = append(conditions, "color = ?")
		args = append(args, params.Color)
	}
	if params.InventoryLevel > 0 { // Filter if InventoryLevel is explicitly positive
		conditions = append(conditions, "inventory_level >= ?")
		args = append(args, params.InventoryLevel)
	}
	if params.Price > 0 { // Filter if Price is explicitly positive
		conditions = append(conditions, "price <= ?")
		args = append(args, params.Price)
	}

	if len(conditions) > 0 {
		baseSelect += " WHERE " + strings.Join(conditions, " AND ")

	}
	query := baseSelect
	// Add LIMIT clause if provided
	if params.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, params.Limit)
	}
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying products: %w", err)
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		product, err := scanProductFromRow(rows)
		if err != nil {
			// scanProductFromRow already wraps errors
			return nil, err
		}
		products = append(products, product)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating product rows: %w", err)
	}
	return products, nil
}

// CountProducts counts products based on type and filter parameters (ignoring Limit).
func (r *ProductRepo) CountProducts(productType models.ProductType, params repos.ProductFilterParams) (int, error) {
	if r.db == nil {
		return 0, errors.New("database connection is nil")
	}

	baseSelect := "SELECT COUNT(*) FROM products"
	args := []any{}
	conditions := []string{}

	if params.Type != "" {
		conditions = append(conditions, "type = ?")
		args = append(args, params.Type)
	}

	if params.MaxWidth > 0 {
		conditions = append(conditions, "width < ?")
		args = append(args, params.MaxWidth)
	}
	if params.Color != "" {
		conditions = append(conditions, "color = ?")
		args = append(args, params.Color)
	}
	if params.InventoryLevel > 0 { // Filter if InventoryLevel is explicitly positive
		conditions = append(conditions, "inventory_level >= ?")
		args = append(args, params.InventoryLevel)
	}
	if params.Price > 0 { // Filter if Price is explicitly positive
		conditions = append(conditions, "price <= ?")
		args = append(args, params.Price)
	}

	query := baseSelect + " WHERE " + strings.Join(conditions, " AND ")
	var count int
	err := r.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		// sql.ErrNoRows is not expected for COUNT(*), handle other DB errors
		return 0, fmt.Errorf("error counting products: %w", err)
	}
	return count, nil
}

// GetCompatibleExtensionsByGateID retrieves extensions compatible with a given gate ID.
func (r *ProductRepo) GetCompatibleExtensionsByGateID(gateID int) ([]*models.Product, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}

	// Explicitly selecting Extension type for clarity and safety
	query := `SELECT
				p.id, p.type, p.name, p.width, p.price, p.img, p.color, p.tolerance, p.inventory_level
			  FROM products p
			  INNER JOIN compatibles c ON p.id = c.extension_id
			  WHERE c.gate_id = ? AND p.type = ?`

	rows, err := r.db.Query(query, gateID, models.ProductTypeExtension)
	if err != nil {
		return nil, fmt.Errorf("error querying compatible extensions for gate ID %d: %w", gateID, err)
	}
	defer rows.Close()

	var extensions []*models.Product
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

// UpdateProductByID updates an existing product record.
// Assumes the input product object has been validated by the service layer.
func (r *ProductRepo) UpdateProductByID(productID int, product *models.Product) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}

	res, err := r.db.Exec(
		`UPDATE products SET
			type = ?, name = ?, width = ?, price = ?, img = ?,
			color = ?, tolerance = ?, inventory_level = ?
		 WHERE id = ?`,
		product.Type, product.Name, product.Width, product.Price, product.Img,
		product.Color, product.Tolerance, product.InventoryLevel,
		productID, // Use the passed productID for the WHERE clause
	)
	if err != nil {
		return fmt.Errorf("database error updating product with ID %d: %w", productID, err)
	}

	// Check if any row was actually updated
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		// Log or ignore this error, the update might have succeeded anyway
		return fmt.Errorf("warning: could not get rows affected after update for product ID %d: %v", productID, err)
	} else if rowsAffected == 0 {
		// Return a specific error indicating the product wasn't found
		return fmt.Errorf("no product found with ID %d to update", productID) // Or return sql.ErrNoRows
	}
	return nil
}

// DeleteProductByID deletes a product record by its ID.
func (r *ProductRepo) DeleteProductByID(productID int) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}

	res, err := r.db.Exec("DELETE FROM products WHERE id = ?", productID)
	if err != nil {
		return fmt.Errorf("database error deleting product with ID %d: %w", productID, err)
	}

	// Check if any row was actually deleted
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("warning: could not get rows affected after delete for product ID %d: %v", productID, err)
	} else if rowsAffected == 0 {
		// Return a specific error indicating the product wasn't found
		return fmt.Errorf("no product found with ID %d to delete", productID) // Or return sql.ErrNoRows
	}
	return nil
}

// GetProductByID retrieves a single product by its primary key ID.
// Returns sql.ErrNoRows if no product with that ID exists.
func (r *ProductRepo) GetProductByID(productID int) (*models.Product, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}
	product, err := scanProductFromRow(
		r.db.QueryRow("SELECT id, type, name, width, price, img, color, tolerance, inventory_level FROM products WHERE id = ?", productID),
	)
	// scanProductFromRow handles wrapping and sql.ErrNoRows detection
	return product, err
}

// --- Wrapper Functions (Convenience) ---
// These call GetProducts with specific types. The logic to set Qty=1 might
// arguably belong in the service or application layer, but kept here for now
// as simple convenience wrappers based on previous code.

func (r *ProductRepo) GetGates(params repos.ProductFilterParams) ([]*models.Product, error) {
	params.Type = models.ProductTypeGate
	gates, err := r.GetProducts(params)
	if err != nil {
		return nil, err // Error already wrapped
	}
	for i := range gates {
		gates[i].Qty = 1 // Consider moving this logic elsewhere
	}
	return gates, nil
}

func (r *ProductRepo) GetExtensions(params repos.ProductFilterParams) ([]*models.Product, error) {
	params.Type = models.ProductTypeExtension
	extensions, err := r.GetProducts(params)
	if err != nil {
		return nil, err // Error already wrapped
	}
	for i := range extensions {
		extensions[i].Qty = 1 // Consider moving this logic elsewhere
	}
	return extensions, nil
}

func (r *ProductRepo) GetBundles(params repos.ProductFilterParams) ([]*models.Product, error) {
	params.Type = models.ProductTypeBundle
	// Bundles may not naturally have a single Qty=1 concept.
	return r.GetProducts(params) // Error already wrapped
}

// NOTE: The CreateProduct function and CreateProductParams struct have been removed.
// Responsibility for validating parameters and then calling repo.InsertProduct
// now lies within your Service layer (e.g., ProductService).
// Add this function to your ProductRepo

// CountProductByID counts products matching a specific ID.
// Returns 1 if the product exists, 0 otherwise.
func (r *ProductRepo) CountProductByID(productID int) (int, error) {
	if r.db == nil {
		return 0, errors.New("count product by ID requires a non nil db pointer")
	}

	var count int
	query := "SELECT SUM(inventory_level) FROM products WHERE id = ?"
	err := r.db.QueryRow(query, productID).Scan(&count)
	if err != nil {
		// sql.ErrNoRows should NOT occur for COUNT(*), but handle other errors
		return 0, fmt.Errorf("error counting product with ID %d: %w", productID, err)
	}

	// count will be 0 or 1 because id is a primary key
	return count, nil
}

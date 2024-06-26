package repositories

import (
	"database/sql"

	"github.com/seanomeara96/gates/models"
)

// Define a custom type for the product
type ProductType string

// Define constants representing the product values
const (
	Gate      ProductType = "gate"
	Extension ProductType = "extension"
	Bundle    ProductType = "bundle"
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func scanProductFromRow(row *sql.Row) (*models.Product, error) {
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
	)
	return &product, err
}

func scanProductFromRows(rows *sql.Rows) (*models.Product, error) {
	var product models.Product
	err := rows.Scan(
		&product.Id,
		&product.Type,
		&product.Name,
		&product.Width,
		&product.Price,
		&product.Img,
		&product.Color,
		&product.Tolerance,
	)
	return &product, err
}

func (r *ProductRepository) Create(product *models.Product) (sql.Result, error) {
	return r.db.Exec(
		`INSERT INTO products (type, name, width, price, img, color, tolerance) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		product.Type,
		product.Name,
		product.Width,
		product.Price,
		product.Img,
		product.Color,
		product.Tolerance,
	)

}

func (r *ProductRepository) GetByID(id int) (*models.Product, error) {
	return scanProductFromRow(
		r.db.QueryRow("SELECT id, type, name, width, price, img, color, tolerance FROM products WHERE id = ?", id),
	)
}

func (r *ProductRepository) GetPrice(id int) (float64, error) {
	var price float64
	if err := r.db.QueryRow("SELECT price FROM products WHERE id = ?", id).Scan(&price); err != nil {
		return 0, err
	}
	return price, nil
}

func (r *ProductRepository) GetByName(name string) (*models.Product, error) {
	return scanProductFromRow(
		r.db.QueryRow("SELECT id, type, name, width, price, img, color, tolerance FROM products WHERE name = ?", name),
	)
}

type ProductFilterParams struct {
	MaxWidth float32
	Limit    int
}

func (r *ProductRepository) GetProducts(productType ProductType, params ProductFilterParams) ([]*models.Product, error) {
	filters := []any{productType}

	baseQuery := "SELECT id, type, name, width, price,  img, color, tolerance FROM products WHERE type = ?"
	if params.MaxWidth > 0 {
		baseQuery = baseQuery + " AND width < ?"
		filters = append(filters, params.MaxWidth)
	}

	if params.Limit > 0 {
		baseQuery += " LIMIT ?"
		filters = append(filters, params.Limit)
	}

	var gates []*models.Product
	rows, err := r.db.Query(baseQuery, filters...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		product, err := scanProductFromRows(rows)
		if err != nil {
			return nil, err
		}
		gates = append(gates, product)
	}
	return gates, nil
}

func (r *ProductRepository) GetCompatibleExtensions(gateID int) ([]*models.Product, error) {
	var extensions []*models.Product
	rows, err := r.db.Query(
		"SELECT p.id, p.type, p.name, p.width, p.price, p.img, p.color, p.tolerance FROM products p INNER JOIN compatibles c ON p.id = c.extension_id WHERE gate_id = ?",
		gateID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		extension, err := scanProductFromRows(rows)
		if err != nil {
			return nil, err
		}
		extensions = append(extensions, extension)
	}
	return extensions, nil
}

func (r *ProductRepository) Update(product *models.Product) error {
	// Code to update an existing user in the database
	// using the provided SQL database connection (r.db)
	return nil
}

func (r *ProductRepository) Delete(productID int) error {
	// Code to delete a user from the database
	// based on the provided user ID (userID)
	// using the provided SQL database connection (r.db)
	_, err := r.db.Exec("DELETE FROM products WHERE id = ?", productID)
	return err
}

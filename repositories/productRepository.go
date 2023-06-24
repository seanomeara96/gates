package repositories

import (
	"database/sql"

	"github.com/seanomeara96/gates/models"
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(product *models.Product) error {
	// Code to insert a new user into the database
	// using the provided SQL database connection (r.db)
	return nil
}

func productFields(product *models.Product) (*int, *string, *string, *float32, *float32, *string, *string, *float32) {
	return &product.Id,
		&product.Type,
		&product.Name,
		&product.Width,
		&product.Price,
		&product.Img,
		&product.Color,
		&product.Tolerance
}

func scanProductFromRow(row *sql.Row, product *models.Product) (*models.Product, error) {
	err := row.Scan(productFields(product))
	if err != nil {
		return nil, err
	}
	return product, nil
}
func scanProductFromRows(rows *sql.Rows, product *models.Product) (*models.Product, error) {
	err := rows.Scan(productFields(product))
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (r *ProductRepository) GetByID(productID int) (*models.Product, error) {
	var product *models.Product
	// Code to fetch a user from the database
	// based on the provided user ID (userID)
	// using the provided SQL database connection (r.db)
	row := r.db.QueryRow("SELECT id, type, name, width, price, img, color, tolerance FROM products WHERE id = ?", productID)
	product, err := scanProductFromRow(row, product)
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (r *ProductRepository) GetByName(name string) (*models.Product, error) {
	var product *models.Product
	row := r.db.QueryRow("SELECT id, type, name, width, price, img, color, tolerance FROM products WHERE name = ?", name)
	product, err := scanProductFromRow(row, product)
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (r *ProductRepository) GetGates() ([]*models.Product, error) {
	var gates []*models.Product
	rows, err := r.db.Query("SELECT id, type, name, width, price,  img, color, tolerance FROM products WHERE type = 'gate'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		product, err := scanProductFromRows(rows, &models.Product{})
		if err != nil {
			return nil, err
		}
		gates = append(gates, product)
	}
	return gates, nil
}

func (r *ProductRepository) GetExtensions() ([]*models.Product, error) {
	var extensions []*models.Product
	rows, err := r.db.Query("SELECT id, type, name, width, price, img, color, tolerance FROM products where type = 'extension'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		product, err := scanProductFromRows(rows, &models.Product{})
		if err != nil {
			return nil, err
		}
		extensions = append(extensions, product)
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
	if err != nil {
		return err
	}
	return nil
}

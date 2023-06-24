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

func (r *ProductRepository) GetByID(productID int) (*models.Product, error) {
	var product models.Product
	// Code to fetch a user from the database
	// based on the provided user ID (userID)
	// using the provided SQL database connection (r.db)
	err := r.db.QueryRow("SELECT id, name, width, price, img, color WHERE id = ?", productID).Scan(
		&product.Id,
		&product.Name,
		&product.Width,
		&product.Price,
		&product.Img,
		&product.Color,
	)
	if err != nil {
		return nil, err
	}

	return &product, nil
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

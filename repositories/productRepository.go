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

func scanProductFromRow(row *sql.Row, product *models.Product) (*models.Product, error) {
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
	if err != nil {
		return nil, err
	}
	return product, nil
}
func scanProductFromRows(rows *sql.Rows, product *models.Product) (*models.Product, error) {
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
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (r *ProductRepository) GetByID(productID int) (*models.Product, error) {
	// Code to fetch a user from the database
	// based on the provided user ID (userID)
	// using the provided SQL database connection (r.db)
	row := r.db.QueryRow("SELECT id, type, name, width, price, img, color, tolerance FROM products WHERE id = ?", productID)
	product, err := scanProductFromRow(row, &models.Product{})
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (r *ProductRepository) GetByName(name string) (*models.Product, error) {
	row := r.db.QueryRow("SELECT id, type, name, width, price, img, color, tolerance FROM products WHERE name = ?", name)
	product, err := scanProductFromRow(row, &models.Product{})
	if err != nil {
		return nil, err
	}
	return product, nil
}

type ProductFilterParams struct {
	MaxWidth float32
	Limit    int
}

func (r *ProductRepository) GetGates(params ProductFilterParams) ([]*models.Product, error) {
	filters := []any{}
	baseQuery := "SELECT id, type, name, width, price,  img, color, tolerance FROM products WHERE type = 'gate'"
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
		product, err := scanProductFromRows(rows, &models.Product{})
		if err != nil {
			return nil, err
		}
		gates = append(gates, product)
	}
	return gates, nil
}

func (r *ProductRepository) GetCompatibleExtensions(gateID int) ([]*models.Product, error) {
	var extensions []*models.Product
	rows, err := r.db.Query("SELECT p.id, p.type, p.name, p.width, p.price, p.img, p.color, p.tolerance FROM products p INNER JOIN compatibles c ON p.id = c.extension_id WHERE gate_id = ?", gateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		extension, err := scanProductFromRows(rows, &models.Product{})
		if err != nil {
			return nil, err
		}
		extensions = append(extensions, extension)
	}
	return extensions, nil
}

func (r *ProductRepository) GetExtensions(params ProductFilterParams) ([]*models.Product, error) {
	filters := []any{}
	query := "SELECT id, type, name, width, price, img, color, tolerance FROM products where type = 'extension'"
	if params.MaxWidth > 0 {
		query += " AND width < ?"
		filters = append(filters, params.MaxWidth)
	}

	if params.Limit > 0 {
		query += " LIMIT ?"
		filters = append(filters, params.Limit)
	}

	var extensions []*models.Product
	rows, err := r.db.Query(query, filters...)
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

func (r *ProductRepository) GetBundles(params ProductFilterParams) ([]*models.Product, error) {
	filters := []any{}
	query := "SELECT id, type, name, width, price,  img, color, tolerance FROM products WHERE type = 'bundle'"
	if params.MaxWidth > 0 {
		query = query + " AND width < ?"
		filters = append(filters, params.MaxWidth)
	}

	if params.Limit > 0 {
		query += " LIMIT ?"
		filters = append(filters, params.Limit)
	}

	var bundles []*models.Product
	rows, err := r.db.Query(query, filters...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		product, err := scanProductFromRows(rows, &models.Product{})
		if err != nil {
			return nil, err
		}
		bundles = append(bundles, product)
	}
	return bundles, nil
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

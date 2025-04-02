package repos

import (
	"database/sql"
	"errors"
	"fmt"

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
	)
	if err != nil {
		return nil, fmt.Errorf("could not scan product from row(s): %w", err)
	}
	return &product, nil
}

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
				tolerance
			)
		VALUES
			(?, ?, ?, ?, ?, ?, ?)`,
		product.Type,
		product.Name,
		product.Width,
		product.Price,
		product.Img,
		product.Color,
		product.Tolerance,
	)
	if err != nil {
		return res, fmt.Errorf("error inserting product into database: %w", err)
	}

	return res, nil

}

func (r *ProductRepo) GetProductPrice(id int) (float32, error) {
	if r.db == nil {
		return 0, errors.New("get product price requires a non nil db pointer")
	}

	var price float32
	if err := r.db.QueryRow("SELECT price FROM products WHERE id = ?", id).Scan(&price); err != nil {
		return 0, err
	}
	return price, nil
}

func (r *ProductRepo) GetProductByName(name string) (*models.Product, error) {
	if r.db == nil {
		return nil, errors.New("need a non nil db to get products by name")
	}
	return scanProductFromRow(
		r.db.QueryRow("SELECT id, type, name, width, price, img, color, tolerance FROM products WHERE name = ?", name),
	)
}

type ProductFilterParams struct {
	MaxWidth float32
	Limit    int
}

func (r *ProductRepo) GetProducts(productType ProductType, params ProductFilterParams) ([]*models.Product, error) {
	if r.db == nil {
		return nil, errors.New("get products requires a non nil db pointer")
	}

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

	var products []*models.Product
	rows, err := r.db.Query(baseQuery, filters...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		product, err := scanProductFromRow(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, nil
}

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
			p.tolerance
		FROM
			products p
		INNER JOIN
			compatibles c
		ON
			p.id = c.extension_id
		WHERE
			gate_id = ?`,
		gateID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		extension, err := scanProductFromRow(rows)
		if err != nil {
			return nil, err
		}
		extensions = append(extensions, extension)
	}
	return extensions, nil
}

func (r *ProductRepo) UpdateProductByID(productID int, product *models.Product) error {
	if r.db == nil {
		return errors.New("update product requires a non nil db pointer")
	}

	_, err := r.db.Exec(
		`UPDATE products
		SET
			type = ?,
			name = ?,
			width = ?,
			price = ?,
			img = ?,
			color = ?,
			tolerance = ?
		WHERE id = ?`,
		product.Type,
		product.Name,
		product.Width,
		product.Price,
		product.Img,
		product.Color,
		product.Tolerance,
		productID,
	)

	if err != nil {
		return fmt.Errorf("error updating product with ID %d: %w", productID, err)
	}

	return nil
}

func (r *ProductRepo) DeleteProductByID(productID int) error {
	if r.db == nil {
		return errors.New("delete product requires a non nil db pointer")
	}

	// Code to delete a user from the database
	// based on the provided user ID (userID)
	// using the provided SQL database connection (r.db)
	_, err := r.db.Exec("DELETE FROM products WHERE id = ?", productID)
	return err
}

type CreateProductParams struct {
	Type      string
	Name      string
	Width     float32
	Price     float32
	Img       string
	Tolerance float32
	Color     string
}

func (r *ProductRepo) CreateProduct(params CreateProductParams) (int64, error) {
	validProductTypes := [2]ProductType{
		Gate,
		Extension,
	}
	// Validate input parameters
	if params.Name == "" || params.Type == "" || params.Color == "" {
		return 0, errors.New("name, type, and color are required")
	}

	hasValidType := false
	for _, validProductType := range validProductTypes {
		if params.Type == string(validProductType) {
			hasValidType = true
		}
	}

	if !hasValidType {
		return 0, errors.New("does not have a valid product type")
	}

	if params.Price == 0.0 || params.Width == 0.0 {
		return 0, errors.New("price and width must be greater than 0")
	}

	existingProduct, err := r.GetProductByName(params.Name)
	if err != nil {
		return 0, err
	}

	if existingProduct != nil {
		return 0, errors.New("product already exists")
	}

	product := &models.Product{
		Id:        0,
		Type:      params.Type,
		Name:      params.Name,
		Width:     params.Width,
		Price:     params.Price,
		Img:       params.Img,
		Color:     params.Color,
		Tolerance: params.Tolerance,
	}

	row, err := r.InsertProduct(product)
	if err != nil {
		return 0, err
	}

	return row.LastInsertId()
}

func (r *ProductRepo) GetProductByID(productID int) (*models.Product, error) {
	return scanProductFromRow(
		r.db.QueryRow("SELECT id, type, name, width, price, img, color, tolerance FROM products WHERE id = ?", productID),
	)
}

func (r *ProductRepo) GetGates(params ProductFilterParams) ([]*models.Product, error) {
	gates, err := r.GetProducts(Gate, params)
	if err != nil {
		return nil, err
	}

	for i := range gates {
		gates[i].Qty = 1
	}
	return gates, nil
}

func (r *ProductRepo) GetExtensions(params ProductFilterParams) ([]*models.Product, error) {
	extensions, err := r.GetProducts(Extension, params)
	if err != nil {
		return nil, err
	}

	for i := range extensions {
		extensions[i].Qty = 1
	}

	return extensions, nil
}

func (r *ProductRepo) GetBundles(params ProductFilterParams) ([]*models.Product, error) {
	bundles, err := r.GetProducts(Bundle, params)
	if err != nil {
		return nil, err
	}

	return bundles, nil
}

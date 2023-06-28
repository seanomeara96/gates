package services

import (
	"errors"

	"github.com/seanomeara96/gates/models"
	"github.com/seanomeara96/gates/repositories"
)

type ProductService struct {
	productRepository *repositories.ProductRepository
}

func NewProductService(productRepository *repositories.ProductRepository) *ProductService {
	return &ProductService{productRepository: productRepository}
}

type createProductParams struct {
	Type      string
	Name      string
	Width     float32
	Price     float32
	Img       string
	Tolerance float32
	Color     string
}

type ProductFilterParams = repositories.ProductFilterParams

func (s *ProductService) CreateProduct(params createProductParams) (*models.Product, error) {
	validProductTypes := [2]string{
		"gate",
		"extension",
	}
	// Validate input parameters
	if params.Name == "" || params.Type == "" || params.Color == "" {
		return nil, errors.New("name, type, and color are required")
	}

	hasValidType := false
	for _, validProductType := range validProductTypes {
		if params.Type == validProductType {
			hasValidType = true
		}
	}

	if !hasValidType {
		return nil, errors.New("does not have a valid product type")
	}

	if params.Price == 0.0 || params.Width == 0.0 {
		return nil, errors.New("price and width must be greater than 0")
	}

	// Check if the email is already registered
	existingProduct, err := s.productRepository.GetByName(params.Name)
	if err != nil {
		return nil, err
	}
	if existingProduct != nil {
		return nil, errors.New("email is already registered")
	}

	// Create a new user instance
	product := &models.Product{
		Id:   0,
		Type: params.Type,
	}

	// Save the user to the database
	err = s.productRepository.Create(product)
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (s *ProductService) GetProductByID(productID int) (*models.Product, error) {
	return s.productRepository.GetByID(productID)
}

func (s *ProductService) GetGates(params ProductFilterParams) ([]*models.Product, error) {
	return s.productRepository.GetProducts(repositories.Gate, params)
}

func (s *ProductService) GetExtensions(params ProductFilterParams) ([]*models.Product, error) {
	return s.productRepository.GetProducts(repositories.Extension, params)

}

func (s *ProductService) GetBundles(params ProductFilterParams) ([]*models.Product, error) {
	return s.productRepository.GetProducts(repositories.Bundle, params)
}

func (s *ProductService) GetCompatibleExtensions(gateID int) ([]*models.Product, error) {
	return s.productRepository.GetCompatibleExtensions(gateID)
}

// Other methods for user-related operations (e.g., UpdateUser, DeleteUser, etc.)

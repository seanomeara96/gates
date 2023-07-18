package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/seanomeara96/gates/models"
	"github.com/seanomeara96/gates/repositories"
)

type ProductService struct {
	productRepository *repositories.ProductRepository
	productCache      *cache.Cache
}

func NewProductService(productRepository *repositories.ProductRepository) *ProductService {
	defaultExpiration := time.Minute * 5
	cleanupInterval := time.Minute * 10
	c := cache.New(defaultExpiration, cleanupInterval)
	return &ProductService{productRepository: productRepository, productCache: c}
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
	_, err = s.productRepository.Create(product)
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (s *ProductService) GetProductByID(productID int) (*models.Product, error) {
	return s.productRepository.GetByID(productID)
}

func (s *ProductService) GetGates(params ProductFilterParams) ([]*models.Product, error) {
	cacheString := fmt.Sprintf("gates;max-width:%f;limit:%d;", params.MaxWidth, params.Limit)
	g, gatesFound := s.productCache.Get(cacheString)
	if gatesFound {
		return g.([]*models.Product), nil
	}
	gates, err := s.productRepository.GetProducts(repositories.Gate, params)
	if err != nil {
		return nil, err
	}
	s.productCache.Set(cacheString, gates, time.Minute*5)
	return gates, nil
}

func (s *ProductService) GetExtensions(params ProductFilterParams) ([]*models.Product, error) {
	cacheString := fmt.Sprintf("extensions;max-width:%f;limit:%d;", params.MaxWidth, params.Limit)
	e, extensionsFound := s.productCache.Get(cacheString)
	if extensionsFound {
		return e.([]*models.Product), nil
	}
	extensions, err := s.productRepository.GetProducts(repositories.Extension, params)
	if err != nil {
		return nil, err
	}
	s.productCache.Set(cacheString, extensions, time.Minute*5)
	return extensions, nil
}

func (s *ProductService) GetBundles(params ProductFilterParams) ([]*models.Product, error) {
	cacheString := fmt.Sprintf("bundles;max-width:%f;limit:%d;", params.MaxWidth, params.Limit)
	b, bundlesFound := s.productCache.Get(cacheString)
	if bundlesFound {
		return b.([]*models.Product), nil
	}
	bundles, err := s.productRepository.GetProducts(repositories.Bundle, params)
	if err != nil {
		return nil, err
	}
	s.productCache.Set(cacheString, bundles, time.Minute*5)
	return bundles, nil
}

func (s *ProductService) GetCompatibleExtensions(gateID int) ([]*models.Product, error) {
	return s.productRepository.GetCompatibleExtensions(gateID)
}

// Other methods for user-related operations (e.g., UpdateUser, DeleteUser, etc.)

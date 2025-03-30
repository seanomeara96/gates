package cachedrepos

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/seanomeara96/gates/models"
	"github.com/seanomeara96/gates/repos"
)

type CachedProductRepo struct {
	cache       *cache.Cache
	productRepo *repos.ProductRepo
}

func NewCachedProductRepo(productRepo *repos.ProductRepo) *CachedProductRepo {
	defaultExpiration := time.Minute * 5
	cleanupInterval := time.Minute * 10
	cache := cache.New(defaultExpiration, cleanupInterval)
	return &CachedProductRepo{cache, productRepo}
}

// Using ProductType from repos package
type ProductType = repos.ProductType

// Using constants from repos package
const (
	Gate      = repos.Gate
	Extension = repos.Extension
	Bundle    = repos.Bundle
)

func (r *CachedProductRepo) InsertProduct(product *models.Product) (sql.Result, error) {
	// Clear cache on insert
	r.cache.Flush()
	return r.productRepo.InsertProduct(product)
}

func (r *CachedProductRepo) GetProductPrice(id int) (float64, error) {
	cacheKey := fmt.Sprintf("product_price_%d", id)
	if cachedPrice, found := r.cache.Get(cacheKey); found {
		return cachedPrice.(float64), nil
	}

	price, err := r.productRepo.GetProductPrice(id)
	if err != nil {
		return 0, err
	}

	r.cache.Set(cacheKey, price, cache.DefaultExpiration)
	return price, nil
}

func (r *CachedProductRepo) GetProductByName(name string) (*models.Product, error) {
	cacheKey := fmt.Sprintf("product_by_name_%s", name)
	if cachedProduct, found := r.cache.Get(cacheKey); found {
		return cachedProduct.(*models.Product), nil
	}

	product, err := r.productRepo.GetProductByName(name)
	if err != nil {
		return nil, err
	}

	r.cache.Set(cacheKey, product, cache.DefaultExpiration)
	return product, nil
}

func (r *CachedProductRepo) GetProducts(productType repos.ProductType, params repos.ProductFilterParams) ([]*models.Product, error) {
	cacheKey := fmt.Sprintf("products_%s_maxwidth_%f_limit_%d", productType, params.MaxWidth, params.Limit)
	if cachedProducts, found := r.cache.Get(cacheKey); found {
		return cachedProducts.([]*models.Product), nil
	}

	products, err := r.productRepo.GetProducts(productType, params)
	if err != nil {
		return nil, err
	}

	r.cache.Set(cacheKey, products, cache.DefaultExpiration)
	return products, nil
}

func (r *CachedProductRepo) GetCompatibleExtensionsByGateID(gateID int) ([]*models.Product, error) {
	cacheKey := fmt.Sprintf("compatible_extensions_%d", gateID)
	if cachedExtensions, found := r.cache.Get(cacheKey); found {
		return cachedExtensions.([]*models.Product), nil
	}

	extensions, err := r.productRepo.GetCompatibleExtensionsByGateID(gateID)
	if err != nil {
		return nil, err
	}

	r.cache.Set(cacheKey, extensions, cache.DefaultExpiration)
	return extensions, nil
}

func (r *CachedProductRepo) UpdateProductByID(productID int, product *models.Product) error {
	// Clear cache on update
	r.cache.Flush()
	return r.productRepo.UpdateProductByID(productID, product)
}

func (r *CachedProductRepo) DeleteProductByID(productID int) error {
	// Clear cache on delete
	r.cache.Flush()
	return r.productRepo.DeleteProductByID(productID)
}

func (r *CachedProductRepo) GetProductByID(productID int) (*models.Product, error) {
	cacheKey := fmt.Sprintf("product_by_id_%d", productID)
	if cachedProduct, found := r.cache.Get(cacheKey); found {
		return cachedProduct.(*models.Product), nil
	}

	product, err := r.productRepo.GetProductByID(productID)
	if err != nil {
		return nil, err
	}

	r.cache.Set(cacheKey, product, cache.DefaultExpiration)
	return product, nil
}

func (r *CachedProductRepo) GetGates(params repos.ProductFilterParams) ([]*models.Product, error) {
	cacheKey := fmt.Sprintf("gates_maxwidth_%f_limit_%d", params.MaxWidth, params.Limit)
	if cachedGates, found := r.cache.Get(cacheKey); found {
		return cachedGates.([]*models.Product), nil
	}

	gates, err := r.productRepo.GetGates(params)
	if err != nil {
		return nil, err
	}

	r.cache.Set(cacheKey, gates, cache.DefaultExpiration)
	return gates, nil
}

func (r *CachedProductRepo) GetExtensions(params repos.ProductFilterParams) ([]*models.Product, error) {
	cacheKey := fmt.Sprintf("extensions_maxwidth_%f_limit_%d", params.MaxWidth, params.Limit)
	if cachedExtensions, found := r.cache.Get(cacheKey); found {
		return cachedExtensions.([]*models.Product), nil
	}

	extensions, err := r.productRepo.GetExtensions(params)
	if err != nil {
		return nil, err
	}

	r.cache.Set(cacheKey, extensions, cache.DefaultExpiration)
	return extensions, nil
}

func (r *CachedProductRepo) GetBundles(params repos.ProductFilterParams) ([]*models.Product, error) {
	cacheKey := fmt.Sprintf("bundles_maxwidth_%f_limit_%d", params.MaxWidth, params.Limit)
	if cachedBundles, found := r.cache.Get(cacheKey); found {
		return cachedBundles.([]*models.Product), nil
	}

	bundles, err := r.productRepo.GetBundles(params)
	if err != nil {
		return nil, err
	}

	r.cache.Set(cacheKey, bundles, cache.DefaultExpiration)
	return bundles, nil
}

func (r *CachedProductRepo) CreateProduct(params repos.CreateProductParams) (int64, error) {
	// Clear cache on create
	r.cache.Flush()
	return r.productRepo.CreateProduct(params)
}

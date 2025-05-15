package cache

import (
	// Import errors for sql.ErrNoRows check potentially
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/seanomeara96/gates/models"
	"github.com/seanomeara96/gates/repos"
)

type CachedProductRepo struct {
	cache       *cache.Cache
	productRepo repos.ProductRepository // The underlying non-cached repository
}

// NewCachedProductRepo creates a new caching wrapper around a ProductRepo.
func NewCachedProductRepo(productRepo repos.ProductRepository) *CachedProductRepo {
	if productRepo == nil {
		panic("underlying productRepo cannot be nil for CachedProductRepo")
	}
	defaultExpiration := time.Minute * 5
	cleanupInterval := time.Minute * 10
	c := cache.New(defaultExpiration, cleanupInterval)
	return &CachedProductRepo{
		cache:       c,
		productRepo: productRepo,
	}
}

// --- Method Implementations ---

// InsertProduct clears relevant caches and calls the underlying repository's InsertProduct.
func (r *CachedProductRepo) InsertProduct(product *models.Product) (int, error) {
	// Cache Invalidation: Flush is simple but potentially broad.
	// More granular invalidation could delete specific keys related to 'product.Name', 'product.Type', etc.
	// For now, Flush ensures correctness.
	r.cache.Flush()
	return r.productRepo.InsertProduct(product)
}

// GetProductPrice checks cache first, otherwise fetches from underlying repo and caches the result.
func (r *CachedProductRepo) GetProductPrice(id int) (float32, error) {
	cacheKey := fmt.Sprintf("product_price_%d", id)
	if cachedPrice, found := r.cache.Get(cacheKey); found {
		if price, ok := cachedPrice.(float32); ok {
			return price, nil
		}
		// If type assertion fails, treat as cache miss (or log error)
	}

	price, err := r.productRepo.GetProductPrice(id)
	if err != nil {
		// Don't cache errors like "not found"
		return 0, err
	}

	r.cache.Set(cacheKey, price, cache.DefaultExpiration)
	return price, nil
}

// GetProductByName checks cache first, otherwise fetches from underlying repo and caches the result.
func (r *CachedProductRepo) GetProductByName(name string) (*models.Product, error) {
	cacheKey := fmt.Sprintf("product_by_name_%s", name) // Consider case sensitivity if needed
	if cachedProduct, found := r.cache.Get(cacheKey); found {
		if product, ok := cachedProduct.(*models.Product); ok {
			return product, nil
		}
	}

	product, err := r.productRepo.GetProductByName(name)
	if err != nil {
		// Don't cache errors, especially sql.ErrNoRows
		return nil, err
	}

	// Cache the found product (make sure product is not nil here)
	if product != nil {
		r.cache.Set(cacheKey, product, cache.DefaultExpiration)
	}
	return product, nil
}

// Helper function to generate cache key for product list filters
func generateProductListCacheKey(prefix string, productType models.ProductType, params repos.ProductFilterParams) string {
	// Ensure consistent key format, handling zero values appropriately
	return fmt.Sprintf("%s_%s_maxwidth_%.2f_color_%s_invlvl_%d_price_%.2f_limit_%d",
		prefix,
		productType,
		params.MaxWidth,
		params.Color, // Empty string is handled fine
		params.InventoryLevel,
		params.Price,
		params.Limit, // Limit included for GetProducts, ignored logically by CountProducts but part of params
	)
}

// GetProducts checks cache first based on *all* filter params, otherwise fetches and caches.
func (r *CachedProductRepo) GetProducts(productType models.ProductType, params repos.ProductFilterParams) ([]*models.Product, error) {
	cacheKey := generateProductListCacheKey("products", productType, params)
	if cachedProducts, found := r.cache.Get(cacheKey); found {
		if products, ok := cachedProducts.([]*models.Product); ok {
			return products, nil
		}
	}

	products, err := r.productRepo.GetProducts(productType, params)
	if err != nil {
		return nil, err
	}

	// Cache the result (even if it's an empty slice, that's a valid result)
	r.cache.Set(cacheKey, products, cache.DefaultExpiration)
	return products, nil
}

// CountProducts checks cache first, otherwise counts via underlying repo and caches.
func (r *CachedProductRepo) CountProducts(productType models.ProductType, params repos.ProductFilterParams) (int, error) {
	// Use the same key generation logic but maybe a different prefix
	// Note: Limit in params is ignored by the underlying CountProducts, but included in key for consistency with params struct
	cacheKey := generateProductListCacheKey("count", productType, params) // Use "count" prefix

	if cachedCount, found := r.cache.Get(cacheKey); found {
		if count, ok := cachedCount.(int); ok {
			return count, nil
		}
	}

	count, err := r.productRepo.CountProducts(productType, params)
	if err != nil {
		// Don't cache errors
		return 0, err
	}

	r.cache.Set(cacheKey, count, cache.DefaultExpiration)
	return count, nil
}

// GetCompatibleExtensionsByGateID checks cache first, otherwise fetches and caches.
func (r *CachedProductRepo) GetCompatibleExtensionsByGateID(gateID int) ([]*models.Product, error) {
	cacheKey := fmt.Sprintf("compatible_extensions_%d", gateID)
	if cachedExtensions, found := r.cache.Get(cacheKey); found {
		if extensions, ok := cachedExtensions.([]*models.Product); ok {
			return extensions, nil
		}
	}

	extensions, err := r.productRepo.GetCompatibleExtensionsByGateID(gateID)
	if err != nil {
		return nil, err
	}

	r.cache.Set(cacheKey, extensions, cache.DefaultExpiration)
	return extensions, nil
}

// UpdateProductByID flushes the cache and calls the underlying repository's UpdateProductByID.
func (r *CachedProductRepo) UpdateProductByID(productID int, product *models.Product) error {
	// Cache Invalidation: Flush is simple. Granular would involve deleting keys for:
	// - product_by_id_...
	// - product_by_name_... (if name changed)
	// - potentially relevant list keys (difficult without knowing which lists it affected)
	r.cache.Flush()
	return r.productRepo.UpdateProductByID(productID, product)
}

// DeleteProductByID flushes the cache and calls the underlying repository's DeleteProductByID.
func (r *CachedProductRepo) DeleteProductByID(productID int) error {
	// Cache Invalidation: Flush is simple. Granular would involve deleting keys for:
	// - product_by_id_...
	// - product_by_name_... (need to fetch name before delete or pass it)
	// - potentially relevant list keys
	r.cache.Flush()
	return r.productRepo.DeleteProductByID(productID)
}

// GetProductByID checks cache first, otherwise fetches from underlying repo and caches the result.
func (r *CachedProductRepo) GetProductByID(productID int) (*models.Product, error) {
	cacheKey := fmt.Sprintf("product_by_id_%d", productID)
	if cachedProduct, found := r.cache.Get(cacheKey); found {
		if product, ok := cachedProduct.(*models.Product); ok {
			return product, nil
		}
	}

	product, err := r.productRepo.GetProductByID(productID)
	if err != nil {
		// Don't cache errors like sql.ErrNoRows
		return nil, err
	}

	if product != nil {
		r.cache.Set(cacheKey, product, cache.DefaultExpiration)
	}
	return product, nil
}

// --- Convenience Wrappers (GetGates, GetExtensions, GetBundles) ---
// These now correctly generate cache keys based on the full ProductFilterParams

func (r *CachedProductRepo) GetGates(params repos.ProductFilterParams) ([]*models.Product, error) {
	cacheKey := generateProductListCacheKey("gates", models.ProductTypeGate, params) // Use helper
	if cachedGates, found := r.cache.Get(cacheKey); found {
		if gates, ok := cachedGates.([]*models.Product); ok {
			return gates, nil
		}
	}

	gates, err := r.productRepo.GetGates(params)
	if err != nil {
		return nil, err
	}

	r.cache.Set(cacheKey, gates, cache.DefaultExpiration)
	return gates, nil
}

func (r *CachedProductRepo) GetExtensions(params repos.ProductFilterParams) ([]*models.Product, error) {
	cacheKey := generateProductListCacheKey("extensions", models.ProductTypeExtension, params) // Use helper
	if cachedExtensions, found := r.cache.Get(cacheKey); found {
		if extensions, ok := cachedExtensions.([]*models.Product); ok {
			return extensions, nil
		}
	}

	extensions, err := r.productRepo.GetExtensions(params)
	if err != nil {
		return nil, err
	}

	r.cache.Set(cacheKey, extensions, cache.DefaultExpiration)
	return extensions, nil
}

func (r *CachedProductRepo) GetBundles(params repos.ProductFilterParams) ([]*models.Product, error) {
	cacheKey := generateProductListCacheKey("bundles", models.ProductTypeBundle, params) // Use helper
	if cachedBundles, found := r.cache.Get(cacheKey); found {
		if bundles, ok := cachedBundles.([]*models.Product); ok {
			return bundles, nil
		}
	}

	bundles, err := r.productRepo.GetBundles(params)
	if err != nil {
		return nil, err
	}

	r.cache.Set(cacheKey, bundles, cache.DefaultExpiration)
	return bundles, nil
}

// NOTE: CreateProduct(params repos.CreateProductParams) has been REMOVED
// as it no longer exists in the underlying repos.ProductRepo.
// The service layer should perform validation and call InsertProduct.

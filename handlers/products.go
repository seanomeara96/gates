package handlers

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strconv"
	"strings"

	"github.com/seanomeara96/gates/models"
	"github.com/seanomeara96/gates/repos"
	"github.com/seanomeara96/gates/repos/sqlite"
	"github.com/seanomeara96/gates/views/pages"
	"github.com/seanomeara96/gates/views/partials"
)

func (h *Handler) UpdateProduct(cart models.Cart, w http.ResponseWriter, r *http.Request) error {
	// Verify request method
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		return fmt.Errorf("UpdateProduct: unsupported HTTP method %s (path=%s)", r.Method, r.URL.Path)
	}

	// Parse product ID from path
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return fmt.Errorf("UpdateProduct: invalid product ID format '%s' (path=%s): %w", idStr, r.URL.Path, err)
	}

	// Retrieve existing product
	product, err := h.productCache.GetProductByID(id)
	if err != nil {
		return fmt.Errorf("UpdateProduct: failed to retrieve product (ID: %d, path=%s): %w", id, r.URL.Path, err)
	}

	// Parse form data with size limit to prevent DOS attacks
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("UpdateProduct: failed to parse form data (productID=%d, path=%s): %w", id, r.URL.Path, err)
	}

	// Update string fields with sanitization
	product.Name = html.EscapeString(strings.TrimSpace(r.Form.Get("name")))
	product.Color = html.EscapeString(strings.TrimSpace(r.Form.Get("color")))
	product.Img = html.EscapeString(strings.TrimSpace(r.Form.Get("img")))

	// Update numeric fields with validation
	inventoryLevelStr := r.Form.Get("inventory_level")
	if inventoryLevelStr != "" {
		inventoryLevel, err := strconv.Atoi(inventoryLevelStr)
		if err != nil {
			return fmt.Errorf("UpdateProduct: invalid inventory_level format '%s' (productID=%d, path=%s): %w",
				inventoryLevelStr, id, r.URL.Path, err)
		}
		product.InventoryLevel = inventoryLevel
	}

	priceStr := r.Form.Get("price")
	if priceStr != "" {
		price, err := strconv.ParseFloat(priceStr, 32)
		if err != nil {
			return fmt.Errorf("UpdateProduct: invalid price format '%s' (productID=%d, path=%s): %w",
				priceStr, id, r.URL.Path, err)
		}
		product.Price = float32(price)
	}

	qtyStr := r.Form.Get("qty")
	if qtyStr != "" {
		qty, err := strconv.Atoi(qtyStr)
		if err != nil {
			return fmt.Errorf("UpdateProduct: invalid qty format '%s' (productID=%d, path=%s): %w",
				qtyStr, id, r.URL.Path, err)
		}
		product.Qty = qty
	}

	toleranceStr := r.Form.Get("tolerance")
	if toleranceStr != "" {
		tolerance, err := strconv.ParseFloat(toleranceStr, 32)
		if err != nil {
			return fmt.Errorf("UpdateProduct: invalid tolerance format '%s' (productID=%d, path=%s): %w",
				toleranceStr, id, r.URL.Path, err)
		}
		product.Tolerance = float32(tolerance)
	}

	widthStr := r.Form.Get("width")
	if widthStr != "" {
		width, err := strconv.ParseFloat(widthStr, 32)
		if err != nil {
			return fmt.Errorf("UpdateProduct: invalid width format '%s' (productID=%d, path=%s): %w",
				widthStr, id, r.URL.Path, err)
		}
		product.Width = float32(width)
	}

	// Validate product type
	typeStr := r.Form.Get("type")
	productType := models.ProductType(typeStr)
	validTypes := map[models.ProductType]bool{
		models.ProductTypeGate:      true,
		models.ProductTypeExtension: true,
		models.ProductTypeBundle:    true,
	}

	if validTypes[productType] {
		product.Type = productType
	} else if typeStr != "" {
		return fmt.Errorf("UpdateProduct: invalid product type '%s' (productID=%d, path=%s)", typeStr, id, r.URL.Path)
	}

	// Persist updated product
	if err := h.productCache.UpdateProductByID(id, product); err != nil {
		return fmt.Errorf("UpdateProduct: failed to update product in database (ID: %d, path=%s): %w", id, r.URL.Path, err)
	}

	return nil
}

func (h *Handler) GetGatesPage(cart models.Cart, w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return fmt.Errorf("GetGatesPage: unsupported HTTP method %s (path=%s)", r.Method, r.URL.Path)
	}

	gates, err := h.productCache.GetGates(repos.ProductFilterParams{})
	if err != nil {
		return fmt.Errorf("GetGatesPage: failed to retrieve gates from product cache (path=%s): %w", r.URL.Path, err)
	}

	if h.cfg.UseTempl {
		props := pages.ProductsPageProps{
			BaseProps: pages.BaseProps{
				PageTitle:       "Shop gates at baby safety gates ireland",
				MetaDescription: "Shop our full range of baby safety gates in Ireland, including pressure-fit and screw-fit options to keep your home safe.",
				Cart:            cart,
				Env:             h.cfg.Mode,
			},
			Heading:  "Shop All Gates",
			Products: gates,
		}
		return pages.Products(props).Render(r.Context(), w)
	}

	data := map[string]any{
		"Heading":         "Shop All Gates",
		"PageTitle":       "Shop All Gates",
		"MetaDescription": "Shop our full range of gates",
		"Products":        gates,
		"Cart":            cart,
		"Env":             h.cfg.Mode,
	}

	return h.rndr.Page(w, "products", data)
}

func (h *Handler) GetGatePage(cart models.Cart, w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return fmt.Errorf("GetGatePage: unsupported HTTP method %s (path=%s)", r.Method, r.URL.Path)
	}

	gateIDStr := r.PathValue("gate_id")
	gateID, err := strconv.Atoi(gateIDStr)
	if err != nil {
		return fmt.Errorf("GetGatePage: failed to convert gate_id '%s' to integer (path=%s): %w", gateIDStr, r.URL.Path, err)
	}

	gate, err := h.productCache.GetProductByID(gateID)
	if err != nil {
		return fmt.Errorf("GetGatePage: failed to retrieve gate from product cache (ID: %d, path=%s): %w", gateID, r.URL.Path, err)
	}

	if h.cfg.UseTempl {
		props := pages.ProductPageProps{
			BaseProps: pages.BaseProps{
				PageTitle:       gate.Name,
				MetaDescription: gate.Name,
				Cart:            cart,
				Env:             h.cfg.Mode,
			},
			Product: gate,
		}
		return pages.Product(props).Render(r.Context(), w)
	}

	data := map[string]any{
		"PageTitle":       gate.Name,
		"MetaDescription": gate.Name,
		"Product":         gate,
		"Cart":            cart,
		"Env":             h.cfg.Mode,
	}

	return h.rndr.Page(w, "product", data)
}

func (h *Handler) GetExtensionsPage(cart models.Cart, w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return fmt.Errorf("GetExtensionsPage: unsupported HTTP method %s (path=%s)", r.Method, r.URL.Path)
	}

	extensions, err := h.productCache.GetExtensions(repos.ProductFilterParams{})
	if err != nil {
		return fmt.Errorf("GetExtensionsPage: failed to retrieve extensions from product cache (path=%s): %w", r.URL.Path, err)
	}

	if h.cfg.UseTempl {
		props := pages.ProductsPageProps{
			BaseProps: pages.BaseProps{
				PageTitle:       "Baby Gate Extensions | Extend Your Safety Gate",
				MetaDescription: "Shop baby gate extensions to widen your safety gate for doorways, stairs and wider openings. Find the right extension size for your gate.",
				Cart:            cart,
				Env:             h.cfg.Mode,
			},
			Heading:  "Shop Gate Extensions",
			Products: extensions,
		}
		return pages.Products(props).Render(r.Context(), w)
	}

	data := map[string]any{
		"Heading":         "All extensions",
		"PageTitle":       "All extensions",
		"MetaDescription": "Shop all extensions",
		"Products":        extensions,
		"Cart":            cart,
		"Env":             h.cfg.Mode,
	}

	return h.rndr.Page(w, "products", data)
}

func (h *Handler) GetExtensionPage(cart models.Cart, w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return fmt.Errorf("GetExtensionPage: unsupported HTTP method %s (path=%s)", r.Method, r.URL.Path)
	}

	extensionIDStr := r.PathValue("extension_id")
	extensionID, err := strconv.Atoi(extensionIDStr)
	if err != nil {
		return fmt.Errorf("GetExtensionPage: failed to convert extension_id '%s' to integer (path=%s): %w", extensionIDStr, r.URL.Path, err)
	}

	extension, err := h.productCache.GetProductByID(extensionID)
	if err != nil {
		return fmt.Errorf("GetExtensionPage: failed to retrieve extension from product cache (ID: %d, path=%s): %w", extensionID, r.URL.Path, err)
	}

	if h.cfg.UseTempl {
		props := pages.ProductPageProps{
			BaseProps: pages.BaseProps{
				PageTitle:       extension.Name,
				MetaDescription: extension.Name,
				Cart:            cart,
				Env:             h.cfg.Mode,
			},
			Product: extension,
		}
		return pages.Product(props).Render(r.Context(), w)
	}

	data := map[string]any{
		"PageTitle":       extension.Name,
		"MetaDescription": extension.Name,
		"Product":         extension,
		"Cart":            cart,
		"Env":             h.cfg.Mode,
	}

	return h.rndr.Page(w, "product", data)
}

func (h *Handler) GetCartJSON(cart models.Cart, w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return fmt.Errorf("GetCartJSON: unsupported HTTP method %s (path=%s, cartID=%s)", r.Method, r.URL.Path, cart.ID)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	bytes, err := json.Marshal(cart)
	if err != nil {
		return fmt.Errorf("GetCartJSON: failed to marshal cart (ID: %s, path=%s): %w", cart.ID, r.URL.Path, err)
	}

	if _, err := w.Write(bytes); err != nil {
		return fmt.Errorf("GetCartJSON: failed to write response for cart (ID: %s, path=%s): %w", cart.ID, r.URL.Path, err)
	}
	return nil
}

/*This should move ?? */
func AddItemToCart(cartRepo *sqlite.CartRepo, cartID string, cartItem models.CartItem) error {
	if cartID == "" {
		return fmt.Errorf("AddItemToCart: empty cartID provided (itemID=%s)", cartItem.ID)
	}

	exists, err := cartRepo.DoesCartItemExist(cartID, cartItem.ID)
	if err != nil {
		return fmt.Errorf("AddItemToCart: failed to check if cart item exists (cartID=%s, itemID=%s): %w",
			cartID, cartItem.ID, err)
	}

	if !exists {
		if err := cartRepo.InsertCartItem(cartItem); err != nil {
			return fmt.Errorf("AddItemToCart: failed to insert cart item (cartID=%s, itemID=%s, components=%d): %w",
				cartID, cartItem.ID, len(cartItem.Components), err)
		}
		if err := cartRepo.SaveCartItemComponents(cartItem.Components); err != nil {
			return fmt.Errorf("AddItemToCart: failed to save item components (cartID=%s, itemID=%s, components=%d): %w",
				cartID, cartItem.ID, len(cartItem.Components), err)
		}
	} else {
		if err := cartRepo.IncrementCartItem(cartID, cartItem.ID); err != nil {
			return fmt.Errorf("AddItemToCart: failed to increment cart item (cartID=%s, itemID=%s): %w",
				cartID, cartItem.ID, err)
		}
	}

	if err := cartRepo.SetLastUpdated(cartID); err != nil {
		return fmt.Errorf("AddItemToCart: failed to update last_updated field for cart (ID: %s, itemID=%s): %w",
			cartID, cartItem.ID, err)
	}

	return nil
}

func (h *Handler) AddItemToCart(cart models.Cart, w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return fmt.Errorf("AddItemToCart: unsupported HTTP method %s (path=%s, cartID=%s)", r.Method, r.URL.Path, cart.ID)
	}

	// Limit form size to prevent DOS attacks
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit

	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("AddItemToCart: failed to parse form (cartID=%s, path=%s): %w", cart.ID, r.URL.Path, err)
	}

	formData := r.Form["data"]
	if len(formData) < 1 {
		return h.rndr.Partial(w, "cart-modal", cart)
	}

	components := []models.CartItemComponent{}

	for i, d := range formData {
		// Validate JSON to prevent injection attacks
		if !json.Valid([]byte(d)) {
			return fmt.Errorf("AddItemToCart: invalid JSON in form data (index=%d, cartID=%s, path=%s)", i, cart.ID, r.URL.Path)
		}

		component := models.NewCartItemComponent(cart.ID)
		if err := json.Unmarshal([]byte(d), &component); err != nil {
			return fmt.Errorf("AddItemToCart: failed to unmarshal cart item component (index=%d, cartID=%s, path=%s): %w",
				i, cart.ID, r.URL.Path, err)
		}
		components = append(components, component)
	}

	if err := AddItemToCart(h.cartRepo, cart.ID, models.NewCartItem(cart.ID, components)); err != nil {
		return fmt.Errorf("AddItemToCart: failed to add item to cart (cartID=%s, path=%s): %w", cart.ID, r.URL.Path, err)
	}

	cart, found, err := h.cartRepo.GetCartByID(cart.ID)
	if err != nil || !found {
		return fmt.Errorf("AddItemToCart: failed to retrieve updated cart (cartID=%s, found=%t, path=%s): %w", cart.ID, found, r.URL.Path, err)
	}

	if h.cfg.UseTempl {
		return partials.CartModal(cart).Render(r.Context(), w)
	}

	return h.rndr.Partial(w, "cart-modal", cart)
}

package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/seanomeara96/gates/models"
	"github.com/seanomeara96/gates/repos"
)

func (h *Handler) GetGatesPage(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
	gates, err := h.productCache.GetGates(repos.ProductFilterParams{})
	if err != nil {
		return fmt.Errorf("gates page: failed to get gates: %w", err)
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

func (h *Handler) GetGatePage(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
	gateID, err := strconv.Atoi(r.PathValue("gate_id"))
	if err != nil {
		return fmt.Errorf("gate details: failed to convert gate_id to integer: %w", err)
	}

	gate, err := h.productCache.GetProductByID(gateID)
	if err != nil {
		return fmt.Errorf("gate details: failed to retrieve gate: %w", err)
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

func (h *Handler) GetExtensionsPage(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
	extensions, err := h.productCache.GetExtensions(repos.ProductFilterParams{})
	if err != nil {
		return fmt.Errorf("extensions page: failed to get extensions: %w", err)
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

func (h *Handler) Handler(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
	extensionID, err := strconv.Atoi(r.PathValue("extension_id"))
	if err != nil {
		return fmt.Errorf("extension details: failed to convert extension_id to integer: %w", err)
	}

	extension, err := h.productCache.GetProductByID(extensionID)
	if err != nil {
		return fmt.Errorf("extension details: failed to retrieve extension: %w", err)
	}

	data := map[string]any{
		"PageTitle":       extension.Name,
		"MetaDescription": extension.Name,
		"Cart":            cart,
		"Env":             h.cfg.Mode,
	}

	return h.rndr.Page(w, "products", data)
}

func (h *Handler) GetCartJSON(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
	bytes, err := json.Marshal(cart)
	if err != nil {
		return fmt.Errorf("cart JSON endpoint: failed to marshal cart: %w", err)
	}

	if _, err := w.Write(bytes); err != nil {
		return fmt.Errorf("cart JSON endpoint: failed to write response: %w", err)
	}
	return nil
}

func (h *Handler) AddItemToCart(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("cart add: failed to parse form: %w", err)
	}

	formData := r.Form["data"]
	if len(formData) < 1 {
		return h.rndr.Partial(w, "cart-modal", cart)
	}

	components := []models.CartItemComponent{}

	for _, d := range formData {
		component := models.NewCartItemComponent(cart.ID)
		if err := json.Unmarshal([]byte(d), &component); err != nil {
			return fmt.Errorf("cart add: failed to unmarshal cart item component %s: %w", d, err)
		}
		components = append(components, component)
	}

	if err := AddItemToCart(h.cartRepo, cart.ID, models.NewCartItem(cart.ID, components)); err != nil {
		return fmt.Errorf("cart add: failed to add item to cart: %w", err)
	}

	cart, err := h.cartRepo.GetCartByID(cart.ID)
	if err != nil {
		return fmt.Errorf("cart add: failed to retrieve updated cart: %w", err)
	}

	return h.rndr.Partial(w, "cart-modal", cart)
}

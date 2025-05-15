package handlers

import (
	"fmt"
	"net/http"

	"github.com/seanomeara96/gates/models"
)

func (h *Handler) GetCartPage(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
	data := map[string]any{
		"PageTitle":       "Your shopping cart",
		"MetaDescription": "",
		"Cart":            cart,
		"Env":             h.cfg.Mode,
	}

	return h.rndr.Page(w, "cart", data)
}
func (h *Handler) AdjustCartItemQty(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {

	mode := r.PathValue("mode")

	if mode != "increment" && mode != "decrement" {
		return nil
	}

	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("cart item update: failed to parse form: %w", err)
	}

	cartItemID := r.Form.Get("cart_item_id")
	if cartItemID == "" {
		return fmt.Errorf("cart item update: cart_item_id is blank")
	}

	cartItem, err := h.cartRepo.SelectCartItem(cart.ID, cartItemID)
	if err != nil {
		return fmt.Errorf("cart item update: failed to select cart item: %w", err)
	}

	if mode == "increment" {
		if err := h.cartRepo.IncrementCartItem(cart.ID, cartItem.ID); err != nil {
			return fmt.Errorf("cart item update: failed to increment cart item: %w", err)
		}
	} else {
		if cartItem.Qty < 2 {
			w.WriteHeader(http.StatusBadRequest)
			return nil
		}
		if err := h.cartRepo.DecrementCartItem(cart.ID, cartItem.ID); err != nil {
			return fmt.Errorf("cart item update: failed to decrement cart item: %w", err)
		}
	}

	cart, err = h.cartRepo.GetCartByID(cart.ID)
	if err != nil {
		return fmt.Errorf("cart item update: failed to retrieve updated cart: %w", err)
	}
	if err := h.rndr.Partial(w, "cart-main", cart); err != nil {
		return fmt.Errorf("cart item update: failed to render partial (cart-main): %w", err)
	}
	if err := h.rndr.Partial(w, "cart-modal-oob", cart); err != nil {
		return fmt.Errorf("cart item update: failed to render partial (cart-modal-oob): %w", err)
	}

	return nil
}

func (h *Handler) RemoveItemFromCart(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("cart item delete: failed to parse form: %w", err)
	}

	cartItemID := r.Form.Get("id")
	if cartItemID == "" {
		return fmt.Errorf("cart item delete: no cart item id supplied")
	}

	if _, err := h.db.Exec("DELETE FROM cart_item WHERE id = ? AND cart_id = ?", cartItemID, cart.ID); err != nil {
		return fmt.Errorf("cart item delete: failed to delete cart item: %w", err)
	}

	cart, err := h.cartRepo.GetCartByID(cart.ID)
	if err != nil {
		return fmt.Errorf("cart item delete: failed to retrieve updated cart: %w", err)
	}

	if err := h.rndr.Partial(w, "cart-main", cart); err != nil {
		return fmt.Errorf("cart item delete: failed to render partial (cart-main): %w", err)
	}

	if err := h.rndr.Partial(w, "cart-modal-oob", cart); err != nil {
		return fmt.Errorf("cart item delete: failed to render partial (cart-modal-oob): %w", err)
	}

	return nil
}

func (h *Handler) RemoveItemFromCart(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("cart item remove: failed to parse form: %w", err)
	}

	cartItemID := r.Form.Get("cart_item_id")
	if cartItemID == "" {
		return nil
	}

	if _, err := h.db.Exec(`DELETE FROM cart_item WHERE id = ? AND cart_id = ?`, cartItemID, cart.ID); err != nil {
		return fmt.Errorf("cart item remove: failed to delete cart item: %w", err)
	}

	cart, err := h.cartRepo.GetCartByID(cart.ID)
	if err != nil {
		return fmt.Errorf("cart item remove: failed to retrieve updated cart: %w", err)
	}

	if err := h.rndr.Partial(w, "cart-main", cart); err != nil {
		return fmt.Errorf("cart item remove: failed to render partial (cart-main): %w", err)
	}

	return nil
}

func (h *Handler) ClearItemsFromCart(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
	if _, err := h.db.Exec(`DELETE FROM cart_item WHERE cart_id = ?`, cart.ID); err != nil {
		return fmt.Errorf("cart clear: could not delete cart_item for cart_id %s: %w", cart.ID, err)
	}

	if _, err := h.db.Exec("DELETE FROM cart_item_component WHERE cart_id = ?", cart.ID); err != nil {
		return fmt.Errorf("cart clear: could not delete cart_item_component for cart_id %s: %w", cart.ID, err)
	}

	cart, err := h.cartRepo.GetCartByID(cart.ID)
	if err != nil {
		return fmt.Errorf("cart clear: failed to retrieve updated cart: %w", err)
	}

	if err := renderPartial(w, "cart-main", cart); err != nil {
		return fmt.Errorf("cart clear: failed to render partial (cart-main): %w", err)
	}

	if err := renderPartial(w, "cart-modal-oob", cart); err != nil {
		return fmt.Errorf("cart clear: failed to render partial (cart-modal-oob): %w", err)
	}

	return nil
}

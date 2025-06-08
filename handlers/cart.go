package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
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
		return fmt.Errorf("cart item delete: failed to render partial (cart-main): %w", err)
	}

	if err := h.rndr.Partial(w, "cart-modal-oob", cart); err != nil {
		return fmt.Errorf("cart item delete: failed to render partial (cart-modal-oob): %w", err)
	}

	return nil
}

func (h *Handler) ClearItemsFromCart(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {

	tx, err := h.db.Begin()
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return rbErr
		}
		return err
	}

	if _, err := tx.Exec(`DELETE FROM cart_item WHERE cart_id = ?`, cart.ID); err != nil {
		return fmt.Errorf("cart clear: could not delete cart_item for cart_id %s: %w", cart.ID, err)
	}

	if _, err := tx.Exec("DELETE FROM cart_item_component WHERE cart_id = ?", cart.ID); err != nil {
		return fmt.Errorf("cart clear: could not delete cart_item_component for cart_id %s: %w", cart.ID, err)
	}

	if err := tx.Commit(); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return rbErr
		}
		return err
	}

	cart, err = h.cartRepo.GetCartByID(cart.ID)
	if err != nil {
		return fmt.Errorf("cart clear: failed to retrieve updated cart: %w", err)
	}

	if err := h.rndr.Partial(w, "cart-main", cart); err != nil {
		return fmt.Errorf("cart clear: failed to render partial (cart-main): %w", err)
	}

	if err := h.rndr.Partial(w, "cart-modal-oob", cart); err != nil {
		return fmt.Errorf("cart clear: failed to render partial (cart-modal-oob): %w", err)
	}

	return nil
}

func (h *Handler) newCart() (*models.Cart, error) {
	cart := models.NewCart()
	if _, err := h.cartRepo.SaveCart(cart); err != nil {
		return nil, err
	}
	return &cart, nil
}

/*returns a new session if the session does not exist*/
func getCartSession(r *http.Request, store *sessions.CookieStore) (*sessions.Session, error) {
	session, err := store.Get(r, "cart-session")
	if err != nil {
		return nil, err
	}
	return session, nil
}

func getCartID(session *sessions.Session) (string, bool, error) {
	if session == nil {
		return "", false, errors.New("cart Session is nil")
	}
	cartID, found := session.Values["cart_id"]
	if !found {
		return "", false, nil
	}

	cartIDString, ok := cartID.(string)
	if !ok {
		return "", false, errors.New("could not convert cartID interface to string")
	}

	return cartIDString, found, nil
}

func attachNewCartToSession(cart *models.Cart, session *sessions.Session, w http.ResponseWriter, r *http.Request) error {

	session.Values["cart_id"] = cart.ID
	if err := session.Save(r, w); err != nil {
		return err
	}
	return nil
}

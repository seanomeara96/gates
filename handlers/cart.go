package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/seanomeara96/gates/models"
	"github.com/seanomeara96/gates/views/pages"
)

func (h *Handler) GetCartPage(cart models.Cart, w http.ResponseWriter, r *http.Request) error {

	if h.cfg.UseTempl {
		props := pages.CartPageProps{
			BaseProps: pages.BaseProps{
				PageTitle: "Cart Page",
				Env:       h.cfg.Mode,
				Cart:      cart,
			},
			Cart: cart,
		}
		return pages.Cart(props).Render(r.Context(), w)
	}

	data := map[string]any{
		"PageTitle":       "Your shopping cart",
		"MetaDescription": "",
		"Cart":            cart,
		"Env":             h.cfg.Mode,
	}

	if err := h.rndr.Page(w, "cart", data); err != nil {
		return fmt.Errorf("get cart page: render page (cart): %w", err)
	}
	return nil
}
func (h *Handler) AdjustCartItemQty(cart models.Cart, w http.ResponseWriter, r *http.Request) error {

	mode := r.PathValue("mode")

	if mode != "increment" && mode != "decrement" {
		return nil
	}

	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("cart item update: parse form: %w", err)
	}

	cartItemID := r.Form.Get("cart_item_id")
	if cartItemID == "" {
		return fmt.Errorf("cart item update: cart_item_id is blank (cart_id=%s, mode=%s)", cart.ID, mode)
	}

	cartItem, err := h.cartRepo.SelectCartItem(cart.ID, cartItemID)
	if err != nil {
		return fmt.Errorf("cart item update: select cart item (cart_id=%s, cart_item_id=%s): %w", cart.ID, cartItemID, err)
	}

	if mode == "increment" {
		if err := h.cartRepo.IncrementCartItem(cart.ID, cartItem.ID); err != nil {
			return fmt.Errorf("cart item update: increment cart item (cart_id=%s, cart_item_id=%s): %w", cart.ID, cartItem.ID, err)
		}
	} else {
		if cartItem.Qty < 2 {
			w.WriteHeader(http.StatusBadRequest)
			return nil
		}
		if err := h.cartRepo.DecrementCartItem(cart.ID, cartItem.ID); err != nil {
			return fmt.Errorf("cart item update: decrement cart item (cart_id=%s, cart_item_id=%s): %w", cart.ID, cartItem.ID, err)
		}
	}

	cart, found, err := h.cartRepo.GetCartByID(cart.ID)
	if err != nil {
		return fmt.Errorf("cart item update: retrieve updated cart (cart_id=%s): %w", cart.ID, err)
	}
	if !found {
		return fmt.Errorf("cart item update: retrieve updated cart (cart_id=%s): not found", cart.ID)
	}
	if err := h.rndr.Partial(w, "cart-main", cart); err != nil {
		return fmt.Errorf("cart item update: render partial (cart-main) (cart_id=%s): %w", cart.ID, err)
	}
	if err := h.rndr.Partial(w, "cart-modal-oob", cart); err != nil {
		return fmt.Errorf("cart item update: render partial (cart-modal-oob) (cart_id=%s): %w", cart.ID, err)
	}

	return nil
}

func (h *Handler) RemoveItemFromCart(cart models.Cart, w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("cart item remove: parse form: %w", err)
	}

	cartItemID := r.Form.Get("cart_item_id")
	if cartItemID == "" {
		return nil
	}

	if _, err := h.db.Exec(`DELETE FROM cart_item WHERE id = ? AND cart_id = ?`, cartItemID, cart.ID); err != nil {
		return fmt.Errorf("cart item remove: delete cart item (cart_id=%s, cart_item_id=%s): %w", cart.ID, cartItemID, err)
	}

	cart, found, err := h.cartRepo.GetCartByID(cart.ID)
	if err != nil {
		return fmt.Errorf("cart item remove: retrieve updated cart (cart_id=%s): %w", cart.ID, err)
	}
	if !found {
		return fmt.Errorf("cart item remove: retrieve updated cart (cart_id=%s): not found", cart.ID)
	}

	if err := h.rndr.Partial(w, "cart-main", cart); err != nil {
		return fmt.Errorf("cart item delete: render partial (cart-main) (cart_id=%s): %w", cart.ID, err)
	}

	if err := h.rndr.Partial(w, "cart-modal-oob", cart); err != nil {
		return fmt.Errorf("cart item delete: render partial (cart-modal-oob) (cart_id=%s): %w", cart.ID, err)
	}

	return nil
}

func (h *Handler) ClearItemsFromCart(cart models.Cart, w http.ResponseWriter, r *http.Request) error {

	tx, err := h.db.Begin()
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("cart clear: rollback after begin failure (cart_id=%s): %w", cart.ID, rbErr)
		}
		return fmt.Errorf("cart clear: begin transaction (cart_id=%s): %w", cart.ID, err)
	}

	if _, err := tx.Exec(`DELETE FROM cart_item WHERE cart_id = ?`, cart.ID); err != nil {
		return fmt.Errorf("cart clear: delete cart_item (cart_id=%s): %w", cart.ID, err)
	}

	if _, err := tx.Exec("DELETE FROM cart_item_component WHERE cart_id = ?", cart.ID); err != nil {
		return fmt.Errorf("cart clear: delete cart_item_component (cart_id=%s): %w", cart.ID, err)
	}

	if err := tx.Commit(); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("cart clear: rollback after commit failure (cart_id=%s): %w", cart.ID, rbErr)
		}
		return fmt.Errorf("cart clear: commit transaction (cart_id=%s): %w", cart.ID, err)
	}

	cart, found, err := h.cartRepo.GetCartByID(cart.ID)
	if err != nil {
		return fmt.Errorf("cart clear: retrieve updated cart (cart_id=%s): %w", cart.ID, err)
	}
	if !found {
		return fmt.Errorf("cart clear: retrieve updated cart (cart_id=%s): not found", cart.ID)
	}

	if err := h.rndr.Partial(w, "cart-main", cart); err != nil {
		return fmt.Errorf("cart clear: render partial (cart-main) (cart_id=%s): %w", cart.ID, err)
	}

	if err := h.rndr.Partial(w, "cart-modal-oob", cart); err != nil {
		return fmt.Errorf("cart clear: render partial (cart-modal-oob) (cart_id=%s): %w", cart.ID, err)
	}

	return nil
}

func (h *Handler) newCart() (models.Cart, error) {
	cart := models.NewCart()
	if _, err := h.cartRepo.SaveCart(cart); err != nil {
		return models.Cart{}, fmt.Errorf("new cart: save cart (cart_id=%s): %w", cart.ID, err)
	}
	return cart, nil
}

/*returns a new session if the session does not exist*/
func getCartSession(r *http.Request, store *sessions.CookieStore) (*sessions.Session, error) {
	session, err := store.Get(r, "cart-session")
	if err != nil {
		return nil, fmt.Errorf("get cart session: store get (name=%q): %w", "cart-session", err)
	}
	return session, nil
}

func getCartID(session *sessions.Session) (string, bool, error) {
	if session == nil {
		return "", false, errors.New("get cart id: session is nil")
	}
	cartID, found := session.Values["cart_id"]
	if !found {
		return "", false, nil
	}

	cartIDString, ok := cartID.(string)
	if !ok {
		return "", false, fmt.Errorf("get cart id: could not convert cart_id to string (type=%T)", cartID)
	}

	return cartIDString, found, nil
}

func attachNewCartToSession(cart models.Cart, session *sessions.Session, w http.ResponseWriter, r *http.Request) error {

	session.Values["cart_id"] = cart.ID
	if err := session.Save(r, w); err != nil {
		return fmt.Errorf("attach new cart to session: save session (cart_id=%s): %w", cart.ID, err)
	}
	return nil
}

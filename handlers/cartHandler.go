package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/seanomeara96/gates/models"
	"github.com/seanomeara96/gates/render"
	"github.com/seanomeara96/gates/services"
)

type CartHandler struct {
	store       *sessions.CookieStore
	cartService *services.CartService
	renderer    *render.Renderer
}

func NewCartHandler(cartService *services.CartService, renderer *render.Renderer, store *sessions.CookieStore) *CartHandler {
	return &CartHandler{
		cartService: cartService,
		renderer:    renderer,
		store:       store,
	}
}

func (h *CartHandler) getSession(r *http.Request) (*sessions.Session, error) {
	session, err := h.store.Get(r, "cart-session")
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (h *CartHandler) getCartID(session *sessions.Session) (interface{}, error) {
	if session == nil {
		return nil, errors.New("Cart Session is nil")
	}
	return session.Values["cart_id"], nil

}

func (h *CartHandler) MiddleWare(w http.ResponseWriter, r *http.Request) (bool, error) {
	session, err := h.getSession(r)
	if err != nil {
		return false, err
	}

	cartID, err := h.getCartID(session)
	if err != nil {
		return false, err
	}

	if cartID != nil {
		return true, nil
	}

	cartID, err = h.cartService.NewCart()
	if err != nil {
		return false, err
	}

	session.Values["cart_id"] = cartID
	if err := session.Save(r, w); err != nil {
		return false, err
	}

	return true, nil
}

func validateCartID(cartID interface{}) (valid bool) {
	if _, ok := cartID.(string); !ok {
		return false
	}
	return true
}

func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) error {
	session, err := h.getSession(r)
	if err != nil {
		return err
	}

	cartID, err := h.getCartID(session)
	if err != nil {
		return err
	}

	if ok := validateCartID(cartID); !ok {
		return errors.New("Invalid cart id")
	}

	r.ParseForm()

	components := []models.CartItemComponent{}

	for _, d := range r.Form["data"] {
		var component models.CartItemComponent
		if err := json.Unmarshal([]byte(d), &component); err != nil {
			return err
		}
		components = append(components, component)
	}

	if err := h.cartService.AddItem(cartID.(string), components); err != nil {
		return err
	}

	return nil
}

func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) error {
	session, err := h.getSession(r)
	if err != nil {
		return err
	}

	cartID, err := h.getCartID(session)
	if err != nil {
		return err
	}

	if ok := validateCartID(cartID); !ok {
		return errors.New("Invalid cart id")
	}

	r.ParseForm()

	itemID := r.Form.Get("item_id")

	if err := h.cartService.RemoveItem(cartID.(string), itemID); err != nil {
		return err
	}

	return nil
}

func (h *CartHandler) View(w http.ResponseWriter, r *http.Request) error {
	return errors.New("Not implemented.")
}

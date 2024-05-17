package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/sessions"
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
func (h *CartHandler) Update(w http.ResponseWriter, r *http.Request) error {
	session, err := h.getSession(r)
	if err != nil {
		return err
	}

	cartID, err := h.getCartID(session)
	if err != nil {
		return err
	}

	r.ParseForm()

	type Item struct {
		ID  int `json:"id"`
		Qty int `json:"qty"`
	}

	items := []Item{}

	for _, d := range r.Form["data"] {
		var item Item
		if err := json.Unmarshal([]byte(d), &item); err != nil {
			return err
		}
		items = append(items, item)
	}

	if err := h.cartService.AddItems(items); err != nil {
		return err
	}

	return nil
}

func (h *CartHandler) View(w http.ResponseWriter, r *http.Request) error {
	return errors.New("Not implemented.")
}

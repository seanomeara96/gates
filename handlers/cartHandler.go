package handlers

import (
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

func (h *CartHandler) MiddleWare(w http.ResponseWriter, r *http.Request, fn func(w http.ResponseWriter, r *http.Request) error) error {
	session, err := h.store.Get(r, "cart-session")
	if err != nil {
		return err
	}

	if cartID := session.Values["cart_id"]; cartID != nil {
		return fn(w, r)
	}

	cartID, err := h.cartService.NewCart()
	if err != nil {
		return err
	}

	session.Values["cart_id"] = cartID
	if err := session.Save(r, w); err != nil {
		return err
	}

	return fn(w, r)
}
func (h *CartHandler) Update(w http.ResponseWriter, r *http.Request) error {
	return errors.New("Not implemented.")
}
func (h *CartHandler) View(w http.ResponseWriter, r *http.Request) error {
	return errors.New("Not implemented.")
}

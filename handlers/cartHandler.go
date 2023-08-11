package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/seanomeara96/gates/render"
	"github.com/seanomeara96/gates/services"
)

type CartHandler struct {
	cartService *services.CartService
	renderer    *render.Renderer
}

func NewCartHandler(cartService *services.CartService, renderer *render.Renderer) *CartHandler {
	return &CartHandler{
		cartService: cartService,
		renderer:    renderer,
	}
}

func (h *CartHandler) New(w http.ResponseWriter, r *http.Request) {
	cart, err := h.cartService.NewCart(1)
	if err != nil {
		http.Error(w, "Could not create new cart", http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(cart)
	if err != nil {
		http.Error(w, "Could not encode json", http.StatusInternalServerError)
		return
	}
}
func (h *CartHandler) Add(w http.ResponseWriter, r *http.Request) {

}
func (h *CartHandler) Update(w http.ResponseWriter, r *http.Request) {}
func (h *CartHandler) View(w http.ResponseWriter, r *http.Request)   {}

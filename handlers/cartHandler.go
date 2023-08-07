package handlers

import (
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

func (h *CartHandler) New()    {}
func (h *CartHandler) Add()    {}
func (h *CartHandler) Remove() {}
func (h *CartHandler) View()   {}

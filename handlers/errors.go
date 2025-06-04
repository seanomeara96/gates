package handlers

import (
	"net/http"

	"github.com/seanomeara96/gates/models"
)

func (h *Handler) InternalError(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
	return h.rndr.Page(w, "internal-error", map[string]any{})
}

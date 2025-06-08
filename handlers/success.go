package handlers

import (
	"net/http"
	"strconv"

	"github.com/seanomeara96/gates/models"
)

func (h *Handler) GetSuccessPage(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
	orderID, err := strconv.Atoi(r.URL.Query().Get("order_id"))
	if err != nil {
		return err
	}
	return h.rndr.Page(w, "success", map[string]any{
		"OrderID": orderID,
	})
}

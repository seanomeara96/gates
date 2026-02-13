package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/seanomeara96/gates/models"
	"github.com/seanomeara96/gates/views/pages"
)

func (h *Handler) GetSuccessPage(cart models.Cart, w http.ResponseWriter, r *http.Request) error {
	orderIDStr := r.URL.Query().Get("order_id")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		return fmt.Errorf("invalid order_id query parameter %q: %w", orderIDStr, err)
	}
	if h.cfg.UseTempl {
		props := pages.OrderSuccessPageProps{
			BaseProps: pages.BaseProps{
				PageTitle:       fmt.Sprintf("Order #%d Confirmed | Thank You", orderID),
				MetaDescription: fmt.Sprintf("Thank you for your order. Your order #%d has been confirmed and is being processed.", orderID),
				Cart:            cart,
				Env:             h.cfg.Mode,
			},
		}
		return pages.OrderSuccess(props).Render(r.Context(), w)
	}
	return h.rndr.Page(w, "success", map[string]any{
		"OrderID": orderID,
	})
}

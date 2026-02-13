package handlers

import (
	"fmt"
	"net/http"

	"github.com/seanomeara96/gates/models"
	"github.com/seanomeara96/gates/views/pages"
)

func (h *Handler) InternalError(cart models.Cart, w http.ResponseWriter, r *http.Request) error {
	if h.cfg.UseTempl {
		props := pages.InternalErrorPageProps{
			BaseProps: pages.BaseProps{PageTitle: "Oops Something Went Wrong"},
		}
		return pages.InternalError(props).Render(r.Context(), w)
	}
	if err := h.rndr.Page(w, "internal-error", map[string]any{}); err != nil {
		return fmt.Errorf("render internal-error page: %w", err)
	}
	return nil
}

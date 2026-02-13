package handlers

import (
	"fmt"
	"net/http"

	"github.com/seanomeara96/gates/models"
)

func (h *Handler) Test(cart models.Cart, w http.ResponseWriter, r *http.Request) error {
	if r.Method == http.MethodGet {
		return h.rndr.Page(w, "test", map[string]any{})
	}
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			return fmt.Errorf("handlers.Test: failed to parse form (method=%s, path=%s): %w", r.Method, r.URL.Path, err)
		}
		fmt.Printf("%+v\n", r.Form["data"])
		for k, v := range r.Form["data"] {
			fmt.Printf("k:%d v:%s\n", k, v)
		}
		return nil
	}
	return nil
}

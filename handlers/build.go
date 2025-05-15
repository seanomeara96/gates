package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/seanomeara96/gates/models"
)

func (h *Handler) BuildBundle(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("build endpoint: failed to parse form: %w", err)
	}

	desiredWidth, err := strconv.ParseFloat(r.Form.Get("desired-width"), 32)
	if err != nil {
		return fmt.Errorf("build endpoint: failed to parse desired width: %w", err)
	}

	/*
		static max witdth for now
		could consider adding a max width to the gates in the DB and enforcing the
		constraint from there
	*/
	maxWidth := 220.00
	if desiredWidth > maxWidth {
		desiredWidth = maxWidth
	}

	if err := SaveRequestedBundleSize(db, float32(desiredWidth)); err != nil {
		return fmt.Errorf("build endpoint: failed to save requested bundle size: %w", err)
	}

	bundles, err := BuildPressureFitBundles(productCache, float32(desiredWidth))
	if err != nil {
		return fmt.Errorf("build endpoint: failed to build pressure fit bundles: %w", err)
	}

	data := map[string]any{
		"RequestedBundleSize": float32(desiredWidth),
		"Bundles":             bundles,
		"Env":                 cfg.Mode,
	}

	return renderPage(w, "build-results", data)
}

package handlers

import (
	"net/http"
	"strconv"

	"github.com/seanomeara96/gates/render"
	"github.com/seanomeara96/gates/services"
)

type BuildHandler struct {
	bundleService *services.BundleService
	render        *render.Renderer
}

func NewBuildHandler(bundleService *services.BundleService, renderer *render.Renderer) *BuildHandler {
	return &BuildHandler{
		bundleService: bundleService,
		render:        renderer,
	}
}

func (h *BuildHandler) Build(w http.ResponseWriter, r *http.Request) error {
	r.ParseForm()

	_desiredWidth := r.Form["desired-width"][0]

	desiredWidth, err := strconv.ParseFloat(_desiredWidth, 32)
	if err != nil {
		return err
	}

	// TODO keep track of user inputs(valid ones). From there we can generate "popular bundles"
	if err := h.bundleService.SaveRequestedBundleSize(float32(desiredWidth)); err != nil {
		return err
	}

	bundles, err := h.bundleService.BuildPressureFitBundles(float32(desiredWidth))
	if err != nil {
		return err
	}

	templateData := h.render.NewBundleBuildResultsData(
		float32(desiredWidth),
		bundles,
	)

	if err = h.render.BundleBuildResults(w, templateData); err != nil {
		return err
	}

	return nil
}

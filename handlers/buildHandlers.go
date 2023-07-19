package handlers

import (
	"fmt"
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

func (h *BuildHandler) Build(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()
		value := r.Form["desired-width"][0]
		desiredWidth, err := strconv.ParseFloat(value, 32)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// not critical to rest of function
		go func() {
			// TODO keep track of user inputs(valid ones). From there we can generate "popular bundles"
			err := h.bundleService.SaveRequestedBundleSize(float32(desiredWidth))
			if err != nil {
				fmt.Println(err)
			}
		}()

		bundles, err := h.bundleService.BuildPressureFitBundles(float32(desiredWidth))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// w.Header().Set("Content-Type", "application/json")
		templateData := render.BundleBuildResultsData{
			RequestedBundleSize: float32(desiredWidth),
			Bundles:             bundles,
		}
		err = h.render.BundleBuildResults(w, templateData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}
	http.Error(w, "Invalid Request", http.StatusBadRequest)
}

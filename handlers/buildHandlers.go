package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/seanomeara96/gates/services"
)

type BuildHandler struct {
	bundleService *services.BundleService
}

func NewBuildHandler(bundleService *services.BundleService) *BuildHandler {
	return &BuildHandler{
		bundleService: bundleService,
	}
}

func (h *BuildHandler) Build(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var data struct {
			DesiredWidth float32 `json:"width"`
		}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// not critical to rest of function
		go func() {
			// TODO keep track of user inputs(valid ones). From there we can generate "popular bundles"
			err := h.bundleService.SaveRequestedBundleSize(data.DesiredWidth)
			if err != nil {
				fmt.Println(err)
			}
		}()

		bundles, err := h.bundleService.BuildPressureFitBundles(data.DesiredWidth)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(bundles)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}
	http.Error(w, "Invalid Request", http.StatusBadRequest)
}

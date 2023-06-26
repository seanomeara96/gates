package handlers

/*
import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/seanomeara96/gates/models"
)

type BuildHandler struct {
	productService *services.ProductService
	tmpl           *template.Template
}

func NewBuildHandler(productService *services.ProductServices, tmpl *template.Template) *BuildHandler {
	return &BuildHandler{
		productService: productService,
		tmpl:           tmpl,
	}
}

func (h *BuildHandler) BundleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {

		// if theres a query for a specific custom bundle
		query := r.URL.Query()
		if len(query) > 0 {
			q := query.Get("gate")
			e := query.Get("extensions")

			type ItemQuantities struct {
				Id  int `json:"id"`
				Qty int `json:"qty"`
			}
			var gateQuantity ItemQuantities

			err := json.Unmarshal([]byte(q), &gateQuantity)
			if err != nil {
				InternalStatusError("error decoding gate data", err, w, h.tmpl)
				return
			}

			var extensionQuantities []ItemQuantities
			err = json.Unmarshal([]byte(e), &extensionQuantities)
			if err != nil {
				InternalStatusError("error decoding extensions", err, w, h.tmpl)
				return
			}

			var bundle models.Bundle

			var gate models.Product
			gate, err := h.productService.GetProductByID(gateQuantity.Id)
			if err != nil {
				InternalStatusError("error fetching gate from db for route /bundles/", err, w, h.tmpl)
				return
			}
			gate.Qty = gateQuantity.Qty

			bundle.Gates = append(bundle.Gates, gate)

			var extensions []models.Product
			for _, extensionQuantity := range extensionQuantities {
				extension, err := h.productService.GetProductByID(extensionQuantity.Id)
				if err != nil {
					InternalStatusError("error fetching extension from db route /build/", err, w)
					return
				}

				extension.Qty = extensionQuantity.Qty
				extensions = append(extensions, extension)
			}
			bundle.Extensions = extensions
			// add bundle meta data
			bundle.ComputeMetaData()

			type SingleBundlePageData struct {
				BasePageData
				Bundle Bundle
			}

			pageData := SingleBundlePageData{
				BasePageData: BasePageData{
					PageTitle:       "Single Bundle: " + bundle.Name,
					MetaDescription: "Buy Bundle " + bundle.Name + " Online and enjoy super fast delivery",
				},
				Bundle: bundle,
			}

			err = tmpl.ExecuteTemplate(w, "single-bundle.tmpl", pageData)
			if err != nil {
				InternalStatusError("error creating bundle page", err, w, h.tmpl)
				return
			}
			return
		}

		/*if r.URL.Path == "/bundles/" {
			popularBundles, err := fetchPopularBundles()
			if err != nil {
				internalStatusError("error fetching popular bundles for route /bundles/", err, w)
				return
			}
			pageData := struct {
				PopularBundles CachedBundles
			}{
				PopularBundles: popularBundles,
			}

			err = tmpl.ExecuteTemplate(w, "bundles.tmpl", pageData)
			if err != nil {
				internalStatusError("error executing bundles template", err, w)
				return
			}

			return
		}*/ /*
	}
	NotFound(w, h.tmpl)
}*/

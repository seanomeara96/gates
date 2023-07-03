package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/seanomeara96/gates/models"
	"github.com/seanomeara96/gates/services"
)

type User struct {
	Email string
}

type BasePageData struct {
	PageTitle       string
	MetaDescription string
	User            User
}

type HomePageData struct {
	FeaturedGates  []*models.Product
	PopularBundles []*models.Product
	BasePageData
}

func InValidRequest(w http.ResponseWriter) {
	//templateErr := tmpl.ExecuteTemplate(w, "inavlidRequest.tmpl", nil)
	//if templateErr != nil {
	http.Error(w, "Invalid Request Method", http.StatusMethodNotAllowed)
	//}
}

func InternalStatusError(description string, err error, w http.ResponseWriter, tmpl *template.Template) {
	fmt.Println(description)
	fmt.Println(err)
	templateErr := tmpl.ExecuteTemplate(w, "notFound.tmpl", nil)
	if templateErr != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func NotFound(w http.ResponseWriter, tmpl *template.Template) {
	w.WriteHeader(http.StatusNotFound)
	err := tmpl.ExecuteTemplate(w, "notFound.tmpl", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type PageHandler struct {
	productService *services.ProductService
	tmpl           *template.Template
}

func NewPageHandler(productService *services.ProductService, templates *template.Template) *PageHandler {
	return &PageHandler{
		productService: productService,
		tmpl:           templates,
	}
}
func (h *PageHandler) Home(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet && r.URL.Path == "/" {
		featuredGates, err := h.productService.GetGates(services.ProductFilterParams{})
		if err != nil {
			InternalStatusError("could not fetch gates from db", err, w, h.tmpl)
			return
		}

		popularBundles, err := h.productService.GetBundles(services.ProductFilterParams{Limit: 3})
		if err != nil {
			InternalStatusError("could not fetch bundles from db", err, w, h.tmpl)
			return
		}

		pageData := HomePageData{
			BasePageData: BasePageData{
				PageTitle:       "Build your own safety gate",
				MetaDescription: "This is a place to build the perfect safety gate for your home",
				User: User{
					"sean@example.com",
				},
			},
			FeaturedGates:  featuredGates,
			PopularBundles: popularBundles,
		}

		err = h.tmpl.ExecuteTemplate(w, "index.tmpl", pageData)
		if err != nil {
			InternalStatusError("could not execute templete fo homepage", err, w, h.tmpl)
		}
		return
	}

	NotFound(w, h.tmpl)
}

func (h *PageHandler) Gates(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet && r.URL.Path == "/gates/" {
		gates, err := h.productService.GetGates(services.ProductFilterParams{})
		if err != nil {
			InternalStatusError("error fetching gates for route /gates/", err, w, h.tmpl)
			return
		}

		pageData := struct {
			Heading  string
			Products []*models.Product
		}{
			Heading:  "Shop Gates",
			Products: gates,
		}

		h.tmpl.ExecuteTemplate(w, "products.tmpl", pageData)
		return
	}
	NotFound(w, h.tmpl)
}

func (h *PageHandler) Extensions(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet && r.URL.Path == "/extensions/" {

		extensions, err := h.productService.GetExtensions(services.ProductFilterParams{})
		if err != nil {
			InternalStatusError("error fetching extensions for route /extensions/", err, w, h.tmpl)
			return
		}

		pageData := struct {
			Heading  string
			Products []*models.Product
		}{
			Heading:  "Shop Extensions",
			Products: extensions,
		}

		h.tmpl.ExecuteTemplate(w, "products.tmpl", pageData)
		return
	}
	NotFound(w, h.tmpl)
}

func (h *PageHandler) Bundles(w http.ResponseWriter, r *http.Request) {
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

			gate, err := h.productService.GetProductByID(gateQuantity.Id)
			if err != nil {
				InternalStatusError("error fetching gate from db for route /bundles/", err, w, h.tmpl)
				return
			}
			gate.Qty = gateQuantity.Qty

			bundle.Gates = append(bundle.Gates, *gate)

			var extensions []models.Product
			for _, extensionQuantity := range extensionQuantities {
				extension, err := h.productService.GetProductByID(extensionQuantity.Id)
				if err != nil {
					InternalStatusError("error fetching extension from db route /build/", err, w, h.tmpl)
					return
				}

				extension.Qty = extensionQuantity.Qty
				extensions = append(extensions, *extension)
			}
			bundle.Extensions = extensions
			// add bundle meta data
			bundle.ComputeMetaData()

			type SingleBundlePageData struct {
				BasePageData BasePageData
				Bundle       models.Bundle
			}

			pageData := SingleBundlePageData{
				BasePageData: BasePageData{
					PageTitle:       "Single Bundle: " + bundle.Name,
					MetaDescription: "Buy Bundle " + bundle.Name + " Online and enjoy super fast delivery",
				},
				Bundle: bundle,
			}

			err = h.tmpl.ExecuteTemplate(w, "single-bundle.tmpl", pageData)
			if err != nil {
				InternalStatusError("error creating bundle page", err, w, h.tmpl)
				return
			}
			return
		}

		if r.URL.Path == "/bundles/" {
			popularBundles, err := h.productService.GetBundles(services.ProductFilterParams{})
			if err != nil {
				InternalStatusError("error fetching popular bundles for route /bundles/", err, w, h.tmpl)
				return
			}
			pageData := struct {
				PopularBundles []*models.Product
			}{
				PopularBundles: popularBundles,
			}

			err = h.tmpl.ExecuteTemplate(w, "bundles.tmpl", pageData)
			if err != nil {
				InternalStatusError("error executing bundles template", err, w, h.tmpl)
				return
			}

			return
		}
	}
	NotFound(w, h.tmpl)

}

package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/seanomeara96/gates/models"
	"github.com/seanomeara96/gates/render"
	"github.com/seanomeara96/gates/services"
)

func InValidRequest(w http.ResponseWriter) {
	//templateErr := tmpl.ExecuteTemplate(w, "inavlidRequest.tmpl", nil)
	//if templateErr != nil {
	http.Error(w, "Invalid Request Method", http.StatusMethodNotAllowed)
	//}
}

func InternalStatusError(description string, err error, w http.ResponseWriter, r *render.Renderer) {
	fmt.Println(description)
	fmt.Println(err)
	t_err := r.NotFoundPage(w, render.NotFoundPageData{})
	if t_err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func NotFound(w http.ResponseWriter, r *render.Renderer) {
	w.WriteHeader(http.StatusNotFound)
	err := r.NotFoundPage(w, render.NotFoundPageData{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type PageHandler struct {
	productService *services.ProductService
	render         *render.Renderer
}

func NewPageHandler(productService *services.ProductService, renderer *render.Renderer) *PageHandler {
	return &PageHandler{
		productService: productService,
		render:         renderer,
	}
}
func (h *PageHandler) Home(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet && r.URL.Path == "/" {
		featuredGates, err := h.productService.GetGates(services.ProductFilterParams{})
		if err != nil {
			InternalStatusError("could not fetch gates from db", err, w, h.render)
			return
		}

		popularBundles, err := h.productService.GetBundles(services.ProductFilterParams{Limit: 3})
		if err != nil {
			InternalStatusError("could not fetch bundles from db", err, w, h.render)
			return
		}
		user := render.User{
			Email: "sean@example.com",
		}
		pageData := render.HomePageData{
			BasePageData: render.BasePageData{
				PageTitle:       "Build your own safety gate",
				MetaDescription: "This is a place to build the perfect safety gate for your home",
				User:            user,
			},
			FeaturedGates:  featuredGates,
			PopularBundles: popularBundles,
		}

		err = h.render.HomePage(w, pageData)
		if err != nil {
			InternalStatusError("could not execute templete fo homepage", err, w, h.render)
		}
		return
	}

	NotFound(w, h.render)
}

func (h *PageHandler) Gates(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet && r.URL.Path == "/gates/" {
		gates, err := h.productService.GetGates(services.ProductFilterParams{})
		if err != nil {
			InternalStatusError("error fetching gates for route /gates/", err, w, h.render)
			return
		}

		pageData := render.ProductsPageData{
			Heading:  "Shop Individual Gates",
			Products: gates,
		}

		err = h.render.ProductsPage(w, pageData)
		if err != nil {
			InternalStatusError("error rendering gates page", err, w, h.render)
			return
		}
		return
	}

	splitPath := strings.Split(r.URL.Path, "/")
	gateID, err := strconv.Atoi(splitPath[len(splitPath)-1])

	if r.Method == http.MethodGet && err == nil {
		gate, err := h.productService.GetProductByID(gateID)
		if err != nil {
			InternalStatusError("error fetching gate", err, w, h.render)
			return
		}
		pageData := render.ProductPageData{
			BasePageData: render.BasePageData{
				PageTitle:       gate.Name,
				MetaDescription: gate.Name,
			},
			Product: gate,
		}
		err = h.render.ProductPage(w, pageData)
		if err != nil {
			InternalStatusError("error rendering gate page", err, w, h.render)
			return
		}
		return
	}

	NotFound(w, h.render)
}

func (h *PageHandler) Extensions(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet && r.URL.Path == "/extensions/" {

		extensions, err := h.productService.GetExtensions(services.ProductFilterParams{})
		if err != nil {
			InternalStatusError("error fetching extensions for route /extensions/", err, w, h.render)
			return
		}

		data := render.ProductsPageData{
			Heading:  "Shop Individual Extensions",
			Products: extensions,
		}

		err = h.render.ProductsPage(w, data)
		if err != nil {
			InternalStatusError("cant render extensions  page", err, w, h.render)
		}
		return
	}

	splitPath := strings.Split(r.URL.Path, "/")
	extensionID, err := strconv.Atoi(splitPath[len(splitPath)-1])

	if r.Method == http.MethodGet && err == nil {
		extension, err := h.productService.GetProductByID(extensionID)
		if err != nil {
			InternalStatusError("error fetching extension", err, w, h.render)
			return
		}
		pageData := render.ProductPageData{
			BasePageData: render.BasePageData{
				PageTitle:       extension.Name,
				MetaDescription: extension.Name,
			},
			Product: extension,
		}
		err = h.render.ProductPage(w, pageData)
		if err != nil {
			InternalStatusError("error rendering extension page", err, w, h.render)
			return
		}
		return
	}

	NotFound(w, h.render)
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
				InternalStatusError("error decoding gate data", err, w, h.render)
				return
			}

			var extensionQuantities []ItemQuantities
			err = json.Unmarshal([]byte(e), &extensionQuantities)
			if err != nil {
				InternalStatusError("error decoding extensions", err, w, h.render)
				return
			}

			var bundle models.Bundle

			gate, err := h.productService.GetProductByID(gateQuantity.Id)
			if err != nil {
				InternalStatusError("error fetching gate from db for route /bundles/", err, w, h.render)
				return
			}
			gate.Qty = gateQuantity.Qty

			bundle.Gates = append(bundle.Gates, *gate)

			var extensions []models.Product
			for _, extensionQuantity := range extensionQuantities {
				extension, err := h.productService.GetProductByID(extensionQuantity.Id)
				if err != nil {
					InternalStatusError("error fetching extension from db route /build/", err, w, h.render)
					return
				}

				extension.Qty = extensionQuantity.Qty
				extensions = append(extensions, *extension)
			}
			bundle.Extensions = extensions
			// add bundle meta data
			bundle.ComputeMetaData()

			pageData := render.BundlePageData{
				BasePageData: render.BasePageData{
					PageTitle:       "Single Bundle: " + bundle.Name,
					MetaDescription: "Buy Bundle " + bundle.Name + " Online and enjoy super fast delivery",
				},
				Bundle: &bundle,
			}

			err = h.render.BundlePage(w, pageData)
			if err != nil {
				InternalStatusError("error creating bundle page", err, w, h.render)
				return
			}
			return
		}

		if r.URL.Path == "/bundles/" {
			popularBundles, err := h.productService.GetBundles(services.ProductFilterParams{})
			if err != nil {
				InternalStatusError("error fetching popular bundles for route /bundles/", err, w, h.render)
				return
			}
			pageData := render.ProductsPageData{
				Products: popularBundles,
			}

			err = h.render.ProductsPage(w, pageData)
			if err != nil {
				InternalStatusError("error executing bundles template", err, w, h.render)
				return
			}

			return
		}
	}
	NotFound(w, h.render)

}

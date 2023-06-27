package handlers

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/seanomeara96/gates/models"
	"github.com/seanomeara96/gates/services"
)

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
	if r.Method == http.MethodGet && r.URL.Path == "/bundles/" {
		bundles, err := h.productService.GetBundles(services.ProductFilterParams{})
		if err != nil {
			InternalStatusError("error fetching bundles for route '/bundles/'", err, w, h.tmpl)
			return
		}

		pageData := struct {
			Heading  string
			Products []*models.Product
		}{
			Heading:  "Shop Bundles",
			Products: bundles,
		}

		h.tmpl.ExecuteTemplate(w, "products.tmpl", pageData)
		return
	}
	NotFound(w, h.tmpl)
}

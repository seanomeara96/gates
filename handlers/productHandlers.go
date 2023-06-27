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

type ProductHandler struct {
	productService *services.ProductService
	tmpl           *template.Template
}

func NewProductHandler(productService *services.ProductService, templates *template.Template) *ProductHandler {
	return &ProductHandler{
		productService: productService,
		tmpl:           templates,
	}
}

func (h *ProductHandler) GetGates(w http.ResponseWriter, r *http.Request) {
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

func (h *ProductHandler) GetExtensions(w http.ResponseWriter, r *http.Request) {
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

package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/seanomeara96/gates/models"
	"github.com/seanomeara96/gates/render"
	"github.com/seanomeara96/gates/services"
)

func InValidRequest(w http.ResponseWriter) {
	http.Error(w, "Invalid Request Method", http.StatusMethodNotAllowed)
}

type PageHandler struct {
	productService *services.ProductService
	cartService    *services.CartService
	render         *render.Renderer
}

func NewPageHandler(
	productService *services.ProductService,
	cartService *services.CartService,
	renderer *render.Renderer,
) *PageHandler {
	return &PageHandler{
		productService: productService,
		render:         renderer,
	}
}

func (h *PageHandler) CustomBundles(w http.ResponseWriter, r *http.Request) error {
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
			return err
		}

		var extensionQuantities []ItemQuantities
		err = json.Unmarshal([]byte(e), &extensionQuantities)
		if err != nil {
			return err
		}

		var bundle models.Bundle

		gate, err := h.productService.GetProductByID(gateQuantity.Id)
		if err != nil {
			return err
		}
		gate.Qty = gateQuantity.Qty

		bundle.Gates = append(bundle.Gates, *gate)

		var extensions []models.Product
		for _, extensionQuantity := range extensionQuantities {
			extension, err := h.productService.GetProductByID(extensionQuantity.Id)
			if err != nil {
				return err
			}

			extension.Qty = extensionQuantity.Qty
			extensions = append(extensions, *extension)
		}
		bundle.Extensions = extensions
		// add bundle meta data
		bundle.ComputeMetaData()

		pageTitle := "Single Bundle: " + bundle.Name
		metaDescription := "Buy Bundle " + bundle.Name + " Online and enjoy super fast delivery"
		user := models.User{}

		basePageData := h.render.NewBasePageData(pageTitle, metaDescription, user)
		pageData := h.render.NewBundlePageData(basePageData, &bundle)

		err = h.render.BundlePage(w, pageData)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *PageHandler) SomethingWentWrong(w http.ResponseWriter, r *http.Request) error {
	basePageData := h.render.NewBasePageData("Something went wrong", "There was an problem processing your request. Please try again later", models.User{})
	if err := h.render.SomethingWentWrong(w, basePageData); err != nil {
		return err
	}

	return nil
}

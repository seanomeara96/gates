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
	http.Error(w, "Invalid Request Method", http.StatusMethodNotAllowed)
}

func NotFound(w http.ResponseWriter, r *render.Renderer) error {
	w.WriteHeader(http.StatusNotFound)
	pageTile := "Page Not Found"
	metaDescription := "Unable to find page"
	user := models.User{}

	basePageData := r.NewBasePageData(pageTile, metaDescription, user)
	err := r.NotFoundPage(w, r.NotFoundPageData(basePageData))
	if err != nil {
		return err
	}
	return nil
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
func (h *PageHandler) Home(w http.ResponseWriter, r *http.Request) error {
	if r.URL.Path == "/" {
		featuredGates, err := h.productService.GetGates(services.ProductFilterParams{})
		if err != nil {
			return err
		}

		popularBundles, err := h.productService.GetBundles(services.ProductFilterParams{Limit: 3})
		if err != nil {
			return err
		}

		pageTitle := "Home Page"
		metaDescription := "Welcome to the home page"
		user := models.User{
			Email: "sean@example.com",
		}

		basePageData := h.render.NewBasePageData(pageTitle, metaDescription, user)

		homepageData := h.render.NewHomePageData(featuredGates, popularBundles, basePageData)
		if err = h.render.HomePage(w, homepageData); err != nil {
			return err
		}
		return nil
	}

	return NotFound(w, h.render)
}

func (h *PageHandler) Gates(w http.ResponseWriter, r *http.Request) error {
	if r.Method == http.MethodGet && r.URL.Path == "/gates/" {
		gates, err := h.productService.GetGates(services.ProductFilterParams{})
		if err != nil {
			return err
		}

		pageTitle := "Shop All Gates"
		metaDescription := "Shop our full range of gates"
		user := models.User{}

		basePageData := h.render.NewBasePageData(pageTitle, metaDescription, user)
		productsPageData := h.render.NewProductsPageData(basePageData, pageTitle, gates)

		err = h.render.ProductsPage(w, productsPageData)
		if err != nil {
			return err
		}
		return nil
	}

	splitPath := strings.Split(r.URL.Path, "/")
	gateID, err := strconv.Atoi(splitPath[len(splitPath)-1])

	if r.Method == http.MethodGet && err == nil {
		gate, err := h.productService.GetProductByID(gateID)
		if err != nil {
			return err
		}

		pageTitle := gate.Name
		metaDescription := gate.Name
		user := models.User{}

		basePageDate := h.render.NewBasePageData(pageTitle, metaDescription, user)
		productPageData := h.render.NewProductPageData(basePageDate, gate)

		err = h.render.ProductPage(w, productPageData)
		if err != nil {
			return err
		}
		return nil
	}

	return NotFound(w, h.render)
}

func (h *PageHandler) Extensions(w http.ResponseWriter, r *http.Request) error {
	if r.Method == http.MethodGet && r.URL.Path == "/extensions/" {

		extensions, err := h.productService.GetExtensions(services.ProductFilterParams{})
		if err != nil {
			return err
		}

		pageTitle := "All extensions"
		metaDescription := "Shop all extensions"
		user := models.User{}

		basePageData := h.render.NewBasePageData(pageTitle, metaDescription, user)

		heading := pageTitle
		productsPageData := h.render.NewProductsPageData(basePageData, heading, extensions)

		err = h.render.ProductsPage(w, productsPageData)
		if err != nil {
			return err
		}
		return nil
	}

	splitPath := strings.Split(r.URL.Path, "/")
	extensionID, err := strconv.Atoi(splitPath[len(splitPath)-1])

	if r.Method == http.MethodGet && err == nil {
		extension, err := h.productService.GetProductByID(extensionID)
		if err != nil {
			return err
		}

		pageTitle := extension.Name
		metaDescription := extension.Name
		user := models.User{}

		basePageData := h.render.NewBasePageData(pageTitle, metaDescription, user)
		productPageData := h.render.NewProductPageData(basePageData, extension)
		err = h.render.ProductPage(w, productPageData)
		if err != nil {
			return err
		}
		return nil
	}

	return NotFound(w, h.render)
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

func (h *PageHandler) Bundles(w http.ResponseWriter, r *http.Request) error {

	popularBundles, err := h.productService.GetBundles(services.ProductFilterParams{})
	if err != nil {
		return err
	}
	pageTitle := "Bundles Page"
	metaDescription := "Shop All Bundles"
	// todo remove
	user := models.User{}
	basePageData := h.render.NewBasePageData(pageTitle, metaDescription, user)
	pageData := h.render.NewProductsPageData(basePageData, pageTitle, popularBundles)

	if err = h.render.ProductsPage(w, pageData); err != nil {
		return err
	}

	return nil
}

func (h *PageHandler) Cart(w http.ResponseWriter, r *http.Request) error {
	cart := &models.Cart{}
	user := models.User{}
	pageTitle := "Your shopping cart"
	metaDescription := ""
	basePageData := h.render.NewBasePageData(pageTitle, metaDescription, user)
	cartPageData := h.render.NewCartPageData(basePageData, cart)

	err := h.render.CartPage(w, cartPageData)
	if err != nil {
		return fmt.Errorf("proble rendering cart page. %w", err)
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

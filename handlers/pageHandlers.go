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
	pageTile := "Internal Status Error"
	metaDescription := "Unable to load page"
	user := models.User{}

	basePageData := r.NewBasePageData(pageTile, metaDescription, user)
	t_err := r.NotFoundPage(w, r.NotFoundPageData(basePageData))
	if t_err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func NotFound(w http.ResponseWriter, r *render.Renderer) {
	w.WriteHeader(http.StatusNotFound)
	pageTile := "Page Not Found"
	metaDescription := "Unable to find page"
	user := models.User{}

	basePageData := r.NewBasePageData(pageTile, metaDescription, user)
	err := r.NotFoundPage(w, r.NotFoundPageData(basePageData))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

		pageTitle := "Home Page"
		metaDescription := "Welcome to the home page"
		user := models.User{
			Email: "sean@example.com",
		}

		basePageData := h.render.NewBasePageData(pageTitle, metaDescription, user)

		homepageData := h.render.NewHomePageData(featuredGates, popularBundles, basePageData)
		err = h.render.HomePage(w, homepageData)
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

		pageTitle := "Shop All Gates"
		metaDescription := "Shop our full range of gates"
		user := models.User{}

		basePageData := h.render.NewBasePageData(pageTitle, metaDescription, user)
		productsPageData := h.render.NewProductsPageData(basePageData, pageTitle, gates)

		err = h.render.ProductsPage(w, productsPageData)
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

		pageTitle := gate.Name
		metaDescription := gate.Name
		user := models.User{}

		basePageDate := h.render.NewBasePageData(pageTitle, metaDescription, user)
		productPageData := h.render.NewProductPageData(basePageDate, gate)

		err = h.render.ProductPage(w, productPageData)
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

		pageTitle := "All extensions"
		metaDescription := "Shop all extensions"
		user := models.User{}

		basePageData := h.render.NewBasePageData(pageTitle, metaDescription, user)

		heading := pageTitle
		productsPageData := h.render.NewProductsPageData(basePageData, heading, extensions)

		err = h.render.ProductsPage(w, productsPageData)
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

		pageTitle := extension.Name
		metaDescription := extension.Name
		user := models.User{}

		basePageData := h.render.NewBasePageData(pageTitle, metaDescription, user)
		productPageData := h.render.NewProductPageData(basePageData, extension)
		err = h.render.ProductPage(w, productPageData)
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

			pageTitle := "Single Bundle: " + bundle.Name
			metaDescription := "Buy Bundle " + bundle.Name + " Online and enjoy super fast delivery"
			user := models.User{}

			basePageData := h.render.NewBasePageData(pageTitle, metaDescription, user)
			pageData := h.render.NewBundlePageData(basePageData, &bundle)

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
			pageTitle := "Bundles Page"
			metaDescription := "Shop All Bundles"
			// todo remove
			user := models.User{}
			basePageData := h.render.NewBasePageData(pageTitle, metaDescription, user)
			pageData := h.render.NewProductsPageData(basePageData, pageTitle, popularBundles)

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

func (h *PageHandler) Cart(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/cart/" && r.Method == http.MethodGet {
		cart := &models.Cart{}
		cartItems := []*models.CartItem{}
		user := models.User{}
		pageTitle := "Your shopping cart"
		metaDescription := ""
		basePageData := h.render.NewBasePageData(pageTitle, metaDescription, user)
		cartPageData := h.render.NewCartPageData(basePageData, cart, cartItems)

		err := h.render.CartPage(w, cartPageData)
		if err != nil {
			InternalStatusError("error executing cart page template", err, w, h.render)
		}
		return
	}
	NotFound(w, h.render)
}

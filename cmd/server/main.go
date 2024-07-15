package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"github.com/seanomeara96/gates/handlers"
	"github.com/seanomeara96/gates/models"
	"github.com/seanomeara96/gates/render"
	"github.com/seanomeara96/gates/repositories"
	"github.com/seanomeara96/gates/services"
	"github.com/seanomeara96/gates/utils"
)

type User struct {
	Email string
}

type BasePageData struct {
	MetaDescription string
	PageTitle       string
	User            User
}
type customHandleFunc func(w http.ResponseWriter, r *http.Request) error
type middlewareFn func(w http.ResponseWriter, r *http.Request) (execNextFunc bool, err error)
type middlewaresFunc func(w http.ResponseWriter, r *http.Request, fn customHandleFunc) error
type customHandler func(path string, fn customHandleFunc)

func main() {

	portValue := flag.String("port", "", "port to listen on")
	createTable := flag.String("create-table", "", "table to create")
	env := flag.String("env", "dev", "dev or prod")

	flag.Parse()

	if *portValue == "" {
		log.Fatal("no port supplied")
	}

	environment := Development
	if *env == "prod" {
		environment = Production
	}

	port := *portValue

	db := utils.SqliteOpen("main.db")
	defer db.Close()

	store := sessions.NewCookieStore([]byte(`secret-key`))

	router := http.NewServeMux()

	renderer := render.NewRenderer("templates/**/*.tmpl", environment)

	productRepo := repositories.NewProductRepository(db)
	bundleRepo := repositories.NewBundleRepository(db)
	cartRepo := repositories.NewCartRepository(db)

	productService := services.NewProductService(productRepo)
	bundleService := services.NewBundleService(productRepo, bundleRepo)
	cartService := services.NewCartService(cartRepo, productRepo)

	pageHandler := handlers.NewPageHandler(productService, cartService, renderer)

	// ROUTING LOGIC
	// TODO put back in the cart middleware
	middlewareFuncs := []middlewareFn{}

	/*
		Call the middlewares func for each request
		Would be better if we could also do a per route basis
	*/
	middlewares := registerMiddlewares(middlewareFuncs)

	handle := createCustomHandler(environment, router, middlewares, pageHandler.SomethingWentWrong)

	handle("/", func(w http.ResponseWriter, r *http.Request) error {
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

			basePageData := h.render.NewBasePageData(pageTitle, metaDescription)

			homepageData := h.render.NewHomePageData(featuredGates, popularBundles, basePageData)

			if err = h.render.HomePage(w, homepageData); err != nil {
				return err
			}
			return nil
		}

		return NotFound(w, h.render)
	}) // cant use 'get' because it causes conflicts

	/*
		Build enpoint. Currently only handling build for pressure gates
	*/
	handle.post("/build/", func(w http.ResponseWriter, r *http.Request) error {
		if err := r.ParseForm(); err != nil {
			return err
		}

		_desiredWidth := r.Form["desired-width"][0]

		desiredWidth, err := strconv.ParseFloat(_desiredWidth, 32)
		if err != nil {
			return err
		}

		// TODO keep track of user inputs(valid ones). From there we can generate "popular bundles"
		if err := h.bundleService.SaveRequestedBundleSize(float32(desiredWidth)); err != nil {
			return err
		}

		bundles, err := h.bundleService.BuildPressureFitBundles(float32(desiredWidth))
		if err != nil {
			return err
		}

		templateData := h.render.NewBundleBuildResultsData(
			float32(desiredWidth),
			bundles,
		)

		if err = h.render.BundleBuildResults(w, templateData); err != nil {
			return err
		}

		return nil
	})

	/*
		Product page endpoints
	*/
	//handle.get("/bundles/", pageHandler.Bundles)
	//handle.get("/bundles/new", pageHandler.Bundles)
	handle.get("/gates/", func(w http.ResponseWriter, r *http.Request) error {
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
	})

	handle.get("/gates/{gate_id}", func(w http.ResponseWriter, r *http.Request) error {
		gateID := r.PathValue("gate_id")
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

	})

	handle.get("/extensions/", func(w http.ResponseWriter, r *http.Request) error {
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
	})

	handle.get("/extensions/{extension_id}", func(w http.ResponseWriter, r *http.Request) error {
		extensionID := r.PathValue("extension_id")
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
	})

	/*
		cart endpoints
	*/
	handle.get("/cart/", func(w http.ResponseWriter, r *http.Request) error {
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
	})

	handle.post("/cart/add", func(w http.ResponseWriter, r *http.Request) error {
		session, err := h.getCartSession(r)
		if err != nil {
			return err
		}

		cartID, err := h.getCartID(session)
		if err != nil {
			return err
		}

		if ok := validateCartID(cartID); !ok {
			return errors.New("invalid cart id")
		}

		if err := r.ParseForm(); err != nil {
			return err
		}

		components := []models.CartItemComponent{}

		for _, d := range r.Form["data"] {
			var component models.CartItemComponent
			if err := json.Unmarshal([]byte(d), &component); err != nil {
				return err
			}
			components = append(components, component)
		}

		if err := h.cartService.AddItem(cartID.(string), components); err != nil {
			return err
		}

		return nil
	}) // TODO consolidate add & update methods

	handle.post("/cart/remove", func(w http.ResponseWriter, r *http.Request) error {
		session, err := getCartSession(r)
		if err != nil {
			return err
		}

		cartID, err := getCartID(session)
		if err != nil {
			return err
		}

		if ok := validateCartID(cartID); !ok {
			return errors.New("invalid cart id")
		}

		r.ParseForm()

		itemID := r.Form.Get("item_id")

		if err := cartService.RemoveItem(cartID.(string), itemID); err != nil {
			return err
		}

		return nil
	})
	handle.post("/cart/clear", cartHandler.ClearCart)

	handle.delete("/test/{int}", func(w http.ResponseWriter, r *http.Request) error {
		fmt.Println("Testing ", r.PathValue("int"))
		w.WriteHeader(http.StatusOK)
		return nil
	})

	// init assets dir
	assetsDirPath := "/assets/"
	httpFileSystem := http.Dir("assets")
	staticFileHttpHandler := http.FileServer(httpFileSystem)
	assetsPathHandler := http.StripPrefix(assetsDirPath, staticFileHttpHandler)
	router.Handle(assetsDirPath, assetsPathHandler)

	log.Println("Listening on http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func registerMiddlewares(middlewareFuncs []middlewareFn) func(w http.ResponseWriter, r *http.Request, fn customHandleFunc) error {
	return func(w http.ResponseWriter, r *http.Request, fn customHandleFunc) error {
		for _, middle := range middlewareFuncs {
			next, err := middle(w, r)
			if err != nil {
				return err
			}
			if !next {
				return nil
			}
		}
		return fn(w, r)
	}
}

func sendErr(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func createCustomHandler(environment Environment, router *http.ServeMux, executeMiddlewares middlewaresFunc, errPageHandler customHandleFunc) customHandler {
	return func(path string, fn customHandleFunc) {
		router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {

			if environment == Development {
				rMsg := "[INFO] %s request made to %s"
				log.Printf(rMsg, r.Method, r.URL.Path)
			}

			// custom handler get passed throgh the carthandler middleware first to
			// ensure there is a cart session
			err := executeMiddlewares(w, r, fn)
			if err != nil {
				log.Printf("[ERROR] Failed  %s request to %s. %v", r.Method, path, err)

				// ideally I would get notified of an error here
				if environment == Development {
					sendErr(w, err)
				} else {
					err := errPageHandler(w, r)
					if err != nil {
						sendErr(w, err)
					}
				}

			}
		})
	}

}

func (handle customHandler) get(path string, fn customHandleFunc) {
	handle("GET "+path, fn)
}
func (handle customHandler) post(path string, fn customHandleFunc) {
	handle("POST "+path, fn)
}
func (handle customHandler) put(path string, fn customHandleFunc) {
	handle("PUT "+path, fn)
}
func (handle customHandler) delete(path string, fn customHandleFunc) {
	handle("DELETE"+path, fn)
}

func validateCartID(cartID interface{}) (valid bool) {
	if _, ok := cartID.(string); !ok {
		return false
	}
	return true
}

func getCartSession(r *http.Request, store *sessions.CookieStore) (*sessions.Session, error) {
	session, err := store.Get(r, "cart-session")
	if err != nil {
		return nil, err
	}
	return session, nil
}

func getCartID(session *sessions.Session) (interface{}, error) {
	if session == nil {
		return nil, errors.New("Cart Session is nil")
	}
	return session.Values["cart_id"], nil

}

func CartMiddleWare(store *sessions.CookieStore, w http.ResponseWriter, r *http.Request) (bool, error) {
	session, err := getCartSession(r, store)
	if err != nil {
		return false, err
	}

	cartID, err := getCartID(session)
	if err != nil {
		return false, err
	}

	if cartID != nil {
		return true, nil
	}

	cartID, err = h.cartService.NewCart()
	if err != nil {
		return false, err
	}

	session.Values["cart_id"] = cartID
	if err := session.Save(r, w); err != nil {
		return false, err
	}

	return true, nil
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

package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"

	"github.com/seanomeara96/gates/config"
	"github.com/seanomeara96/gates/handlers"
	"github.com/seanomeara96/gates/models"
	"github.com/seanomeara96/gates/repos"
	"github.com/seanomeara96/gates/repos/cache"
	"github.com/seanomeara96/gates/repos/sqlite"
)

type customHandleFunc func(cart *models.Cart, w http.ResponseWriter, r *http.Request) error
type middlewareFunc func(next customHandleFunc) customHandleFunc
type customHandler func(path string, fn customHandleFunc)

func app() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	handler, err := handlers.DefaultHandler(cfg)
	if err != nil {
		return err
	}
	defer handler.Close()

	router := http.NewServeMux()
	getCartFromRequest := NewCartFromSessionGetter(cartRepo, productCache, store)
	// ROUTING LOGIC
	// middleware executed in reverse order; i = 0 executes last
	middleware := []middlewareFunc{
		// Example for middleware usage:
	}

	handle := createCustomHandler(cfg.Mode, router, middleware, getCartFromRequest)

	handle("/webhook", handler.StripeWebhook)
	handle("/", handler.GetHomePage)
	handle.get("/admin/login", handler.GetAdminLoginPage)
	handle.post("/admin/login", handler.AdminLogin)
	handle.get("/admin", handler.GetAdminDashboard)
	handle.get("/contact/", handler.GetContactPage)
	handle.get("/checkout/", handler.GetCheckoutPage)
	handle.post("/contact/", handler.ProcessContactFormSumbission)
	// Build endpoint. Currently only handling builds for pressure gates.
	handle.post("/build/", handler.BuildBundle)
	// Product page endpoints.
	handle.get("/gates/", handler.GetGatesPage)
	handle.get("/gates/{gate_id}", handler.GetGatePage)
	handle.get("/extensions/", handler.GetExtensionsPage)
	handle.get("/extensions/{extension_id}", handler.GetExtensionPage)
	// Cart endpoints.
	handle.get("/cart/", handler.GetCartPage)
	handle.get("/cart/json", handler.GetCartJSON)

	if cfg.Mode == config.Development {
		handle("/test", handler.Test)
	}
	handle.post("/cart/add", handlers.AddItemToCart)
	handle.post("/cart/item/{mode}", handlers.AdjustCartItemQty)
	handle.delete("/cart/item/", handler.RemoveItemFromCart)
	handle.post("/cart/clear", handler.ClearItemsFromCart)

	// Initialize assets dir.
	assetsDirPath := "/assets/"
	httpFileSystem := http.Dir("assets")
	staticFileHttpHandler := http.FileServer(httpFileSystem)
	assetsPathHandler := http.StripPrefix(assetsDirPath, staticFileHttpHandler)
	router.Handle(assetsDirPath, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(24*int(time.Hour.Seconds())))
		assetsPathHandler.ServeHTTP(w, r)
	}))

	srv := http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	errChan := make(chan error)
	go func() {
		log.Println("Listening on http://localhost:" + cfg.Port)
		if err := srv.ListenAndServe(); err != nil {
			errChan <- fmt.Errorf("server startup: failed to listen and serve on env.config.Port %s: %w", cfg.Port, err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errChan:
		close(errChan)
		return fmt.Errorf("server errored: %w", err)
	case <-quit:
		close(quit)
		log.Println("shutting down server gracefully")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			return fmt.Errorf("server forced to shut down %w", err)
		}
		log.Println("server shut down gracefully")
		return nil
	}
}

func main() {
	if err := app(); err != nil {
		log.Fatal(err)
	}
}

func sendErr(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func attachNewCartToSession(cart *models.Cart, session *sessions.Session, w http.ResponseWriter, r *http.Request) error {

	session.Values["cart_id"] = cart.ID
	if err := session.Save(r, w); err != nil {
		return err
	}
	return nil
}

func NewCartFromSessionGetter(cartRepo *sqlite.CartRepo, products *cache.CachedProductRepo, store *sessions.CookieStore) func(w http.ResponseWriter, r *http.Request) (*models.Cart, error) {
	return func(w http.ResponseWriter, r *http.Request) (*models.Cart, error) {
		session, err := getCartSession(r, store)
		if err != nil {
			return nil, err
		}

		cartID, err := getCartID(session)
		if err != nil {
			return nil, err
		}

		if cartID == nil {
			cart, err := NewCart(cartRepo)
			if err != nil {
				return nil, err
			}
			if err := attachNewCartToSession(cart, session, w, r); err != nil {
				return nil, err
			}
			return cart, nil
		}

		if valid := validateCartID(cartID); !valid {
			return nil, fmt.Errorf("cart id is invalid")
		}

		exists, err := cartRepo.CartExists(cartID.(string))
		if err != nil {
			return nil, err
		}
		if !exists {
			cart, err := NewCart(cartRepo)
			if err != nil {
				return nil, err
			}
			if err := attachNewCartToSession(cart, session, w, r); err != nil {
				return nil, err
			}
			return cart, nil
		}

		cart, err := cartRepo.GetCartByID(cartID.(string))
		if err != nil {
			return nil, err
		}

		return cart, nil
	}
}

func createCustomHandler(environment config.Environment, router *http.ServeMux, middleware []middlewareFunc, getCartFromRequest func(w http.ResponseWriter, r *http.Request) (*models.Cart, error)) customHandler {
	return func(path string, fn customHandleFunc) {

		// wrap fn in middleware funcs
		for i := range middleware {
			fn = middleware[i](fn)
		}

		router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			if environment == config.Development {
				rMsg := "[INFO] %s request made to %s"
				log.Printf(rMsg, r.Method, r.URL.Path)
			}

			cart, err := getCartFromRequest(w, r)
			if err != nil {
				errMsg := "[ERROR] Failed  %s request to %s. %v"
				log.Printf(errMsg, r.Method, path, err)

				// ideally I would get notified of an error here
				if environment == config.Development {
					sendErr(w, err)
				} else {
					// TODO handle production env err handling different
					sendErr(w, err)
				}
				return
			}

			// custom handler get passed throgh the cart handler middleware first to
			// ensure there is a cart session
			if err := fn(cart, w, r); err != nil {
				errMsg := "[ERROR] Failed  %s request to %s. %v"
				log.Printf(errMsg, r.Method, path, err)

				// ideally I would get notified of an error here
				if environment == config.Development {
					sendErr(w, err)
				} else {
					// TODO handle production env err handling different
					sendErr(w, err)
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

//	func (handle customHandler) put(path string, fn customHandleFunc) {
//		handle("PUT "+path, fn)
//	}
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
		return nil, errors.New("cart Session is nil")
	}
	return session.Values["cart_id"], nil

}

/* service funcs start */

func BuildPressureFitBundles(products *cache.CachedProductRepo, limit float32) ([]models.Bundle, error) {
	var bundles []models.Bundle

	gates, err := products.GetProducts(models.ProductTypeGate, repos.ProductFilterParams{MaxWidth: limit})
	if err != nil {
		return bundles, err
	}
	if len(gates) < 1 {
		return bundles, nil
	}

	for _, gate := range gates {
		compatibleExtensions, err := products.GetCompatibleExtensionsByGateID(gate.Id)
		if err != nil {
			return bundles, err
		}

		bundle, err := BuildPressureFitBundle(limit, gate, compatibleExtensions)
		if err != nil {
			return bundles, err
		}
		bundles = append(bundles, bundle)
	}

	return bundles, nil
}

func BuildPressureFitBundle(limit float32, gate *models.Product, extensions []*models.Product) (models.Bundle, error) {
	widthLimit := limit

	var bundle = models.Bundle{}

	// returning a single bundle
	bundle.Qty = 1

	//  add gate to the bundle. Ensure Qty is at least 1
	if gate.Width > widthLimit {
		return bundle, errors.New("gate too big")
	}

	if gate.Qty < 1 {
		gate.Qty = 1
	}

	bundle.Components = append(bundle.Components, *gate)

	widthLimit -= gate.Width

	// sort extensions to ensure width descending
	sort.Slice(extensions, func(i int, j int) bool {
		return extensions[i].Width > extensions[j].Width
	})

	extensionIndex := 0
	for widthLimit > 0 {

		// we want to add one more extension if the width remaining > 0 but we've reached the last extension
		var override bool = false
		if extensionIndex >= len(extensions) {
			extensionIndex--
			override = true
		}

		extension := extensions[extensionIndex]
		if extension.Width > widthLimit && !override {
			//  extension too big, try next extension size down
			extensionIndex++
			continue
		}

		// check if extension already exists in the bundle and if so, increment the qty, else add it with a qty of 1
		var existingExtension *models.Product
		for ii := 1; ii < len(bundle.Components); ii++ {
			var bundleExtension *models.Product = &bundle.Components[ii]

			if bundleExtension.Id == extension.Id {
				existingExtension = bundleExtension
			}
		}

		if existingExtension != nil {
			existingExtension.Qty++
			widthLimit -= existingExtension.Width
		} else {
			extension.Qty = 1
			bundle.Components = append(bundle.Components, *extension)
			widthLimit -= extension.Width
		}
	}
	bundle.ComputeMetaData()
	return bundle, nil
}

func TotalValue(products *cache.CachedProductRepo, cart models.Cart) (float32, error) {
	var value float32 = 0.0
	for _, item := range cart.Items {
		for _, component := range item.Components {
			productPrice, err := products.GetProductPrice(component.Product.Id)
			if err != nil {
				return 0, err
			}
			value += (productPrice * float32(component.Qty))
		}
	}
	return value, nil
}

func NewCart(cartRepo *sqlite.CartRepo) (*models.Cart, error) {
	cart := models.NewCart()
	if _, err := cartRepo.SaveCart(cart); err != nil {
		return nil, err
	}
	return &cart, nil
}

func AddItemToCart(cartRepo *sqlite.CartRepo, cartID string, cartItem models.CartItem) error {
	exists, err := cartRepo.DoesCartItemExist(cartID, cartItem.ID)
	if err != nil {
		return err
	}
	if !exists {
		if err := cartRepo.InsertCartItem(cartItem); err != nil {
			return fmt.Errorf("adding item to cart failed at insert cartitem: %w", err)
		}
		if err := cartRepo.SaveCartItemComponents(cartItem.Components); err != nil {
			return fmt.Errorf("adding item components failed: %w", err)
		}
	} else {
		if err := cartRepo.IncrementCartItem(cartID, cartItem.ID); err != nil {
			return fmt.Errorf("adding item to cart failed at increment cart item %w", err)
		}
	}
	if err := cartRepo.SetLastUpdated(cartID); err != nil {
		return fmt.Errorf("failed to update last_updated field from main.go; %w", err)
	}
	return nil
}

func RemoveItem(cartRepo *sqlite.CartRepo, cartID, itemID string) error {
	// TODO add transactions
	if err := cartRepo.RemoveCartItem(cartID, itemID); err != nil {
		return fmt.Errorf("failed to remove item. %w", err)
	}
	if err := cartRepo.RemoveCartItemComponents(itemID); err != nil {
		return fmt.Errorf("failed to remove item components. %w", err)
	}
	return nil
}

// Other methods for user-related operations (e.g., UpdateUser, DeleteUser, etc.)

/* service funcs end */

/* repository funcs start */

func SaveRequestedBundleSize(db *sql.DB, desiredWidth float32) error {
	_, err := db.Exec("INSERT INTO bundle_sizes (type, size) VALUES ('pressure fit', ?)", desiredWidth)
	if err != nil {
		return err
	}
	return nil
}

type renderPageFunc func(w http.ResponseWriter, templateName string, templateData map[string]any) error
type renderPartialFunc func(w http.ResponseWriter, templateName string, templateData any) error

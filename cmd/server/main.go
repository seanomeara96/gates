package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"text/template"
	"time"

	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"github.com/patrickmn/go-cache"
	"github.com/seanomeara96/gates/models"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type key string

const cartKey key = "cart"

type Environment string

const (
	Development Environment = "development"
	Production  Environment = "production"
)

type customHandleFunc func(w http.ResponseWriter, r *http.Request) error

type middlewareFunc func(next customHandleFunc) customHandleFunc
type customHandler func(path string, fn customHandleFunc)

type CustomRequest struct {
	*http.Request
	Cart models.Cart
}

func main() {

	portValue := flag.String("port", "", "port to listen on")

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

	db := SqliteOpen("main.db")
	defer db.Close()

	defaultExpiration := time.Minute * 5
	cleanupInterval := time.Minute * 10
	productCache := cache.New(defaultExpiration, cleanupInterval)

	store := sessions.NewCookieStore([]byte(`secret-key`))

	router := http.NewServeMux()

	tmpl := templateParser(environment)
	renderPage := NewPageRenderer(tmpl)
	renderPartial := NewPartialRenderer(tmpl)

	// ROUTING LOGIC
	//cartMiddleware := newCartMiddlware(db, store)

	/*
		executed in reverse order. where i = 0 will execute last
	*/
	middleware := []middlewareFunc{
		// cart middleware
		func(next customHandleFunc) customHandleFunc {
			return func(w http.ResponseWriter, r *http.Request) error {
				session, err := getCartSession(r, store)
				if err != nil {
					return err
				}

				cartID, err := getCartID(session)
				if err != nil {
					return err
				}

				ctx := r.Context()

				var cart *models.Cart

				newCartSession := func() error {
					cart, err = NewCart(db)
					if err != nil {
						return err
					}
					session.Values["cart_id"] = cart.ID
					if err := session.Save(r.WithContext(ctx), w); err != nil {
						return err
					}
					return nil
				}

				if cartID != nil {
					if valid := validateCartID(cartID); !valid {
						return fmt.Errorf("cart id is invalid")
					}

					exists, err := CartExists(db, cartID.(string))
					if err != nil {
						return err
					}
					if !exists {
						if err := newCartSession(); err != nil {
							return err
						}
					} else {
						cart, err = GetCartByID(db, cartID.(string))
						if err != nil {
							return err
						}
					}

				} else {
					if err := newCartSession(); err != nil {
						return err
					}

				}
				ctx = context.WithValue(ctx, cartKey, cart)
				return next(w, r.WithContext(ctx))
			}
		},
	}

	handle := createCustomHandler(environment, router, middleware)

	handle("/", func(w http.ResponseWriter, r *http.Request) error {
		if r.URL.Path == "/" {
			featuredGates, err := GetGates(db, productCache, ProductFilterParams{})
			if err != nil {
				return err
			}

			popularBundles, err := GetBundles(db, productCache, ProductFilterParams{Limit: 3})
			if err != nil {
				return err
			}

			cart, err := ctxCart(r)
			if err != nil {
				return err
			}

			data := map[string]any{
				"PageTitle":       "Home Page",
				"MetaDescription": "Welcome to the home page",
				"FeaturedGates":   featuredGates,
				"PopularBundles":  popularBundles,
				"Cart":            cart,
			}

			return renderPage(w, "home", data)
		}

		return NotFoundPage(w, renderPage)
	}) // cant use 'get' because it causes conflicts

	handle.get("/contact/", func(w http.ResponseWriter, r *http.Request) error {
		cart, err := ctxCart(r)
		if err != nil {
			return err
		}
		data := map[string]any{
			"PageTitle":       "Contact BabyGate Builders",
			"MetaDescription": "Contact form for Babygate builders",
			"Cart":            cart,
		}
		return renderPage(w, "contact", data)
	})

	handle.post("/contact/", func(w http.ResponseWriter, r *http.Request) error {
		if err := r.ParseForm(); err != nil {
			return err
		}
		email := r.Form.Get("email")
		name := r.Form.Get("name")
		message := r.Form.Get("message")

		emailRegex, err := regexp.Compile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if err != nil {
			return fmt.Errorf("could not compile email validation regex. %w", err)
		}

		if message == "" {
			// need some error state on this form
			return renderPage(w, "contact", map[string]any{
				"PageTitle":       "Contact us | No Message Provided",
				"MetaDescription": "Please provide a message",
				"Cart":            nil,
			})
		}

		if !emailRegex.MatchString(email) || email == "" {
			// need some error state on this form
			return renderPage(w, "contact", map[string]any{
				"PageTitle":       "Contact us | Invalid Email",
				"MetaDescription": "Please provide a  valid email address",
				"Cart":            nil,
			})
		}

		email = template.HTMLEscapeString(email)
		name = template.HTMLEscapeString(name)
		message = template.HTMLEscapeString(message)

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		q := "INSERT INTO contact (name, email, message, timestamp) VALUES (?, ?, ?, ?)"
		_, err = db.ExecContext(ctx, q, email, name, message, time.Now())
		if err != nil {
			return err
		}
		cart, err := ctxCart(r)
		if err != nil {
			return err
		}
		data := map[string]any{
			"PageTitle":       "Contact BabyGate Builders",
			"MetaDescription": "Contact form for Babygate builders",
			"Cart":            cart,
		}
		return renderPage(w, "contact", data)
	})

	/*
		Build enpoint. Currently only handling build for pressure gates
	*/
	handle.post("/build/", func(w http.ResponseWriter, r *http.Request) error {
		if err := r.ParseForm(); err != nil {
			return err
		}

		desiredWidth, err := strconv.ParseFloat(r.Form.Get("desired-width"), 32)
		if err != nil {
			return err
		}

		// keep track of user inputs(valid ones). From there we can generate "popular bundles"
		if err := SaveRequestedBundleSize(db, float32(desiredWidth)); err != nil {
			return err
		}

		bundles, err := BuildPressureFitBundles(db, float32(desiredWidth))
		if err != nil {
			return err
		}

		data := map[string]any{
			"RequestedBundleSize": float32(desiredWidth),
			"Bundles":             bundles,
		}

		return renderPage(w, "build-results", data)
	})

	/*
		Product page endpoints
	*/
	//handle.get("/bundles/", pageHandler.Bundles)
	//handle.get("/bundles/new", pageHandler.Bundles)
	handle.get("/gates/", func(w http.ResponseWriter, r *http.Request) error {
		cart, err := ctxCart(r)
		if err != nil {
			return err
		}

		gates, err := GetGates(db, productCache, ProductFilterParams{})
		if err != nil {
			return err
		}

		data := map[string]any{
			"Heading":         "Shop All Gates",
			"PageTitle":       "Shop All Gates",
			"MetaDescription": "Shop our full range of gates",
			"Products":        gates,
			"Cart":            cart,
		}

		return renderPage(w, "products", data)
	})

	handle.get("/gates/{gate_id}", func(w http.ResponseWriter, r *http.Request) error {
		cart, err := ctxCart(r)
		if err != nil {
			return err
		}

		gateID, err := strconv.Atoi(r.PathValue("gate_id"))
		if err != nil {
			return err
		}

		gate, err := GetProductByID(db, gateID)
		if err != nil {
			return err
		}

		data := map[string]any{
			"PageTitle":       gate.Name,
			"MetaDescription": gate.Name,
			"Product":         gate,
			"Cart":            cart,
		}

		return renderPage(w, "product", data)

	})

	handle.get("/extensions/", func(w http.ResponseWriter, r *http.Request) error {
		cart, err := ctxCart(r)
		if err != nil {
			return err
		}
		extensions, err := GetExtensions(db, productCache, ProductFilterParams{})
		if err != nil {
			return err
		}

		data := map[string]any{
			"Heading":         "All extensions",
			"PageTitle":       "All extensions",
			"MetaDescription": "Shop all extensions",
			"Products":        extensions,
			"Cart":            cart,
		}

		return renderPage(w, "products", data)
	})

	handle.get("/extensions/{extension_id}", func(w http.ResponseWriter, r *http.Request) error {
		cart, err := ctxCart(r)
		if err != nil {
			return err
		}

		extensionID, err := strconv.Atoi(r.PathValue("extension_id"))
		if err != nil {
			return err
		}

		extension, err := GetProductByID(db, extensionID)
		if err != nil {
			return err
		}

		data := map[string]any{
			"PageTitle":       extension.Name,
			"MetaDescription": extension.Name,
			"Cart":            cart,
		}

		return renderPage(w, "products", data)
	})

	/*
		cart endpoints
	*/
	handle.get("/cart/", func(w http.ResponseWriter, r *http.Request) error {
		cart, err := ctxCart(r)
		if err != nil {
			return err
		}

		data := map[string]any{
			"PageTitle":       "Your shopping cart",
			"MetaDescription": "",
			"Cart":            cart,
		}

		return renderPage(w, "cart", data)
	})

	handle.post("/cart/add", func(w http.ResponseWriter, r *http.Request) error {
		cart, err := ctxCart(r)
		if err != nil {
			return err
		}

		if err := r.ParseForm(); err != nil {
			return err
		}

		if (len(r.Form["data"])) < 1 {
			return renderPartial(w, "cart-modal", cart)
		}

		components := []models.CartItemComponent{}

		for _, d := range r.Form["data"] {
			component := models.NewCartItemComponent(cart.ID)
			if err := json.Unmarshal([]byte(d), &component); err != nil {
				return err
			}
			components = append(components, component)
		}

		if err := AddItemToCart(db, cart.ID, models.NewCartItem(cart.ID, components)); err != nil {
			return err
		}

		cart, err = GetCartByID(db, cart.ID)
		if err != nil {
			return err
		}

		return renderPartial(w, "cart-modal", cart)
	})

	handle.post("/cart/item/remove", func(w http.ResponseWriter, r *http.Request) error {
		if err := r.ParseForm(); err != nil {
			return err
		}

		cart, err := ctxCart(r)
		if err != nil {
			return err
		}

		cartItemID := r.Form.Get("cart_item_id")
		if cartItemID == "" {
			return nil
		}

		if _, err := db.Exec(`DELETE FROM cart_item WHERE id = ? AND cart_id = ?`, cartItemID, cart.ID); err != nil {
			return err
		}

		return nil
	})

	handle.post("/cart/item/{mode}", func(w http.ResponseWriter, r *http.Request) error {

		mode := r.PathValue("mode")

		if mode != "increment" && mode != "decrement" {
			return nil
		}

		if err := r.ParseForm(); err != nil {
			return err
		}

		cart, err := ctxCart(r)
		if err != nil {
			return err
		}

		cartItemID := r.Form.Get("cart_item_id")
		if cartItemID == "" {
			return fmt.Errorf("cartItemID is blank")
		}

		incrementOrDecremement := func(item *models.CartItem) error {
			if mode == "increment" {
				item.Qty++
				if _, err := db.Exec(`UPDATE 
					cart_item 
				SET qty = qty + 1 
					WHERE id = ? 
				AND 
					cart_id = ?`, item.ID, cart.ID); err != nil {
					return err
				}
			} else {
				if item.Qty < 2 {
					return nil
				}
				item.Qty--
				if _, err := db.Exec(`UPDATE 
					cart_item 
				SET qty = qty - 1 
					WHERE id = ? 
				AND 
					cart_id = ?`, item.ID, cart.ID); err != nil {
					return err
				}
			}
			return nil
		}

		for i := range cart.Items {
			if cart.Items[i].ID == cartItemID {
				err := incrementOrDecremement(&cart.Items[i])
				if err != nil {
					return err
				}
				return renderPartial(w, "item-details", cart.Items[i])
			}
		}

		return nil
	})

	handle.delete("/cart/item/", func(w http.ResponseWriter, r *http.Request) error {
		cart, err := ctxCart(r)
		if err != nil {
			return err
		}

		if err := r.ParseForm(); err != nil {
			return err
		}

		cartItemID := r.Form.Get("id")
		if cartItemID == "" {
			return fmt.Errorf("no cart item id supplied")
		}

		if _, err := db.Exec("DELETE FROM cart_item WHERE id = ?", cartItemID); err != nil {
			return err
		}

		cart, err = GetCartByID(db, cart.ID)
		if err != nil {
			return err
		}

		return renderPartial(w, "cart-main", cart)
	})

	handle.post("/cart/clear", func(w http.ResponseWriter, r *http.Request) error {
		cart, err := ctxCart(r)
		if err != nil {
			return err
		}

		_, err = db.Exec("DELETE FROM cart_item WHERE cart_id = ?", cart.ID)
		if err != nil {
			return fmt.Errorf("could not delete cart_item where cart_id = %s: %w", cart.ID, err)
		}

		_, err = db.Exec("DELETE FROM cart_item_component WHERE cart_id = ?", cart.ID)
		if err != nil {
			return fmt.Errorf("could not delete  cart_item_component where cart_id = %s: %w", cart.ID, err)
		}

		cart, err = GetCartByID(db, cart.ID)
		if err != nil {
			return err
		}

		return renderPartial(w, "cart-main", cart)

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

func sendErr(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func createCustomHandler(environment Environment, router *http.ServeMux, middleware []middlewareFunc) customHandler {
	return func(path string, fn customHandleFunc) {

		// wrap fn in middleware funcs
		for i := range middleware {
			fn = middleware[i](fn)
		}

		router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			if environment == Development {
				rMsg := "[INFO] %s request made to %s"
				log.Printf(rMsg, r.Method, r.URL.Path)
			}

			// custom handler get passed throgh the cart handler middleware first to
			// ensure there is a cart session
			if err := fn(w, r); err != nil {
				errMsg := "[ERROR] Failed  %s request to %s. %v"
				log.Printf(errMsg, r.Method, path, err)

				// ideally I would get notified of an error here
				if environment == Development {
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

func (handle customHandler) put(path string, fn customHandleFunc) {
	handle("PUT "+path, fn)
}
func (handle customHandler) delete(path string, fn customHandleFunc) {
	handle("DELETE"+path, fn)
}

func ctxCart(r *http.Request) (*models.Cart, error) {
	cart, ok := r.Context().Value(cartKey).(*models.Cart)
	if !ok {
		return nil, errors.New("could not get cart from request context")
	}
	return cart, nil
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

func NotFoundPage(w http.ResponseWriter, renderPage renderPageFunc) error {
	w.WriteHeader(http.StatusNotFound)
	data := map[string]any{
		"pageTile":        "Page Not Found",
		"MetaDescription": "Unable to find page",
	}
	return renderPage(w, "not-found", data)
}

/* service funcs start */

func BuildPressureFitBundles(db *sql.DB, limit float32) ([]models.Bundle, error) {
	var bundles []models.Bundle

	gates, err := GetProducts(db, Gate, ProductFilterParams{MaxWidth: limit})
	if err != nil {
		return bundles, err
	}
	if len(gates) < 1 {
		return bundles, nil
	}

	for _, gate := range gates {
		compatibleExtensions, err := GetCompatibleExtensions(db, gate.Id)
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

	var bundle models.Bundle = models.Bundle{}
	// returning a single bundle
	bundle.Qty = 1

	//  add gate to the bundle. Ensure Qty is at least 1
	if gate.Width > widthLimit {
		return bundle, errors.New("gate too big")
	}

	if gate.Qty < 1 {
		gate.Qty = 1
	}

	bundle.Gates = append(bundle.Gates, *gate)

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
		for ii := 0; ii < len(bundle.Extensions); ii++ {
			var bundleExtension *models.Product = &bundle.Extensions[ii]

			if bundleExtension.Id == extension.Id {
				existingExtension = bundleExtension
			}
		}

		if existingExtension != nil {
			existingExtension.Qty++
			widthLimit -= existingExtension.Width
		} else {
			extension.Qty = 1
			bundle.Extensions = append(bundle.Extensions, *extension)
			widthLimit -= extension.Width
		}
	}

	bundle.ComputeMetaData()
	return bundle, nil
}

func SaveBundle(db *sql.DB, bundle models.Bundle) (int64, error) {
	bundleId, err := SaveBundleAsProduct(db, bundle.Product)
	if err != nil {
		return 0, err
	}
	for iii := 0; iii < len(bundle.Gates); iii++ {
		gate := bundle.Gates[iii]
		err = SaveBundleGate(db, gate.Id, bundleId, gate.Qty)
		if err != nil {
			return 0, err
		}
	}
	for ii := 0; ii < len(bundle.Extensions); ii++ {
		extension := bundle.Extensions[ii]
		err := SaveBundleExtension(db, extension.Id, bundleId, extension.Qty)
		if err != nil {
			return 0, err
		}
	}
	return bundleId, nil
}

func TotalValue(db *sql.DB, cart models.Cart) (float64, error) {
	value := 0.0
	for _, item := range cart.Items {
		for _, component := range item.Components {
			productPrice, err := GetProductPrice(db, component.ProductID)
			if err != nil {
				return 0, err
			}
			value += (productPrice * float64(component.Qty))
		}
	}
	return value, nil
}

func NewCart(db *sql.DB) (*models.Cart, error) {
	cart := models.NewCart()
	if _, err := SaveCart(db, cart); err != nil {
		return nil, err
	}
	return &cart, nil
}

func AddItemToCart(db *sql.DB, cartID string, cartItem models.CartItem) error {
	if err := InsertOrIncrementCartItem(db, cartItem); err != nil {
		return fmt.Errorf("adding item to cart failed at insert or increment cart item: %w", err)
	}
	if err := SaveCartItemComponents(db, cartItem.Components); err != nil {
		return fmt.Errorf("adding item components failed: %w", err)
	}
	if _, err := db.Exec("UPDATE cart SET last_updated_at = ? WHERE id = ?", time.Now(), cartID); err != nil {
		return fmt.Errorf("could not update last updated at on cart: %w", err)
	}
	return nil
}

func RemoveItem(db *sql.DB, cartID, itemID string) error {
	if err := RemoveCartItem(db, cartID, itemID); err != nil {
		return fmt.Errorf("failed to remove item. %w", err)
	}
	if err := RemoveCartItemComponents(db, itemID); err != nil {
		return fmt.Errorf("failed to remove item components. %w", err)
	}
	return nil
}

type createProductParams struct {
	Type      string
	Name      string
	Width     float32
	Price     float32
	Img       string
	Tolerance float32
	Color     string
}

func CreateProduct(db *sql.DB, params createProductParams) (int64, error) {
	validProductTypes := [2]ProductType{
		Gate,
		Extension,
	}
	// Validate input parameters
	if params.Name == "" || params.Type == "" || params.Color == "" {
		return 0, errors.New("name, type, and color are required")
	}

	hasValidType := false
	for _, validProductType := range validProductTypes {
		if params.Type == string(validProductType) {
			hasValidType = true
		}
	}

	if !hasValidType {
		return 0, errors.New("does not have a valid product type")
	}

	if params.Price == 0.0 || params.Width == 0.0 {
		return 0, errors.New("price and width must be greater than 0")
	}

	existingProduct, err := GetProductByName(db, params.Name)
	if err != nil {
		return 0, err
	}

	if existingProduct != nil {
		return 0, errors.New("product already exists")
	}

	product := &models.Product{
		Id:        0,
		Type:      params.Type,
		Name:      params.Name,
		Width:     params.Width,
		Price:     params.Price,
		Img:       params.Img,
		Color:     params.Color,
		Tolerance: params.Tolerance,
	}

	row, err := InsertProduct(db, product)
	if err != nil {
		return 0, err
	}

	return row.LastInsertId()
}

func GetProductByID(db *sql.DB, productID int) (*models.Product, error) {
	return scanProductFromRow(
		db.QueryRow("SELECT id, type, name, width, price, img, color, tolerance FROM products WHERE id = ?", productID),
	)
}

func GetGates(db *sql.DB, productCache *cache.Cache, params ProductFilterParams) ([]*models.Product, error) {
	cacheString := fmt.Sprintf("gates;max-width:%f;limit:%d;", params.MaxWidth, params.Limit)
	if productCache != nil {
		cachedGates, found := productCache.Get(cacheString)
		if found {
			return cachedGates.([]*models.Product), nil
		}
	}

	gates, err := GetProducts(db, Gate, params)
	if err != nil {
		return nil, err
	}

	for i := range gates {
		gates[i].Qty = 1
	}

	if productCache != nil {
		productCache.Set(cacheString, gates, time.Minute*5)
	}
	return gates, nil
}

func GetExtensions(db *sql.DB, productCache *cache.Cache, params ProductFilterParams) ([]*models.Product, error) {
	cacheString := fmt.Sprintf("extensions;max-width:%f;limit:%d;", params.MaxWidth, params.Limit)

	cachedExtensions, found := productCache.Get(cacheString)
	if found {
		return cachedExtensions.([]*models.Product), nil
	}

	extensions, err := GetProducts(db, Extension, params)
	if err != nil {
		return nil, err
	}

	for i := range extensions {
		extensions[i].Qty = 1
	}

	productCache.Set(cacheString, extensions, time.Minute*5)

	return extensions, nil
}

func GetBundles(db *sql.DB, productCache *cache.Cache, params ProductFilterParams) ([]*models.Product, error) {
	cacheString := fmt.Sprintf("bundles;max-width:%f;limit:%d;", params.MaxWidth, params.Limit)

	cachedBundles, found := productCache.Get(cacheString)
	if found {
		return cachedBundles.([]*models.Product), nil
	}

	bundles, err := GetProducts(db, Bundle, params)
	if err != nil {
		return nil, err
	}

	productCache.Set(cacheString, bundles, time.Minute*5)

	return bundles, nil
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

func SaveBundleGate(db *sql.DB, gate_id int, bundle_id int64, qty int) error {
	_, err := db.Exec("INSERT INTO bundle_gates(gate_id, bundle_id, qty) VALUES (?, ?, ?)", gate_id, bundle_id, qty)
	if err != nil {
		return err
	}
	return nil
}

func SaveBundleExtension(db *sql.DB, extension_id int, bundle_id int64, qty int) error {
	_, err := db.Exec(`INSERT INTO 
		bundle_extensions(
			extension_id, 
			bundle_id, 
			qty
			) 
		VALUES 
			(?,?,?)`,
		extension_id,
		bundle_id,
		qty,
	)
	if err != nil {
		return err
	}
	return nil
}

func SaveBundleAsProduct(db *sql.DB, bundleProductValues models.Product) (int64, error) {
	result, err := db.Exec(
		`INSERT INTO 
			products(
				type, 
				name, 
				width, 
				price, 
				img, 
				tolerance, 
				color
			) 
		VALUES 
			(?, ?, ?, ?, ?, ?, ?)`,
		"bundle",
		bundleProductValues.Name,
		bundleProductValues.Width,
		bundleProductValues.Price,
		bundleProductValues.Img,
		bundleProductValues.Tolerance,
		bundleProductValues.Color,
	)
	if err != nil {
		return 0, err
	}
	lastInsertId, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return lastInsertId, nil
}

func SaveCart(db *sql.DB, cart models.Cart) (*sql.Result, error) {
	res, err := db.Exec(`INSERT INTO 
		cart(
			id,
			total_value,
			created_at, 
			last_updated_at) 
		VALUES 
			(?, ?, ?, ?)`,
		cart.ID,
		cart.TotalValue,
		cart.CreatedAt,
		cart.LastUpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to save cart. %v", err)
	}
	return &res, nil
}

func CartExists(db *sql.DB, id string) (bool, error) {
	var count int
	err := db.QueryRow(`SELECT count(id) AS count FROM cart WHERE id = ?`, id).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func GetCartByID(db *sql.DB, id string) (*models.Cart, error) {
	cart, err := selectCart(db, id)
	if err != nil {
		return nil, err
	}
	if cart.Items, err = selectCartItems(db, cart.ID); err != nil {
		return nil, err
	}
	for i := range cart.Items {
		if cart.Items[i].Components, err = selectCartItemComponents(
			db,
			cart.ID,
			cart.Items[i].ID,
		); err != nil {
			return nil, err
		}
	}
	return &cart, nil
}

func selectCart(db *sql.DB, id string) (models.Cart, error) {
	row := db.QueryRow(`
	SELECT 
		id, 
		created_at, 
		last_updated_at, 
		total_value 
	FROM 
		cart 
	WHERE 
		id = ?`, id)
	var cart models.Cart
	if err := row.Scan(
		&cart.ID,
		&cart.CreatedAt,
		&cart.LastUpdatedAt,
		&cart.TotalValue,
	); err != nil {
		return models.Cart{}, err
	}
	return cart, nil
}

func selectCartItems(db *sql.DB, cartID string) ([]models.CartItem, error) {
	rows, err := db.Query(`
	SELECT 
		id, 
		cart_id, 
		name, 
		sale_price, 
		qty, 
		created_at 
	FROM 
		cart_item 
	WHERE 
		cart_id = ?`, cartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cartItems := []models.CartItem{}
	for rows.Next() {
		var cartItem models.CartItem
		if err := rows.Scan(
			&cartItem.ID,
			&cartItem.CartID,
			&cartItem.Name,
			&cartItem.SalePrice,
			&cartItem.Qty,
			&cartItem.CreatedAt,
		); err != nil {
			return nil, err
		}
		cartItems = append(cartItems, cartItem)
	}
	return cartItems, nil
}

func selectCartItemComponents(db *sql.DB, cartID, cartItemID string) ([]models.CartItemComponent, error) {
	rows, err := db.Query(`
	SELECT 
		cart_item_id, 
		cart_id, 
		product_id, 
		qty, 
		name, 
		created_at 
	FROM 
		cart_item_component 
	WHERE 
		cart_item_id = ? 
	AND 
		cart_id = ?`,
		cartItemID, cartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	components := []models.CartItemComponent{}
	for rows.Next() {
		var component models.CartItemComponent
		if err := rows.Scan(
			&component.CartItemID,
			&component.CartID,
			&component.ProductID,
			&component.Qty,
			&component.Name,
			&component.CreatedAt,
		); err != nil {
			return nil, err
		}
		components = append(components, component)
	}
	return components, nil
}

func InsertOrIncrementCartItem(db *sql.DB, cartItem models.CartItem) error {
	var count int
	err := db.QueryRow(`
	SELECT 
		count(id) as count 
	FROM 
		cart_item 
	WHERE 
		id = ? 
	AND 
		cart_id = ?`,
		cartItem.ID, cartItem.CartID).Scan(&count)
	if err != nil {
		return fmt.Errorf("could not count cart_item %w", err)
	}

	if count > 0 {
		_, err := db.Exec(`
		UPDATE 
			cart_item 
		SET 
			qty = qty + 1 
		WHERE 
			id = ? 
		AND 
			cart_id = ?`,
			cartItem.ID, cartItem.CartID)
		if err != nil {
			return fmt.Errorf("could not increment cart item: %w", err)
		}
		return nil
	}

	q := `
	INSERT INTO 
		cart_item (
			id, 
			cart_id, 
			name, 
			sale_price, 
			qty, 
			created_at
		) 
	VALUES 
		(?, ?, ?, ?, ?, ?)`
	_, err = db.Exec(
		q,
		cartItem.ID,
		cartItem.CartID,
		cartItem.Name,
		cartItem.SalePrice,
		cartItem.Qty,
		cartItem.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("could not insert cart item in db %w", err)
	}
	return nil
}

func GetCartByUserID(db *sql.DB, userID int) (*models.Cart, error) {
	var cart models.Cart
	err := db.QueryRow(`
		SELECT 
			id, created_at, last_updated_at 
		FROM 
			carts 
		WHERE 
			user_id = ?`,
		userID,
	).Scan(
		&cart.ID,
		&cart.CreatedAt,
		&cart.LastUpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &cart, nil
}

func GetCartItemsByCartID(db *sql.DB, cartID string) ([]*models.CartItem, error) {
	return nil, errors.New("getting cart items not implemented")
}

func SaveCartItemComponents(db *sql.DB, components []models.CartItemComponent) error {
	for _, c := range components {
		q := `
		INSERT INTO 
			cart_item_component (
				cart_item_id, 
				cart_id, 
				product_id, 
				qty, 
				name, 
				created_at
			) 
		VALUES 
			(?, ?, ?, ?, ?, ?)`
		if _, err := db.Exec(q,
			c.CartItemID,
			c.CartID,
			c.ProductID,
			c.Qty,
			c.Name,
			c.CreatedAt,
		); err != nil {
			return err
		}
	}
	return nil
}

func RemoveCartItem(db *sql.DB, cartID, itemID string) error {
	return errors.New("remove cart item not implemented")
}

func RemoveCartItemComponents(db *sql.DB, itemID string) error {
	return errors.New("remove Cart Item Components not yet implemented")
}

// Define a custom type for the product
type ProductType string

// Define constants representing the product values
const (
	Gate      ProductType = "gate"
	Extension ProductType = "extension"
	Bundle    ProductType = "bundle"
)

type scannable interface {
	Scan(dest ...any) error
}

func scanProductFromRow(row scannable) (*models.Product, error) {
	var product models.Product
	err := row.Scan(
		&product.Id,
		&product.Type,
		&product.Name,
		&product.Width,
		&product.Price,
		&product.Img,
		&product.Color,
		&product.Tolerance,
	)
	return &product, err
}

func InsertProduct(db *sql.DB, product *models.Product) (sql.Result, error) {
	return db.Exec(
		`INSERT INTO products (type, name, width, price, img, color, tolerance) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		product.Type,
		product.Name,
		product.Width,
		product.Price,
		product.Img,
		product.Color,
		product.Tolerance,
	)

}

func GetProductPrice(db *sql.DB, id int) (float64, error) {
	var price float64
	if err := db.QueryRow("SELECT price FROM products WHERE id = ?", id).Scan(&price); err != nil {
		return 0, err
	}
	return price, nil
}

func GetProductByName(db *sql.DB, name string) (*models.Product, error) {
	if db == nil {
		return nil, errors.New("need a non nil db to get products by name")
	}
	return scanProductFromRow(
		db.QueryRow("SELECT id, type, name, width, price, img, color, tolerance FROM products WHERE name = ?", name),
	)
}

type ProductFilterParams struct {
	MaxWidth float32
	Limit    int
}

func GetProducts(db *sql.DB, productType ProductType, params ProductFilterParams) ([]*models.Product, error) {
	if db == nil {
		return nil, errors.New("get products requires a non nil db pointer")
	}

	filters := []any{productType}

	baseQuery := "SELECT id, type, name, width, price,  img, color, tolerance FROM products WHERE type = ?"
	if params.MaxWidth > 0 {
		baseQuery = baseQuery + " AND width < ?"
		filters = append(filters, params.MaxWidth)
	}

	if params.Limit > 0 {
		baseQuery += " LIMIT ?"
		filters = append(filters, params.Limit)
	}

	var products []*models.Product
	rows, err := db.Query(baseQuery, filters...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		product, err := scanProductFromRow(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, nil
}

func GetCompatibleExtensions(db *sql.DB, gateID int) ([]*models.Product, error) {
	var extensions []*models.Product
	rows, err := db.Query(
		"SELECT p.id, p.type, p.name, p.width, p.price, p.img, p.color, p.tolerance FROM products p INNER JOIN compatibles c ON p.id = c.extension_id WHERE gate_id = ?",
		gateID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		extension, err := scanProductFromRow(rows)
		if err != nil {
			return nil, err
		}
		extensions = append(extensions, extension)
	}
	return extensions, nil
}

func Update(db *sql.DB, product *models.Product) error {
	// Code to update an existing user in the database
	// using the provided SQL database connection (r.db)
	return errors.New("update product not implemented")
}

func Delete(db *sql.DB, productID int) error {
	// Code to delete a user from the database
	// based on the provided user ID (userID)
	// using the provided SQL database connection (r.db)
	_, err := db.Exec("DELETE FROM products WHERE id = ?", productID)
	return err
}

/* repository funcs end */

func SqliteOpen(path string) *sql.DB {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		panic(err)
	}

	return db
}

type renderPageFunc func(w http.ResponseWriter, templateName string, templateData map[string]any) error
type renderPartialFunc func(w http.ResponseWriter, templateName string, templateData any) error

func NewPageRenderer(tmpl tmplFunc) renderPageFunc {
	return func(w http.ResponseWriter, templateName string, templateData map[string]any) error {

		data := map[string]any{
			"MetaDescription": "default meta description",
			"PageTitle":       "default page title",
		}

		for k, v := range templateData {
			data[k] = v
		}

		var buffer bytes.Buffer
		if err := tmpl().ExecuteTemplate(&buffer, templateName, data); err != nil {
			return fmt.Errorf("problem executing template %s: %w", templateName, err)
		}

		if _, err := w.Write(buffer.Bytes()); err != nil {
			return err
		}

		return nil
	}
}

func NewPartialRenderer(tmpl tmplFunc) renderPartialFunc {
	return func(w http.ResponseWriter, templateName string, templateData any) error {

		var buffer bytes.Buffer
		if err := tmpl().ExecuteTemplate(&buffer, templateName, templateData); err != nil {
			return fmt.Errorf("problem executing partial template %s: %w", templateName, err)
		}

		if _, err := w.Write(buffer.Bytes()); err != nil {
			return err
		}

		return nil
	}
}

type tmplFunc func() *template.Template

func templateParser(env Environment) tmplFunc {
	var tmpl *template.Template
	return func() *template.Template {
		if env == Production && tmpl != nil {
			return tmpl
		}
		tmpl = template.Must(template.New("").Funcs(template.FuncMap{
			"bundleJSON": func(bundle models.Bundle) string {
				data := []struct {
					ProductID int `json:"product_id"`
					Qty       int `json:"qty"`
				}{}
				for _, g := range bundle.Gates {
					data = append(data, struct {
						ProductID int `json:"product_id"`
						Qty       int `json:"qty"`
					}{g.Id, g.Qty})
				}
				for _, g := range bundle.Extensions {
					data = append(data, struct {
						ProductID int `json:"product_id"`
						Qty       int `json:"qty"`
					}{g.Id, g.Qty})
				}
				bytes, err := json.Marshal(data)
				if err != nil {
					return ""
				}
				return string(bytes)
			},
			"sizeRange": func(width, tolerance float32) float32 {
				return width - tolerance
			},
			"title": func(str string) string {
				return cases.Title(language.AmericanEnglish).String(str)
			},
		}).ParseGlob("templates/**/*.tmpl"))
		return tmpl
	}
}

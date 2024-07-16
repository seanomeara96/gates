package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
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

type templateData struct {
	Env Environment
}

func NewTemplateData() templateData {
	return templateData{
		Env: Development,
	}
}

type BasePageData struct {
	templateData
	PageTitle       string
	MetaDescription string
}

func NewBasePageData(pageTitle string, metaDescription string) BasePageData {
	templateData := NewTemplateData()
	return BasePageData{
		templateData:    templateData,
		PageTitle:       pageTitle,
		MetaDescription: metaDescription,
	}
}

type customHandleFunc func(w http.ResponseWriter, r *http.Request) error
type middlewareFn func(w http.ResponseWriter, r *http.Request) (execNextFunc bool, err error)
type middlewaresFunc func(w http.ResponseWriter, r *http.Request, fn customHandleFunc) error
type customHandler func(path string, fn customHandleFunc)

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

	tmpl, err := template.New("").Funcs(template.FuncMap{
		"sizeRange": func(width, tolerance float32) float32 {
			return width - tolerance
		},
		"title": func(str string) string {
			return cases.Title(language.AmericanEnglish).String(str)
		},
	}).ParseGlob("templates/**/*.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	renderPage := NewPageRenderer(tmpl)

	// ROUTING LOGIC
	// TODO put back in the cart middleware
	middlewareFuncs := []middlewareFn{}

	/*
		Call the middlewares func for each request
		Would be better if we could also do a per route basis
	*/
	middlewares := registerMiddlewares(middlewareFuncs)

	handle := createCustomHandler(environment, router, middlewares)

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

			data := map[string]any{
				"pageTitle":       "Home Page",
				"metaDescription": "Welcome to the home page",
				"FeaturedGates":   featuredGates,
				"PopularBundles":  popularBundles,
			}

			return renderPage(w, "home", data)
		}

		return NotFoundPage(w, renderPage)
	}) // cant use 'get' because it causes conflicts

	handle.get("/contact/", func(w http.ResponseWriter, r *http.Request) error {
		data := map[string]any{
			"pageTitle":       "Contact BabyGate Builders",
			"metaDescription": "Contact form for Babygate builders",
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

		// TODO keep track of user inputs(valid ones). From there we can generate "popular bundles"
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
		gates, err := GetGates(db, productCache, ProductFilterParams{})
		if err != nil {
			return err
		}

		data := map[string]any{
			"Heading":         "Shop All Gates",
			"pageTitle":       "Shop All Gates",
			"metaDescription": "Shop our full range of gates",
			"Products":        gates,
		}

		return renderPage(w, "products", data)
	})

	handle.get("/gates/{gate_id}", func(w http.ResponseWriter, r *http.Request) error {
		gateID, err := strconv.Atoi(r.PathValue("gate_id"))
		if err != nil {
			return err
		}

		gate, err := GetProductByID(db, gateID)
		if err != nil {
			return err
		}

		data := map[string]any{
			"pageTitle":       gate.Name,
			"metaDescription": gate.Name,
			"Product":         gate,
		}

		return renderPage(w, "product", data)

	})

	handle.get("/extensions/", func(w http.ResponseWriter, r *http.Request) error {
		extensions, err := GetExtensions(db, productCache, ProductFilterParams{})
		if err != nil {
			return err
		}

		data := map[string]any{
			"Heading":         "All extensions",
			"pageTitle":       "All extensions",
			"metaDescription": "Shop all extensions",
			"Products":        extensions,
		}

		return renderPage(w, "products", data)
	})

	handle.get("/extensions/{extension_id}", func(w http.ResponseWriter, r *http.Request) error {
		extensionID, err := strconv.Atoi(r.PathValue("extension_id"))
		if err != nil {
			return err
		}

		extension, err := GetProductByID(db, extensionID)
		if err != nil {
			return err
		}

		data := map[string]any{
			"pageTitle":       extension.Name,
			"metaDescription": extension.Name,
		}

		return renderPage(w, "products", data)
	})

	/*
		cart endpoints
	*/
	handle.get("/cart/", func(w http.ResponseWriter, r *http.Request) error {
		cart := models.Cart{}

		data := map[string]any{
			"pageTitle":       "Your shopping cart",
			"metaDescription": "",
			"Cart":            cart,
		}

		return renderPage(w, "cart", data)
	})

	handle.post("/cart/add", func(w http.ResponseWriter, r *http.Request) error {
		session, err := getCartSession(r, store)
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

		if err := AddItemToCart(db, cartID.(string), components); err != nil {
			return err
		}

		return nil
	}) // TODO consolidate add & update methods

	handle.post("/cart/remove", func(w http.ResponseWriter, r *http.Request) error {
		session, err := getCartSession(r, store)
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

		if err := r.ParseForm(); err != nil {
			return err
		}

		itemID := r.Form.Get("item_id")

		if err := RemoveItem(db, cartID.(string), itemID); err != nil {
			return err
		}

		return nil
	})

	handle.post("/cart/clear", func(w http.ResponseWriter, r *http.Request) error {
		return fmt.Errorf("clear cart not implemented")
	})

	handle.get("/test/{int}", func(w http.ResponseWriter, r *http.Request) error {
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

func createCustomHandler(environment Environment, router *http.ServeMux, executeMiddlewares middlewaresFunc) customHandler {
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

/*func (handle customHandler) put(path string, fn customHandleFunc) {
	handle("PUT "+path, fn)
}
func (handle customHandler) delete(path string, fn customHandleFunc) {
	handle("DELETE"+path, fn)
}*/

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

func CartMiddleWare(db *sql.DB, store *sessions.CookieStore, w http.ResponseWriter, r *http.Request) (bool, error) {
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

	cartID, err = NewCart(db)
	if err != nil {
		return false, err
	}

	session.Values["cart_id"] = cartID
	if err := session.Save(r, w); err != nil {
		return false, err
	}

	return true, nil
}

func NotFoundPage(w http.ResponseWriter, renderPage renderPageFunc) error {
	w.WriteHeader(http.StatusNotFound)
	data := map[string]any{
		"pageTile":        "Page Not Found",
		"metaDescription": "Unable to find page",
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

func NewCart(db *sql.DB) (string, error) {
	cart := models.NewCart()
	if _, err := SaveCart(db, cart); err != nil {
		return "", err
	}
	return cart.ID, nil
}

func GetCart(db *sql.DB, userID int) (*models.Cart, []*models.CartItem, error) {
	//maybe if there is not cartby that user id we should auto call new cart
	cart, err := GetCartByUserID(db, userID)
	if err != nil {
		return nil, nil, err
	}
	items, err := GetCartItemsByCartID(db, cart.ID)
	if err != nil {
		return nil, nil, err
	}
	return cart, items, nil
}

func AddItemToCart(db *sql.DB, cartID string, components []models.CartItemComponent) error {
	cartItem := models.NewCartItem(cartID)
	if err := SaveCartItem(db, cartItem); err != nil {
		return err
	}
	if err := SaveCartItemComponents(db, cartItem.ID, components); err != nil {
		return err
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

	cachedGates, found := productCache.Get(cacheString)
	if found {
		return cachedGates.([]*models.Product), nil
	}

	gates, err := GetProducts(db, Gate, params)
	if err != nil {
		return nil, err
	}

	productCache.Set(cacheString, gates, time.Minute*5)
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

func CreateBundleTables(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS bundle_gates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		gate_id INTEGER NOT NULL,
		bundle_id INTEGER NOT NULL,
		qty INTEGER NOT NULL,
		FOREIGN KEY (gate_id) REFERENCES products(id),
		FOREIGN KEY (bundle_id) REFERENCES products(id)
	)`)
	if err != nil {
		return err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS bundle_extensions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		extension_id INTEGER NOT NULL,
		bundle_id INTEGER NOT NULL,
		qty INTEGER NOT NULL,
		FOREIGN KEY (extension_id) REFERENCES products(id),
		FOREIGN KEY (bundle_id) REFERENCES products(id)
	)`)
	if err != nil {
		return err
	}
	return nil
}

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
	_, err := db.Exec("INSERT INTO bundle_extensions(extension_id, bundle_id, qty) VALUES (?,?,?)", extension_id, bundle_id, qty)
	if err != nil {
		return err
	}
	return nil
}

func SaveBundleAsProduct(db *sql.DB, bundleProductValues models.Product) (int64, error) {
	result, err := db.Exec(
		"INSERT INTO products(type, name, width, price, img, tolerance, color) VALUES (?, ?, ?, ?, ?, ?, ?)",
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

func CreateCartTables(db *sql.DB) (sql.Result, error) {
	res, err := db.Exec(`CREATE TABLE IF NOT EXISTS carts(
		id STRING PRIMARY KEY,
		created_at DATETIME,
		last_updated_at DATETIME
	)`)

	if err != nil {
		return res, err
	}

	res, err = db.Exec(`CREATE TABLE IF NOT EXISTS cart_items(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		cart_id STRING NOT NULL,
		product_id INTEGER NOT NULL,
		quantity INTEGER DEFAULT 1,
		created_at DATETIME,
		FOREIGN KEY (cart_id) REFERENCES carts(id),
		FOREIGN KEY(product_id) REFERENCES products(id)
	)`)

	return res, err
}

func SaveCart(db *sql.DB, cart models.Cart) (*sql.Result, error) {
	res, err := db.Exec(`INSERT INTO 
		carts(
			id, 
			created_at, 
			last_updated_at) 
		VALUES 
			(?, ?, ?)`,
		cart.ID,
		cart.CreatedAt,
		cart.LastUpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to save cart. %v", err)
	}
	return &res, nil
}

func SaveCartItem(db *sql.DB, cartItem models.CartItem) error {
	return errors.New("save cart item not implemented")
}

func GetCartByUserID(db *sql.DB, userID int) (*models.Cart, error) {
	var cart models.Cart
	err := db.QueryRow(`SELECT 
			id, created_at, last_updated_at 
		FROM carts 
		WHERE user_id = ?`,
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

func SaveCartItemComponents(db *sql.DB, cartID string, components []models.CartItemComponent) error {

	return errors.New("save Cart Item Components not yet implemented")

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
	return scanProductFromRow(
		db.QueryRow("SELECT id, type, name, width, price, img, color, tolerance FROM products WHERE name = ?", name),
	)
}

type ProductFilterParams struct {
	MaxWidth float32
	Limit    int
}

func GetProducts(db *sql.DB, productType ProductType, params ProductFilterParams) ([]*models.Product, error) {
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

	var gates []*models.Product
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
		gates = append(gates, product)
	}
	return gates, nil
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

func NewPageRenderer(tmpl *template.Template) renderPageFunc {
	return func(w http.ResponseWriter, templateName string, templateData map[string]any) error {

		data := map[string]any{
			"metaDescription": "default meta description",
			"pageTitle":       "default page title",
		}

		for k, v := range templateData {
			data[k] = v
		}

		var buffer bytes.Buffer
		if err := tmpl.ExecuteTemplate(&buffer, templateName, data); err != nil {
			return fmt.Errorf("problem executing template %s: %w", templateName, err)
		}

		if _, err := w.Write(buffer.Bytes()); err != nil {
			return err
		}

		return nil
	}
}

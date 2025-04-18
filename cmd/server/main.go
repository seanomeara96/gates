package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"sort"
	"strconv"
	"syscall"
	"text/template"
	"time"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/seanomeara96/auth"
	"github.com/seanomeara96/gates/cachedrepos"
	"github.com/seanomeara96/gates/models"
	"github.com/seanomeara96/gates/repos"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Environment string
type customHandleFunc func(cart *models.Cart, w http.ResponseWriter, r *http.Request) error
type middlewareFunc func(next customHandleFunc) customHandleFunc
type customHandler func(path string, fn customHandleFunc)

const (
	Development Environment = "development"
	Production  Environment = "production"
)

type config struct {
	Port                 string
	Mode                 Environment
	DBPath               string
	CookieStoreSecretKey string
	AdminUserID          string
	AdminUserPassword    string
	JWTSecretKey         string
}

func configEnv() (*config, error) {

	var config config
	var errs []error

	config.Port = os.Getenv("PORT")
	if config.Port == "" {
		errs = append(errs, errors.New("env PORT value not set in env"))
	}

	config.Mode = Environment(os.Getenv("MODE"))
	if config.Mode != Development && config.Mode != Production {
		errs = append(errs, errors.New("env MODE not set in env"))
	}

	config.DBPath = os.Getenv("DB_FILE_PATH")
	if config.DBPath == "" {
		config.DBPath = "main.db"
	}

	config.AdminUserID = os.Getenv("ADMIN_USER_ID")
	if config.AdminUserID == "" {
		errs = append(errs, errors.New("env ADMIN_USER_ID not set"))
	}

	config.AdminUserPassword = os.Getenv("ADMIN_USER_PASSWORD")
	if config.AdminUserPassword == "" {
		errs = append(errs, errors.New("env ADMIN_USER_PASSWORD not set"))
	}

	stripe.Key = os.Getenv("STRIPE_API_KEY")

	config.CookieStoreSecretKey = os.Getenv("COOKIE_SECRET")
	if config.CookieStoreSecretKey == "" {
		errs = append(errs, errors.New("env COOKIE_SECRET not set"))
	}

	config.JWTSecretKey = os.Getenv("JWT_SECRET_KEY")
	if config.JWTSecretKey == "" {
		errs = append(errs, errors.New("env JWT_SECRET_KEY not set"))
	}

	if len(errs) > 0 {
		var err = errors.New("env config err(s)")
		for _, e := range errs {
			err = fmt.Errorf("%w|%w", err, e)
		}
		return nil, err
	}

	return &config, nil
}

func configCookieStore(config *config) (*sessions.CookieStore, error) {

	if config.CookieStoreSecretKey == "" {
		if config.Mode == Development {
			config.CookieStoreSecretKey = "suprSecrtStoreKey"
		} else {
			return nil, errors.New("cookie secret not set in env")
		}
	}
	return sessions.NewCookieStore([]byte(config.CookieStoreSecretKey)), nil
}

func app() error {

	if err := godotenv.Load(); err != nil {
		return err
	}

	config, err := configEnv()
	if err != nil {
		return err
	}

	db := SqliteOpen(config.DBPath)
	defer db.Close()

	authConfig := auth.AuthConfig{DB: db, JWTSecretKey: config.JWTSecretKey}

	auth, err := auth.Init(authConfig)
	if err != nil {
		return err
	}

	auth.Register(context.Background(), config.AdminUserID, config.AdminUserPassword)

	productRepo := repos.NewProductRepo(db)
	productCache := cachedrepos.NewCachedProductRepo(productRepo)

	store, err := configCookieStore(config)
	if err != nil {
		return err
	}

	tmpl := templateParser(config.Mode)
	renderPage := NewPageRenderer(tmpl)
	renderPartial := NewPartialRenderer(tmpl)

	router := http.NewServeMux()
	getCartFromRequest := NewCartFromSessionGetter(db, productCache, store)
	// ROUTING LOGIC
	// middleware executed in reverse order; i = 0 executes last
	middleware := []middlewareFunc{
		// Example for middleware usage:
		func(next customHandleFunc) customHandleFunc {
			return func(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
				// Do something with the cart ...
				// like refresh cart details
				if cart != nil {
					if err := refreshCartDetails(cart, productCache); err != nil {
						return err
					}
				}
				return next(cart, w, r)
			}
		},
	}

	handle := createCustomHandler(config.Mode, router, middleware, getCartFromRequest)

	handle.get("/admin/login", func(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
		accessToken, refreshToken, _ := auth.GetTokensFromRequest(r)

		_, err = auth.ValidateToken(accessToken)
		if err == nil {
			accessToken, refreshToken, err = auth.Refresh(r.Context(), refreshToken)
			if err != nil {
				return err
			}
			auth.SetTokens(w, accessToken, refreshToken)
			http.Redirect(w, r, "/admin", http.StatusTemporaryRedirect)
			return nil
		}

		data := map[string]any{
			"PageTitle":       "Home Page",
			"MetaDescription": "Welcome to the home page",

			"Cart": cart,
			"Env":  config.Mode,
		}

		return renderPage(w, "admin-login", data)
	})

	handle.post("/admin/login", func(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
		if err := r.ParseForm(); err != nil {
			return err
		}

		accessToken, refreshToken, err := auth.Login(r.Context(), r.Form.Get("user_id"), r.Form.Get("password"))
		if err != nil {
			return err
		}
		auth.SetTokens(w, accessToken, refreshToken)

		http.Redirect(w, r, "/admin", http.StatusTemporaryRedirect)
		return nil
	})

	handle.get("/admin", func(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {

		accessToken, refreshToken, err := auth.GetTokensFromRequest(r)
		if err != nil {
			return err
		}

		_, err = auth.ValidateToken(accessToken)
		if err != nil {
			return err
		}

		accessToken, refreshToken, err = auth.Refresh(r.Context(), refreshToken)
		if err != nil {
			return err
		}
		auth.SetTokens(w, accessToken, refreshToken)

		data := map[string]any{
			"PageTitle":       "Home Page",
			"MetaDescription": "Welcome to the home page",

			"Cart": cart,
			"Env":  config.Mode,
		}

		return renderPage(w, "home", data)
	})

	handle("/", func(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
		if r.URL.Path == "/" {
			featuredGates, err := productCache.GetGates(repos.ProductFilterParams{})
			if err != nil {
				return fmt.Errorf("home page: failed to get featured gates: %w", err)
			}

			extensions, err := productCache.GetExtensions(repos.ProductFilterParams{Limit: 2})
			if err != nil {
				return fmt.Errorf("home page: failed to get featured extensions: %w", err)
			}

			data := map[string]any{
				"PageTitle":          "Home Page",
				"MetaDescription":    "Welcome to the home page",
				"FeaturedGates":      featuredGates,
				"FeaturedExtensions": extensions,
				"Cart":               cart,
				"Env":                config.Mode,
			}

			return renderPage(w, "home", data)
		}

		return NotFoundPage(w, renderPage)
	})

	handle.get("/contact/", func(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
		data := map[string]any{
			"PageTitle":       "Contact BabyGate Builders",
			"MetaDescription": "Contact form for Babygate builders",
			"Cart":            cart,
			"Env":             config.Mode,
		}
		return renderPage(w, "contact", data)
	})

	handle.get("/checkout/", func(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {

		// reset the prices in the cart object in case there has been some manipulation on the client side
		cart.TotalValue = 0
		for i := range cart.Items {
			cartItem := &cart.Items[i]
			cartItem.SalePrice = 0
			for ii := range cartItem.Components {
				component := &cartItem.Components[ii]
				count, err := productRepo.CountProductByID(component.Id)
				if err != nil {
					return err
				}
				insufficientStock := count < component.Qty
				if insufficientStock {
					return fmt.Errorf("insufficient stock of %d expected more than %d but only have  %d", component.Id, component.Qty, count)
				}
				price, err := productRepo.GetProductPrice(component.Id)
				if err != nil {
					return err
				}
				component.Price = price
				cartItem.SalePrice += ((component.Price) * float32(component.Qty))
			}
			cart.TotalValue += (cartItem.SalePrice * float32(cartItem.Qty))
		}

		if os.Getenv("STRIPE_API_KEY") == "" {
			if err := json.NewEncoder(w).Encode(cart); err != nil {
				return err
			}
			return nil
		}

		// continue validating cart
		lineItems := []*stripe.CheckoutSessionLineItemParams{}
		for _, item := range cart.Items {
			lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
				Quantity: stripe.Int64(int64(item.Qty)),
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					UnitAmount: stripe.Int64(int64(item.SalePrice * 100)),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(item.Name),
					},
					Currency: stripe.String("EUR"),
				},
			},
			)
		}

		domain := "http://localhost:3000"
		params := &stripe.CheckoutSessionParams{
			LineItems:  lineItems,
			Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
			SuccessURL: stripe.String(domain + "/success"),
			CancelURL:  stripe.String(domain + "/cart"),
			ShippingAddressCollection: &stripe.CheckoutSessionShippingAddressCollectionParams{
				AllowedCountries: []*string{stripe.String("IE")},
			},
			PhoneNumberCollection: &stripe.CheckoutSessionPhoneNumberCollectionParams{
				Enabled: stripe.Bool(true),
			},
		}
		s, err := session.New(params)
		if err != nil {
			log.Printf("session.New: %v", err)
		}

		http.Redirect(w, r, s.URL, http.StatusSeeOther)
		return nil
	})

	emailRegex, err := regexp.Compile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if err != nil {
		return fmt.Errorf("contact page: could not compile email validation regex: %w", err)
	}
	handle.post("/contact/", func(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
		if err := r.ParseForm(); err != nil {
			return fmt.Errorf("contact page: failed to parse form: %w", err)
		}
		email := r.Form.Get("email")
		name := r.Form.Get("name")
		message := r.Form.Get("message")

		if message == "" {
			return renderPage(w, "contact", map[string]any{
				"PageTitle":       "Contact us | No Message Provided",
				"MetaDescription": "Please provide a message",
				"Cart":            nil,
			})
		}

		if !emailRegex.MatchString(email) || email == "" {
			return renderPage(w, "contact", map[string]any{
				"PageTitle":       "Contact us | Invalid Email",
				"MetaDescription": "Please provide a valid email address",
				"Cart":            nil,
				"Env":             config.Mode,
			})
		}

		var contact struct {
			Email   string
			Name    string
			Message string
		}

		contact.Email = template.HTMLEscapeString(email)
		contact.Name = template.HTMLEscapeString(name)
		contact.Message = template.HTMLEscapeString(message)

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if err := InsertContactForm(ctx, db, contact); err != nil {
			return err
		}

		data := map[string]any{
			"PageTitle":       "Contact BabyGate Builders",
			"MetaDescription": "Contact form for Babygate builders",
			"Cart":            cart,
			"Env":             config.Mode,
		}
		return renderPage(w, "contact", data)
	})

	// Build endpoint. Currently only handling builds for pressure gates.
	handle.post("/build/", func(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
		if err := r.ParseForm(); err != nil {
			return fmt.Errorf("build endpoint: failed to parse form: %w", err)
		}

		desiredWidth, err := strconv.ParseFloat(r.Form.Get("desired-width"), 32)
		if err != nil {
			return fmt.Errorf("build endpoint: failed to parse desired width: %w", err)
		}

		/*
			static max witdth for now
			could consider adding a max width to the gates in the DB and enforcing the
			constraint from there
		*/
		maxWidth := 220.00
		if desiredWidth > maxWidth {
			desiredWidth = maxWidth
		}

		if err := SaveRequestedBundleSize(db, float32(desiredWidth)); err != nil {
			return fmt.Errorf("build endpoint: failed to save requested bundle size: %w", err)
		}

		bundles, err := BuildPressureFitBundles(productCache, float32(desiredWidth))
		if err != nil {
			return fmt.Errorf("build endpoint: failed to build pressure fit bundles: %w", err)
		}

		data := map[string]any{
			"RequestedBundleSize": float32(desiredWidth),
			"Bundles":             bundles,
			"Env":                 config.Mode,
		}

		return renderPage(w, "build-results", data)
	})

	// Product page endpoints.
	handle.get("/gates/", func(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
		gates, err := productCache.GetGates(repos.ProductFilterParams{})
		if err != nil {
			return fmt.Errorf("gates page: failed to get gates: %w", err)
		}

		data := map[string]any{
			"Heading":         "Shop All Gates",
			"PageTitle":       "Shop All Gates",
			"MetaDescription": "Shop our full range of gates",
			"Products":        gates,
			"Cart":            cart,
			"Env":             config.Mode,
		}

		return renderPage(w, "products", data)
	})

	handle.get("/gates/{gate_id}", func(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
		gateID, err := strconv.Atoi(r.PathValue("gate_id"))
		if err != nil {
			return fmt.Errorf("gate details: failed to convert gate_id to integer: %w", err)
		}

		gate, err := productCache.GetProductByID(gateID)
		if err != nil {
			return fmt.Errorf("gate details: failed to retrieve gate: %w", err)
		}

		data := map[string]any{
			"PageTitle":       gate.Name,
			"MetaDescription": gate.Name,
			"Product":         gate,
			"Cart":            cart,
			"Env":             config.Mode,
		}

		return renderPage(w, "product", data)
	})

	handle.get("/extensions/", func(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
		extensions, err := productCache.GetExtensions(repos.ProductFilterParams{})
		if err != nil {
			return fmt.Errorf("extensions page: failed to get extensions: %w", err)
		}

		data := map[string]any{
			"Heading":         "All extensions",
			"PageTitle":       "All extensions",
			"MetaDescription": "Shop all extensions",
			"Products":        extensions,
			"Cart":            cart,
			"Env":             config.Mode,
		}

		return renderPage(w, "products", data)
	})

	handle.get("/extensions/{extension_id}", func(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
		extensionID, err := strconv.Atoi(r.PathValue("extension_id"))
		if err != nil {
			return fmt.Errorf("extension details: failed to convert extension_id to integer: %w", err)
		}

		extension, err := productCache.GetProductByID(extensionID)
		if err != nil {
			return fmt.Errorf("extension details: failed to retrieve extension: %w", err)
		}

		data := map[string]any{
			"PageTitle":       extension.Name,
			"MetaDescription": extension.Name,
			"Cart":            cart,
			"Env":             config.Mode,
		}

		return renderPage(w, "products", data)
	})

	// Cart endpoints.
	handle.get("/cart/", func(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
		data := map[string]any{
			"PageTitle":       "Your shopping cart",
			"MetaDescription": "",
			"Cart":            cart,
			"Env":             config.Mode,
		}

		return renderPage(w, "cart", data)
	})

	handle.get("/cart/json", func(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
		bytes, err := json.Marshal(cart)
		if err != nil {
			return fmt.Errorf("cart JSON endpoint: failed to marshal cart: %w", err)
		}

		if _, err := w.Write(bytes); err != nil {
			return fmt.Errorf("cart JSON endpoint: failed to write response: %w", err)
		}
		return nil
	})

	if config.Mode == Development {
		handle("/test", func(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
			if r.Method == http.MethodGet {
				return renderPage(w, "test", map[string]any{})
			}
			if r.Method == http.MethodPost {
				if err := r.ParseForm(); err != nil {
					return fmt.Errorf("cart add: failed to parse form: %w", err)
				}
				fmt.Printf("%+v\n", r.Form["data"])
				for k, v := range r.Form["data"] {
					fmt.Printf("k:%d v:%s\n", k, v)
				}
				return nil
			}
			return nil
		})
	}

	handle.post("/cart/add", func(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
		if err := r.ParseForm(); err != nil {
			return fmt.Errorf("cart add: failed to parse form: %w", err)
		}

		formData := r.Form["data"]
		if len(formData) < 1 {
			return renderPartial(w, "cart-modal", cart)
		}

		components := []models.CartItemComponent{}

		for _, d := range formData {
			component := models.NewCartItemComponent(cart.ID)
			if err := json.Unmarshal([]byte(d), &component); err != nil {
				return fmt.Errorf("cart add: failed to unmarshal cart item component %s: %w", d, err)
			}
			components = append(components, component)
		}

		if err := AddItemToCart(db, cart.ID, models.NewCartItem(cart.ID, components)); err != nil {
			return fmt.Errorf("cart add: failed to add item to cart: %w", err)
		}

		cart, err := GetCartByID(db, productCache, cart.ID)
		if err != nil {
			return fmt.Errorf("cart add: failed to retrieve updated cart: %w", err)
		}

		return renderPartial(w, "cart-modal", cart)
	})

	handle.post("/cart/item/remove", func(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
		if err := r.ParseForm(); err != nil {
			return fmt.Errorf("cart item remove: failed to parse form: %w", err)
		}

		cartItemID := r.Form.Get("cart_item_id")
		if cartItemID == "" {
			return nil
		}

		if _, err := db.Exec(`DELETE FROM cart_item WHERE id = ? AND cart_id = ?`, cartItemID, cart.ID); err != nil {
			return fmt.Errorf("cart item remove: failed to delete cart item: %w", err)
		}

		cart, err := GetCartByID(db, productCache, cart.ID)
		if err != nil {
			return fmt.Errorf("cart item remove: failed to retrieve updated cart: %w", err)
		}

		if err := renderPartial(w, "cart-main", cart); err != nil {
			return fmt.Errorf("cart item remove: failed to render partial (cart-main): %w", err)
		}

		return nil
	})

	handle.post("/cart/item/{mode}", func(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {

		mode := r.PathValue("mode")

		if mode != "increment" && mode != "decrement" {
			return nil
		}

		if err := r.ParseForm(); err != nil {
			return fmt.Errorf("cart item update: failed to parse form: %w", err)
		}

		cartItemID := r.Form.Get("cart_item_id")
		if cartItemID == "" {
			return fmt.Errorf("cart item update: cart_item_id is blank")
		}

		cartItem, err := selectCartItem(db, cart.ID, cartItemID)
		if err != nil {
			return fmt.Errorf("cart item update: failed to select cart item: %w", err)
		}

		if mode == "increment" {
			if err := IncrementCartItem(db, cart.ID, cartItem.ID); err != nil {
				return fmt.Errorf("cart item update: failed to increment cart item: %w", err)
			}
		} else {
			if cartItem.Qty < 2 {
				w.WriteHeader(http.StatusBadRequest)
				return nil
			}
			if err := DecrementCartItem(db, cart.ID, cartItem.ID); err != nil {
				return fmt.Errorf("cart item update: failed to decrement cart item: %w", err)
			}
		}

		cart, err = GetCartByID(db, productCache, cart.ID)
		if err != nil {
			return fmt.Errorf("cart item update: failed to retrieve updated cart: %w", err)
		}
		if err := renderPartial(w, "cart-main", cart); err != nil {
			return fmt.Errorf("cart item update: failed to render partial (cart-main): %w", err)
		}
		if err := renderPartial(w, "cart-modal-oob", cart); err != nil {
			return fmt.Errorf("cart item update: failed to render partial (cart-modal-oob): %w", err)
		}

		return nil
	})

	handle.delete("/cart/item/", func(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
		if err := r.ParseForm(); err != nil {
			return fmt.Errorf("cart item delete: failed to parse form: %w", err)
		}

		cartItemID := r.Form.Get("id")
		if cartItemID == "" {
			return fmt.Errorf("cart item delete: no cart item id supplied")
		}

		if _, err := db.Exec("DELETE FROM cart_item WHERE id = ? AND cart_id = ?", cartItemID, cart.ID); err != nil {
			return fmt.Errorf("cart item delete: failed to delete cart item: %w", err)
		}

		cart, err := GetCartByID(db, productCache, cart.ID)
		if err != nil {
			return fmt.Errorf("cart item delete: failed to retrieve updated cart: %w", err)
		}

		if err := renderPartial(w, "cart-main", cart); err != nil {
			return fmt.Errorf("cart item delete: failed to render partial (cart-main): %w", err)
		}

		if err := renderPartial(w, "cart-modal-oob", cart); err != nil {
			return fmt.Errorf("cart item delete: failed to render partial (cart-modal-oob): %w", err)
		}

		return nil
	})

	handle.post("/cart/clear", func(cart *models.Cart, w http.ResponseWriter, r *http.Request) error {
		if _, err := db.Exec(`DELETE FROM cart_item WHERE cart_id = ?`, cart.ID); err != nil {
			return fmt.Errorf("cart clear: could not delete cart_item for cart_id %s: %w", cart.ID, err)
		}

		if _, err := db.Exec("DELETE FROM cart_item_component WHERE cart_id = ?", cart.ID); err != nil {
			return fmt.Errorf("cart clear: could not delete cart_item_component for cart_id %s: %w", cart.ID, err)
		}

		cart, err := GetCartByID(db, productCache, cart.ID)
		if err != nil {
			return fmt.Errorf("cart clear: failed to retrieve updated cart: %w", err)
		}

		if err := renderPartial(w, "cart-main", cart); err != nil {
			return fmt.Errorf("cart clear: failed to render partial (cart-main): %w", err)
		}

		if err := renderPartial(w, "cart-modal-oob", cart); err != nil {
			return fmt.Errorf("cart clear: failed to render partial (cart-modal-oob): %w", err)
		}

		return nil
	})

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
		Addr:    ":" + config.Port,
		Handler: router,
	}

	errChan := make(chan error)
	go func() {
		log.Println("Listening on http://localhost:" + config.Port)
		if err := srv.ListenAndServe(); err != nil {
			errChan <- fmt.Errorf("server startup: failed to listen and serve on env.config.Port %s: %w", config.Port, err)
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

func refreshCartDetails(cart *models.Cart, productCache *cachedrepos.CachedProductRepo) error {
	for i := range cart.Items {
		item := &cart.Items[i]
		for j := range item.Components {
			component := &item.Components[j]
			product, err := productCache.GetProductByID(component.Id)
			if err != nil {
				return err
			}
			originalQty := component.Qty
			component.Product = *product
			component.Product.Qty = originalQty
		}
		item.SetName()
	}
	return nil
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

func NewCartFromSessionGetter(db *sql.DB, products *cachedrepos.CachedProductRepo, store *sessions.CookieStore) func(w http.ResponseWriter, r *http.Request) (*models.Cart, error) {
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
			cart, err := NewCart(db)
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

		exists, err := CartExists(db, cartID.(string))
		if err != nil {
			return nil, err
		}
		if !exists {
			cart, err := NewCart(db)
			if err != nil {
				return nil, err
			}
			if err := attachNewCartToSession(cart, session, w, r); err != nil {
				return nil, err
			}
			return cart, nil
		}

		cart, err := GetCartByID(db, products, cartID.(string))
		if err != nil {
			return nil, err
		}

		return cart, nil
	}
}

func createCustomHandler(environment Environment, router *http.ServeMux, middleware []middlewareFunc, getCartFromRequest func(w http.ResponseWriter, r *http.Request) (*models.Cart, error)) customHandler {
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

			cart, err := getCartFromRequest(w, r)
			if err != nil {
				errMsg := "[ERROR] Failed  %s request to %s. %v"
				log.Printf(errMsg, r.Method, path, err)

				// ideally I would get notified of an error here
				if environment == Development {
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

func NotFoundPage(w http.ResponseWriter, renderPage renderPageFunc) error {
	w.WriteHeader(http.StatusNotFound)
	data := map[string]any{
		"pageTile":        "Page Not Found",
		"MetaDescription": "Unable to find page",
	}
	return renderPage(w, "not-found", data)
}

/* service funcs start */

func BuildPressureFitBundles(products *cachedrepos.CachedProductRepo, limit float32) ([]models.Bundle, error) {
	var bundles []models.Bundle

	gates, err := products.GetProducts(repos.Gate, repos.ProductFilterParams{MaxWidth: limit})
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

func SaveBundle(db *sql.DB, bundle models.Bundle) (int64, error) {
	bundleId, err := SaveBundleAsProduct(db, bundle.Product)
	if err != nil {
		return 0, err
	}
	for iii := 0; iii < len(bundle.Components); iii++ {
		component := bundle.Components[iii]
		err = SaveBundleComponent(db, component.Id, component.Type, bundleId, component.Qty)
		if err != nil {
			return 0, err
		}
	}
	return bundleId, nil
}

func TotalValue(products *cachedrepos.CachedProductRepo, cart models.Cart) (float32, error) {
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

func NewCart(db *sql.DB) (*models.Cart, error) {
	cart := models.NewCart()
	if _, err := SaveCart(db, cart); err != nil {
		return nil, err
	}
	return &cart, nil
}

func AddItemToCart(db *sql.DB, cartID string, cartItem models.CartItem) error {
	exists, err := doesCartItemExist(db, cartID, cartItem.ID)
	if err != nil {
		return err
	}
	if !exists {
		if err := InsertCartItem(db, cartItem); err != nil {
			return fmt.Errorf("adding item to cart failed at insert or increment cart item: %w", err)
		}
		if err := SaveCartItemComponents(db, cartItem.Components); err != nil {
			return fmt.Errorf("adding item components failed: %w", err)
		}
	} else {
		if err := IncrementCartItem(db, cartID, cartItem.ID); err != nil {
			return err
		}
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

// Other methods for user-related operations (e.g., UpdateUser, DeleteUser, etc.)

/* service funcs end */

/* repository funcs start */

func InsertContactForm(ctx context.Context, db *sql.DB, contact struct {
	Email   string
	Name    string
	Message string
}) error {
	q := "INSERT INTO contact (name, email, message, timestamp) VALUES (?, ?, ?, ?)"
	_, err := db.ExecContext(ctx, q, contact.Email, contact.Name, contact.Message, time.Now())
	if err != nil {
		return fmt.Errorf("contact page: failed to insert contact into database: %w", err)
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

func SaveBundleComponent(db *sql.DB, productID int, productType string, bundleID int64, qty int) error {
	_, err := db.Exec(
		"INSERT INTO bundle_components(product_id, product_type, bundle_id, qty) VALUES (?, ?, ?, ?)",
		productID, productType, bundleID, qty,
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

func structToString(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "error: marshalling struct to string"
	}

	return string(b)
}

func templateParser(env Environment) tmplFunc {
	var tmpl *template.Template
	return func() *template.Template {
		if env == Production && tmpl != nil {
			return tmpl
		}
		tmpl = template.Must(template.New("").Funcs(template.FuncMap{
			"toString": structToString,
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

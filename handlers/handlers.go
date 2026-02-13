package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"text/template"
	"time"

	"github.com/gorilla/sessions"
	"github.com/seanomeara96/auth"
	"github.com/seanomeara96/gates/config"
	"github.com/seanomeara96/gates/models"
	"github.com/seanomeara96/gates/render"
	"github.com/seanomeara96/gates/repos"
	"github.com/seanomeara96/gates/repos/cache"
	"github.com/seanomeara96/gates/repos/sqlite"
	"github.com/seanomeara96/gates/views/pages"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/webhook"
	"golang.org/x/time/rate"

	_ "github.com/mattn/go-sqlite3"
)

type Handler struct {
	db           *sql.DB
	cfg          *config.Config
	auth         *auth.Authenticator
	orderRepo    *sqlite.OrderRepo
	cartRepo     *sqlite.CartRepo
	productRepo  *sqlite.ProductRepo
	productCache *cache.CachedProductRepo
	cookieStore  *sessions.CookieStore
	emailRegex   *regexp.Regexp
	rndr         *render.Render
}

type CustomHandleFunc func(cart models.Cart, w http.ResponseWriter, r *http.Request) error

func (h *Handler) Close() {
	if h.db != nil {
		h.db.Close()
	}
}

func SqliteOpen(path string) *sql.DB {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		panic(err)
	}
	return db
}

func configCookieStore(cfg *config.Config) (*sessions.CookieStore, error) {

	if cfg.CookieStoreSecretKey == "" {
		if cfg.Mode == config.Development {
			cfg.CookieStoreSecretKey = "suprSecrtStoreKey"
		} else {
			return nil, errors.New("cookie secret not set in env")
		}
	}
	return sessions.NewCookieStore([]byte(cfg.CookieStoreSecretKey)), nil
}

func DefaultHandler(cfg *config.Config) (*Handler, error) {

	var h Handler
	var err error

	stripe.Key = cfg.StripeAPIKey

	h.db = SqliteOpen(cfg.DBPath)
	h.cfg = cfg
	authConfig := auth.AuthConfig{
		DB:           h.db,
		JWTSecretKey: cfg.JWTSecretKey,
	}
	h.auth, err = auth.Init(authConfig)
	if err != nil {
		return nil, fmt.Errorf("default handler: init auth: %w", err)
	}
	h.auth.Register(context.Background(), cfg.AdminUserID, cfg.AdminUserPassword)

	h.productRepo = sqlite.NewProductRepo(h.db)
	h.productCache = cache.NewCachedProductRepo(h.productRepo)

	h.cartRepo = sqlite.NewCartRepo(h.db, h.productRepo)
	h.orderRepo = sqlite.NewOrderRepo(h.db)
	h.cookieStore, err = configCookieStore(cfg)
	if err != nil {
		return nil, fmt.Errorf("default handler: config cookie store: %w", err)
	}

	h.emailRegex, err = regexp.Compile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if err != nil {
		return nil, fmt.Errorf("contact page: could not compile email validation regex: %w", err)
	}

	h.rndr = render.DefaultRender(cfg)

	return &h, nil
}

func (h *Handler) StripeWebhook(cart models.Cart, w http.ResponseWriter, r *http.Request) error {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading request body: %v\n", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return nil
	}

	endpointSecret := h.cfg.StripeWebhookSecret
	// Pass the request body and Stripe-Signature header to ConstructEvent, along
	// with the webhook signing key.
	event, err := webhook.ConstructEvent(
		payload,
		r.Header.Get("Stripe-Signature"),
		endpointSecret,
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error verifying webhook signature: %v\n", err)
		w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
		return nil
	}

	// Unmarshal the event data into an appropriate struct depending on its Type
	switch event.Type {
	case "payment_intent.succeeded":
		if event.Data == nil {
			w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
			log.Println("[WARNING] event data for payment_intent.succeeded is nil")
			return nil
		}

		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
			return fmt.Errorf("stripe webhook: payment_intent.succeeded: unmarshal payment intent: %w", err)
		}

		_id, found := paymentIntent.Metadata["order_id"]
		if !found {
			return fmt.Errorf("stripe webhook: payment_intent.succeeded: missing order_id in payment intent metadata (payment_intent_id=%s)", paymentIntent.ID)
		}

		id, err := strconv.Atoi(_id)
		if err != nil {
			return fmt.Errorf("stripe webhook: payment_intent.succeeded: parse order_id %q (payment_intent_id=%s): %w", _id, paymentIntent.ID, err)
		}

		if err := h.orderRepo.UpdateStatus(id, models.OrderStatusProcessing); err != nil {
			return fmt.Errorf("stripe webhook: payment_intent.succeeded: update order status to %s (order_id=%d, payment_intent_id=%s): %w", models.OrderStatusProcessing, id, paymentIntent.ID, err)
		}

	case "checkout.session.completed":
		if event.Data == nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Println("[WARNING] event data for checkout.session.completed  is nil")
			return nil
		}
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return fmt.Errorf("stripe webhook: checkout.session.completed: unmarshal checkout session: %w", err)
		}

		_id, found := session.Metadata["order_id"]
		if !found {
			log.Println("[WARNING] order id not found on chckout session object")
		}

		id, err := strconv.Atoi(_id)
		if err != nil {
			return fmt.Errorf("stripe webhook: checkout.session.completed: parse order_id %q (session_id=%s): %w", _id, session.ID, err)
		}

		details := session.CustomerDetails
		if details == nil {
			log.Printf("[WARNING] customer details on checkout session is nil order %d", id)
		} else if found {
			order, err := h.orderRepo.GetOrderByID(id)
			if err != nil {
				return fmt.Errorf("stripe webhook: checkout.session.completed: get order by id (order_id=%d, session_id=%s): %w", id, session.ID, err)
			}

			if details.Name != "" {
				order.CustomerName = sql.NullString{
					String: details.Name,
					Valid:  true,
				}
			}

			if details.Email != "" {
				order.CustomerEmail = sql.NullString{
					String: details.Email,
					Valid:  true,
				}
			}

			if details.Phone != "" {
				order.CustomerPhone = sql.NullString{
					String: details.Phone,
					Valid:  true,
				}
			}

			b, err := json.Marshal(details.Address)
			if err != nil {
				log.Printf("[WARNING] could not marshal json for customer details shipping address order: %d", id)
			} else {
				order.ShippingAddress = sql.NullString{
					String: string(b),
					Valid:  true,
				}
			}

			if err := h.orderRepo.UpdateOrder(order); err != nil {
				return fmt.Errorf("stripe webhook: checkout.session.completed: update order with customer details (order_id=%d, session_id=%s): %w", id, session.ID, err)
			}

		}

	default:
		fmt.Fprintf(os.Stderr, "Unhandled event type: %s\n", event.Type)
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func (h *Handler) AdminLoginPage(cart models.Cart, w http.ResponseWriter, r *http.Request) error {
	accessToken, refreshToken, _ := h.auth.GetTokensFromRequest(r)

	_, err := h.auth.ValidateToken(accessToken)
	if err == nil {
		accessToken, refreshToken, err = h.auth.Refresh(r.Context(), refreshToken)
		if err != nil {
			return fmt.Errorf("admin login page: refresh tokens: %w", err)
		}
		h.auth.SetTokens(w, accessToken, refreshToken)
		http.Redirect(w, r, "/admin", http.StatusTemporaryRedirect)
		return nil
	}

	if h.cfg.UseTempl {
		props := pages.AdminLoginPageProps{
			BaseProps: pages.BaseProps{
				PageTitle:       "Admin Login Page",
				MetaDescription: "Admin Login Page",
			},
		}
		return pages.AdminLogin(props).Render(r.Context(), w)
	}
	data := map[string]any{
		"PageTitle":       "Home Page",
		"MetaDescription": "Welcome to the home page",

		"Cart": cart,
		"Env":  h.cfg.Mode,
	}

	return h.rndr.Page(w, "admin-login", data)
}

func (h *Handler) GetHomePage(cart models.Cart, w http.ResponseWriter, r *http.Request) error {
	if r.URL.Path == "/" {
		featuredGates, err := h.productCache.GetGates(repos.ProductFilterParams{Type: models.ProductTypeGate})
		if err != nil {
			return fmt.Errorf("home page: failed to get featured gates: %w", err)
		}

		extensions, err := h.productCache.GetExtensions(repos.ProductFilterParams{Limit: 2, Type: models.ProductTypeExtension})
		if err != nil {
			return fmt.Errorf("home page: failed to get featured extensions: %w", err)
		}

		if h.cfg.UseTempl {
			props := pages.HomeProps{
				BaseProps: pages.BaseProps{
					PageTitle:       "Welcome to Baby Safety Gates Ireland",
					MetaDescription: "The best service for finding the exact size gate for your requirements.",
					Cart:            cart,
					Env:             h.cfg.Mode,
				},
				FeaturedGates:      featuredGates,
				FeaturedExtensions: extensions,
			}
			return pages.Home(props).Render(r.Context(), w)
		}

		data := map[string]any{
			"PageTitle":          "Home Page",
			"MetaDescription":    "Welcome to the home page",
			"FeaturedGates":      featuredGates,
			"FeaturedExtensions": extensions,
			"Cart":               cart,
			"Env":                h.cfg.Mode,
		}

		return h.rndr.Page(w, "home", data)
	}

	return h.NotFoundPage(w)
}

func (h *Handler) GetContactPage(cart models.Cart, w http.ResponseWriter, r *http.Request) error {
	if h.cfg.UseTempl {
		props := pages.ContactPageProps{
			BaseProps: pages.BaseProps{
				PageTitle:       "Contact Baby Safety Gates Ireland",
				MetaDescription: "Contact Baby Safety Gates Ireland for product questions, sizing advice, orders, and support. Send us a message and we’ll get back to you.",
				Cart:            cart,
				Env:             h.cfg.Mode,
			},
		}
		return pages.Contact(props).Render(r.Context(), w)
	}
	data := map[string]any{
		"PageTitle":       "Contact BabyGate Builders",
		"MetaDescription": "Contact form for Babygate builders",
		"Cart":            cart,
		"Env":             h.cfg.Mode,
	}
	return h.rndr.Page(w, "contact", data)
}

func (h *Handler) GetCheckoutPage(cart models.Cart, w http.ResponseWriter, r *http.Request) error {

	// reset the prices in the cart object in case there has been some manipulation on the client side
	cart.TotalValue = 0
	for i := range cart.Items {
		cartItem := &cart.Items[i]
		cartItem.SalePrice = 0
		for ii := range cartItem.Components {
			component := &cartItem.Components[ii]
			count, err := h.productRepo.CountProductByID(component.Id)
			if err != nil {
				return fmt.Errorf("checkout: count product by id %d: %w", component.Id, err)
			}
			insufficientStock := count < component.Qty
			if insufficientStock {
				return fmt.Errorf("checkout: insufficient stock (product_id=%d, required_qty=%d, available_qty=%d)", component.Id, component.Qty, count)
			}
			price, err := h.productRepo.GetProductPrice(component.Id)
			if err != nil {
				return fmt.Errorf("checkout: get product price (product_id=%d): %w", component.Id, err)
			}
			component.Price = price
			cartItem.SalePrice += ((component.Price) * float32(component.Qty))
		}
		cart.TotalValue += (cartItem.SalePrice * float32(cartItem.Qty))
	}

	if h.cfg.StripeAPIKey == "" {
		if err := json.NewEncoder(w).Encode(cart); err != nil {
			return fmt.Errorf("checkout: encode cart json: %w", err)
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
					Images: func() (images []*string) {
						for _, component := range item.Components {
							images = append(images, stripe.String(component.Img))
						}
						return images
					}(),
				},
				Currency: stripe.String("EUR"),
			},
		},
		)
	}

	id, err := h.orderRepo.New(cart)
	if err != nil {
		return fmt.Errorf("checkout: create new order: %w", err)
	}

	params := &stripe.CheckoutSessionParams{
		ClientReferenceID: stripe.String(strconv.Itoa(id)),
		LineItems:         lineItems,
		Mode:              stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:        stripe.String(h.cfg.Domain + fmt.Sprintf("/success?order_id=%d", id)),
		CancelURL:         stripe.String(h.cfg.Domain + "/cart"),
		ShippingAddressCollection: &stripe.CheckoutSessionShippingAddressCollectionParams{
			AllowedCountries: []*string{stripe.String("IE")},
		},
		PhoneNumberCollection: &stripe.CheckoutSessionPhoneNumberCollectionParams{
			Enabled: stripe.Bool(true),
		},
		Currency: stripe.String("EUR"),
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			Description: stripe.String(fmt.Sprintf("Order: #%d", id)),
			Metadata: map[string]string{
				"order_id": strconv.Itoa(id),
			},
		},
		Metadata: map[string]string{
			"order_id": strconv.Itoa(id),
		},
	}

	s, err := session.New(params)
	if err != nil {
		return fmt.Errorf("checkout: create stripe checkout session (order_id=%d): %w", id, err)
	}

	if err := h.orderRepo.UpdateStripeRef(id, s.ID); err != nil {
		return fmt.Errorf("checkout: update order with stripe ref (order_id=%d, session_id=%s): %w", id, s.ID, err)
	}

	http.Redirect(w, r, s.URL, http.StatusSeeOther)
	return nil
}

var contactFormRateLimiter = rate.NewLimiter(1, 3)

func (h *Handler) ProcessContactFormSumbission(cart models.Cart, w http.ResponseWriter, r *http.Request) error {

	if !contactFormRateLimiter.Allow() {
		log.Printf("[WARNING] rate limit for contact form submission exceeded")
		return nil
	}

	// Limit request body size to prevent DoS attacks
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit

	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("contact page: failed to parse form: %w", err)
	}

	email := r.Form.Get("email")
	name := r.Form.Get("name")
	message := r.Form.Get("message")

	// Check for message length limitations
	// TODO implment an error message on the front end
	if len(message) > 5000 {
		if h.cfg.UseTempl {
			props := pages.ContactPageProps{
				BaseProps: pages.BaseProps{
					PageTitle:       "Contact Baby Safety Gates Ireland",
					MetaDescription: "Contact Baby Safety Gates Ireland for product questions, sizing advice, orders, and support. Send us a message and we’ll get back to you.",
					Cart:            cart,
					Env:             h.cfg.Mode,
				},
			}
			return pages.Contact(props).Render(r.Context(), w)
		}
		return h.rndr.Page(w, "contact", map[string]any{
			"PageTitle":       "Contact us | Message Too Long",
			"MetaDescription": "Please provide a shorter message",
			"Cart":            cart,
			"Env":             h.cfg.Mode,
			"Error":           "Message exceeds maximum length",
		})
	}

	// Validate required fields
	if message == "" {
		if h.cfg.UseTempl {
			props := pages.ContactPageProps{
				BaseProps: pages.BaseProps{
					PageTitle:       "Contact Baby Safety Gates Ireland",
					MetaDescription: "Contact Baby Safety Gates Ireland for product questions, sizing advice, orders, and support. Send us a message and we’ll get back to you.",
					Cart:            cart,
					Env:             h.cfg.Mode,
				},
			}
			return pages.Contact(props).Render(r.Context(), w)
		}
		return h.rndr.Page(w, "contact", map[string]any{
			"PageTitle":       "Contact us | No Message Provided",
			"MetaDescription": "Please provide a message",
			"Cart":            cart,
			"Env":             h.cfg.Mode,
			"Error":           "Message is required",
		})
	}

	// Strict email validation
	if !h.emailRegex.MatchString(email) || email == "" {
		if h.cfg.UseTempl {
			props := pages.ContactPageProps{
				BaseProps: pages.BaseProps{
					PageTitle:       "Contact Baby Safety Gates Ireland",
					MetaDescription: "Contact Baby Safety Gates Ireland for product questions, sizing advice, orders, and support. Send us a message and we’ll get back to you.",
					Cart:            cart,
					Env:             h.cfg.Mode,
				},
			}
			return pages.Contact(props).Render(r.Context(), w)
		}
		return h.rndr.Page(w, "contact", map[string]any{
			"PageTitle":       "Contact us | Invalid Email",
			"MetaDescription": "Please provide a valid email address",
			"Cart":            cart,
			"Env":             h.cfg.Mode,
			"Error":           "Valid email is required",
		})
	}

	// Name validation
	if len(name) > 100 || name == "" {
		if h.cfg.UseTempl {
			props := pages.ContactPageProps{
				BaseProps: pages.BaseProps{
					PageTitle:       "Contact Baby Safety Gates Ireland",
					MetaDescription: "Contact Baby Safety Gates Ireland for product questions, sizing advice, orders, and support. Send us a message and we’ll get back to you.",
					Cart:            cart,
					Env:             h.cfg.Mode,
				},
			}
			return pages.Contact(props).Render(r.Context(), w)
		}
		return h.rndr.Page(w, "contact", map[string]any{
			"PageTitle":       "Contact us | Invalid Name",
			"MetaDescription": "Please provide a valid name",
			"Cart":            cart,
			"Env":             h.cfg.Mode,
			"Error":           "Valid name is required (maximum 100 characters)",
		})
	}

	// Sanitize inputs before storing
	var contact struct {
		Email   string
		Name    string
		Message string
	}

	contact.Email = template.HTMLEscapeString(email)
	contact.Name = template.HTMLEscapeString(name)
	contact.Message = template.HTMLEscapeString(message)

	// Use context with timeout for database operations
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := InsertContactForm(ctx, h.db, contact); err != nil {
		// Don't expose database errors to the client
		log.Printf("Contact form database error: %v", err)
		if h.cfg.UseTempl {
			props := pages.ContactPageProps{
				BaseProps: pages.BaseProps{
					PageTitle:       "Contact Baby Safety Gates Ireland",
					MetaDescription: "Contact Baby Safety Gates Ireland for product questions, sizing advice, orders, and support. Send us a message and we’ll get back to you.",
					Cart:            cart,
					Env:             h.cfg.Mode,
				},
			}
			return pages.Contact(props).Render(r.Context(), w)
		}
		return h.rndr.Page(w, "contact", map[string]any{
			"PageTitle":       "Contact Error",
			"MetaDescription": "An error occurred processing your request",
			"Cart":            cart,
			"Env":             h.cfg.Mode,
			"Error":           "Unable to process your request at this time",
		})
	}
	// TODO implement a success message
	if h.cfg.UseTempl {
		props := pages.ContactPageProps{
			BaseProps: pages.BaseProps{
				PageTitle:       "Contact Baby Safety Gates Ireland",
				MetaDescription: "Contact Baby Safety Gates Ireland for product questions, sizing advice, orders, and support. Send us a message and we’ll get back to you.",
				Cart:            cart,
				Env:             h.cfg.Mode,
			},
		}
		return pages.Contact(props).Render(r.Context(), w)
	}
	data := map[string]any{
		"PageTitle":       "Contact BabyGate Builders",
		"MetaDescription": "Contact form for Babygate builders",
		"Cart":            cart,
		"Env":             h.cfg.Mode,
		"Success":         "Your message has been sent successfully",
	}
	return h.rndr.Page(w, "contact", data)
}

func (h *Handler) NotFoundPage(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusNotFound)
	if h.cfg.UseTempl {
		props := pages.NotFoundPageProps{
			BaseProps: pages.BaseProps{
				PageTitle:       "Page Not Found",
				MetaDescription: "",
				Env:             h.cfg.Mode,
			},
		}
		return pages.NotFound(props).Render(context.Background(), w)
	}
	data := map[string]any{
		"pageTile":        "Page Not Found",
		"MetaDescription": "Unable to find page",
	}
	return h.rndr.Page(w, "not-found", data)
}

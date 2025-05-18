package router

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/seanomeara96/gates/config"
	"github.com/seanomeara96/gates/handlers"
	"github.com/seanomeara96/gates/models"
)

type customHandleFunc func(cart *models.Cart, w http.ResponseWriter, r *http.Request) error
type middlewareFunc func(next customHandleFunc) customHandleFunc
type customHandler func(path string, fn customHandleFunc)

type Router struct {
	cfg        *config.Config
	handler    *handlers.Handler
	mux        *http.ServeMux
	middleware []middlewareFunc
}

func (r *Router) Mux() *http.ServeMux {
	return r.mux
}

func (r *Router) Close() {
	if r.handler != nil {
		r.handler.Close()
	}
}

func DefaultRouter(cfg *config.Config) (*Router, error) {
	var r Router
	var err error

	r.cfg = cfg
	if r.cfg == nil {
		return nil, fmt.Errorf("config passed to default router cannot be nil")
	}

	r.handler, err = handlers.DefaultHandler(cfg)
	if err != nil {
		return nil, err
	}

	r.mux = http.NewServeMux()

	r.Handle("/webhook", r.handler.StripeWebhook)
	r.Handle("/", r.handler.GetHomePage)
	r.Get("/admin/login", r.handler.GetAdminLoginPage)
	r.Post("/admin/login", r.handler.AdminLogin)
	r.Get("/admin/logout", r.handler.AdminLogout)
	r.Get("/admin", r.handler.GetAdminDashboard)
	r.Get("/contact", r.handler.GetContactPage)
	r.Get("/checkout", r.handler.GetCheckoutPage)
	r.Post("/contact", r.handler.ProcessContactFormSumbission)
	// Build endpoint. Currently only handling builds for pressure gates.
	r.Post("/build", r.handler.BuildBundle)
	// Product page endpoints.
	r.Get("/gates", r.handler.GetGatesPage)
	r.Get("/gates/{gate_id}", r.handler.GetGatePage)
	r.Get("/extensions", r.handler.GetExtensionsPage)
	r.Get("/extensions/{extension_id}", r.handler.GetExtensionPage)
	// Cart endpoints.
	r.Get("/cart", r.handler.GetCartPage)
	r.Get("/cart/json", r.handler.GetCartJSON)

	if cfg.Mode == config.Development {
		r.Handle("/test", r.handler.Test)
	}
	r.Post("/cart/add", r.handler.AddItemToCart)
	r.Post("/cart/item/{mode}", r.handler.AdjustCartItemQty)
	r.Delete("/cart/item", r.handler.RemoveItemFromCart)
	r.Post("/cart/clear", r.handler.ClearItemsFromCart)

	// Initialize assets dir.
	assetsDirPath := "/assets/"
	httpFileSystem := http.Dir("assets")
	staticFileHttpHandler := http.FileServer(httpFileSystem)
	assetsPathHandler := http.StripPrefix(assetsDirPath, staticFileHttpHandler)
	r.mux.Handle(assetsDirPath, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(24*int(time.Hour.Seconds())))
		assetsPathHandler.ServeHTTP(w, r)
	}))

	return &r, nil

}
func sendErr(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func (r *Router) Handle(pattern string, fn customHandleFunc) {

	// wrap fn in middleware funcs
	for i := range r.middleware {
		fn = r.middleware[i](fn)
	}

	r.mux.HandleFunc(pattern, func(w http.ResponseWriter, req *http.Request) {
		if r.cfg.Mode == config.Development {
			rMsg := "[INFO] %s request made to %s"
			log.Printf(rMsg, req.Method, req.URL.Path)
		}

		cart, err := r.handler.GetCartFromRequest(w, req)
		if err != nil {
			errMsg := "[ERROR] Failed  %s request to %s. %v"
			log.Printf(errMsg, req.Method, pattern, err)

			// ideally I would get notified of an error here
			if r.cfg.Mode == config.Development {
				sendErr(w, err)
			} else {
				// TODO handle production env err handling different
				sendErr(w, err)
			}
			return
		}

		// custom handler get passed throgh the cart handler middleware first to
		// ensure there is a cart session
		if err := fn(cart, w, req); err != nil {
			errMsg := "[ERROR] Failed  %s request to %s. %v"
			log.Printf(errMsg, req.Method, pattern, err)

			// ideally I would get notified of an error here
			if r.cfg.Mode == config.Development {
				sendErr(w, err)
			} else {
				// TODO handle production env err handling different
				sendErr(w, err)
			}
		}
	})

}

func (r *Router) Get(path string, fn customHandleFunc) {
	r.Handle("GET "+path, fn)
}
func (r *Router) Post(path string, fn customHandleFunc) {
	r.Handle("POST "+path, fn)
}

//	func (r *Router) put(path string, fn customHandleFunc) {
//		r.Handle("PUT "+path, fn)
//	}

func (r *Router) Delete(path string, fn customHandleFunc) {
	r.Handle("DELETE"+path, fn)
}

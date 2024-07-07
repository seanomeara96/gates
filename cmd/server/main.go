package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"github.com/seanomeara96/gates/config"
	"github.com/seanomeara96/gates/handlers"
	"github.com/seanomeara96/gates/render"
	"github.com/seanomeara96/gates/repositories"
	"github.com/seanomeara96/gates/services"
	"github.com/seanomeara96/gates/utils"
)

var db *sql.DB

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

	environment := config.Development
	if *env == "prod" {
		environment = config.Production
	}

	port := *portValue

	db = utils.SqliteOpen("main.db")
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

	if *createTable == "carts" {
		if _, err := cartRepo.CreateTables(); err != nil {
			log.Fatal(err)
		}
	}

	buildHandler := handlers.NewBuildHandler(bundleService, renderer)
	cartHandler := handlers.NewCartHandler(cartService, renderer, store)
	pageHandler := handlers.NewPageHandler(productService, cartService, renderer)

	// ROUTING LOGIC
	middlewareFuncs := []middlewareFn{cartHandler.MiddleWare}

	/*
		Call the middlewares func for each request
		Would be better if we could also do a per route basis
	*/
	middlewares := registerMiddlewares(middlewareFuncs)

	handle := createCustomHandler(environment, router, middlewares, pageHandler.SomethingWentWrong)

	handle("/", pageHandler.Home) // cant use 'get' because it causes conflicts

	/*
		Build enpoint. Currently only handling build for pressure gates
	*/
	handle.post("/build/", buildHandler.Build)

	/*
		Product page endpoints
	*/
	handle.get("/bundles/", pageHandler.Bundles)
	handle.get("/bundles/new", pageHandler.Bundles)
	handle.get("/gates/", pageHandler.Gates)
	handle.get("/extensions/", pageHandler.Extensions)

	/*
		cart endpoints
	*/
	handle.get("/cart/", pageHandler.Cart)
	handle.post("/cart/add", cartHandler.AddItem) // TODO consolidate add & update methods
	handle.post("/cart/remove", cartHandler.RemoveItem)
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

func createCustomHandler(environment config.Environment, router *http.ServeMux, executeMiddlewares middlewaresFunc, errPageHandler customHandleFunc) customHandler {
	return func(path string, fn customHandleFunc) {
		router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {

			if environment == config.Development {
				rMsg := "[INFO] %s request made to %s"
				log.Printf(rMsg, r.Method, r.URL.Path)
			}

			// custom handler get passed throgh the carthandler middleware first to
			// ensure there is a cart session
			err := executeMiddlewares(w, r, fn)
			if err != nil {
				log.Printf("[ERROR] Failed  %s request to %s. %v", r.Method, path, err)

				// ideally I would get notified of an error here
				if environment == config.Development {
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

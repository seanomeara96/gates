package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"github.com/seanomeara96/gates/config"
	"github.com/seanomeara96/gates/handlers"
	"github.com/seanomeara96/gates/render"
	"github.com/seanomeara96/gates/repositories"
	"github.com/seanomeara96/gates/services"
)

var db *sql.DB
var err error

type User struct {
	Email string
}

type BasePageData struct {
	MetaDescription string
	PageTitle       string
	User            User
}

func main() {

	portValue := flag.String("port", "", "port to listen on")
	createTable := flag.String("create-table", "", "table to create")

	flag.Parse()

	if *portValue == "" {
		panic("no port supplied")
	}

	environment := config.Development
	port := *portValue

	db, err = sql.Open("sqlite3", "main.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	store := sessions.NewCookieStore([]byte(`secret-key`))

	router := http.NewServeMux()

	renderer := render.NewRenderer("templates/**/*.tmpl", environment)

	productRepo := repositories.NewProductRepository(db)
	bundleRepo := repositories.NewBundleRepository(db)
	cartRepo := repositories.NewCartRepository(db)

	productService := services.NewProductService(productRepo)
	bundleService := services.NewBundleService(productRepo, bundleRepo)
	cartService := services.NewCartService(cartRepo)

	if *createTable == "cart" {
		if _, err := cartRepo.CreateTables(); err != nil {
			log.Fatal(err)
		}
	}

	buildHandler := handlers.NewBuildHandler(bundleService, renderer)
	cartHandler := handlers.NewCartHandler(cartService, renderer, store)
	pageHandler := handlers.NewPageHandler(productService, cartService, renderer)

	// ROUTING LOGIC

	handle := func(path string, fn customHandleFunc) {
		router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			log.Printf("[INFO] %s request made to %s", r.Method, r.URL.Path)
			if err := cartHandler.MiddleWare(w, r, fn); err != nil {
				log.Printf("Error while making request to %s. %v", path, err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})
	}

	get := func(path string, fn customHandleFunc) {
		handle("GET "+path, fn)
	}
	post := func(path string, fn customHandleFunc) {
		handle("POST "+path, fn)
	}
	put := func(path string, fn customHandleFunc) {
		handle("PUT "+path, fn)
	}

	handle("/", pageHandler.Home) // cant use 'get' because it causes conflicts
	post("/build/", buildHandler.Build)
	get("/bundles/", pageHandler.Bundles)
	get("/bundles/new", pageHandler.Bundles)
	get("/gates/", pageHandler.Gates)
	get("/extensions/", pageHandler.Extensions)
	get("/cart/", pageHandler.Cart)
	put("/cart/", cartHandler.Update) // TODO consolidate add & update methods
	// init assets dir
	assetsDirPath := "/assets/"
	httpFileSystem := http.Dir("assets")
	staticFileHttpHandler := http.FileServer(httpFileSystem)
	assetsPathHandler := http.StripPrefix(assetsDirPath, staticFileHttpHandler)
	router.Handle(assetsDirPath, assetsPathHandler)

	log.Println("Listening on http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

type customHandleFunc func(w http.ResponseWriter, r *http.Request) error

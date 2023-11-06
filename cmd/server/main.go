package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"

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

	router := http.NewServeMux()
	// init assets dir
	assetsDirPath := "/assets/"
	httpFileSystem := http.Dir("assets")
	staticFileHttpHandler := http.FileServer(httpFileSystem)
	assetsPathHandler := http.StripPrefix(assetsDirPath, staticFileHttpHandler)
	router.Handle(assetsDirPath, assetsPathHandler)

	renderer := render.NewRenderer("templates/*.tmpl", environment)

	productRepo := repositories.NewProductRepository(db)
	bundleRepo := repositories.NewBundleRepository(db)
	cartRepo := repositories.NewCartRepository(db)

	productService := services.NewProductService(productRepo)
	bundleService := services.NewBundleService(productRepo, bundleRepo)
	cartService := services.NewCartService(cartRepo)

	buildHandler := handlers.NewBuildHandler(bundleService, renderer)
	cartHandler := handlers.NewCartHandler(cartService, renderer)
	pageHandler := handlers.NewPageHandler(productService, cartService, renderer)

	router.HandleFunc("/build/", buildHandler.Build)
	router.HandleFunc("/", pageHandler.Home)
	router.HandleFunc("/bundles/", pageHandler.Bundles)
	router.HandleFunc("/gates/", pageHandler.Gates)
	router.HandleFunc("/extensions/", pageHandler.Extensions)
	router.HandleFunc("/cart/", pageHandler.Cart)
	router.HandleFunc("/cart/add", cartHandler.Add)
	router.HandleFunc("/cart/update", cartHandler.Update)
	router.HandleFunc("/cart/new", cartHandler.New)

	fmt.Println("Listening on http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

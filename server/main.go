package main

import (
	"database/sql"
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

	environment := config.Development
	port := "3000"

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

	pageHandler := handlers.NewPageHandler(productService, renderer)
	buildHandler := handlers.NewBuildHandler(bundleService, renderer)
	cartHandler := handlers.NewCartHandler(cartService, renderer)

	router.HandleFunc("/build/", buildHandler.Build)
	router.HandleFunc("/", pageHandler.Home)
	router.HandleFunc("/bundles/", pageHandler.Bundles)
	router.HandleFunc("/gates/", pageHandler.Gates)
	router.HandleFunc("/extensions/", pageHandler.Extensions)
	router.HandleFunc("/cart/", cartHandler.View)

	fmt.Println("Listening on http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

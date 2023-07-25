package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	"github.com/seanomeara96/gates/config"
	"github.com/seanomeara96/gates/handlers"
	"github.com/seanomeara96/gates/render"
	"github.com/seanomeara96/gates/repositories"
	"github.com/seanomeara96/gates/services"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var tmpl *template.Template
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

	funcMap := template.FuncMap{
		"sizeRange": func(width, tolerance float32) float32 {
			return width - tolerance
		},
		"title": func(str string) string {
			return cases.Title(language.AmericanEnglish).String(str)
		},
	}

	tmpl = template.New("gate-builder").Funcs(funcMap)
	tmpl = template.Must(tmpl.ParseGlob("templates/*.tmpl"))
	renderer := render.NewRenderer(tmpl, environment)

	productRepo := repositories.NewProductRepository(db)
	bundleRepo := repositories.NewBundleRepository(db)

	productService := services.NewProductService(productRepo)
	bundleService := services.NewBundleService(productRepo, bundleRepo)

	pageHandler := handlers.NewPageHandler(productService, renderer)
	buildHandler := handlers.NewBuildHandler(bundleService, renderer)

	router.HandleFunc("/build/", buildHandler.Build)
	router.HandleFunc("/", pageHandler.Home)
	router.HandleFunc("/bundles/", pageHandler.Bundles)
	router.HandleFunc("/gates/", pageHandler.Gates)
	router.HandleFunc("/extensions/", pageHandler.Extensions)

	fmt.Println("Listening on http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", router))
}

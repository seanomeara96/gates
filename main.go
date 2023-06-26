package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	"github.com/seanomeara96/gates/handlers"
	"github.com/seanomeara96/gates/repositories"
	"github.com/seanomeara96/gates/services"
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
	db, err = sql.Open("sqlite3", "main.db")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer db.Close()
	router := http.NewServeMux()
	// init assets dir
	assetsDirPath := "/assets/"
	httpFileSystem := http.Dir("assets")
	staticFileHttpHandler := http.FileServer(httpFileSystem)
	assetsPathHandler := http.StripPrefix(assetsDirPath, staticFileHttpHandler)
	router.Handle(assetsDirPath, assetsPathHandler)

	tmpl = template.Must(template.ParseGlob("./templates/*.tmpl"))

	productRepo := repositories.NewProductRepository(db)
	productService := services.NewProductService(productRepo)
	productHandler := handlers.NewProductHandler(productService, tmpl)

	//	http.HandleFunc("/", handlers.HomeHandler)

	//	http.HandleFunc("/build/", handlers.BuildHandler)

	//	http.HandleFunc("/bundles/", handlers.BundleHandler)

	router.HandleFunc("/gates/", productHandler.GetGates)
	router.HandleFunc("/extensions/", productHandler.GetExtensions)

	fmt.Println("Listening on http://localhost:3000")
	err := http.ListenAndServe(":3000", router)
	if err != nil {
		panic(err)
	}
}

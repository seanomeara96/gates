package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/seanomeara96/gates/repositories"
)

func main() {
	db, err := sql.Open("sqlite3", "main.db")
	if err != nil {
		log.Fatal("could not connect to database")
	}

	productRepository := repositories.NewProductRepository(db)

	userService := services.NewProductService(productRepository)

	productHandler := handler.NewProductHandler(productService)

	router := http.NewServeMux()

	router.HandleFunc("/products", productHandler.GetProducts).Methods("GET")
	port := "3000"
	log.Println("Server  is running on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, router))

}

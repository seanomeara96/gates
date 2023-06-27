package tests

import (
	"database/sql"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/seanomeara96/gates/handlers"
	"github.com/seanomeara96/gates/repositories"
	"github.com/seanomeara96/gates/services"
)

func TestGetGates(t *testing.T) {
	db, err := sql.Open("sqlite3", "../main.db")
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	tmpl := template.Must(template.ParseGlob("../templates/*.tmpl"))
	productRepo := repositories.NewProductRepository(db)
	productService := services.NewProductService(productRepo)
	productHandler := handlers.NewPageHandler(productService, tmpl)

	req, err := http.NewRequest("GET", "/gates/", nil)
	if err != nil {
		t.Error(err)
	}

	recorder := httptest.NewRecorder()

	productHandler.Gates(recorder, req)

	resp := recorder.Result()

	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status code %d, got %d instead", http.StatusOK, resp.StatusCode)
	}

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}

}

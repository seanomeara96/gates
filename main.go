package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

var tmpl *template.Template

func inValidRequest(w http.ResponseWriter) {
	templateErr := tmpl.ExecuteTemplate(w, "inavlidRequest.tmpl", nil)
	if templateErr != nil {
		http.Error(w, "Invalid Request Method", http.StatusMethodNotAllowed)
	}
}

func internalStatusError(description string, err error, w http.ResponseWriter) {
	fmt.Println(description)
	fmt.Println(err)
	templateErr := tmpl.ExecuteTemplate(w, "notfound.tmpl", nil)
	if templateErr != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	db, err := sql.Open("sqlite3", "main.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	tmpl, err := template.ParseGlob("./templates/*.tmpl")
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/" {
			type Gate struct {
				Id        int
				Name      string
				Width     float32
				Price     float32
				Img       string
				Tolerance float32
				Color     string
			}
			type Gates []Gate

			rows, err := db.Query("SELECT id, name, width, price, img, tolerance, color FROM gates")
			if err != nil {
				internalStatusError("could not fetch gates from db", err, w)
				return
			}
			var featuredGates Gates
			for rows.Next() {
				var gate Gate
				err := rows.Scan(&gate.Id, &gate.Name, &gate.Width, &gate.Price, &gate.Img, &gate.Tolerance, &gate.Color)
				if err != nil {
					internalStatusError("error scanning gates", err, w)
					return
				}
				featuredGates = append(featuredGates, gate)
			}
			rows.Close()

			pageData := struct {
				FeaturedGates Gates
			}{
				FeaturedGates: featuredGates,
			}

			err = tmpl.ExecuteTemplate(w, "index.tmpl", pageData)
			if err != nil {
				panic(err)
			}
			return
		}

		w.WriteHeader(http.StatusNotFound)
		tmpl.ExecuteTemplate(w, "notFound.tmpl", nil)

	})

	http.ListenAndServe(":3000", nil)
}

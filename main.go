package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	"github.com/seanomeara96/gates/build"
	"github.com/seanomeara96/gates/components"
)

var tmpl *template.Template

func inValidRequest(w http.ResponseWriter) {
	//templateErr := tmpl.ExecuteTemplate(w, "inavlidRequest.tmpl", nil)
	//if templateErr != nil {
	http.Error(w, "Invalid Request Method", http.StatusMethodNotAllowed)
	//}
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

	// init assets dir
	assetsDirPath := "/assets/"
	httpFileSystem := http.Dir(assetsDirPath)
	staticFileHttpHandler := http.FileServer(httpFileSystem)
	assetsPathHandler := http.StripPrefix(assetsDirPath, staticFileHttpHandler)
	http.Handle(assetsDirPath, assetsPathHandler)

	tmpl := template.Must(template.ParseGlob("./templates/*.tmpl"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/" {

			rows, err := db.Query("SELECT id, name, width, price, img, tolerance, color FROM gates")
			if err != nil {
				internalStatusError("could not fetch gates from db", err, w)
				return
			}
			var featuredGates components.Gates
			for rows.Next() {
				var gate components.Gate
				err := rows.Scan(&gate.Id, &gate.Name, &gate.Width, &gate.Price, &gate.Img, &gate.Tolerance, &gate.Color)
				if err != nil {
					internalStatusError("error scanning gates", err, w)
					return
				}
				featuredGates = append(featuredGates, gate)
			}
			rows.Close()

			pageData := struct {
				FeaturedGates components.Gates
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

	http.HandleFunc("/build/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var data struct {
				DesiredWidth int `json:"width"`
			}
			err := json.NewDecoder(r.Body).Decode(&data)
			if err != nil {
				internalStatusError("something went wrong while decding json", err, w)
				return
			}
			fmt.Println(data.DesiredWidth)
			// fetch gates & compatible extensions from db
			rows, err := db.Query("SELECT id, name, width, price, img, tolerance, color FROM gates")
			if err != nil {
				internalStatusError("failed to query gates from db", err, w)
				return
			}
			defer rows.Close()
			var gates components.Gates
			for rows.Next() {
				var gate components.Gate
				err := rows.Scan(&gate.Id, &gate.Name, &gate.Width, &gate.Price, &gate.Img, &gate.Tolerance, &gate.Color)
				if err != nil {
					//internalStatusError("somehting went wrong while scanning gate rows", err, w)
					// return
					// maybe just print?
					fmt.Println(err)
					continue
				}
				gates = append(gates, gate)
			}
			var bundles components.Bundles
			for i := 0; i < len(gates); i++ {
				gate := gates[i]
				rows, err := db.Query("SELECT e.id, name, width, price, img, color FROM extensions e INNER JOIN compatibles c ON e.id = c.extension_id WHERE c.gate_id = ?", gate.Id)
				if err != nil {
					internalStatusError("could not query compatible extensions", err, w)
					return
				}
				defer rows.Close()
				var compatibleExtensions components.Extensions
				for rows.Next() {
					var extension components.Extension
					err := rows.Scan(&extension.Id, &extension.Name, &extension.Width, &extension.Price, &extension.Img, &extension.Color)
					if err != nil {
						fmt.Println("something went wrong while scanning extension rows", err)
						continue
					}
					compatibleExtensions = append(compatibleExtensions, extension)
				}
				bundle := build.Bundle(gate, compatibleExtensions, float32(data.DesiredWidth))
				if bundle.Qty > 0 {
					bundles = append(bundles, bundle)
				}
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			byt, err := json.Marshal(bundles)
			if err != nil {
				internalStatusError("error marshalling bundles", err, w)
				return
			}
			err = json.NewEncoder(w).Encode(string(byt))
			if err != nil {
				internalStatusError("something went wrong while responding with json", err, w)
				return
			}

			// encode each bundle into a json array and send it back

			// temporary

			return
		}
		inValidRequest(w)
	})

	http.ListenAndServe(":3000", nil)
}

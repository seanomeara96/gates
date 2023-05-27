package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	"github.com/seanomeara96/gates/build"
	"github.com/seanomeara96/gates/types"
	"github.com/seanomeara96/gates/utils"
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

func fetchAllGates() (types.Gates, error) {
	var featuredGates types.Gates
	rows, err := db.Query("SELECT id, name, width, price, img, tolerance, color FROM products WHERE type = 'gate'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	featuredGates, err = utils.ParseGates(rows)
	if err != nil {
		return nil, err
	}

	return featuredGates, nil
}

func fetchAllExtensions() (types.Extensions, error) {
	rows, err := db.Query("SELECT id, name, width, price, img, color FROM products WHERE type = 'extension'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	extensions, err := utils.ParseExtensions(rows)
	if err != nil {
		return nil, err
	}
	return extensions, nil

}

func inValidRequest(w http.ResponseWriter) {
	//templateErr := tmpl.ExecuteTemplate(w, "inavlidRequest.tmpl", nil)
	//if templateErr != nil {
	http.Error(w, "Invalid Request Method", http.StatusMethodNotAllowed)
	//}
}

func internalStatusError(description string, err error, w http.ResponseWriter) {
	fmt.Println(description)
	fmt.Println(err)
	templateErr := tmpl.ExecuteTemplate(w, "notFound.tmpl", nil)
	if templateErr != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func notFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	err := tmpl.ExecuteTemplate(w, "notFound.tmpl", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	db, err = sql.Open("sqlite3", "main.db")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer db.Close()

	// init assets dir
	assetsDirPath := "/assets/"
	httpFileSystem := http.Dir("assets")
	staticFileHttpHandler := http.FileServer(httpFileSystem)
	assetsPathHandler := http.StripPrefix(assetsDirPath, staticFileHttpHandler)
	http.Handle(assetsDirPath, assetsPathHandler)

	tmpl = template.Must(template.ParseGlob("./templates/*.tmpl"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/" {
			featuredGates, err := fetchAllGates()
			if err != nil {
				internalStatusError(
					"could not fetch gates from db",
					err,
					w,
				)
				return
			}

			var popularBundles types.Bundles
			rows, err := db.Query("SELECT id, name, width, img, price, color FROM bundles LIMIT 3")
			if err != nil {
				internalStatusError(
					"could not fetch bundles from db",
					err,
					w,
				)
				return
			}

			for rows.Next() {
				var bundle types.Bundle
				if err := rows.Scan(
					&bundle.Id,
					&bundle.Name,
					&bundle.Width,
					&bundle.Img,
					&bundle.Price,
					&bundle.Color,
				); err != nil {
					internalStatusError(
						"could not scan bundle to bundle struct",
						err,
						w,
					)
					return
				}
				popularBundles = append(popularBundles, bundle)
			}

			for i := 0; i < len(popularBundles); i++ {
				var currentBundle *types.Bundle = &popularBundles[i]

				// get the bundle's gates
				rows, err = db.Query(
					"SELECT gate_id, qty FROM bundle_gates WHERE bundle_id = ?",
					currentBundle.Id,
				)
				if err != nil {
					internalStatusError(
						"Something ent wrong while looking for the bundle's gates",
						err,
						w,
					)
					return
				}
				for rows.Next() {
					var gate types.Gate
					if err := rows.Scan(
						&gate.Id,
						&gate.Qty,
					); err != nil {
						internalStatusError(
							"could not scan gates to bundle",
							err,
							w,
						)
						return
					}
					currentBundle.Gates = append(currentBundle.Gates, gate)
				}

				// get the bundle's extensions
				rows, err = db.Query(
					"SELECT extension_id, qty FROM bundle_extensions WHERE bundle_id = ?",
					currentBundle.Id,
				)
				if err != nil {
					internalStatusError(
						"something went wrong while looking for bundle's extensions",
						err,
						w,
					)
					return
				}
				for rows.Next() {
					var extension types.Extension
					if err := rows.Scan(
						&extension.Id,
						&extension.Qty,
					); err != nil {
						internalStatusError(
							"could not scan extensions to bundle",
							err,
							w,
						)
						return
					}
					currentBundle.Extensions = append(currentBundle.Extensions, extension)
				}

				// make sure to call compute metadata
				currentBundle.ComputeMetaData()
			}

			type HomePageData struct {
				FeaturedGates  types.Gates
				PopularBundles types.Bundles
				BasePageData
			}

			pageData := HomePageData{
				BasePageData: BasePageData{
					PageTitle:       "Build your own safety gate",
					MetaDescription: "This is a place to build the perfect safety gate for your home",
					User: User{
						"sean@example.com",
					},
				},
				FeaturedGates:  featuredGates,
				PopularBundles: popularBundles,
			}

			err = tmpl.ExecuteTemplate(w, "index.tmpl", pageData)
			if err != nil {
				fmt.Println(err)
				panic(err)
			}
			return
		}

		notFound(w)
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

			// not critical to rest of function
			go func() {
				// TODO keep track of user inputs(valid ones). From there we can generate "popular bundles"
				_, err := db.Exec(
					"INSERT INTO bundle_sizes(type, size) VALUES ('pressure fit', ?)",
					data.DesiredWidth,
				)
				if err != nil {
					fmt.Println(err)
				}
			}()

			// fetch gates & compatible extensions from db
			rows, err := db.Query(
				"SELECT id, name, width, price, img, tolerance, color FROM products WHERE width < ? AND type = 'gate'",
				data.DesiredWidth,
			)
			if err != nil {
				internalStatusError("failed to query gates from db", err, w)
				return
			}
			defer rows.Close()

			gates, err := utils.ParseGates(rows)
			if err != nil {
				internalStatusError(
					"failed to parse gates",
					err,
					w,
				)
				return
			}

			var bundles types.Bundles
			for i := 0; i < len(gates); i++ {
				gate := gates[i]
				rows, err := db.Query(
					"SELECT p.id, name, width, price, img, color FROM products p INNER JOIN compatibles c ON p.id = c.extension_id WHERE c.gate_id = ?",
					gate.Id,
				)
				if err != nil {
					internalStatusError("could not query compatible extensions", err, w)
					return
				}
				defer rows.Close()
				var compatibleExtensions types.Extensions
				compatibleExtensions, err = utils.ParseExtensions(rows)
				if err != nil {
					internalStatusError("could not parse extensions", err, w)
				}

				bundle, err := build.BuildPressureFitBundle(float32(data.DesiredWidth), gate, compatibleExtensions)
				if err != nil {
					log.Print("error building bundle")
					continue
				}
				if bundle.Qty > 0 {
					bundles = append(bundles, bundle)
				}
			}
			w.Header().Set("Content-Type", "application/json")
			err = json.NewEncoder(w).Encode(bundles)
			if err != nil {
				internalStatusError("something went wrong while responding with json", err, w)
				return
			}

			return
		}
		inValidRequest(w)
	})

	http.HandleFunc("/bundles/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {

			// if theres a query for a specific custom bundle
			query := r.URL.Query()
			if len(query) > 0 {
				q := query.Get("gate")
				e := query.Get("extensions")

				type ItemQuantities struct {
					Id  int `json:"id"`
					Qty int `json:"qty"`
				}
				var gateQuantity ItemQuantities

				err := json.Unmarshal([]byte(q), &gateQuantity)
				if err != nil {
					internalStatusError("error decoding gate data", err, w)
					return
				}

				var extensionQuantities []ItemQuantities
				err = json.Unmarshal([]byte(e), &extensionQuantities)
				if err != nil {
					internalStatusError("error decoding extensions", err, w)
					return
				}

				var bundle types.Bundle
				var gate types.Gate
				err = db.QueryRow(
					"SELECT id, name, width, price, img, tolerance, color FROM products WHERE id = ? AND type = 'gate'",
					gateQuantity.Id,
				).Scan(
					&gate.Id,
					&gate.Name,
					&gate.Width,
					&gate.Price,
					&gate.Img,
					&gate.Tolerance,
					&gate.Color,
				)
				if err != nil {
					internalStatusError("error fetching gate from db for route /bundles/", err, w)
					return
				}
				gate.Qty = gateQuantity.Qty
				bundle.Gates = append(bundle.Gates, gate)

				var extensions types.Extensions
				for _, extensionQuantity := range extensionQuantities {
					var extension types.Extension
					err := db.QueryRow(
						"SELECT id, name, width, price, img, color FROM products WHERE id = ? AND type = 'extension'",
						extensionQuantity.Id,
					).Scan(
						&extension.Id,
						&extension.Name,
						&extension.Width,
						&extension.Price,
						&extension.Img,
						&extension.Color,
					)
					if err != nil {
						internalStatusError("error fetching extension from db route /build/", err, w)
						return
					}
					extension.Qty = extensionQuantity.Qty
					extensions = append(extensions, extension)
				}
				bundle.Extensions = extensions
				// add bundle meta data
				bundle.ComputeMetaData()

				type SingleBundlePageData struct {
					BasePageData
					Bundle types.Bundle
				}

				pageData := SingleBundlePageData{
					BasePageData: BasePageData{
						PageTitle:       "Single Bundle: " + bundle.Name,
						MetaDescription: "Buy Bundle " + bundle.Name + " Online and enjoy super fast delivery",
					},
					Bundle: bundle,
				}

				err = tmpl.ExecuteTemplate(w, "single-bundle.tmpl", pageData)
				if err != nil {
					internalStatusError("error creating bundle page", err, w)
					return
				}
				return
			}

			/*if r.URL.Path == "/bundles/" {
				popularBundles, err := fetchPopularBundles()
				if err != nil {
					internalStatusError("error fetching popular bundles for route /bundles/", err, w)
					return
				}
				pageData := struct {
					PopularBundles types.CachedBundles
				}{
					PopularBundles: popularBundles,
				}

				err = tmpl.ExecuteTemplate(w, "bundles.tmpl", pageData)
				if err != nil {
					internalStatusError("error executing bundles template", err, w)
					return
				}

				return
			}*/
		}
		notFound(w)
	})

	http.HandleFunc("/gates/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/gates/" {

			gates, err := fetchAllGates()
			if err != nil {
				internalStatusError("error fetching gates for route /gates/", err, w)
				return
			}
			pageData := struct {
				Heading  string
				Products types.Gates
			}{
				Heading:  "Shop Gates",
				Products: gates,
			}

			tmpl.ExecuteTemplate(w, "products.tmpl", pageData)

			return
		}
		notFound(w)
	})

	http.HandleFunc("/extensions/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/extensions/" {

			extensions, err := fetchAllExtensions()
			if err != nil {
				internalStatusError("error fetching extensions for route /extensions/", err, w)
				return
			}
			pageData := struct {
				Heading  string
				Products types.Extensions
			}{
				Heading:  "Shop Extensions",
				Products: extensions,
			}

			tmpl.ExecuteTemplate(w, "products.tmpl", pageData)

			return
		}
		notFound(w)
	})

	fmt.Println("Lsitening on 3000")
	http.ListenAndServe(":3000", nil)
}

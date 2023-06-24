package handlers

import "net/http"

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet && r.URL.Path == "/" {
		featuredGates, err := fetchAllGates()
		if err != nil {
			internalStatusError("could not fetch gates from db", err, w)
			return
		}

		var popularBundles Bundles
		rows, err := db.Query("SELECT id, name, width, img, price, color FROM bundles LIMIT 3")
		if err != nil {
			internalStatusError("could not fetch bundles from db", err, w)
			return
		}

		for rows.Next() {
			var bundle Bundle
			err := rows.Scan(&bundle.Id, &bundle.Name, &bundle.Width, &bundle.Img, &bundle.Price, &bundle.Color)
			if err != nil {
				internalStatusError("could not scan bundle to bundle struct", err, w)
				return
			}
			popularBundles = append(popularBundles, bundle)
		}

		for i := 0; i < len(popularBundles); i++ {
			var currentBundle *Bundle = &popularBundles[i]

			// get the bundle's gates
			rows, err = db.Query("SELECT gate_id, qty FROM bundle_gates WHERE bundle_id = ?", currentBundle.Id)
			if err != nil {
				internalStatusError("Something ent wrong while looking for the bundle's gates", err, w)
				return
			}

			for rows.Next() {
				var gate Gate

				err := rows.Scan(&gate.Id, &gate.Qty)
				if err != nil {
					internalStatusError("could not scan gates to bundle", err, w)
					return
				}

				currentBundle.Gates = append(currentBundle.Gates, gate)
			}

			// get the bundle's extensions
			rows, err = db.Query("SELECT extension_id, qty FROM bundle_extensions WHERE bundle_id = ?", currentBundle.Id)
			if err != nil {
				internalStatusError("something went wrong while looking for bundle's extensions", err, w)
				return
			}

			for rows.Next() {
				var extension Extension
				err := rows.Scan(&extension.Id, &extension.Qty)
				if err != nil {
					internalStatusError("could not scan extensions to bundle", err, w)
					return
				}
				currentBundle.Extensions = append(currentBundle.Extensions, extension)
			}

			// make sure to call compute metadata
			currentBundle.ComputeMetaData()
		}

		type HomePageData struct {
			FeaturedGates  Gates
			PopularBundles Bundles
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
			internalStatusError("could not execute templete fo homepage", err, w)
		}
		return
	}

	notFound(w)
}

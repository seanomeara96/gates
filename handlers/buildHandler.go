package handlers

/*
import (
	"encoding/json"
	"fmt"
	"net/http"
)

func BuildHandler(w http.ResponseWriter, r *http.Request) {
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
			_, err := db.Exec("INSERT INTO bundle_sizes(type, size) VALUES ('pressure fit', ?)", data.DesiredWidth)
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

		gates, err := ParseGates(rows)
		if err != nil {
			internalStatusError("failed to parse gates", err, w)
			return
		}

		var bundles Bundles
		for i := 0; i < len(gates); i++ {
			gate := gates[i]
			rows, err := db.Query(`SELECT p.id, name, width, price, img, color
				FROM products p INNER JOIN compatibles c ON p.id = c.extension_id
				WHERE c.gate_id = ?`, gate.Id)

			if err != nil {
				internalStatusError("could not query compatible extensions", err, w)
				return
			}
			defer rows.Close()

			compatibleExtensions, err := ParseExtensions(rows)
			if err != nil {
				internalStatusError("could not parse extensions", err, w)
				return
			}

			bundle, err := BuildPressureFitBundle(float32(data.DesiredWidth), gate, compatibleExtensions)
			if err != nil {
				internalStatusError("error building bundle", err, w)
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
}*/

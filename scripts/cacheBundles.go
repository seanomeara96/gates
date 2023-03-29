package main

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"

	_ "github.com/mattn/go-sqlite3"
	"github.com/seanomeara96/gates/build"
	"github.com/seanomeara96/gates/components"
)

func main() {
	db, err := sql.Open("sqlite3", "main.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// todo either drop tables or clear all rows before this flow
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS bundles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		size REAL NOT NULL,
		price REAL,
		color TEXT
	)`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS bundle_gates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		gate_id INTEGER NOT NULL,
		bundle_id INTEGER NOT NULL,
		qty INTEGER NOT NULL,
		FOREIGN KEY (gate_id) REFERENCES gates(id),
		FOREIGN KEY (bundle_id) REFERENCES bundles(id)
	)`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS bundle_extensions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		extension_id INTEGER NOT NULL,
		bundle_id INTEGER NOT NULL,
		qty INTEGER NOT NULL,
		FOREIGN KEY (extension_id) REFERENCES extensions(id),
		FOREIGN KEY (bundle_id) REFERENCES bundles(id)
	)`)
	if err != nil {
		log.Fatal(err)
	}

	// get most searched for sizes
	rows, err := db.Query("SELECT size, COUNT(*) AS count FROM bundle_sizes GROUP BY size ORDER BY count DESC LIMIT 3")
	if err != nil {
		log.Fatal(err)
	}

	type QueryData struct {
		Size  float32
		Count int
	}
	var data []QueryData
	for rows.Next() {
		var query QueryData
		err := rows.Scan(&query.Size, &query.Count)
		if err != nil {
			fmt.Println(err)
			continue
		}
		data = append(data, query)
	}
	rows.Close()

	getGates, gStmtErr := db.Prepare("SELECT id, name, width, price, img, tolerance, color FROM gates WHERE width < ?")
	if gStmtErr != nil {
		log.Fatal(err)
	}
	getExtensions, eStmtErr := db.Prepare("SELECT e.id, name, width, price, img, color FROM extensions e INNER JOIN compatibles c on e.id = c.extension_id WHERE c.gate_id = ?")
	if eStmtErr != nil {
		log.Fatal(err)
	}
	defer getGates.Close()
	defer getExtensions.Close()

	var bundles components.Bundles

	for i := 0; i < len(data); i++ {
		desiredWidth := data[i].Size
		rows, err := getGates.Query(desiredWidth)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		var gates []components.Gate
		for rows.Next() {
			var gate components.Gate
			err := rows.Scan(&gate.Id, &gate.Name, &gate.Width, &gate.Price, &gate.Img, &gate.Tolerance, &gate.Color)
			if err != nil {
				log.Fatal(err)
			}
			gates = append(gates, gate)
		}

		for i := 0; i < len(gates); i++ {
			gate := gates[i]
			rows, err := getExtensions.Query(gate.Id)
			if err != nil {
				log.Fatal(err)
			}
			defer rows.Close()

			var extensions components.Extensions
			for rows.Next() {
				var extension components.Extension
				err := rows.Scan(&extension.Id, &extension.Name, &extension.Width, &extension.Price, &extension.Img, &extension.Color)
				if err != nil {
					log.Fatal(err)
				}
				extensions = append(extensions, extension)
			}

			bundle := build.Bundle(gate, extensions, float32(desiredWidth))
			bundles = append(bundles, bundle)
		}
	}
	// todo filter duplicate bundles && save bundles to the database
	var uniqueBundles components.Bundles
	for i := 0; i < len(bundles); i++ {
		bundle := bundles[i]
		encountered := false
		for _, value := range uniqueBundles {
			if reflect.DeepEqual(bundle, value) {
				encountered = true
			}
		}
		if encountered == false {
			uniqueBundles = append(uniqueBundles, bundle)
		}
	}
	insertExtensionStmt, err := db.Prepare("INSERT INTO bundle_extensions(extension_id, bundle_id, qty) VALUES (?,?,?)")
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < len(uniqueBundles); i++ {
		bundle := uniqueBundles[i]
		result, err := db.Exec("INSERT INTO bundles(size, price, color) VALUES (?,?,?)", bundle.MaxLength, bundle.Price, bundle.Gate.Color)
		if err != nil {
			log.Fatal("something went wrong adding bundle to db", err)
		}
		lastInsertId, err := result.LastInsertId()
		if err != nil {
			log.Fatal(err)
		}
		bundleId := lastInsertId
		_, err = db.Exec("INSERT INTO bundle_gates(gate_id, bundle_id, qty) VALUES (?, ?, ?)", bundle.Gate.Id, bundleId, bundle.Gate.Qty)
		if err != nil {
			log.Panic("somehting went wrong while inserting bundle gates into db", err)
		}
		for ii := 0; ii < len(bundle.Extensions); ii++ {
			extension := bundle.Extensions[ii]
			_, err := insertExtensionStmt.Exec(extension.Id, bundleId, extension.Qty)
			if err != nil {
				log.Fatal("something went wrong inserting extensions into db", err)
			}
		}
	}
}

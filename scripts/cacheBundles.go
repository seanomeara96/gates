package main

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

func CacheBundles(db *sql.DB) {

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS bundles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		width REAL NOT NULL,
		img TEXT DEFAULT '',
		price REAL,
		color TEXT
	)`)
	if err != nil {
		fmt.Println("Error creating bundles table")
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
		fmt.Println("Error creating bundle gates table")
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
		fmt.Println("error dropping bundle extensions table")
		log.Fatal(err)
	}

	// drop tables or clear all rows before this flow
	_, err = db.Exec(`DELETE FROM bundles`)
	if err != nil {
		fmt.Println("error clearing bundles table")
		log.Fatal(err)
	}
	_, err = db.Exec(`DELETE FROM bundle_gates`)
	if err != nil {
		fmt.Println("error clearing bundle gates table")
		log.Fatal(err)
	}
	_, err = db.Exec(`DELETE FROM bundle_extensions`)
	if err != nil {
		fmt.Println("error clearing bundle extensions table")
		log.Fatal(err)
	}

	// get most searched for sizes
	rows, err := db.Query("SELECT size, COUNT(*) AS count FROM bundle_sizes WHERE size > 0 GROUP BY size ORDER BY count DESC LIMIT 3")
	if err != nil {
		fmt.Println("Error finding most searched-for sizes")
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

	getGates, gStmtErr := db.Prepare("SELECT id, name, width, price, img, tolerance, color FROM products WHERE width < ? AND type = 'gate'")
	if gStmtErr != nil {
		fmt.Println("Error occured trying to prepare select gate statement", err)
		panic(err)
	}
	defer getGates.Close()

	getExtensions, eStmtErr := db.Prepare("SELECT p.id, name, width, price, img, color FROM products p INNER JOIN compatibles c on p.id = c.extension_id WHERE c.gate_id = ?")
	if eStmtErr != nil {
		fmt.Println("Error occured trying to prepare select extension statement", err)
		panic(err)
	}
	defer getExtensions.Close()

	var bundles Bundles

	// for each common size request build a bundle and save it to the database
	for i := 0; i < len(data); i++ {

		desiredWidth := data[i].Size
		fmt.Println("desiredWidth", desiredWidth)

		rows, err := getGates.Query(desiredWidth)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		var gates []Gate
		for rows.Next() {
			var gate Gate
			err := rows.Scan(&gate.Id, &gate.Name, &gate.Width, &gate.Price, &gate.Img, &gate.Tolerance, &gate.Color)
			if err != nil {
				log.Fatal(err)
			}
			gates = append(gates, gate)
		}

		// no gates matched the request
		if len(gates) < 1 {
			continue
		}

		for i := 0; i < len(gates); i++ {
			gate := gates[i]
			rows, err := getExtensions.Query(gate.Id)
			if err != nil {
				log.Fatal(err)
			}
			defer rows.Close()

			var extensions Extensions
			for rows.Next() {
				var extension Extension
				err := rows.Scan(&extension.Id, &extension.Name, &extension.Width, &extension.Price, &extension.Img, &extension.Color)
				if err != nil {
					log.Fatal(err)
				}
				extensions = append(extensions, extension)
			}

			bundle, err := BuildPressureFitBundle(desiredWidth, gate, extensions)
			if err != nil {
				log.Print("errorbuilding bundle")
				continue
			}
			bundles = append(bundles, bundle)
		}
	}

	for _, bundle := range bundles {
		fmt.Println(bundle.Width, bundle.Color)
	}

	// filter duplicate bundles && save unique bundles to the database
	var uniqueBundles Bundles
	for i := 0; i < len(bundles); i++ {
		bundle := bundles[i]
		encountered := false
		for _, value := range uniqueBundles {
			if reflect.DeepEqual(bundle, value) {
				encountered = true
			}
		}
		if !encountered {
			uniqueBundles = append(uniqueBundles, bundle)
		}
	}

	insertExtensionStmt, err := db.Prepare("INSERT INTO bundle_extensions(extension_id, bundle_id, qty) VALUES (?,?,?)")
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < len(uniqueBundles); i++ {
		bundle := uniqueBundles[i]
		extensionsQtyTotal := 0
		for _, extension := range bundle.Extensions {
			extensionsQtyTotal += extension.Qty
		}

		bundleName := bundle.Gates[0].Name
		if extensionsQtyTotal > 0 {
			var trailing string = " Extension"
			if extensionsQtyTotal > 1 {
				trailing = " Extensions"
			}
			bundleName = bundleName + " and " + strconv.Itoa(extensionsQtyTotal) + trailing
		}
		bundle.Name = bundleName

		result, err := db.Exec(
			"INSERT INTO bundles(name, width, img, price, color) VALUES (?, ?, ?, ?, ?)",
			bundle.Name,
			bundle.Width,
			bundle.Gates[0].Img,
			bundle.Price,
			bundle.Gates[0].Color,
		)
		if err != nil {
			log.Fatal("something went wrong adding bundle to db", err)
		}
		lastInsertId, err := result.LastInsertId()
		if err != nil {
			log.Fatal(err)
		}
		bundleId := lastInsertId
		for iii := 0; iii < len(bundle.Gates); iii++ {
			gate := bundle.Gates[iii]
			_, err = db.Exec("INSERT INTO bundle_gates(gate_id, bundle_id, qty) VALUES (?, ?, ?)", gate.Id, bundleId, gate.Qty)
			if err != nil {
				log.Panic("somehting went wrong while inserting bundle gates into db", err)
			}
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

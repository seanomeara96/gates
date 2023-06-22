package main

import (
	"database/sql"
)

func ParseExtensions(rows *sql.Rows) (Extensions, error) {
	var extensions Extensions
	for rows.Next() {
		var extension Extension
		extension.Qty = 1
		err := rows.Scan(&extension.Id, &extension.Name, &extension.Width, &extension.Price, &extension.Img, &extension.Color)
		if err != nil {
			return nil, err
		}
		extensions = append(extensions, extension)
	}
	return extensions, nil
}

func ParseGates(rows *sql.Rows) (Gates, error) {
	var gates Gates
	for rows.Next() {
		var gate Gate
		gate.Qty = 1
		err := rows.Scan(&gate.Id, &gate.Name, &gate.Width, &gate.Price, &gate.Img, &gate.Tolerance, &gate.Color)
		if err != nil {
			return nil, err
		}
		gates = append(gates, gate)
	}
	return gates, nil
}

package utils

import (
	"database/sql"

	"github.com/seanomeara96/gates/types"
)

func ParseExtensions(rows *sql.Rows) (types.Extensions, error) {
	var extensions types.Extensions
	for rows.Next() {
		var extension types.Extension
		extension.Qty = 1
		err := rows.Scan(&extension.Id, &extension.Name, &extension.Width, &extension.Price, &extension.Img, &extension.Color)
		if err != nil {
			return nil, err
		}
		extensions = append(extensions, extension)
	}
	return extensions, nil
}

func ParseGates(rows *sql.Rows) (types.Gates, error) {
	var gates types.Gates
	for rows.Next() {
		var gate types.Gate
		gate.Qty = 1
		err := rows.Scan(&gate.Id, &gate.Name, &gate.Width, &gate.Price, &gate.Img, &gate.Tolerance, &gate.Color)
		if err != nil {
			return nil, err
		}
		gates = append(gates, gate)
	}
	return gates, nil
}

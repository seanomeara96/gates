package tests

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestSizesQuery(t *testing.T) {
	db, err := sql.Open("sqlite3", "main.db")

	if err != nil {
		t.Error(err)
	}

	rows, err := db.Query("SELECT size, COUNT(*) AS count FROM bundle_sizes WHERE size > 0 GROUP BY size ORDER BY count DESC LIMIT 3")
	if err != nil {
		t.Error(err)
	}
	defer rows.Close()

	type QueryData struct {
		Size  float32
		Count int
	}

	var data []QueryData
	for rows.Next() {
		var query QueryData
		err := rows.Scan(&query.Size, &query.Count)
		if err != nil {
			t.Error(err)
		}

		data = append(data, query)
	}

	fmt.Println(data)
}

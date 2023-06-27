package repositories

import (
	"database/sql"

	"github.com/seanomeara96/gates/models"
)

type BundleRepository struct {
	db *sql.DB
}

func NewBundleRepository(db *sql.DB) *BundleRepository {
	return &BundleRepository{
		db,
	}
}

func (r *BundleRepository) CreateTables() error {
	_, err := r.db.Exec(`CREATE TABLE IF NOT EXISTS bundle_gates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		gate_id INTEGER NOT NULL,
		bundle_id INTEGER NOT NULL,
		qty INTEGER NOT NULL,
		FOREIGN KEY (gate_id) REFERENCES products(id),
		FOREIGN KEY (bundle_id) REFERENCES products(id)
	)`)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`CREATE TABLE IF NOT EXISTS bundle_extensions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		extension_id INTEGER NOT NULL,
		bundle_id INTEGER NOT NULL,
		qty INTEGER NOT NULL,
		FOREIGN KEY (extension_id) REFERENCES products(id),
		FOREIGN KEY (bundle_id) REFERENCES products(id)
	)`)
	if err != nil {
		return err
	}
	return nil
}

func (r *BundleRepository) ClearAll() error {
	// drop tables or clear all rows before this flow
	_, err := r.db.Exec(`DELETE FROM products WHERE type = 'bundle'`)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`DELETE FROM bundle_gates`)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`DELETE FROM bundle_extensions`)
	if err != nil {
		return err
	}
	return nil
}

type PopularSize struct {
	Size  float32
	Count int
}

func (r *BundleRepository) PopularSizes(limit int) ([]PopularSize, error) {
	var popularSizes []PopularSize
	// get most searched for sizes
	rows, err := r.db.Query("SELECT size, COUNT(*) AS count FROM bundle_sizes WHERE size > 0 GROUP BY size ORDER BY count DESC LIMIT ?", limit)
	if err != nil {
		return popularSizes, err
	}
	defer rows.Close()

	for rows.Next() {
		var query PopularSize
		err := rows.Scan(&query.Size, &query.Count)
		if err != nil {
			return popularSizes, err
		}
		popularSizes = append(popularSizes, query)
	}
	return popularSizes, nil
}

func (r *BundleRepository) SaveBundleGate(gate_id int, bundle_id int64, qty int) error {
	_, err := r.db.Exec("INSERT INTO bundle_gates(gate_id, bundle_id, qty) VALUES (?, ?, ?)", gate_id, bundle_id, qty)
	if err != nil {
		return err
	}
	return nil
}

func (r *BundleRepository) SaveBundleExtension(extension_id int, bundle_id int64, qty int) error {
	_, err := r.db.Exec("INSERT INTO bundle_extensions(extension_id, bundle_id, qty) VALUES (?,?,?)", extension_id, bundle_id, qty)
	if err != nil {
		return err
	}
	return nil
}

func (r *BundleRepository) SaveBundleAsProduct(bundleProductValues models.Product) (int64, error) {
	result, err := r.db.Exec(
		"INSERT INTO products(type, name, width, price, img, tolerance, color) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"bundle",
		bundleProductValues.Name,
		bundleProductValues.Width,
		bundleProductValues.Price,
		bundleProductValues.Img,
		bundleProductValues.Tolerance,
		bundleProductValues.Color,
	)
	if err != nil {
		return 0, err
	}
	lastInsertId, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return lastInsertId, nil
}

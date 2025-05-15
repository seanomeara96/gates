package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

func InsertContactForm(ctx context.Context, db *sql.DB, contact struct {
	Email   string
	Name    string
	Message string
}) error {
	q := "INSERT INTO contact (name, email, message, timestamp) VALUES (?, ?, ?, ?)"
	_, err := db.ExecContext(ctx, q, contact.Email, contact.Name, contact.Message, time.Now())
	if err != nil {
		return fmt.Errorf("contact page: failed to insert contact into database: %w", err)
	}
	return nil
}

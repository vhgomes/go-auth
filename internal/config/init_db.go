package config

import (
	"database/sql"
	"fmt"
	"log"
)

func InitDB(db *sql.DB) error {
	// SQL statement to create the users table
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(255) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// Execute the SQL statement
	_, err := db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	log.Println("Table 'users' created or already exists.")
	return nil
}

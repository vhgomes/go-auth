package config

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/redis/go-redis/v9"
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

func InitPostgres(cfg *Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DBConfig.Addr)

	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %v", err)
	}

	db.SetMaxOpenConns(cfg.DBConfig.MaxOpenConns)
	db.SetMaxIdleConns(cfg.DBConfig.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.DBConfig.MaxConnLifetime)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	return db, nil
}

func InitRedis(config *RedisConfig) (*redis.Client, error) {

	var ctx = context.Background()

	client := redis.NewClient(&redis.Options{
		Addr:     config.addr,     // Redis server address
		Password: config.password, // No password
		DB:       config.db,       // Default DB
	})

	// Test the connection
	_, err := client.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}

	log.Print("Redis created!")

	return client, nil
}

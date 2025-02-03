package config

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq" // Import the PostgreSQL driver
)

type Application struct {
	Config Config
}

type Config struct {
	Addr     string
	DBConfig DBConfig
}

type DBConfig struct {
	Addr            string
	MaxOpenConns    int
	MaxIdleConns    int
	MaxConnLifetime time.Duration
}

func LoadConfig(port string, dbAddr string) (*Config, error) {
	maxOpenConns, err := getEnvInt("DB_MAX_OPEN_CONNS", 30)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DB_MAX_OPEN_CONNS: %v", err)
	}

	maxIdleConns, err := getEnvInt("DB_MAX_IDLE_CONNS", 30)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DB_MAX_IDLE_CONNS: %v", err)
	}

	maxConnLifetime, err := getEnvDuration("DB_MAX_CONN_LIFETIME", "15m")
	if err != nil {
		return nil, fmt.Errorf("failed to parse DB_MAX_CONN_LIFETIME: %v", err)
	}

	cfg := &Config{
		Addr: port,
		DBConfig: DBConfig{
			Addr:            dbAddr,
			MaxOpenConns:    maxOpenConns,
			MaxIdleConns:    maxIdleConns,
			MaxConnLifetime: maxConnLifetime,
		},
	}

	return cfg, nil
}

func getEnvInt(key string, defaultValue int) (int, error) {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue, nil
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("invalid integer value for %s: %v", key, err)
	}
	return value, nil
}

func getEnvDuration(key string, defaultValue string) (time.Duration, error) {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		valueStr = defaultValue
	}
	duration, err := time.ParseDuration(valueStr)
	if err != nil {
		return 0, fmt.Errorf("invalid duration value for %s: %v", key, err)
	}
	return duration, nil
}

func New(cfg *Config) (*sql.DB, error) {
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

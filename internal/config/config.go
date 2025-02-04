package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq" // Import the PostgreSQL driver
)

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

type RedisConfig struct {
	addr     string
	password string
	db       int
}

func LoadPostgresConfig() (*Config, error) {
	port := os.Getenv("PORT")
	dbAddr := os.Getenv("DB_ADDR")
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

func LoadRedisConfig() (*RedisConfig, error) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		log.Fatal("Redis address is blank")
	}

	password := os.Getenv("REDIS_PASSWORD")
	if password == "" {
		log.Fatal("Redis password is blank")
	}

	db, err := getEnvInt("REDIS_DB", 0)
	if err != nil {
		log.Fatal("failed to parse REDIS_DB")
	}

	redisConfig := &RedisConfig{
		addr:     addr,
		password: password,
		db:       db,
	}

	return redisConfig, nil
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

package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

type Config struct {
	Addr         string
	DBConfig     DBConfig
	CookieSecure bool
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
	err := godotenv.Load()

	if err != nil {
		return nil, fmt.Errorf("failed to load .env file: %v", err)
	}

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

	cookieSecure, err := getEnvBool("COOKIE_SECURE", false)
	if err != nil {
		return nil, fmt.Errorf("failed to parse COOKIE_SECURE: %v", err)
	}

	cfg := &Config{
		Addr: port,
		DBConfig: DBConfig{
			Addr:            dbAddr,
			MaxOpenConns:    maxOpenConns,
			MaxIdleConns:    maxIdleConns,
			MaxConnLifetime: maxConnLifetime,
		},
		CookieSecure: cookieSecure,
	}

	return cfg, nil
}

func LoadRedisConfig() (*RedisConfig, error) {
	err := godotenv.Load()
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		log.Fatal("Redis address is blank")
	}

	password := os.Getenv("REDIS_PASSWORD")

	db, err := getEnvInt("REDIS_DB", 0)
	if err != nil {
		log.Fatal("failed to parse REDIS_DB")
	}

	log.Println(addr, password, db)

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

func getEnvBool(key string, defaultValue bool) (bool, error) {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue, nil
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return false, fmt.Errorf("invalid boolean value for %s: %v", key, err)
	}
	return value, nil
}

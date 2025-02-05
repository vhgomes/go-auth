package repository

import (
	"auth-go/pkg/utils"
	_ "auth-go/pkg/utils"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

type UserRepository struct {
	db    *sql.DB
	redis *redis.Client
}

func NewUserRepository(db *sql.DB, redis *redis.Client) *UserRepository {
	return &UserRepository{
		db:    db,
		redis: redis,
	}
}

func (r *UserRepository) RegisterUser(username string, password string) error {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE username = $1`
	err := r.db.QueryRow(query, username).Scan(&count)

	if err != nil {
		return errors.New("failed to check username existence")
	}

	if count > 0 {
		return errors.New("user already exists")
	}

	hashedPassword, _ := utils.HashPassword(password)

	insertQuery := `INSERT INTO users (username, password) VALUES ($1, $2)`
	_, err = r.db.Exec(insertQuery, username, hashedPassword)
	if err != nil {
		fmt.Printf("%s", err)
		return errors.New("failed to insert user")
	}

	return nil
}

func (r *UserRepository) LoginUser(username string, password string) (string, error) {
	ctx := context.Background()
	var dbpass, userId string
	var queryUser = `SELECT id, username, password from users WHERE username = $1`
	err := r.db.QueryRow(queryUser, username).Scan(&userId, &username, &dbpass)
	if err != nil {
		return "", errors.New("username do not exist")
	}

	if !utils.CheckHash(password, dbpass) {
		return "", errors.New("invalid password")
	}

	token, err := r.GetToken(ctx, userId)
	if token != "" {
		log.Println("Retrieved token is", token)
		return token, nil
	}

	sessionToken := utils.GenerateToken(32)
	log.Println("Generated token is", sessionToken)

	err = r.redis.Set(ctx, userId, sessionToken, time.Hour).Err()

	if err != nil {
		return "", fmt.Errorf("failed to store session in Redis: %v", err)
	}

	log.Println("Token added to the redis", sessionToken)

	return sessionToken, nil
}

func (r *UserRepository) GetToken(ctx context.Context, userId string) (string, error) {
	tokenExists, err := r.redis.Get(ctx, userId).Result()
	if err != nil {
		if err == redis.Nil {
			log.Printf("Token for user ID '%s' does not exist", userId)
			return "", nil
		}
		log.Printf("Failed to get token from Redis for user ID '%s': %v", userId, err)
		return "", err
	}

	if tokenExists == "" {
		log.Printf("Token for user ID '%s' is empty", userId)
		return "", nil
	}

	log.Printf("Retrieved token '%s' for user ID '%s'", tokenExists, userId)

	return tokenExists, nil
}

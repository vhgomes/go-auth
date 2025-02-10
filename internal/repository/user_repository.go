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

func (r *UserRepository) LoginUser(username, password string) (string, error) {
	ctx := context.Background()
	var dbpass, userID string

	queryUser := `SELECT id, password FROM users WHERE username = $1`
	err := r.db.QueryRowContext(ctx, queryUser, username).Scan(&userID, &dbpass)
	if err != nil {
		return "", errors.New("username does not exist")
	}

	if !utils.CheckHash(password, dbpass) {
		return "", errors.New("invalid password")
	}

	existingToken, err := r.GetTokenByUserId(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get existing token: %v", err)
	}

	if existingToken != "" {
		log.Println("Retrieved existing token:", existingToken)
		return existingToken, nil
	}

	sessionToken := utils.GenerateToken(32)
	log.Println("Generated new token:", sessionToken)

	err = r.redis.HSet(ctx, "user_sessions", sessionToken, userID).Err()
	if err != nil {
		return "", fmt.Errorf("failed to store session in Redis: %v", err)
	}

	err = r.redis.Expire(ctx, sessionToken, time.Hour).Err()
	if err != nil {
		return "", fmt.Errorf("failed to set session expiration: %v", err)
	}

	log.Println("Token added to Redis:", sessionToken)
	return sessionToken, nil
}

func (r *UserRepository) GetTokenByUserId(ctx context.Context, userId string) (string, error) {
	log.Println("UserID:", userId)

	var sessionToken string
	iter := r.redis.HScan(ctx, "user_sessions", 0, "*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		if iter.Next(ctx) {
			value := iter.Val()
			if value == userId {
				sessionToken = key
				break
			}
		}
	}

	if err := iter.Err(); err != nil {
		log.Printf("Failed to scan Redis hash: %v", err)
		return "", err
	}

	if sessionToken == "" {
		log.Printf("No session token found for user ID: '%s'", userId)
		return "", nil
	}

	log.Printf("Retrieved token '%s' for user ID '%s'", sessionToken, userId)
	return sessionToken, nil
}

//func (r *UserRepository) GetTokenBySession(ctx context.Context, sessionToken string) (string, error) {
//	log.Println("Session Token:", sessionToken)
//
//	userID, err := r.redis.HGet(ctx, "user_sessions", sessionToken).Result()
//	if err != nil {
//		if err == redis.Nil {
//			log.Printf("No session token found: '%s'", sessionToken)
//			return "", nil
//		}
//		log.Printf("Failed to get token from Redis: '%s', Error %v", sessionToken, err)
//		return "", err
//	}
//
//	if userID == "" {
//		log.Printf("Session token not found")
//		return "", nil
//	}
//
//	log.Printf("Retrieved user ID '%s' for token '%s'", userID, sessionToken)
//	return userID, nil
//}

func (r *UserRepository) LogoutUser(sessionToken string) (string, error) {
	ctx := context.Background()

	userID, err := r.redis.HGet(ctx, "user_sessions", sessionToken).Result()
	if err != nil {
		if err == redis.Nil {
			log.Printf("No session token found: '%s'", sessionToken)
			return "", nil
		}
		return "", err
	}

	_, err = r.redis.HDel(ctx, "user_sessions", sessionToken).Result()
	if err != nil {
		log.Printf("Failed to delete session token: %v", err)
		return "", err
	}

	log.Printf("Session token deleted: '%s' for user ID '%s'", sessionToken, userID)
	return userID, nil
}

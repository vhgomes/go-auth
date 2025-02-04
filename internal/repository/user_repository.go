package repository

import (
	"auth-go/pkg/utils"
	_ "auth-go/pkg/utils"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
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
	var queryUser = `SELECT id, username, password from users WHERE username = $1`
	var dbpass string
	var userId string
	err := r.db.QueryRow(queryUser, username).Scan(&userId, &username, &dbpass)
	if err != nil {
		return "", errors.New("username do not exist")
	}

	if !utils.CheckHash(password, dbpass) {
		return "", errors.New("invalid password")
	}

	sessionToken := utils.GenerateToken(32)

	err = r.redis.Set(ctx, sessionToken, userId, time.Hour).Err()

	if err != nil {
		return "", fmt.Errorf("failed to store session in Redis: %v", err)
	}

	return sessionToken, nil
}

func (r *UserRepository) ValidateSession(sessionToken string) (int, error) {
	ctx := context.Background()
	userID, err := r.redis.Get(ctx, sessionToken).Int()
	if err != nil {
		if err == redis.Nil {
			return 0, errors.New("invalid session token")
		}
		return 0, err
	}

	return userID, nil
}

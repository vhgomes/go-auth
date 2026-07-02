package repository

import (
	"auth-go/pkg/utils"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

var ErrUserAlreadyExists = errors.New("user already exists")

type pgUserRepository struct {
	db    *sql.DB
	redis *redis.Client
}

func NewPgUserRepository(db *sql.DB, redis *redis.Client) *pgUserRepository {
	return &pgUserRepository{
		db:    db,
		redis: redis,
	}
}

func (r *pgUserRepository) RegisterUser(ctx context.Context, username, password string) error {
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	insertQuery := `INSERT INTO users (username, password) VALUES ($1, $2)`
	_, err = r.db.ExecContext(ctx, insertQuery, username, hashedPassword)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrUserAlreadyExists
		}
		log.Printf("failed to insert user: %v", err)
		return fmt.Errorf("failed to insert user: %w", err)
	}
	return nil
}

func (r *pgUserRepository) LoginUser(ctx context.Context, username, password string) (string, error) {
	var dbpass, userID string
	queryUser := `SELECT id, password FROM users WHERE username = $1`
	err := r.db.QueryRowContext(ctx, queryUser, username).Scan(&userID, &dbpass)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("invalid credentials")
		}
		return "", fmt.Errorf("failed to query user: %w", err)
	}

	if !utils.CheckHash(password, dbpass) {
		return "", errors.New("invalid credentials")
	}

	existingToken, err := r.GetTokenByUserId(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get existing token: %w", err)
	}
	if existingToken != "" {
		log.Println("Retrieved existing token for user:", userID)
		return existingToken, nil
	}

	sessionToken := utils.GenerateToken(32)
	if err := r.createSession(ctx, userID, sessionToken); err != nil {
		return "", fmt.Errorf("failed to store session: %w", err)
	}

	log.Println("Generated new session for user:", userID)
	return sessionToken, nil
}

func (r *pgUserRepository) createSession(ctx context.Context, userID, token string) error {
	pipe := r.redis.TxPipeline()
	pipe.Set(ctx, "session:"+token, userID, time.Hour)
	pipe.Set(ctx, "user_session:"+userID, token, time.Hour)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *pgUserRepository) GetTokenByUserId(ctx context.Context, userId string) (string, error) {
	token, err := r.redis.Get(ctx, "user_session:"+userId).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get session token for user %s: %w", userId, err)
	}
	return token, nil
}

func (r *pgUserRepository) GetUserIDByToken(ctx context.Context, sessionToken string) (string, error) {
	userID, err := r.redis.Get(ctx, "session:"+sessionToken).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get user for token: %w", err)
	}
	return userID, nil
}

func (r *pgUserRepository) LogoutUser(ctx context.Context, sessionToken string) (string, error) {
	userID, err := r.GetUserIDByToken(ctx, sessionToken)
	if err != nil {
		return "", fmt.Errorf("failed to resolve session: %w", err)
	}
	if userID == "" {
		log.Printf("No session found for token: '%s'", sessionToken)
		return "", nil
	}

	pipe := r.redis.TxPipeline()
	pipe.Del(ctx, "session:"+sessionToken)
	pipe.Del(ctx, "user_session:"+userID)
	if _, err := pipe.Exec(ctx); err != nil {
		return "", fmt.Errorf("failed to delete session: %w", err)
	}

	log.Printf("Session deleted for user ID '%s'", userID)
	return userID, nil
}

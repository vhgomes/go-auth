package repository

import (
	"auth-go/pkg/utils"
	_ "auth-go/pkg/utils"
	"database/sql"
	"errors"
	"fmt"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
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
	var queryUser = `SELECT username, password from users WHERE username = $1`
	var dbpass string
	err := r.db.QueryRow(queryUser, username).Scan(&username, &dbpass)
	if err != nil {
		return "", errors.New("username do not exist")
	}

	if !utils.CheckHash(password, dbpass) {
		return "", errors.New("invalid password")
	}

	sessionToken := utils.GenerateToken(32)

	return sessionToken, nil
}

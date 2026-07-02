package repository

import "context"

type UserRepository interface {
	RegisterUser(ctx context.Context, username, password string) error
	LoginUser(ctx context.Context, username, password string) (string, error)
	GetTokenByUserId(ctx context.Context, userId string) (string, error)
	GetUserIDByToken(ctx context.Context, sessionToken string) (string, error)
	LogoutUser(ctx context.Context, sessionToken string) (string, error)
}

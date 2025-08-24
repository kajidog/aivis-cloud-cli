package domain

import "context"

type UserRepository interface {
	GetMe(ctx context.Context) (*UserMe, error)
	GetUserByHandle(ctx context.Context, handle string) (*User, error)
}
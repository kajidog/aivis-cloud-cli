package usecase

import (
	"context"
	"fmt"

	"github.com/kajidog/aivis-cloud-cli/client/users/domain"
)

type UserUsecase struct {
	repo domain.UserRepository
}

func NewUserUsecase(repo domain.UserRepository) *UserUsecase {
	return &UserUsecase{
		repo: repo,
	}
}

func (u *UserUsecase) GetMe(ctx context.Context) (*domain.UserMe, error) {
	return u.repo.GetMe(ctx)
}

func (u *UserUsecase) GetUserByHandle(ctx context.Context, handle string) (*domain.User, error) {
	if handle == "" {
		return nil, fmt.Errorf("handle is required")
	}
	
	return u.repo.GetUserByHandle(ctx, handle)
}
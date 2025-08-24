package infrastructure

import (
	"context"
	"fmt"

	"github.com/kajidog/aivis-cloud-cli/client/common/http"
	"github.com/kajidog/aivis-cloud-cli/client/users/domain"
)

type UserAPI struct {
	client http.HTTPClient
}

func NewUserAPI(client http.HTTPClient) *UserAPI {
	return &UserAPI{
		client: client,
	}
}

func (a *UserAPI) GetMe(ctx context.Context) (*domain.UserMe, error) {
	endpoint := "/v1/users/me"
	
	var response domain.UserMe
	err := a.client.Get(ctx, endpoint, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}
	
	return &response, nil
}

func (a *UserAPI) GetUserByHandle(ctx context.Context, handle string) (*domain.User, error) {
	endpoint := fmt.Sprintf("/v1/users/%s", handle)
	
	var response domain.User
	err := a.client.Get(ctx, endpoint, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by handle %s: %w", handle, err)
	}
	
	return &response, nil
}
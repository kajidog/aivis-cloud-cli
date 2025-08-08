package infrastructure

import (
	"context"
	"net/url"
	"strconv"

	"github.com/kajidog/aiviscloud-mcp/client/common/http"
	"github.com/kajidog/aiviscloud-mcp/client/models/domain"
)

// ModelAPIRepository implements the ModelRepository interface using HTTP API calls
type ModelAPIRepository struct {
	httpClient *http.Client
}

// NewModelAPIRepository creates a new model API repository
func NewModelAPIRepository(httpClient *http.Client) *ModelAPIRepository {
	return &ModelAPIRepository{
		httpClient: httpClient,
	}
}

// SearchModels searches for available models
func (r *ModelAPIRepository) SearchModels(ctx context.Context, request *domain.ModelSearchRequest) (*domain.ModelSearchResponse, error) {
	query := url.Values{}

	// Add query parameters
	if request.Query != nil {
		query.Set("query", *request.Query)
	}

	if len(request.Tags) > 0 {
		for _, tag := range request.Tags {
			query.Add("tags", tag)
		}
	}

	if request.Author != nil {
		query.Set("author", *request.Author)
	}

	if request.Language != nil {
		query.Set("language", *request.Language)
	}

	if request.IsPublic != nil {
		query.Set("is_public", strconv.FormatBool(*request.IsPublic))
	}

	if request.ModelType != nil {
		query.Set("model_type", *request.ModelType)
	}

	if request.Page != nil {
		query.Set("page", strconv.Itoa(*request.Page))
	}

	if request.PageSize != nil {
		query.Set("page_size", strconv.Itoa(*request.PageSize))
	}

	if request.SortBy != nil {
		query.Set("sort_by", *request.SortBy)
	}

	if request.SortOrder != nil {
		query.Set("sort_order", *request.SortOrder)
	}

	httpReq := &http.Request{
		Method: "GET",
		Path:   "/v1/aivm-models/search",
		Query:  query,
	}

	var response domain.ModelSearchResponse
	err := r.httpClient.DoJSON(ctx, httpReq, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// GetModel retrieves a specific model by UUID
func (r *ModelAPIRepository) GetModel(ctx context.Context, modelUUID string) (*domain.Model, error) {
	httpReq := &http.Request{
		Method: "GET",
		Path:   "/v1/aivm-models/" + modelUUID,
	}

	var model domain.Model
	err := r.httpClient.DoJSON(ctx, httpReq, &model)
	if err != nil {
		return nil, err
	}

	return &model, nil
}

// GetModelSpeakers retrieves speakers for a specific model
func (r *ModelAPIRepository) GetModelSpeakers(ctx context.Context, modelUUID string) ([]domain.Speaker, error) {
	httpReq := &http.Request{
		Method: "GET",
		Path:   "/v1/aivm-models/" + modelUUID + "/speakers",
	}

	var speakers []domain.Speaker
	err := r.httpClient.DoJSON(ctx, httpReq, &speakers)
	if err != nil {
		return nil, err
	}

	return speakers, nil
}

package infrastructure

import (
	"context"
	"net/url"
	"strconv"

	"github.com/kajidog/aivis-cloud-cli/client/common/http"
	"github.com/kajidog/aivis-cloud-cli/client/models/domain"
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

	// Add query parameters - using actual API parameter names
	if request.Query != nil {
		query.Set("q", *request.Query)
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
		query.Set("public", strconv.FormatBool(*request.IsPublic))
	}

	if request.ModelType != nil {
		query.Set("model_type", *request.ModelType)
	}

	// Use limit/offset instead of page/page_size
	if request.PageSize != nil {
		query.Set("limit", strconv.Itoa(*request.PageSize))
	}

	if request.Page != nil && request.PageSize != nil {
		offset := (*request.Page - 1) * *request.PageSize
		query.Set("offset", strconv.Itoa(offset))
	}

	if request.SortBy != nil {
		query.Set("sort", *request.SortBy)
	}

	if request.SortOrder != nil {
		query.Set("sort_order", *request.SortOrder)
	}

	httpReq := &http.Request{
		Method: "GET",
		Path:   "/v1/aivm-models/search",
		Query:  query,
	}

	var apiResponse struct {
		Total      int64           `json:"total"`
		AivmModels []domain.Model `json:"aivm_models"`
	}
	
	err := r.httpClient.DoJSON(ctx, httpReq, &apiResponse)
	if err != nil {
		return nil, err
	}

	// Calculate pagination information
	pageSize := 10 // Default page size
	if request.PageSize != nil {
		pageSize = *request.PageSize
	}
	
	currentPage := 1
	if request.Page != nil {
		currentPage = *request.Page
	}
	
	totalPages := int(apiResponse.Total) / pageSize
	if int(apiResponse.Total)%pageSize > 0 {
		totalPages++
	}
	
	pagination := domain.Pagination{
		CurrentPage:  currentPage,
		PageSize:     pageSize,
		TotalPages:   totalPages,
		TotalResults: apiResponse.Total,
		HasNext:      currentPage < totalPages,
		HasPrevious:  currentPage > 1,
	}

	response := &domain.ModelSearchResponse{
		Models:     apiResponse.AivmModels,
		Total:      apiResponse.Total,
		Pagination: pagination,
	}

	return response, nil
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

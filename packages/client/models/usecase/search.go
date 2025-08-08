package usecase

import (
	"context"

	"github.com/kajidog/aiviscloud-mcp/client/models/domain"
)

// ModelSearcher handles model search use cases
type ModelSearcher struct {
	repository domain.ModelRepository
}

// NewModelSearcher creates a new model searcher
func NewModelSearcher(repository domain.ModelRepository) *ModelSearcher {
	return &ModelSearcher{
		repository: repository,
	}
}

// SearchModels searches for available models
func (s *ModelSearcher) SearchModels(ctx context.Context, request *domain.ModelSearchRequest) (*domain.ModelSearchResponse, error) {
	if err := s.validateSearchRequest(request); err != nil {
		return nil, err
	}

	return s.repository.SearchModels(ctx, request)
}

// GetModel retrieves a specific model by UUID
func (s *ModelSearcher) GetModel(ctx context.Context, modelUUID string) (*domain.Model, error) {
	if modelUUID == "" {
		return nil, &ValidationError{Field: "ModelUUID", Message: "Model UUID is required"}
	}

	return s.repository.GetModel(ctx, modelUUID)
}

// GetModelSpeakers retrieves speakers for a specific model
func (s *ModelSearcher) GetModelSpeakers(ctx context.Context, modelUUID string) ([]domain.Speaker, error) {
	if modelUUID == "" {
		return nil, &ValidationError{Field: "ModelUUID", Message: "Model UUID is required"}
	}

	return s.repository.GetModelSpeakers(ctx, modelUUID)
}

// SearchPublicModels searches for public models only
func (s *ModelSearcher) SearchPublicModels(ctx context.Context, query string) (*domain.ModelSearchResponse, error) {
	request := domain.NewModelSearchRequestBuilder().
		WithQuery(query).
		WithPublicOnly().
		Build()

	return s.SearchModels(ctx, request)
}

// SearchModelsByAuthor searches for models by a specific author
func (s *ModelSearcher) SearchModelsByAuthor(ctx context.Context, author string) (*domain.ModelSearchResponse, error) {
	request := domain.NewModelSearchRequestBuilder().
		WithAuthor(author).
		WithPublicOnly().
		Build()

	return s.SearchModels(ctx, request)
}

// SearchModelsByTags searches for models with specific tags
func (s *ModelSearcher) SearchModelsByTags(ctx context.Context, tags ...string) (*domain.ModelSearchResponse, error) {
	request := domain.NewModelSearchRequestBuilder().
		WithTags(tags...).
		WithPublicOnly().
		Build()

	return s.SearchModels(ctx, request)
}

// GetPopularModels retrieves popular models sorted by download count
func (s *ModelSearcher) GetPopularModels(ctx context.Context, limit int) (*domain.ModelSearchResponse, error) {
	request := domain.NewModelSearchRequestBuilder().
		WithPublicOnly().
		WithPageSize(limit).
		SortByDownloadCount().
		Descending().
		Build()

	return s.SearchModels(ctx, request)
}

// GetRecentModels retrieves recently updated models
func (s *ModelSearcher) GetRecentModels(ctx context.Context, limit int) (*domain.ModelSearchResponse, error) {
	request := domain.NewModelSearchRequestBuilder().
		WithPublicOnly().
		WithPageSize(limit).
		SortByUpdatedAt().
		Descending().
		Build()

	return s.SearchModels(ctx, request)
}

// GetTopRatedModels retrieves top-rated models
func (s *ModelSearcher) GetTopRatedModels(ctx context.Context, limit int) (*domain.ModelSearchResponse, error) {
	request := domain.NewModelSearchRequestBuilder().
		WithPublicOnly().
		WithPageSize(limit).
		SortByRating().
		Descending().
		Build()

	return s.SearchModels(ctx, request)
}

// validateSearchRequest validates a model search request
func (s *ModelSearcher) validateSearchRequest(request *domain.ModelSearchRequest) error {
	if request.Page != nil && *request.Page < 1 {
		return &ValidationError{Field: "Page", Message: "Page must be greater than 0"}
	}

	if request.PageSize != nil && (*request.PageSize < 1 || *request.PageSize > 100) {
		return &ValidationError{Field: "PageSize", Message: "PageSize must be between 1 and 100"}
	}

	if request.SortOrder != nil && *request.SortOrder != "asc" && *request.SortOrder != "desc" {
		return &ValidationError{Field: "SortOrder", Message: "SortOrder must be 'asc' or 'desc'"}
	}

	validSortFields := map[string]bool{
		"created_at":     true,
		"updated_at":     true,
		"name":           true,
		"download_count": true,
		"usage_count":    true,
		"rating":         true,
	}

	if request.SortBy != nil && !validSortFields[*request.SortBy] {
		return &ValidationError{Field: "SortBy", Message: "Invalid sort field"}
	}

	return nil
}

// ValidationError represents a request validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

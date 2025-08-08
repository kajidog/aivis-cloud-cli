package domain

import "context"

// ModelRepository defines the interface for model operations
type ModelRepository interface {
	// SearchModels searches for available models
	SearchModels(ctx context.Context, request *ModelSearchRequest) (*ModelSearchResponse, error)

	// GetModel retrieves a specific model by UUID
	GetModel(ctx context.Context, modelUUID string) (*Model, error)

	// GetModelSpeakers retrieves speakers for a specific model
	GetModelSpeakers(ctx context.Context, modelUUID string) ([]Speaker, error)
}

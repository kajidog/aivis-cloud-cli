package domain

import "time"

// ModelSearchRequest represents a model search request
type ModelSearchRequest struct {
	// Search query parameters
	Query     *string  `json:"query,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	Author    *string  `json:"author,omitempty"`
	Language  *string  `json:"language,omitempty"`
	IsPublic  *bool    `json:"is_public,omitempty"`
	ModelType *string  `json:"model_type,omitempty"`

	// Pagination
	Page     *int `json:"page,omitempty"`
	PageSize *int `json:"page_size,omitempty"`

	// Sorting
	SortBy    *string `json:"sort_by,omitempty"`
	SortOrder *string `json:"sort_order,omitempty"`
}

// ModelSearchResponse represents a model search response
type ModelSearchResponse struct {
	Models     []Model    `json:"models"`
	Pagination Pagination `json:"pagination"`
}

// Model represents an AI voice model
type Model struct {
	UUID        string    `json:"uuid"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Author      string    `json:"author"`
	Tags        []string  `json:"tags"`
	Language    string    `json:"language"`
	IsPublic    bool      `json:"is_public"`
	ModelType   string    `json:"model_type"`
	Version     string    `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Model-specific information
	Speakers   []Speaker `json:"speakers,omitempty"`
	SampleRate int       `json:"sample_rate,omitempty"`
	VocabSize  int       `json:"vocab_size,omitempty"`
	ModelSize  int64     `json:"model_size,omitempty"`

	// Usage statistics
	DownloadCount int     `json:"download_count,omitempty"`
	UsageCount    int     `json:"usage_count,omitempty"`
	Rating        float64 `json:"rating,omitempty"`

	// License and attribution
	License     string `json:"license,omitempty"`
	Attribution string `json:"attribution,omitempty"`
}

// Speaker represents a speaker in a multi-speaker model
type Speaker struct {
	UUID        string  `json:"uuid"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Gender      string  `json:"gender,omitempty"`
	Age         int     `json:"age,omitempty"`
	Language    string  `json:"language,omitempty"`
	IsDefault   bool    `json:"is_default"`
	Styles      []Style `json:"styles,omitempty"`
}

// Style represents a speaking style for a speaker
type Style struct {
	LocalID     int    `json:"local_id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	IsDefault   bool   `json:"is_default"`
}

// Pagination represents pagination information
type Pagination struct {
	CurrentPage  int   `json:"current_page"`
	PageSize     int   `json:"page_size"`
	TotalPages   int   `json:"total_pages"`
	TotalResults int64 `json:"total_results"`
	HasNext      bool  `json:"has_next"`
	HasPrevious  bool  `json:"has_previous"`
}

// ModelSearchRequestBuilder helps build model search requests with method chaining
type ModelSearchRequestBuilder struct {
	request *ModelSearchRequest
}

// NewModelSearchRequestBuilder creates a new model search request builder
func NewModelSearchRequestBuilder() *ModelSearchRequestBuilder {
	return &ModelSearchRequestBuilder{
		request: &ModelSearchRequest{},
	}
}

// WithQuery sets the search query
func (b *ModelSearchRequestBuilder) WithQuery(query string) *ModelSearchRequestBuilder {
	b.request.Query = &query
	return b
}

// WithTags sets the tags to filter by
func (b *ModelSearchRequestBuilder) WithTags(tags ...string) *ModelSearchRequestBuilder {
	b.request.Tags = tags
	return b
}

// WithAuthor sets the author to filter by
func (b *ModelSearchRequestBuilder) WithAuthor(author string) *ModelSearchRequestBuilder {
	b.request.Author = &author
	return b
}

// WithLanguage sets the language to filter by
func (b *ModelSearchRequestBuilder) WithLanguage(language string) *ModelSearchRequestBuilder {
	b.request.Language = &language
	return b
}

// WithPublicOnly filters for public models only
func (b *ModelSearchRequestBuilder) WithPublicOnly() *ModelSearchRequestBuilder {
	isPublic := true
	b.request.IsPublic = &isPublic
	return b
}

// WithPrivateOnly filters for private models only
func (b *ModelSearchRequestBuilder) WithPrivateOnly() *ModelSearchRequestBuilder {
	isPublic := false
	b.request.IsPublic = &isPublic
	return b
}

// WithModelType sets the model type to filter by
func (b *ModelSearchRequestBuilder) WithModelType(modelType string) *ModelSearchRequestBuilder {
	b.request.ModelType = &modelType
	return b
}

// WithPage sets the page number for pagination
func (b *ModelSearchRequestBuilder) WithPage(page int) *ModelSearchRequestBuilder {
	b.request.Page = &page
	return b
}

// WithPageSize sets the page size for pagination
func (b *ModelSearchRequestBuilder) WithPageSize(pageSize int) *ModelSearchRequestBuilder {
	b.request.PageSize = &pageSize
	return b
}

// WithSortBy sets the field to sort by
func (b *ModelSearchRequestBuilder) WithSortBy(sortBy string) *ModelSearchRequestBuilder {
	b.request.SortBy = &sortBy
	return b
}

// WithSortOrder sets the sort order (asc or desc)
func (b *ModelSearchRequestBuilder) WithSortOrder(sortOrder string) *ModelSearchRequestBuilder {
	b.request.SortOrder = &sortOrder
	return b
}

// SortByCreatedAt sorts by creation date
func (b *ModelSearchRequestBuilder) SortByCreatedAt() *ModelSearchRequestBuilder {
	sortBy := "created_at"
	b.request.SortBy = &sortBy
	return b
}

// SortByUpdatedAt sorts by update date
func (b *ModelSearchRequestBuilder) SortByUpdatedAt() *ModelSearchRequestBuilder {
	sortBy := "updated_at"
	b.request.SortBy = &sortBy
	return b
}

// SortByName sorts by name
func (b *ModelSearchRequestBuilder) SortByName() *ModelSearchRequestBuilder {
	sortBy := "name"
	b.request.SortBy = &sortBy
	return b
}

// SortByDownloadCount sorts by download count
func (b *ModelSearchRequestBuilder) SortByDownloadCount() *ModelSearchRequestBuilder {
	sortBy := "download_count"
	b.request.SortBy = &sortBy
	return b
}

// SortByUsageCount sorts by usage count
func (b *ModelSearchRequestBuilder) SortByUsageCount() *ModelSearchRequestBuilder {
	sortBy := "usage_count"
	b.request.SortBy = &sortBy
	return b
}

// SortByRating sorts by rating
func (b *ModelSearchRequestBuilder) SortByRating() *ModelSearchRequestBuilder {
	sortBy := "rating"
	b.request.SortBy = &sortBy
	return b
}

// Ascending sets sort order to ascending
func (b *ModelSearchRequestBuilder) Ascending() *ModelSearchRequestBuilder {
	sortOrder := "asc"
	b.request.SortOrder = &sortOrder
	return b
}

// Descending sets sort order to descending
func (b *ModelSearchRequestBuilder) Descending() *ModelSearchRequestBuilder {
	sortOrder := "desc"
	b.request.SortOrder = &sortOrder
	return b
}

// Build returns the built model search request
func (b *ModelSearchRequestBuilder) Build() *ModelSearchRequest {
	return b.request
}

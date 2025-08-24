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
	Models     []Model    `json:"aivm_models"`
	Total      int64      `json:"total"`
	Pagination Pagination `json:"pagination"`
}

// Model represents an AI voice model
type Model struct {
	UUID             string    `json:"aivm_model_uuid"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	DetailedDesc     string    `json:"detailed_description,omitempty"`
	Category         string    `json:"category,omitempty"`
	VoiceTimbre      string    `json:"voice_timbre,omitempty"`
	Visibility       string    `json:"visibility,omitempty"`
	IsTagLocked      bool      `json:"is_tag_locked,omitempty"`
	TotalDownloadCount int     `json:"total_download_count,omitempty"`
	LikeCount        int       `json:"like_count,omitempty"`
	IsLiked          bool      `json:"is_liked,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	// User information
	User *User `json:"user,omitempty"`

	// Model files
	ModelFiles []ModelFile `json:"model_files,omitempty"`

	// Tags
	Tags []Tag `json:"tags,omitempty"`

	// Speakers
	Speakers []Speaker `json:"speakers,omitempty"`

	// Legacy fields for backward compatibility
	Author        string  `json:"author"`
	Language      string  `json:"language"`
	IsPublic      bool    `json:"is_public"`
	ModelType     string  `json:"model_type"`
	Version       string  `json:"version"`
	SampleRate    int     `json:"sample_rate,omitempty"`
	VocabSize     int     `json:"vocab_size,omitempty"`
	ModelSize     int64   `json:"model_size,omitempty"`
	DownloadCount int     `json:"download_count,omitempty"`
	UsageCount    int     `json:"usage_count,omitempty"`
	Rating        float64 `json:"rating,omitempty"`
	License       string  `json:"license,omitempty"`
	Attribution   string  `json:"attribution,omitempty"`
}

// User represents a user who owns models
type User struct {
	Handle       string        `json:"handle"`
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	IconURL      string        `json:"icon_url"`
	AccountType  string        `json:"account_type"`
	AccountStatus string       `json:"account_status"`
	SocialLinks  []SocialLink  `json:"social_links,omitempty"`
}

// SocialLink represents a social media link
type SocialLink struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

// ModelFile represents a model file
type ModelFile struct {
	UUID             string    `json:"aivm_model_uuid"`
	ManifestVersion  string    `json:"manifest_version"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	Creators         []string  `json:"creators"`
	LicenseType      string    `json:"license_type"`
	LicenseText      string    `json:"license_text"`
	ModelType        string    `json:"model_type"`
	ModelArchitecture string   `json:"model_architecture"`
	ModelFormat      string    `json:"model_format"`
	TrainingEpochs   int       `json:"training_epochs"`
	TrainingSteps    int       `json:"training_steps"`
	Version          string    `json:"version"`
	FileSize         int64     `json:"file_size"`
	Checksum         string    `json:"checksum"`
	DownloadCount    int       `json:"download_count"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// Tag represents a model tag
type Tag struct {
	Name string `json:"name"`
}

// Speaker represents a speaker in a multi-speaker model
type Speaker struct {
	UUID               string   `json:"aivm_speaker_uuid"`
	Name               string   `json:"name"`
	IconURL            string   `json:"icon_url,omitempty"`
	SupportedLanguages []string `json:"supported_languages,omitempty"`
	LocalID            int      `json:"local_id"`
	Styles             []Style  `json:"styles,omitempty"`
	
	// Legacy fields for backward compatibility
	Description string `json:"description,omitempty"`
	Gender      string `json:"gender,omitempty"`
	Age         int    `json:"age,omitempty"`
	Language    string `json:"language,omitempty"`
	IsDefault   bool   `json:"is_default"`
}

// Style represents a speaking style for a speaker
type Style struct {
	Name         string        `json:"name"`
	IconURL      string        `json:"icon_url,omitempty"`
	LocalID      int           `json:"local_id"`
	VoiceSamples []VoiceSample `json:"voice_samples,omitempty"`
	
	// Legacy fields for backward compatibility
	Description string `json:"description,omitempty"`
	IsDefault   bool   `json:"is_default"`
}

// VoiceSample represents a voice sample for a style
type VoiceSample struct {
	AudioURL   string `json:"audio_url"`
	Transcript string `json:"transcript"`
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

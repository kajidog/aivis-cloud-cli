package domain

import (
	"context"
	"time"
)

// TTSHistory represents a TTS synthesis history record
type TTSHistory struct {
	// Sequential ID starting from 1
	ID int `json:"id"`
	
	// Internal UUID for session management
	InternalUUID string `json:"internal_uuid"`
	
	// Original TTS request
	Request *TTSRequest `json:"request"`
	
	// Audio file information
	FilePath     string `json:"file_path"`
	FileFormat   string `json:"file_format"` // wav, mp3, etc.
	FileSizeBytes int64  `json:"file_size_bytes"`
	
	// Metadata
	CreatedAt time.Time `json:"created_at"`
	Text      string    `json:"text"` // Copy of request text for quick access
	ModelUUID string    `json:"model_uuid"` // Copy of model UUID for quick access
	
	// Optional billing info if available
	Credits *float64 `json:"credits,omitempty"`
}

// TTSHistorySearchRequest represents search criteria for TTS history
type TTSHistorySearchRequest struct {
	// Pagination
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
	
	// Filters
	ModelUUID    *string    `json:"model_uuid,omitempty"`
	TextContains *string    `json:"text_contains,omitempty"`
	StartDate    *time.Time `json:"start_date,omitempty"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	
	// Sorting
	SortBy    string `json:"sort_by,omitempty"`    // "id", "created_at", "text"
	SortOrder string `json:"sort_order,omitempty"` // "asc", "desc"
}

// TTSHistoryListResponse represents the response for history list requests
type TTSHistoryListResponse struct {
	Histories []*TTSHistory `json:"histories"`
	Total     int           `json:"total"`
	Limit     int           `json:"limit"`
	Offset    int           `json:"offset"`
	HasMore   bool          `json:"has_more"`
}

// TTSHistoryRepository defines the interface for TTS history storage operations
type TTSHistoryRepository interface {
	// Save stores a new history record and returns the assigned ID
	Save(ctx context.Context, history *TTSHistory) (int, error)
	
	// GetByID retrieves a history record by its sequential ID
	GetByID(ctx context.Context, id int) (*TTSHistory, error)
	
	// List retrieves history records based on search criteria
	List(ctx context.Context, request *TTSHistorySearchRequest) (*TTSHistoryListResponse, error)
	
	// Delete removes a history record by ID (including associated files)
	Delete(ctx context.Context, id int) error
	
	// DeleteMultiple removes multiple history records by IDs
	DeleteMultiple(ctx context.Context, ids []int) error
	
	// Clear removes all history records
	Clear(ctx context.Context) error
	
	// Count returns the total number of history records
	Count(ctx context.Context) (int, error)
	
	// GetNextID returns the next available sequential ID
	GetNextID(ctx context.Context) (int, error)
	
	// Cleanup removes old records based on configuration (max count, age, etc.)
	Cleanup(ctx context.Context, maxCount int, maxAge *time.Duration) error
}

// TTSHistoryManager handles business logic for history management
type TTSHistoryManager interface {
	// SaveHistory saves a TTS request and response as history
	SaveHistory(ctx context.Context, request *TTSRequest, filePath string, credits *float64) (*TTSHistory, error)
	
	// GetHistory retrieves history by ID
	GetHistory(ctx context.Context, id int) (*TTSHistory, error)
	
	// ListHistory lists history records with pagination and filtering
	ListHistory(ctx context.Context, request *TTSHistorySearchRequest) (*TTSHistoryListResponse, error)
	
	// PlayHistory replays audio from history
	PlayHistory(ctx context.Context, id int, playbackOptions *PlaybackRequest) error
	
	// DeleteHistory removes a history record
	DeleteHistory(ctx context.Context, id int) error
	
	// ClearHistory removes all history records
	ClearHistory(ctx context.Context) error
	
	// CleanupHistory removes old records based on configuration
	CleanupHistory(ctx context.Context) error
}

// TTSHistoryStats represents statistics about TTS history
type TTSHistoryStats struct {
	TotalRecords  int     `json:"total_records"`
	TotalFileSize int64   `json:"total_file_size"`
	TotalCredits  float64 `json:"total_credits"`
}

// NewTTSHistorySearchRequest creates a new search request builder
func NewTTSHistorySearchRequest() *TTSHistorySearchRequestBuilder {
	return &TTSHistorySearchRequestBuilder{
		request: &TTSHistorySearchRequest{
			Limit:     10, // Default limit
			Offset:    0,
			SortBy:    "id",
			SortOrder: "desc", // Latest first by default
		},
	}
}

// TTSHistorySearchRequestBuilder helps build history search requests
type TTSHistorySearchRequestBuilder struct {
	request *TTSHistorySearchRequest
}

// WithLimit sets the maximum number of results to return
func (b *TTSHistorySearchRequestBuilder) WithLimit(limit int) *TTSHistorySearchRequestBuilder {
	b.request.Limit = limit
	return b
}

// WithOffset sets the number of results to skip
func (b *TTSHistorySearchRequestBuilder) WithOffset(offset int) *TTSHistorySearchRequestBuilder {
	b.request.Offset = offset
	return b
}

// WithModelUUID filters by model UUID
func (b *TTSHistorySearchRequestBuilder) WithModelUUID(modelUUID string) *TTSHistorySearchRequestBuilder {
	b.request.ModelUUID = &modelUUID
	return b
}

// WithTextContains filters by text content
func (b *TTSHistorySearchRequestBuilder) WithTextContains(text string) *TTSHistorySearchRequestBuilder {
	b.request.TextContains = &text
	return b
}

// WithDateRange sets the date range filter
func (b *TTSHistorySearchRequestBuilder) WithDateRange(start, end time.Time) *TTSHistorySearchRequestBuilder {
	b.request.StartDate = &start
	b.request.EndDate = &end
	return b
}

// WithSorting sets the sorting criteria
func (b *TTSHistorySearchRequestBuilder) WithSorting(sortBy, sortOrder string) *TTSHistorySearchRequestBuilder {
	b.request.SortBy = sortBy
	b.request.SortOrder = sortOrder
	return b
}

// Build returns the constructed search request
func (b *TTSHistorySearchRequestBuilder) Build() *TTSHistorySearchRequest {
	return b.request
}

// GetFileExtensionFromFormat returns the file extension for a given output format
func GetFileExtensionFromFormat(format OutputFormat) string {
	switch format {
	case OutputFormatWAV:
		return ".wav"
	case OutputFormatMP3:
		return ".mp3"
	case OutputFormatFLAC:
		return ".flac"
	case OutputFormatAAC:
		return ".aac"
	case OutputFormatOpus:
		return ".opus"
	default:
		return ".wav" // Default to WAV
	}
}

// GetFormatFromFilePath extracts the format from file path extension
func GetFormatFromFilePath(filePath string) string {
	if len(filePath) < 4 {
		return "wav"
	}
	
	ext := filePath[len(filePath)-4:]
	switch ext {
	case ".wav":
		return "wav"
	case ".mp3":
		return "mp3"
	case ".aac":
		return "aac"
	case "opus": // .opus
		return "opus"
	case "flac": // .flac
		return "flac"
	default:
		return "wav"
	}
}
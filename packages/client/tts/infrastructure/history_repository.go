package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kajidog/aivis-cloud-cli/client/tts/domain"
)

// FileHistoryRepository implements TTSHistoryRepository using local file system
type FileHistoryRepository struct {
	basePath string
}

// historyMetadata holds the metadata for all history records
type historyMetadata struct {
	Records []*domain.TTSHistory `json:"records"`
}

// idCounter holds the next available ID
type idCounter struct {
	NextID int `json:"next_id"`
}

// NewFileHistoryRepository creates a new file-based history repository
func NewFileHistoryRepository(basePath string) *FileHistoryRepository {
	return &FileHistoryRepository{
		basePath: basePath,
	}
}

// ensureDirectories creates necessary directories if they don't exist
func (r *FileHistoryRepository) ensureDirectories() error {
	audioDir := filepath.Join(r.basePath, "audio")
	return os.MkdirAll(audioDir, 0755)
}

// getMetadataPath returns the path to the metadata file
func (r *FileHistoryRepository) getMetadataPath() string {
	return filepath.Join(r.basePath, "metadata.json")
}

// getCounterPath returns the path to the ID counter file
func (r *FileHistoryRepository) getCounterPath() string {
	return filepath.Join(r.basePath, "counter.json")
}

// loadMetadata loads the metadata from file
func (r *FileHistoryRepository) loadMetadata() (*historyMetadata, error) {
	metadataPath := r.getMetadataPath()
	
	// If file doesn't exist, return empty metadata
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		return &historyMetadata{Records: []*domain.TTSHistory{}}, nil
	}
	
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}
	
	var metadata historyMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}
	
	return &metadata, nil
}

// saveMetadata saves the metadata to file
func (r *FileHistoryRepository) saveMetadata(metadata *historyMetadata) error {
	if err := r.ensureDirectories(); err != nil {
		return err
	}
	
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	
	metadataPath := r.getMetadataPath()
	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}
	
	return nil
}

// loadCounter loads the ID counter from file
func (r *FileHistoryRepository) loadCounter() (*idCounter, error) {
	counterPath := r.getCounterPath()
	
	// If file doesn't exist, start from ID 1
	if _, err := os.Stat(counterPath); os.IsNotExist(err) {
		return &idCounter{NextID: 1}, nil
	}
	
	data, err := os.ReadFile(counterPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read counter file: %w", err)
	}
	
	var counter idCounter
	if err := json.Unmarshal(data, &counter); err != nil {
		return nil, fmt.Errorf("failed to unmarshal counter: %w", err)
	}
	
	return &counter, nil
}

// saveCounter saves the ID counter to file
func (r *FileHistoryRepository) saveCounter(counter *idCounter) error {
	if err := r.ensureDirectories(); err != nil {
		return err
	}
	
	data, err := json.MarshalIndent(counter, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal counter: %w", err)
	}
	
	counterPath := r.getCounterPath()
	if err := os.WriteFile(counterPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write counter file: %w", err)
	}
	
	return nil
}

// GetNextID returns the next available sequential ID
func (r *FileHistoryRepository) GetNextID(ctx context.Context) (int, error) {
	counter, err := r.loadCounter()
	if err != nil {
		return 0, err
	}
	
	nextID := counter.NextID
	counter.NextID++
	
	if err := r.saveCounter(counter); err != nil {
		return 0, err
	}
	
	return nextID, nil
}

// Save stores a new history record and returns the assigned ID
func (r *FileHistoryRepository) Save(ctx context.Context, history *domain.TTSHistory) (int, error) {
	if history.ID == 0 {
		id, err := r.GetNextID(ctx)
		if err != nil {
			return 0, err
		}
		history.ID = id
	}
	
	// Generate internal UUID if not present
	if history.InternalUUID == "" {
		history.InternalUUID = uuid.New().String()
	}
	
	// Set creation time if not set
	if history.CreatedAt.IsZero() {
		history.CreatedAt = time.Now()
	}
	
	// Load existing metadata
	metadata, err := r.loadMetadata()
	if err != nil {
		return 0, err
	}
	
	// Add new record
	metadata.Records = append(metadata.Records, history)
	
	// Save metadata
	if err := r.saveMetadata(metadata); err != nil {
		return 0, err
	}
	
	return history.ID, nil
}

// GetByID retrieves a history record by its sequential ID
func (r *FileHistoryRepository) GetByID(ctx context.Context, id int) (*domain.TTSHistory, error) {
	metadata, err := r.loadMetadata()
	if err != nil {
		return nil, err
	}
	
	for _, record := range metadata.Records {
		if record.ID == id {
			return record, nil
		}
	}
	
	return nil, fmt.Errorf("history record with ID %d not found", id)
}

// List retrieves history records based on search criteria
func (r *FileHistoryRepository) List(ctx context.Context, request *domain.TTSHistorySearchRequest) (*domain.TTSHistoryListResponse, error) {
	metadata, err := r.loadMetadata()
	if err != nil {
		return nil, err
	}
	
	// Apply filters
	filtered := r.filterRecords(metadata.Records, request)
	
	// Apply sorting
	r.sortRecords(filtered, request)
	
	total := len(filtered)
	
	// Apply pagination
	start := request.Offset
	end := start + request.Limit
	
	if start > len(filtered) {
		start = len(filtered)
	}
	if end > len(filtered) {
		end = len(filtered)
	}
	
	paginatedRecords := filtered[start:end]
	hasMore := end < total
	
	return &domain.TTSHistoryListResponse{
		Histories: paginatedRecords,
		Total:     total,
		Limit:     request.Limit,
		Offset:    request.Offset,
		HasMore:   hasMore,
	}, nil
}

// filterRecords applies search filters to the records
func (r *FileHistoryRepository) filterRecords(records []*domain.TTSHistory, request *domain.TTSHistorySearchRequest) []*domain.TTSHistory {
	var filtered []*domain.TTSHistory
	
	for _, record := range records {
		// Filter by model UUID
		if request.ModelUUID != nil && record.ModelUUID != *request.ModelUUID {
			continue
		}
		
		// Filter by text content
		if request.TextContains != nil {
			if !strings.Contains(strings.ToLower(record.Text), strings.ToLower(*request.TextContains)) {
				continue
			}
		}
		
		// Filter by date range
		if request.StartDate != nil && record.CreatedAt.Before(*request.StartDate) {
			continue
		}
		if request.EndDate != nil && record.CreatedAt.After(*request.EndDate) {
			continue
		}
		
		filtered = append(filtered, record)
	}
	
	return filtered
}

// sortRecords applies sorting to the records
func (r *FileHistoryRepository) sortRecords(records []*domain.TTSHistory, request *domain.TTSHistorySearchRequest) {
	sortBy := request.SortBy
	if sortBy == "" {
		sortBy = "id"
	}
	
	sortOrder := request.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
	}
	
	sort.Slice(records, func(i, j int) bool {
		var less bool
		
		switch sortBy {
		case "id":
			less = records[i].ID < records[j].ID
		case "created_at":
			less = records[i].CreatedAt.Before(records[j].CreatedAt)
		case "text":
			less = strings.ToLower(records[i].Text) < strings.ToLower(records[j].Text)
		default:
			less = records[i].ID < records[j].ID
		}
		
		if sortOrder == "desc" {
			return !less
		}
		return less
	})
}

// Delete removes a history record by ID (including associated files)
func (r *FileHistoryRepository) Delete(ctx context.Context, id int) error {
	metadata, err := r.loadMetadata()
	if err != nil {
		return err
	}
	
	var updatedRecords []*domain.TTSHistory
	var deletedRecord *domain.TTSHistory
	
	for _, record := range metadata.Records {
		if record.ID == id {
			deletedRecord = record
		} else {
			updatedRecords = append(updatedRecords, record)
		}
	}
	
	if deletedRecord == nil {
		return fmt.Errorf("history record with ID %d not found", id)
	}
	
	// Delete associated audio file
	if deletedRecord.FilePath != "" {
		if err := os.Remove(deletedRecord.FilePath); err != nil && !os.IsNotExist(err) {
			// Log error but don't fail the operation
			// The metadata cleanup should succeed even if file deletion fails
		}
	}
	
	// Update metadata
	metadata.Records = updatedRecords
	return r.saveMetadata(metadata)
}

// DeleteMultiple removes multiple history records by IDs
func (r *FileHistoryRepository) DeleteMultiple(ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	
	idMap := make(map[int]bool)
	for _, id := range ids {
		idMap[id] = true
	}
	
	metadata, err := r.loadMetadata()
	if err != nil {
		return err
	}
	
	var updatedRecords []*domain.TTSHistory
	var deletedPaths []string
	
	for _, record := range metadata.Records {
		if idMap[record.ID] {
			if record.FilePath != "" {
				deletedPaths = append(deletedPaths, record.FilePath)
			}
		} else {
			updatedRecords = append(updatedRecords, record)
		}
	}
	
	// Delete associated audio files
	for _, path := range deletedPaths {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			// Log error but don't fail the operation
		}
	}
	
	// Update metadata
	metadata.Records = updatedRecords
	return r.saveMetadata(metadata)
}

// Clear removes all history records
func (r *FileHistoryRepository) Clear(ctx context.Context) error {
	metadata, err := r.loadMetadata()
	if err != nil {
		return err
	}
	
	// Delete all audio files
	audioDir := filepath.Join(r.basePath, "audio")
	if err := filepath.WalkDir(audioDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				// Log error but continue
			}
		}
		return nil
	}); err != nil && !os.IsNotExist(err) {
		// Log error but continue
	}
	
	// Clear metadata
	metadata.Records = []*domain.TTSHistory{}
	if err := r.saveMetadata(metadata); err != nil {
		return err
	}
	
	// Reset counter
	counter := &idCounter{NextID: 1}
	return r.saveCounter(counter)
}

// Count returns the total number of history records
func (r *FileHistoryRepository) Count(ctx context.Context) (int, error) {
	metadata, err := r.loadMetadata()
	if err != nil {
		return 0, err
	}
	
	return len(metadata.Records), nil
}

// Cleanup removes old records based on configuration
func (r *FileHistoryRepository) Cleanup(ctx context.Context, maxCount int, maxAge *time.Duration) error {
	metadata, err := r.loadMetadata()
	if err != nil {
		return err
	}
	
	var toDelete []int
	
	// Remove records older than maxAge if specified
	if maxAge != nil {
		cutoff := time.Now().Add(-*maxAge)
		for _, record := range metadata.Records {
			if record.CreatedAt.Before(cutoff) {
				toDelete = append(toDelete, record.ID)
			}
		}
	}
	
	// If still over maxCount, remove oldest records
	if maxCount > 0 && len(metadata.Records) > maxCount {
		// Sort by creation time (oldest first)
		sorted := make([]*domain.TTSHistory, len(metadata.Records))
		copy(sorted, metadata.Records)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].CreatedAt.Before(sorted[j].CreatedAt)
		})
		
		// Mark excess records for deletion
		excess := len(sorted) - maxCount
		for i := 0; i < excess; i++ {
			found := false
			for _, id := range toDelete {
				if id == sorted[i].ID {
					found = true
					break
				}
			}
			if !found {
				toDelete = append(toDelete, sorted[i].ID)
			}
		}
	}
	
	// Delete marked records
	if len(toDelete) > 0 {
		return r.DeleteMultiple(ctx, toDelete)
	}
	
	return nil
}

// GetAudioFilePath returns the path where an audio file should be stored for a given ID and format
func (r *FileHistoryRepository) GetAudioFilePath(id int, format string) string {
	filename := strconv.Itoa(id) + domain.GetFileExtensionFromFormat(domain.OutputFormat(format))
	return filepath.Join(r.basePath, "audio", filename)
}
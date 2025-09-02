package usecase

import (
    "context"
    "fmt"
    "io"
    "bytes"
    "os"
    "path/filepath"
    "strconv"
    "time"

	"github.com/kajidog/aivis-cloud-cli/client/common/http"
	"github.com/kajidog/aivis-cloud-cli/client/config"
	"github.com/kajidog/aivis-cloud-cli/client/tts/domain"
)

// TTSHistoryManager implements the domain.TTSHistoryManager interface
type TTSHistoryManager struct {
	historyRepo  domain.TTSHistoryRepository
	ttsRepo      domain.TTSRepository
	audioPlayer  domain.AudioPlayer
	config       *config.Config
}

// NewTTSHistoryManager creates a new TTS history manager
func NewTTSHistoryManager(
	historyRepo domain.TTSHistoryRepository,
	ttsRepo domain.TTSRepository,
	audioPlayer domain.AudioPlayer,
	config *config.Config,
) *TTSHistoryManager {
	return &TTSHistoryManager{
		historyRepo: historyRepo,
		ttsRepo:     ttsRepo,
		audioPlayer: audioPlayer,
		config:      config,
	}
}

// SaveHistory saves a TTS request and response as history
func (m *TTSHistoryManager) SaveHistory(ctx context.Context, request *domain.TTSRequest, filePath string, credits *float64) (*domain.TTSHistory, error) {
	if !m.config.HistoryEnabled {
		return nil, nil // History disabled, return without error
	}

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Extract format from request or file path
	format := "wav" // default
	if request.OutputFormat != nil {
		format = string(*request.OutputFormat)
	} else {
		format = domain.GetFormatFromFilePath(filePath)
	}

	// Create history record
	history := &domain.TTSHistory{
		Request:       request,
		FilePath:      filePath,
		FileFormat:    format,
		FileSizeBytes: fileInfo.Size(),
		CreatedAt:     time.Now(),
		Text:          request.Text,
		ModelUUID:     request.ModelUUID,
		Credits:       credits,
	}

	// Save history record and get assigned ID
	id, err := m.historyRepo.Save(ctx, history)
	if err != nil {
		return nil, fmt.Errorf("failed to save history: %w", err)
	}

	history.ID = id

	// Perform cleanup if necessary
	if err := m.CleanupHistory(ctx); err != nil {
		// Log error but don't fail the save operation
		// TODO: Add proper logging
	}

	return history, nil
}

// SaveHistoryWithAudio saves TTS request and audio data as history
func (m *TTSHistoryManager) SaveHistoryWithAudio(ctx context.Context, request *domain.TTSRequest, audioData io.ReadCloser, billingInfo *http.BillingInfo) (*domain.TTSHistory, error) {
	if !m.config.HistoryEnabled {
		if audioData != nil {
			audioData.Close()
		}
		return nil, nil // History disabled, return without error
	}

	// Get next ID for file naming
	id, err := m.historyRepo.GetNextID(ctx)
	if err != nil {
		if audioData != nil {
			audioData.Close()
		}
		return nil, fmt.Errorf("failed to get next ID: %w", err)
	}

	// Determine file format
	format := "wav" // default
	if request.OutputFormat != nil {
		format = string(*request.OutputFormat)
	}

	// Get history store path
	storePath, err := m.config.GetHistoryStorePath()
	if err != nil {
		if audioData != nil {
			audioData.Close()
		}
		return nil, fmt.Errorf("failed to get history store path: %w", err)
	}

	// Create audio directory
	audioDir := filepath.Join(storePath, "audio")
	if err := os.MkdirAll(audioDir, 0755); err != nil {
		if audioData != nil {
			audioData.Close()
		}
		return nil, fmt.Errorf("failed to create audio directory: %w", err)
	}

	// Create audio file
	fileName := fmt.Sprintf("%d%s", id, domain.GetFileExtensionFromFormat(domain.OutputFormat(format)))
	filePath := filepath.Join(audioDir, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		if audioData != nil {
			audioData.Close()
		}
		return nil, fmt.Errorf("failed to create audio file: %w", err)
	}
	defer file.Close()

	// Copy audio data to file
	size, err := io.Copy(file, audioData)
	audioData.Close()
	if err != nil {
		os.Remove(filePath) // Clean up on error
		return nil, fmt.Errorf("failed to write audio data: %w", err)
	}

	// Extract credits from billing info
	var credits *float64
	if billingInfo != nil && billingInfo.CreditsUsed != "" {
		if creditsUsed, err := strconv.ParseFloat(billingInfo.CreditsUsed, 64); err == nil {
			credits = &creditsUsed
		}
	}

	// Create history record
	history := &domain.TTSHistory{
		ID:            id,
		Request:       request,
		FilePath:      filePath,
		FileFormat:    format,
		FileSizeBytes: size,
		CreatedAt:     time.Now(),
		Text:          request.Text,
		ModelUUID:     request.ModelUUID,
		Credits:       credits,
	}

	// Save history record
	if _, err := m.historyRepo.Save(ctx, history); err != nil {
		os.Remove(filePath) // Clean up on error
		return nil, fmt.Errorf("failed to save history: %w", err)
	}

	// Perform cleanup if necessary
	if err := m.CleanupHistory(ctx); err != nil {
		// Log error but don't fail the save operation
		// TODO: Add proper logging
	}

	return history, nil
}

// GetHistory retrieves history by ID
func (m *TTSHistoryManager) GetHistory(ctx context.Context, id int) (*domain.TTSHistory, error) {
	return m.historyRepo.GetByID(ctx, id)
}

// ListHistory lists history records with pagination and filtering
func (m *TTSHistoryManager) ListHistory(ctx context.Context, request *domain.TTSHistorySearchRequest) (*domain.TTSHistoryListResponse, error) {
	return m.historyRepo.List(ctx, request)
}

// PlayHistory replays audio from history
func (m *TTSHistoryManager) PlayHistory(ctx context.Context, id int, playbackOptions *domain.PlaybackRequest) error {
	// Get history record
	history, err := m.GetHistory(ctx, id)
	if err != nil {
		return err
	}

    // Check if audio file exists
    if _, err := os.Stat(history.FilePath); err != nil {
        if os.IsNotExist(err) {
            return fmt.Errorf("audio file not found: %s", history.FilePath)
        }
        return fmt.Errorf("failed to access audio file: %w", err)
    }

	// Create default playback options if not provided
	if playbackOptions == nil {
		playbackOptions = domain.NewPlaybackRequest(history.Request).
			WithMode(domain.PlaybackModeImmediate).
			WithWaitForEnd(true).
			Build()
	}

    // Open and fully read the audio file to avoid premature close/truncation
    f, err := os.Open(history.FilePath)
    if err != nil {
        return fmt.Errorf("failed to open audio file: %w", err)
    }
    data, err := io.ReadAll(f)
    f.Close()
    if err != nil {
        return fmt.Errorf("failed to read audio file: %w", err)
    }

    // Determine the output format from file extension
    format := domain.OutputFormat(history.FileFormat)

    // Detach playback from request context to allow asynchronous completion
    playbackCtx := context.Background()

    // Play from in-memory buffer for stability
    return m.audioPlayer.Play(playbackCtx, bytes.NewReader(data), format)
}

// DeleteHistory removes a history record
func (m *TTSHistoryManager) DeleteHistory(ctx context.Context, id int) error {
	return m.historyRepo.Delete(ctx, id)
}

// ClearHistory removes all history records
func (m *TTSHistoryManager) ClearHistory(ctx context.Context) error {
	return m.historyRepo.Clear(ctx)
}

// CleanupHistory removes old records based on configuration
func (m *TTSHistoryManager) CleanupHistory(ctx context.Context) error {
	if !m.config.HistoryEnabled {
		return nil
	}

	return m.historyRepo.Cleanup(ctx, m.config.HistoryMaxCount, nil)
}

// GetHistoryStats returns statistics about the history
func (m *TTSHistoryManager) GetHistoryStats(ctx context.Context) (*domain.TTSHistoryStats, error) {
	count, err := m.historyRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	// Get recent history to calculate total file size
	searchRequest := domain.NewTTSHistorySearchRequest().
		WithLimit(1000). // Get a reasonable number of recent records
		Build()

	response, err := m.historyRepo.List(ctx, searchRequest)
	if err != nil {
		return nil, err
	}

	var totalSize int64
	var totalCredits float64
	for _, record := range response.Histories {
		totalSize += record.FileSizeBytes
		if record.Credits != nil {
			totalCredits += *record.Credits
		}
	}

	return &domain.TTSHistoryStats{
		TotalRecords:  count,
		TotalFileSize: totalSize,
		TotalCredits:  totalCredits,
	}, nil
}

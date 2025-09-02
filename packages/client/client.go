package client

import (
    "context"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strconv"
    "time"

	"github.com/kajidog/aivis-cloud-cli/client/common/http"
	"github.com/kajidog/aivis-cloud-cli/client/common/logger"
	"github.com/kajidog/aivis-cloud-cli/client/config"
	"github.com/kajidog/aivis-cloud-cli/client/models/domain"
	modelsInfra "github.com/kajidog/aivis-cloud-cli/client/models/infrastructure"
	modelsUsecase "github.com/kajidog/aivis-cloud-cli/client/models/usecase"
	paymentDomain "github.com/kajidog/aivis-cloud-cli/client/payment/domain"
	paymentInfra "github.com/kajidog/aivis-cloud-cli/client/payment/infrastructure"
	paymentUsecase "github.com/kajidog/aivis-cloud-cli/client/payment/usecase"
	ttsDomain "github.com/kajidog/aivis-cloud-cli/client/tts/domain"
	ttsInfra "github.com/kajidog/aivis-cloud-cli/client/tts/infrastructure"
	ttsUsecase "github.com/kajidog/aivis-cloud-cli/client/tts/usecase"
	usersDomain "github.com/kajidog/aivis-cloud-cli/client/users/domain"
	usersInfra "github.com/kajidog/aivis-cloud-cli/client/users/infrastructure"
	usersUsecase "github.com/kajidog/aivis-cloud-cli/client/users/usecase"
)

// Client is the main Aivis Cloud API client
type Client struct {
	config         *config.Config
	logger         logger.Logger
	httpClient     *http.Client
	ttsService     *ttsUsecase.TTSSynthesizer
	historyManager *ttsUsecase.TTSHistoryManager
	modelsService  *modelsUsecase.ModelSearcher
	playerService  *ttsUsecase.AudioPlayerServiceAdapter
	usersService   *usersUsecase.UserUsecase
	paymentService *paymentUsecase.PaymentUsecase
}

// New creates a new Aivis Cloud API client with the provided API key
func New(apiKey string) (*Client, error) {
	cfg := config.NewConfig(apiKey)
	return NewWithConfig(cfg)
}

// NewWithConfig creates a new Aivis Cloud API client with the provided configuration
func NewWithConfig(cfg *config.Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Initialize logger
	logLevel, err := logger.ParseLogLevel(cfg.LogLevel)
	if err != nil {
		logLevel = logger.INFO // fallback to INFO level
	}
	
	logWriter, err := cfg.GetLogWriter()
	if err != nil {
		return nil, err
	}
	
	loggerConfig := &logger.Config{
		Level:  logLevel,
		Output: logWriter,
		Format: logger.Format(cfg.LogFormat),
	}
	
	clientLogger := logger.New(loggerConfig)
	clientLogger.Info("Initializing AivisCloud client", 
		logger.String("log_level", cfg.LogLevel),
		logger.String("log_output", cfg.LogOutput),
		logger.String("log_format", cfg.LogFormat),
	)

	httpClient := http.NewClient(cfg)

	// Initialize repositories
	ttsRepo := ttsInfra.NewTTSAPIRepository(httpClient)
	modelsRepo := modelsInfra.NewModelAPIRepository(httpClient)
	usersRepo := usersInfra.NewUserAPI(httpClient)
	paymentRepo := paymentInfra.NewPaymentAPI(httpClient)
	
	// Initialize history repository if history is enabled
	var historyRepo ttsDomain.TTSHistoryRepository
	if cfg.HistoryEnabled {
		historyStorePath, err := cfg.GetHistoryStorePath()
		if err != nil {
			return nil, err
		}
		historyRepo = ttsInfra.NewFileHistoryRepository(historyStorePath)
	}

	// Initialize use cases
	ttsService := ttsUsecase.NewTTSSynthesizer(ttsRepo)
	modelsService := modelsUsecase.NewModelSearcher(modelsRepo)
	usersService := usersUsecase.NewUserUsecase(usersRepo)
	paymentService := paymentUsecase.NewPaymentUsecase(paymentRepo)
	
	// Initialize audio player with configuration
	playbackConfig := ttsDomain.DefaultPlaybackConfig()
	if cfg.DefaultPlaybackMode != "" {
		switch cfg.DefaultPlaybackMode {
		case "immediate":
			playbackConfig.DefaultMode = ttsDomain.PlaybackModeImmediate
		case "queue":
			playbackConfig.DefaultMode = ttsDomain.PlaybackModeQueue
		case "no_queue":
			playbackConfig.DefaultMode = ttsDomain.PlaybackModeNoQueue
		}
	}
	
	audioPlayer := ttsInfra.NewOSCommandAudioPlayerWithLogger(playbackConfig, clientLogger)
	
	// Initialize global audio player service singleton
	globalPlayerService := ttsUsecase.GetGlobalAudioPlayerService()
	
	// Create player config
	playerConfig := &ttsUsecase.AudioPlayerConfig{
		MaxQueueSize: 100, // Default queue size
	}
	
    globalPlayerService.InitializeWithLogger(ttsService, audioPlayer, playerConfig, clientLogger)
    // Provide factory for creating independent players (for no_queue concurrent playback)
    globalPlayerService.SetNewPlayerFactory(func() ttsDomain.AudioPlayer {
        // Use same playback config and logger
        return ttsInfra.NewOSCommandAudioPlayerWithLogger(playbackConfig, clientLogger)
    })
	
	// Create adapter to maintain compatibility with existing interface
	playerService := ttsUsecase.NewAudioPlayerServiceAdapter(globalPlayerService)
	
	// Initialize history manager if history is enabled
	var historyManager *ttsUsecase.TTSHistoryManager
	if cfg.HistoryEnabled && historyRepo != nil {
		historyManager = ttsUsecase.NewTTSHistoryManager(historyRepo, ttsRepo, audioPlayer, cfg)
	}

	return &Client{
		config:         cfg,
		logger:         clientLogger,
		httpClient:     httpClient,
		ttsService:     ttsService,
		historyManager: historyManager,
		modelsService:  modelsService,
		playerService:  playerService,
		usersService:   usersService,
		paymentService: paymentService,
	}, nil
}

// TTS Service Methods

// Synthesize performs text-to-speech synthesis
func (c *Client) Synthesize(ctx context.Context, request *ttsDomain.TTSRequest) (*ttsDomain.TTSResponse, error) {
	if err := c.ttsService.ValidateRequest(request); err != nil {
		return nil, err
	}
	
	response, err := c.ttsService.Synthesize(ctx, request)
	if err != nil {
		return nil, err
	}
	
	// Note: Regular Synthesize() does not save to history to preserve AudioData stream
	// Use SynthesizeToFileWithHistory() or PlayStreamWithHistory() for automatic history saving
	
	return response, nil
}

// SynthesizeToFile performs text-to-speech synthesis and writes the result to a writer
func (c *Client) SynthesizeToFile(ctx context.Context, request *ttsDomain.TTSRequest, writer io.Writer) error {
	if err := c.ttsService.ValidateRequest(request); err != nil {
		return err
	}
	return c.ttsService.SynthesizeToFile(ctx, request, writer)
}

// SynthesizeToFileWithHistory performs text-to-speech synthesis, writes to a file, and saves to history
func (c *Client) SynthesizeToFileWithHistory(ctx context.Context, request *ttsDomain.TTSRequest, filePath string) (*ttsDomain.TTSResponse, error) {
	if err := c.ttsService.ValidateRequest(request); err != nil {
		return nil, err
	}
	
	// Create file for writing
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()
	
	// Get TTS response
	response, err := c.ttsService.Synthesize(ctx, request)
	if err != nil {
		return nil, err
	}
	defer response.AudioData.Close()
	
	// Write audio data to file
	_, err = io.Copy(file, response.AudioData)
	if err != nil {
		return nil, fmt.Errorf("failed to write audio data to file: %w", err)
	}
	
	// Save to history if history manager is available
	if c.historyManager != nil && c.config.HistoryEnabled {
		var credits *float64
		if response.BillingInfo != nil && response.BillingInfo.CreditsUsed != "" {
			if creditsUsed, err := strconv.ParseFloat(response.BillingInfo.CreditsUsed, 64); err == nil {
				credits = &creditsUsed
			}
		}
		
		history, err := c.historyManager.SaveHistory(ctx, request, filePath, credits)
		if err != nil {
			// Log error but don't fail the operation
			c.logger.Warn("Failed to save TTS history: " + err.Error())
		} else if history != nil {
			response.HistoryID = history.ID
		}
	}
	
	// Clear AudioData since it's been consumed
	response.AudioData = nil
	
	return response, nil
}

// SynthesizeStreamWithHistory performs streaming TTS synthesis with concurrent history saving
func (c *Client) SynthesizeStreamWithHistory(ctx context.Context, request *ttsDomain.TTSRequest, filePath string, handler ttsDomain.TTSStreamHandler) (*ttsDomain.TTSResponse, error) {
	if err := c.ttsService.ValidateRequest(request); err != nil {
		return nil, err
	}
	
	// Create file for concurrent writing
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()
	
	// Create completion channel
	completionChan := make(chan error, 1)
	
	// Create a tee handler that writes to file while forwarding to user handler
	teeHandler := &streamingHistoryHandler{
		userHandler:    handler,
		file:           file,
		completionChan: completionChan,
	}
	
	// Start streaming synthesis
	err = c.ttsService.SynthesizeStream(ctx, request, teeHandler)
	if err != nil {
		return nil, err
	}
	
	// Wait for completion
	select {
	case err := <-completionChan:
		if err != nil {
			return nil, err
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	
	// Create response object for history saving
	response := &ttsDomain.TTSResponse{
		HistoryID: 0, // Will be set by history manager
		// Note: BillingInfo is not available in streaming mode
	}
	
	// Save to history if history manager is available
	if c.historyManager != nil && c.config.HistoryEnabled {
		c.logger.Info("Attempting to save TTS history", 
			logger.String("file_path", filePath),
		)
		history, err := c.historyManager.SaveHistory(ctx, request, filePath, nil)
		if err != nil {
			// Log error but don't fail the operation
			c.logger.Warn("Failed to save TTS history: " + err.Error())
		} else if history != nil {
			c.logger.Info("TTS history saved successfully", 
				logger.Int("history_id", history.ID),
			)
			response.HistoryID = history.ID
		} else {
			c.logger.Warn("TTS history save returned nil without error")
		}
	} else {
		c.logger.Warn("TTS history not saved", 
			logger.Bool("history_manager_nil", c.historyManager == nil),
			logger.Bool("history_enabled", c.config.HistoryEnabled),
		)
	}
	
	return response, nil
}

// streamingHistoryHandler wraps user handler and writes to file concurrently
type streamingHistoryHandler struct {
	userHandler    ttsDomain.TTSStreamHandler
	file           *os.File
	completionChan chan<- error
}

func (h *streamingHistoryHandler) OnChunk(chunk *ttsDomain.TTSStreamChunk) error {
	// Write chunk to file concurrently
	if h.file != nil {
		if _, err := h.file.Write(chunk.Data); err != nil {
			// Don't fail streaming for file write errors, just log
			// (the history save will be skipped if file is corrupted)
		}
	}
	
	// Forward to user handler
	if h.userHandler != nil {
		return h.userHandler.OnChunk(chunk)
	}
	
	return nil
}

func (h *streamingHistoryHandler) OnComplete() error {
	// Forward to user handler first
	var userErr error
	if h.userHandler != nil {
		userErr = h.userHandler.OnComplete()
	}
	
	// Signal completion
	select {
	case h.completionChan <- userErr:
	default:
		// Channel might be full, should not happen
	}
	
	return userErr
}

func (h *streamingHistoryHandler) OnError(err error) {
	// Send error to completion channel
	select {
	case h.completionChan <- err:
	default:
		// Channel might be full, just forward to user handler
	}
	
	// Forward to user handler
	if h.userHandler != nil {
		h.userHandler.OnError(err)
	}
}

// SynthesizeStream performs streaming text-to-speech synthesis
func (c *Client) SynthesizeStream(ctx context.Context, request *ttsDomain.TTSRequest, handler ttsDomain.TTSStreamHandler) error {
	if err := c.ttsService.ValidateRequest(request); err != nil {
		return err
	}
	return c.ttsService.SynthesizeStream(ctx, request, handler)
}

// Models Service Methods

// SearchModels searches for available models
func (c *Client) SearchModels(ctx context.Context, request *domain.ModelSearchRequest) (*domain.ModelSearchResponse, error) {
	return c.modelsService.SearchModels(ctx, request)
}

// GetModel retrieves a specific model by UUID
func (c *Client) GetModel(ctx context.Context, modelUUID string) (*domain.Model, error) {
	return c.modelsService.GetModel(ctx, modelUUID)
}

// GetModelSpeakers retrieves speakers for a specific model
func (c *Client) GetModelSpeakers(ctx context.Context, modelUUID string) ([]domain.Speaker, error) {
	return c.modelsService.GetModelSpeakers(ctx, modelUUID)
}

// SearchPublicModels searches for public models only
func (c *Client) SearchPublicModels(ctx context.Context, query string) (*domain.ModelSearchResponse, error) {
	return c.modelsService.SearchPublicModels(ctx, query)
}

// SearchModelsByAuthor searches for models by a specific author
func (c *Client) SearchModelsByAuthor(ctx context.Context, author string) (*domain.ModelSearchResponse, error) {
	return c.modelsService.SearchModelsByAuthor(ctx, author)
}

// SearchModelsByTags searches for models with specific tags
func (c *Client) SearchModelsByTags(ctx context.Context, tags ...string) (*domain.ModelSearchResponse, error) {
	return c.modelsService.SearchModelsByTags(ctx, tags...)
}

// GetPopularModels retrieves popular models sorted by download count
func (c *Client) GetPopularModels(ctx context.Context, limit int) (*domain.ModelSearchResponse, error) {
	return c.modelsService.GetPopularModels(ctx, limit)
}

// GetRecentModels retrieves recently updated models
func (c *Client) GetRecentModels(ctx context.Context, limit int) (*domain.ModelSearchResponse, error) {
	return c.modelsService.GetRecentModels(ctx, limit)
}

// GetTopRatedModels retrieves top-rated models
func (c *Client) GetTopRatedModels(ctx context.Context, limit int) (*domain.ModelSearchResponse, error) {
	return c.modelsService.GetTopRatedModels(ctx, limit)
}

// Convenience Methods

// NewTTSRequest creates a new TTS request builder
func (c *Client) NewTTSRequest(modelUUID, text string) *ttsDomain.TTSRequestBuilder {
	return ttsDomain.NewTTSRequestBuilder(modelUUID, text)
}

// NewModelSearchRequest creates a new model search request builder
func (c *Client) NewModelSearchRequest() *domain.ModelSearchRequestBuilder {
	return domain.NewModelSearchRequestBuilder()
}

// NewPlaybackRequest creates a new playback request builder
func (c *Client) NewPlaybackRequest(ttsRequest *ttsDomain.TTSRequest) *ttsDomain.PlaybackRequestBuilder {
	return ttsDomain.NewPlaybackRequest(ttsRequest)
}

// Audio Playback Methods

// PlayText plays the given text using TTS synthesis and audio playback
func (c *Client) PlayText(ctx context.Context, text, modelUUID string) error {
	return c.playerService.PlayText(ctx, text, modelUUID, nil)
}

// PlayTextWithOptions plays text with custom playback options
func (c *Client) PlayTextWithOptions(ctx context.Context, text, modelUUID string, options *ttsDomain.PlaybackRequest) error {
	return c.playerService.PlayText(ctx, text, modelUUID, options)
}

// PlayStreamWithHistory performs streaming TTS synthesis with audio playback and history saving
func (c *Client) PlayStreamWithHistory(ctx context.Context, request *ttsDomain.PlaybackRequest, filePath string) (*ttsDomain.TTSResponse, error) {
	c.logger.Info("Using single-pass streaming synthesis with concurrent playback and history saving")
	
	// Use the adapter's streaming method with history (single synthesis only)
	// This performs only ONE synthesis with concurrent playback and file saving
    err := c.playerService.PlayRequestWithHistory(ctx, request, filePath)
    if err != nil {
        return nil, fmt.Errorf("streaming synthesis and playback failed: %w", err)
    }

    // Best-effort: wait briefly until the history file is created by the streaming goroutine
    // to avoid race where SaveHistory fails with file-not-found
    waitDeadline := time.Now().Add(3 * time.Second)
    for time.Now().Before(waitDeadline) {
        if _, statErr := os.Stat(filePath); statErr == nil {
            break
        }
        select {
        case <-ctx.Done():
            // Context cancelled; proceed without blocking further
            break
        case <-time.After(50 * time.Millisecond):
        }
    }

    // Save to history database (file should exist from streaming path)
    historyResponse, historyErr := c.historyManager.SaveHistory(ctx, request.TTSRequest, filePath, nil)
	var historyID int = 0
	if historyErr == nil && historyResponse != nil {
		historyID = historyResponse.ID
	} else {
		c.logger.Warn("Failed to save TTS history metadata", 
			logger.String("error", historyErr.Error()))
	}
	
	response := &ttsDomain.TTSResponse{
		HistoryID: historyID,
	}
	
	c.logger.Info("Single-pass streaming synthesis with concurrent operations completed successfully")
	return response, nil
}

// playbackStreamHandler handles streaming audio for immediate playback
type playbackStreamHandler struct {
	playerService *ttsUsecase.AudioPlayerServiceAdapter
	playbackReq   *ttsDomain.PlaybackRequest
	ctx          context.Context
	started      bool
	pipeWriter   *io.PipeWriter
}

func (h *playbackStreamHandler) OnChunk(chunk *ttsDomain.TTSStreamChunk) error {
	// Start progressive playback on first chunk for MP3 format
	if !h.started && h.playbackReq.TTSRequest != nil {
		h.started = true
		
		// Check if output format is MP3 (supports streaming)
		format := ttsDomain.OutputFormatWAV // default
		if h.playbackReq.TTSRequest.OutputFormat != nil {
			format = *h.playbackReq.TTSRequest.OutputFormat
		}
		
		// Only enable progressive playback for MP3 format
		if format == ttsDomain.OutputFormatMP3 {
			// Create pipe for streaming audio to player
			pipeReader, pipeWriter := io.Pipe()
			h.pipeWriter = pipeWriter
			
			// Start playback in goroutine with pipe reader
			go func() {
				defer pipeReader.Close()
				// Create a temporary file that we'll write to progressively
				tempFile, err := os.CreateTemp("", "progressive_*.mp3")
				if err != nil {
					return
				}
				defer os.Remove(tempFile.Name())
				defer tempFile.Close()
				
				// Start copying from pipe to temp file and player simultaneously
				// This allows OS command to start playing as soon as MP3 frames are available
				go func() {
					_, err := io.Copy(tempFile, pipeReader)
					if err != nil {
						return
					}
				}()
				
				// Give a small delay for first MP3 frames to be written
				time.Sleep(100 * time.Millisecond)
				
				// For progressive MP3 playback, use direct file playback
				// Note: This bypasses queue system - should be improved in future
				if err := h.playerService.PlayAudioFile(h.ctx, tempFile.Name(), format); err != nil {
					// Log error but don't fail the streaming
				}
			}()
		}
	}
	
	// Write chunk data to pipe if progressive playback is active
	if h.pipeWriter != nil {
		if _, err := h.pipeWriter.Write(chunk.Data); err != nil {
			// Pipe closed or error - switch to completion-based playback
			h.pipeWriter.Close()
			h.pipeWriter = nil
		}
	}
	
	return nil
}

func (h *playbackStreamHandler) OnComplete() error {
	// Close pipe writer if progressive playback was active
	if h.pipeWriter != nil {
		h.pipeWriter.Close()
		h.pipeWriter = nil
		// Progressive playback was handled, don't trigger completion-based playback
		return nil
	}
	
	// Trigger playback after synthesis completes (for non-MP3 formats)
	// This will use the file that was written during streaming
	return h.playerService.PlayRequest(h.ctx, h.playbackReq)
}

func (h *playbackStreamHandler) OnError(err error) {
	// Error handling is managed by the caller
}

// historyOnlyStreamHandler handles streaming synthesis for history saving only
type historyOnlyStreamHandler struct {
	filePath string
	file     *os.File
	logger   logger.Logger
}

func (h *historyOnlyStreamHandler) OnChunk(chunk *ttsDomain.TTSStreamChunk) error {
	// This handler doesn't need to do anything - file writing is handled by SynthesizeStreamWithHistory
	return nil
}

func (h *historyOnlyStreamHandler) OnComplete() error {
	return nil
}

func (h *historyOnlyStreamHandler) OnError(err error) {
	h.logger.Error("History synthesis error: " + err.Error())
}

func (c *Client) PlayRequest(ctx context.Context, request *ttsDomain.PlaybackRequest) error {
    return c.playerService.PlayRequest(ctx, request)
}

// PlayRequestWithHistory plays audio with concurrent history saving.
// It auto-generates a history file path under the configured history store.
func (c *Client) PlayRequestWithHistory(ctx context.Context, request *ttsDomain.PlaybackRequest) (*ttsDomain.TTSResponse, error) {
    if c.historyManager == nil || !c.config.HistoryEnabled {
        return nil, fmt.Errorf("history is disabled or not configured")
    }

    // Determine history store directory
    storePath, err := c.config.GetHistoryStorePath()
    if err != nil {
        return nil, fmt.Errorf("failed to get history store path: %w", err)
    }
    audioDir := filepath.Join(storePath, "audio")
    if err := os.MkdirAll(audioDir, 0755); err != nil {
        return nil, fmt.Errorf("failed to create audio directory: %w", err)
    }

    // Determine extension from request format
    ext := ".wav"
    if request != nil && request.TTSRequest != nil && request.TTSRequest.OutputFormat != nil {
        switch *request.TTSRequest.OutputFormat {
        case ttsDomain.OutputFormatWAV:
            ext = ".wav"
        case ttsDomain.OutputFormatMP3:
            ext = ".mp3"
        case ttsDomain.OutputFormatFLAC:
            ext = ".flac"
        case ttsDomain.OutputFormatAAC:
            ext = ".aac"
        case ttsDomain.OutputFormatOpus:
            ext = ".opus"
        }
    }

    // Generate filename
    timestamp := time.Now().Format("20060102_150405")
    filePath := filepath.Join(audioDir, fmt.Sprintf("tts_%s%s", timestamp, ext))

    // Use single-pass streaming synthesis with concurrent playback and file writing
    resp, err := c.PlayStreamWithHistory(ctx, request, filePath)
    if err != nil {
        return nil, err
    }
    return resp, nil
}

// StopPlayback stops current playback and clears queue
func (c *Client) StopPlayback() error {
	return c.playerService.Stop()
}

// PausePlayback pauses current playback
func (c *Client) PausePlayback() error {
	return c.playerService.Pause()
}

// ResumePlayback resumes paused playback
func (c *Client) ResumePlayback() error {
	return c.playerService.Resume()
}

// SetPlaybackVolume sets playback volume (0.0 to 1.0)
func (c *Client) SetPlaybackVolume(volume float64) error {
	return c.playerService.SetVolume(volume)
}

// GetPlaybackStatus returns current playback status
func (c *Client) GetPlaybackStatus() ttsDomain.PlaybackInfo {
	return c.playerService.GetStatus()
}

// ClearPlaybackQueue clears all items from the playback queue
func (c *Client) ClearPlaybackQueue() {
	c.playerService.ClearQueue()
}

// Configuration Methods

// GetConfig returns the current configuration
func (c *Client) GetConfig() *config.Config {
	return c.config
}

// GetLogger returns the current logger
func (c *Client) GetLogger() logger.Logger {
	return c.logger
}

// UpdateConfig updates the client configuration
func (c *Client) UpdateConfig(cfg *config.Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	
	c.config = cfg
	
	// Reinitialize logger
	logLevel, err := logger.ParseLogLevel(cfg.LogLevel)
	if err != nil {
		logLevel = logger.INFO // fallback to INFO level
	}
	
	logWriter, err := cfg.GetLogWriter()
	if err != nil {
		return err
	}
	
	loggerConfig := &logger.Config{
		Level:  logLevel,
		Output: logWriter,
		Format: logger.Format(cfg.LogFormat),
	}
	
	c.logger = logger.New(loggerConfig)
	c.logger.Info("Client configuration updated", 
		logger.String("log_level", cfg.LogLevel),
		logger.String("log_output", cfg.LogOutput),
		logger.String("log_format", cfg.LogFormat),
	)
	
	c.httpClient = http.NewClient(cfg)
	
	// Reinitialize repositories with new HTTP client
	ttsRepo := ttsInfra.NewTTSAPIRepository(c.httpClient)
	modelsRepo := modelsInfra.NewModelAPIRepository(c.httpClient)
	usersRepo := usersInfra.NewUserAPI(c.httpClient)
	paymentRepo := paymentInfra.NewPaymentAPI(c.httpClient)
	
	// Reinitialize services
	c.ttsService = ttsUsecase.NewTTSSynthesizer(ttsRepo)
	c.modelsService = modelsUsecase.NewModelSearcher(modelsRepo)
	c.usersService = usersUsecase.NewUserUsecase(usersRepo)
	c.paymentService = paymentUsecase.NewPaymentUsecase(paymentRepo)
	
	// Reinitialize history repository and manager if history is enabled
	var historyRepo ttsDomain.TTSHistoryRepository
	if cfg.HistoryEnabled {
		historyStorePath, err := cfg.GetHistoryStorePath()
		if err != nil {
			return err
		}
		historyRepo = ttsInfra.NewFileHistoryRepository(historyStorePath)
	}
	
	// Reinitialize audio player service with configuration
	playbackConfig := ttsDomain.DefaultPlaybackConfig()
	if cfg.DefaultPlaybackMode != "" {
		switch cfg.DefaultPlaybackMode {
		case "immediate":
			playbackConfig.DefaultMode = ttsDomain.PlaybackModeImmediate
		case "queue":
			playbackConfig.DefaultMode = ttsDomain.PlaybackModeQueue
		case "no_queue":
			playbackConfig.DefaultMode = ttsDomain.PlaybackModeNoQueue
		}
	}
	
	audioPlayer := ttsInfra.NewOSCommandAudioPlayerWithLogger(playbackConfig, c.logger)
	
	// Use global audio player service singleton
	globalPlayerService := ttsUsecase.GetGlobalAudioPlayerService()
	
	// Create player config
	playerConfig := &ttsUsecase.AudioPlayerConfig{
		MaxQueueSize: 100, // Default queue size
	}
	
	globalPlayerService.InitializeWithLogger(c.ttsService, audioPlayer, playerConfig, c.logger)
	
	// Create adapter to maintain compatibility with existing interface
	c.playerService = ttsUsecase.NewAudioPlayerServiceAdapter(globalPlayerService)
	
	// Reinitialize history manager if history is enabled
	if cfg.HistoryEnabled && historyRepo != nil {
		c.historyManager = ttsUsecase.NewTTSHistoryManager(historyRepo, ttsRepo, audioPlayer, cfg)
	} else {
		c.historyManager = nil
	}
	
	return nil
}

// Users Service Methods

// GetMe retrieves current user's account information
func (c *Client) GetMe(ctx context.Context) (*usersDomain.UserMe, error) {
	return c.usersService.GetMe(ctx)
}

// GetUserByHandle retrieves a user profile by handle
func (c *Client) GetUserByHandle(ctx context.Context, handle string) (*usersDomain.User, error) {
	return c.usersService.GetUserByHandle(ctx, handle)
}

// Payment Service Methods

// GetSubscriptions retrieves user subscriptions
func (c *Client) GetSubscriptions(ctx context.Context, limit, offset int) (*paymentDomain.SubscriptionListResponse, error) {
	return c.paymentService.GetSubscriptions(ctx, limit, offset)
}

// GetCreditTransactions retrieves credit transaction history
func (c *Client) GetCreditTransactions(ctx context.Context, transactionType paymentDomain.TransactionType, status paymentDomain.TransactionStatus, startDate, endDate *time.Time, limit, offset int) (*paymentDomain.CreditTransactionListResponse, error) {
	return c.paymentService.GetCreditTransactions(ctx, transactionType, status, startDate, endDate, limit, offset)
}

// GetAPIKeys retrieves API keys
func (c *Client) GetAPIKeys(ctx context.Context, limit, offset int) (*paymentDomain.APIKeyListResponse, error) {
	return c.paymentService.GetAPIKeys(ctx, limit, offset)
}

// CreateAPIKey creates a new API key
func (c *Client) CreateAPIKey(ctx context.Context, name string) (*paymentDomain.APIKey, error) {
	return c.paymentService.CreateAPIKey(ctx, name)
}

// DeleteAPIKey deletes an API key
func (c *Client) DeleteAPIKey(ctx context.Context, keyID string) error {
	return c.paymentService.DeleteAPIKey(ctx, keyID)
}

// GetUsageSummaries retrieves usage statistics
func (c *Client) GetUsageSummaries(ctx context.Context, period string, startDate, endDate *time.Time, modelID string) (*paymentDomain.UsageSummary, error) {
	return c.paymentService.GetUsageSummaries(ctx, period, startDate, endDate, modelID)
}

// TTS History Management Methods

// GetTTSHistory retrieves a specific TTS history record by ID
func (c *Client) GetTTSHistory(ctx context.Context, id int) (*ttsDomain.TTSHistory, error) {
	if c.historyManager == nil {
		return nil, fmt.Errorf("history management is disabled")
	}
	return c.historyManager.GetHistory(ctx, id)
}

// ListTTSHistory lists TTS history records with pagination and filtering
func (c *Client) ListTTSHistory(ctx context.Context, request *ttsDomain.TTSHistorySearchRequest) (*ttsDomain.TTSHistoryListResponse, error) {
	if c.historyManager == nil {
		return nil, fmt.Errorf("history management is disabled")
	}
	return c.historyManager.ListHistory(ctx, request)
}

// PlayTTSHistory replays audio from TTS history
func (c *Client) PlayTTSHistory(ctx context.Context, id int, playbackOptions *ttsDomain.PlaybackRequest) error {
	if c.historyManager == nil {
		return fmt.Errorf("history management is disabled")
	}
	return c.historyManager.PlayHistory(ctx, id, playbackOptions)
}

// DeleteTTSHistory removes a TTS history record
func (c *Client) DeleteTTSHistory(ctx context.Context, id int) error {
	if c.historyManager == nil {
		return fmt.Errorf("history management is disabled")
	}
	return c.historyManager.DeleteHistory(ctx, id)
}

// ClearTTSHistory removes all TTS history records
func (c *Client) ClearTTSHistory(ctx context.Context) error {
	if c.historyManager == nil {
		return fmt.Errorf("history management is disabled")
	}
	return c.historyManager.ClearHistory(ctx)
}

// GetTTSHistoryStats retrieves statistics about TTS history
func (c *Client) GetTTSHistoryStats(ctx context.Context) (*ttsDomain.TTSHistoryStats, error) {
	if c.historyManager == nil {
		return nil, fmt.Errorf("history management is disabled")
	}
	return c.historyManager.GetHistoryStats(ctx)
}

// NewTTSHistorySearchRequest creates a new TTS history search request builder
func (c *Client) NewTTSHistorySearchRequest() *ttsDomain.TTSHistorySearchRequestBuilder {
	return ttsDomain.NewTTSHistorySearchRequest()
}

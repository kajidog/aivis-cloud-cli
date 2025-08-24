package client

import (
	"context"
	"io"
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
	
	// Create adapter to maintain compatibility with existing interface
	playerService := ttsUsecase.NewAudioPlayerServiceAdapter(globalPlayerService)

	return &Client{
		config:         cfg,
		logger:         clientLogger,
		httpClient:     httpClient,
		ttsService:     ttsService,
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
	return c.ttsService.Synthesize(ctx, request)
}

// SynthesizeToFile performs text-to-speech synthesis and writes the result to a writer
func (c *Client) SynthesizeToFile(ctx context.Context, request *ttsDomain.TTSRequest, writer io.Writer) error {
	if err := c.ttsService.ValidateRequest(request); err != nil {
		return err
	}
	return c.ttsService.SynthesizeToFile(ctx, request, writer)
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

// PlayRequest plays audio based on the playback request
func (c *Client) PlayRequest(ctx context.Context, request *ttsDomain.PlaybackRequest) error {
	return c.playerService.PlayRequest(ctx, request)
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
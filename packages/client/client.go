package client

import (
	"context"
	"io"

	"github.com/kajidog/aiviscloud-mcp/client/common/http"
	"github.com/kajidog/aiviscloud-mcp/client/config"
	"github.com/kajidog/aiviscloud-mcp/client/models/domain"
	modelsInfra "github.com/kajidog/aiviscloud-mcp/client/models/infrastructure"
	modelsUsecase "github.com/kajidog/aiviscloud-mcp/client/models/usecase"
	ttsDomain "github.com/kajidog/aiviscloud-mcp/client/tts/domain"
	ttsInfra "github.com/kajidog/aiviscloud-mcp/client/tts/infrastructure"
	ttsUsecase "github.com/kajidog/aiviscloud-mcp/client/tts/usecase"
)

// Client is the main Aivis Cloud API client
type Client struct {
	config        *config.Config
	httpClient    *http.Client
	ttsService    *ttsUsecase.TTSSynthesizer
	modelsService *modelsUsecase.ModelSearcher
	playerService *ttsUsecase.AudioPlayerService
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

	httpClient := http.NewClient(cfg)

	// Initialize repositories
	ttsRepo := ttsInfra.NewTTSAPIRepository(httpClient)
	modelsRepo := modelsInfra.NewModelAPIRepository(httpClient)

	// Initialize use cases
	ttsService := ttsUsecase.NewTTSSynthesizer(ttsRepo)
	modelsService := modelsUsecase.NewModelSearcher(modelsRepo)
	
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
	
	audioPlayer := ttsInfra.NewOSCommandAudioPlayer(playbackConfig)
	playerService := ttsUsecase.NewAudioPlayerService(ttsService, audioPlayer, playbackConfig)

	return &Client{
		config:        cfg,
		httpClient:    httpClient,
		ttsService:    ttsService,
		modelsService: modelsService,
		playerService: playerService,
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

// UpdateConfig updates the client configuration
func (c *Client) UpdateConfig(cfg *config.Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	
	c.config = cfg
	c.httpClient = http.NewClient(cfg)
	
	// Reinitialize repositories with new HTTP client
	ttsRepo := ttsInfra.NewTTSAPIRepository(c.httpClient)
	modelsRepo := modelsInfra.NewModelAPIRepository(c.httpClient)
	
	// Reinitialize services
	c.ttsService = ttsUsecase.NewTTSSynthesizer(ttsRepo)
	c.modelsService = modelsUsecase.NewModelSearcher(modelsRepo)
	
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
	
	audioPlayer := ttsInfra.NewOSCommandAudioPlayer(playbackConfig)
	c.playerService = ttsUsecase.NewAudioPlayerService(c.ttsService, audioPlayer, playbackConfig)
	
	return nil
}
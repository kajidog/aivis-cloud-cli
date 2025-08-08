package domain

import (
	"context"
	"io"
	"time"
)

// PlaybackMode represents different audio playback strategies
type PlaybackMode string

const (
	PlaybackModeImmediate  PlaybackMode = "immediate"   // Stop current audio and play new one immediately
	PlaybackModeQueue      PlaybackMode = "queue"       // Queue audio and wait for current to complete
	PlaybackModeNoQueue    PlaybackMode = "no_queue"    // Start playback without queue management (allows simultaneous)
)

// PlaybackStatus represents the current status of the audio player
type PlaybackStatus string

const (
	PlaybackStatusIdle    PlaybackStatus = "idle"
	PlaybackStatusPlaying PlaybackStatus = "playing" 
	PlaybackStatusPaused  PlaybackStatus = "paused"
	PlaybackStatusStopped PlaybackStatus = "stopped"
)

// PlaybackRequest represents a request to play audio with specific settings
type PlaybackRequest struct {
	TTSRequest   *TTSRequest   `json:"tts_request"`
	Mode         *PlaybackMode `json:"mode,omitempty"`
	Volume       *float64      `json:"volume,omitempty"`        // 0.0 to 1.0
	StartOffset  *time.Duration `json:"start_offset,omitempty"` // Start playback from specific position
	FadeInDuration  *time.Duration `json:"fade_in_duration,omitempty"`
	FadeOutDuration *time.Duration `json:"fade_out_duration,omitempty"`
}

// PlaybackConfig represents configuration for audio playback
type PlaybackConfig struct {
	DefaultMode     PlaybackMode  `json:"default_mode"`
	DefaultVolume   float64       `json:"default_volume"`        // 0.0 to 1.0
	BufferSize      int           `json:"buffer_size"`           // Audio buffer size
	SampleRate      int           `json:"sample_rate"`           // Audio sample rate
	MaxQueueSize    int           `json:"max_queue_size"`        // Maximum items in queue
}

// DefaultPlaybackConfig returns a default playback configuration
func DefaultPlaybackConfig() *PlaybackConfig {
	return &PlaybackConfig{
		DefaultMode:   PlaybackModeImmediate,
		DefaultVolume: 1.0,
		BufferSize:    512,
		SampleRate:    44100,
		MaxQueueSize:  10,
	}
}

// PlaybackInfo contains information about current playback
type PlaybackInfo struct {
	Status      PlaybackStatus `json:"status"`
	QueueLength int            `json:"queue_length"`
	CurrentText string         `json:"current_text,omitempty"`
	Duration    time.Duration  `json:"duration,omitempty"`
	Position    time.Duration  `json:"position,omitempty"`
	Volume      float64        `json:"volume"`
}

// AudioPlayer defines the interface for audio playback operations
type AudioPlayer interface {
	// Play starts playback of the given audio stream
	Play(ctx context.Context, audioData io.Reader, format OutputFormat) error
	
	// Stop stops current playback
	Stop() error
	
	// Pause pauses current playback
	Pause() error
	
	// Resume resumes paused playback
	Resume() error
	
	// SetVolume sets the playback volume (0.0 to 1.0)
	SetVolume(volume float64) error
	
	// GetStatus returns current playback status
	GetStatus() PlaybackInfo
	
	// IsPlaying returns true if audio is currently playing
	IsPlaying() bool
	
	// Close closes the audio player and releases resources
	Close() error
}

// PlaybackRequestBuilder helps build playback requests with method chaining
type PlaybackRequestBuilder struct {
	request *PlaybackRequest
}

// NewPlaybackRequest creates a new playback request builder
func NewPlaybackRequest(ttsRequest *TTSRequest) *PlaybackRequestBuilder {
	return &PlaybackRequestBuilder{
		request: &PlaybackRequest{
			TTSRequest: ttsRequest,
		},
	}
}

// WithMode sets the playback mode
func (b *PlaybackRequestBuilder) WithMode(mode PlaybackMode) *PlaybackRequestBuilder {
	b.request.Mode = &mode
	return b
}

// WithVolume sets the playback volume
func (b *PlaybackRequestBuilder) WithVolume(volume float64) *PlaybackRequestBuilder {
	b.request.Volume = &volume
	return b
}

// WithStartOffset sets the start offset for playback
func (b *PlaybackRequestBuilder) WithStartOffset(offset time.Duration) *PlaybackRequestBuilder {
	b.request.StartOffset = &offset
	return b
}

// WithFadeIn sets the fade in duration
func (b *PlaybackRequestBuilder) WithFadeIn(duration time.Duration) *PlaybackRequestBuilder {
	b.request.FadeInDuration = &duration
	return b
}

// WithFadeOut sets the fade out duration
func (b *PlaybackRequestBuilder) WithFadeOut(duration time.Duration) *PlaybackRequestBuilder {
	b.request.FadeOutDuration = &duration
	return b
}

// Build returns the built playback request
func (b *PlaybackRequestBuilder) Build() *PlaybackRequest {
	return b.request
}
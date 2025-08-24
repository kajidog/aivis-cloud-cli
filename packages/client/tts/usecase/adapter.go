package usecase

import (
	"context"

	"github.com/kajidog/aivis-cloud-cli/client/tts/domain"
)

// AudioPlayerServiceAdapter adapts the GlobalAudioPlayerService to maintain compatibility
type AudioPlayerServiceAdapter struct {
	globalService *GlobalAudioPlayerService
}

// NewAudioPlayerServiceAdapter creates a new adapter for global audio player service
func NewAudioPlayerServiceAdapter(globalService *GlobalAudioPlayerService) *AudioPlayerServiceAdapter {
	return &AudioPlayerServiceAdapter{
		globalService: globalService,
	}
}

// PlayText plays text with optional playback options
func (a *AudioPlayerServiceAdapter) PlayText(ctx context.Context, text, modelUUID string, options *domain.PlaybackRequest) error {
	// Create TTS request
	ttsRequest := domain.NewTTSRequestBuilder(modelUUID, text).Build()
	
	// Create playback request
	var playbackRequest *domain.PlaybackRequest
	if options != nil {
		playbackRequest = options
		// Update the TTS request in the options
		playbackRequest.TTSRequest = ttsRequest
	} else {
		playbackRequest = domain.NewPlaybackRequest(ttsRequest).Build()
	}
	
	return a.globalService.PlayRequest(ctx, playbackRequest)
}

// PlayRequest plays audio based on the playback request
func (a *AudioPlayerServiceAdapter) PlayRequest(ctx context.Context, request *domain.PlaybackRequest) error {
	return a.globalService.PlayRequest(ctx, request)
}

// Stop stops current playback and clears queue
func (a *AudioPlayerServiceAdapter) Stop() error {
	return a.globalService.Stop()
}

// Pause pauses current playback
func (a *AudioPlayerServiceAdapter) Pause() error {
	return a.globalService.Pause()
}

// Resume resumes paused playback
func (a *AudioPlayerServiceAdapter) Resume() error {
	return a.globalService.Resume()
}

// SetVolume sets playback volume
func (a *AudioPlayerServiceAdapter) SetVolume(volume float64) error {
	return a.globalService.SetVolume(volume)
}

// GetStatus returns current playback status
func (a *AudioPlayerServiceAdapter) GetStatus() domain.PlaybackInfo {
	return a.globalService.GetStatus()
}

// GetQueueLength returns the current queue length
func (a *AudioPlayerServiceAdapter) GetQueueLength() int {
	return a.globalService.GetQueueLength()
}

// ClearQueue clears all items from the queue
func (a *AudioPlayerServiceAdapter) ClearQueue() {
	a.globalService.ClearQueue()
}

// Close closes the audio player service and releases resources
func (a *AudioPlayerServiceAdapter) Close() error {
	return a.globalService.Close()
}
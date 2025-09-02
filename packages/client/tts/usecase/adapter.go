package usecase

import (
	"context"
	"fmt"
	"os"

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

// PlayAudioFile plays an audio file directly (for progressive playback)
// WARNING: This bypasses the queue system entirely - use with caution
func (a *AudioPlayerServiceAdapter) PlayAudioFile(ctx context.Context, filePath string, format domain.OutputFormat) error {
	// Open the file and play using the global service's player
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()
	
	// ISSUE: This bypasses queue modes (immediate/queue/no_queue)
	// For proper queue support, should use PlayRequest with pre-synthesized file
	return a.globalService.player.Play(ctx, file, format)
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

// PlayRequestWithHistory plays audio with history saving using streaming synthesis
func (a *AudioPlayerServiceAdapter) PlayRequestWithHistory(ctx context.Context, request *domain.PlaybackRequest, historyFilePath string) error {
    // Respect playback mode and wait_for_end by delegating to global service
    return a.globalService.PlayRequestWithHistory(ctx, request, historyFilePath)
}

// GetGlobalService returns the underlying GlobalAudioPlayerService
func (a *AudioPlayerServiceAdapter) GetGlobalService() *GlobalAudioPlayerService {
	return a.globalService
}

// getOutputFormatFromRequest extracts output format from playback request
func getOutputFormatFromRequest(request *domain.PlaybackRequest) domain.OutputFormat {
	if request.TTSRequest != nil && request.TTSRequest.OutputFormat != nil {
		return *request.TTSRequest.OutputFormat
	}
	return domain.OutputFormatWAV // default
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

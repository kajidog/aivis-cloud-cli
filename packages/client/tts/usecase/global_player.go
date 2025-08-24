package usecase

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/kajidog/aivis-cloud-cli/client/common/logger"
	"github.com/kajidog/aivis-cloud-cli/client/tts/domain"
)

// AudioPlayerConfig holds configuration for the audio player service
type AudioPlayerConfig struct {
	MaxQueueSize int
}

// GlobalAudioPlayerService manages audio playback with global singleton pattern
type GlobalAudioPlayerService struct {
	ttsService *TTSSynthesizer
	player     domain.AudioPlayer
	config     *AudioPlayerConfig
	logger     logger.Logger

	mu         sync.RWMutex
	queue      []queueItem
	processing bool
	
	// Worker goroutine management
	workerCtx    context.Context
	workerCancel context.CancelFunc
	workerDone   chan struct{}
}

var (
	globalPlayerInstance *GlobalAudioPlayerService
	globalPlayerOnce     sync.Once
)

// GetGlobalAudioPlayerService returns the singleton instance
func GetGlobalAudioPlayerService() *GlobalAudioPlayerService {
	globalPlayerOnce.Do(func() {
		globalPlayerInstance = &GlobalAudioPlayerService{
			queue:      make([]queueItem, 0),
			workerDone: make(chan struct{}),
		}
	})
	return globalPlayerInstance
}

// Initialize initializes the global audio player service
func (s *GlobalAudioPlayerService) Initialize(ttsService *TTSSynthesizer, player domain.AudioPlayer, config *AudioPlayerConfig) {
	s.InitializeWithLogger(ttsService, player, config, logger.NewNoop())
}

// InitializeWithLogger initializes the global audio player service with logger
func (s *GlobalAudioPlayerService) InitializeWithLogger(ttsService *TTSSynthesizer, player domain.AudioPlayer, config *AudioPlayerConfig, log logger.Logger) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.ttsService = ttsService
	s.player = player
	s.config = config
	if log == nil {
		s.logger = logger.NewNoop()
	} else {
		s.logger = log
	}
	
	// Start worker goroutine if not already running
	if s.workerCtx == nil {
		s.workerCtx, s.workerCancel = context.WithCancel(context.Background())
		go s.queueWorker()
	}
}

// PlayRequest plays audio with the specified playback mode (asynchronous or synchronous)
func (s *GlobalAudioPlayerService) PlayRequest(ctx context.Context, request *domain.PlaybackRequest) error {
	if s.ttsService == nil || s.player == nil {
		return fmt.Errorf("audio player service not initialized")
	}
	
	// Handle wait_for_end flag - if true, use synchronous playback regardless of mode
	if request.WaitForEnd != nil && *request.WaitForEnd {
		return s.playSynchronous(ctx, request)
	}
	
	// Normal asynchronous playback
	if request.Mode == nil {
		return s.playImmediate(ctx, request)
	}
	
	switch *request.Mode {
	case domain.PlaybackModeImmediate:
		return s.playImmediate(ctx, request)
	case domain.PlaybackModeQueue:
		return s.addToQueue(ctx, request)
	case domain.PlaybackModeNoQueue:
		return s.playWithoutQueue(ctx, request)
	default:
		return s.playImmediate(ctx, request)
	}
}

// playImmediate stops current playback and plays new audio immediately
func (s *GlobalAudioPlayerService) playImmediate(ctx context.Context, request *domain.PlaybackRequest) error {
	// Stop current playback and clear queue
	s.Stop()
	
	// Play immediately in background
	go func() {
		s.synthesizeAndPlayStream(ctx, request, getOutputFormat(request))
	}()
	
	return nil
}

// addToQueue adds request to queue for sequential playback (asynchronous)
func (s *GlobalAudioPlayerService) addToQueue(ctx context.Context, request *domain.PlaybackRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Check queue size limit
	if len(s.queue) >= s.config.MaxQueueSize {
		return fmt.Errorf("queue is full (max size: %d)", s.config.MaxQueueSize)
	}
	
	// Add to queue (no done channel needed for asynchronous operation)
	item := queueItem{
		request: request,
		ctx:     ctx,
		done:    nil, // Asynchronous - no waiting
	}
	
	s.queue = append(s.queue, item)
	
	// Update queue length if player supports it
	if setter, ok := s.player.(interface{ SetQueueLength(int) }); ok {
		setter.SetQueueLength(len(s.queue))
	}
	
	return nil
}

// playWithoutQueue plays audio without queue management (asynchronous)
func (s *GlobalAudioPlayerService) playWithoutQueue(ctx context.Context, request *domain.PlaybackRequest) error {
	// Play in background without affecting current playback or queue
	go func() {
		s.synthesizeAndPlayStream(ctx, request, getOutputFormat(request))
	}()
	
	return nil
}

// queueWorker processes queue items sequentially in background
func (s *GlobalAudioPlayerService) queueWorker() {
	defer close(s.workerDone)
	
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-s.workerCtx.Done():
			return
		case <-ticker.C:
			s.processNextQueueItem()
		}
	}
}

// processNextQueueItem processes the next item in queue if player is idle
func (s *GlobalAudioPlayerService) processNextQueueItem() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Skip if already processing or queue is empty
	if s.processing || len(s.queue) == 0 {
		return
	}
	
	// Skip if player is currently playing
	if s.player.IsPlaying() {
		return
	}
	
	// Get next item from queue
	item := s.queue[0]
	s.queue = s.queue[1:]
	s.processing = true
	
	// Update queue length
	if setter, ok := s.player.(interface{ SetQueueLength(int) }); ok {
		setter.SetQueueLength(len(s.queue))
	}
	
	// Process item in background
	go func() {
		defer func() {
			s.mu.Lock()
			s.processing = false
			s.mu.Unlock()
		}()
		
		err := s.synthesizeAndPlayStream(item.ctx, item.request, getOutputFormat(item.request))
		if err != nil {
			s.logger.Error("Queue playback error", logger.Error(err))
		}
	}()
}

// synthesizeAndPlayStream performs streaming TTS synthesis with progressive playback
func (s *GlobalAudioPlayerService) synthesizeAndPlayStream(ctx context.Context, request *domain.PlaybackRequest, format domain.OutputFormat) error {
	// Validate TTS request
	if err := s.ttsService.ValidateRequest(request.TTSRequest); err != nil {
		return fmt.Errorf("invalid TTS request: %w", err)
	}
	
	// Set current text for status reporting
	if player, ok := s.player.(interface{ SetCurrentText(string) }); ok {
		player.SetCurrentText(request.TTSRequest.Text)
	}
	
	// Apply volume if specified
	if request.Volume != nil {
		if err := s.player.SetVolume(*request.Volume); err != nil {
			return fmt.Errorf("failed to set volume: %w", err)
		}
	}
	
	// Use streaming synthesis with progressive playback
	return s.streamingSynthesisAndPlay(ctx, request, format)
}

// streamingSynthesisAndPlay performs streaming synthesis and playback
func (s *GlobalAudioPlayerService) streamingSynthesisAndPlay(ctx context.Context, request *domain.PlaybackRequest, format domain.OutputFormat) error {
	// Create a pipe for streaming audio data
	pipeReader, pipeWriter := io.Pipe()
	defer pipeReader.Close()
	
	// Error channel for goroutine communication
	errChan := make(chan error, 2)
	
	// Start playback in a goroutine (will wait for first chunk)
	go func() {
		// Create independent context for audio playback to prevent premature cancellation
		playbackCtx := context.Background()
		err := s.player.Play(playbackCtx, pipeReader, format)
		if err != nil {
			errChan <- fmt.Errorf("playback failed: %w", err)
		} else {
			errChan <- nil
		}
	}()
	
	// Create streaming handler that writes to pipe
	handler := &streamingPlaybackHandler{
		writer:      pipeWriter,
		firstChunk:  true,
		startTime:   time.Now(),
		logger:      s.logger,
	}
	
	// Start streaming synthesis with independent context
	go func() {
		// Create independent context for TTS synthesis to prevent premature cancellation
		synthesisCtx := context.Background()
		err := s.ttsService.SynthesizeStream(synthesisCtx, request.TTSRequest, handler)
		pipeWriter.Close() // Close writer when done
		if err != nil {
			errChan <- fmt.Errorf("synthesis failed: %w", err)
		} else {
			errChan <- nil
		}
	}()
	
	// Wait for both synthesis and playback to complete
	var synthErr, playErr error
	for i := 0; i < 2; i++ {
		if err := <-errChan; err != nil {
			if synthErr == nil {
				synthErr = err
			} else {
				playErr = err
			}
		}
	}
	
	// Return the first error if any
	if synthErr != nil {
		return synthErr
	}
	return playErr
}

// Stop stops current playback and clears queue
func (s *GlobalAudioPlayerService) Stop() error {
	s.mu.Lock()
	s.queue = make([]queueItem, 0)
	s.processing = false
	s.mu.Unlock()
	
	return s.player.Stop()
}

// Pause pauses current playback
func (s *GlobalAudioPlayerService) Pause() error {
	return s.player.Pause()
}

// Resume resumes paused playback
func (s *GlobalAudioPlayerService) Resume() error {
	return s.player.Resume()
}

// SetVolume sets playback volume
func (s *GlobalAudioPlayerService) SetVolume(volume float64) error {
	return s.player.SetVolume(volume)
}

// GetStatus returns current playback status
func (s *GlobalAudioPlayerService) GetStatus() domain.PlaybackInfo {
	s.mu.RLock()
	queueLen := len(s.queue)
	s.mu.RUnlock()
	
	info := s.player.GetStatus()
	info.QueueLength = queueLen
	
	return info
}

// GetQueueLength returns the current queue length
func (s *GlobalAudioPlayerService) GetQueueLength() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.queue)
}

// ClearQueue clears all items from the queue
func (s *GlobalAudioPlayerService) ClearQueue() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.queue = make([]queueItem, 0)
}

// Close closes the audio player service and releases resources
func (s *GlobalAudioPlayerService) Close() error {
	s.Stop()
	
	// Cancel worker goroutine
	if s.workerCancel != nil {
		s.workerCancel()
		<-s.workerDone
	}
	
	return s.player.Close()
}

// playSynchronous performs synchronous playback that waits for completion
func (s *GlobalAudioPlayerService) playSynchronous(ctx context.Context, request *domain.PlaybackRequest) error {
	// For synchronous playback, determine the mode
	mode := domain.PlaybackModeImmediate // default
	if request.Mode != nil {
		mode = *request.Mode
	}
	
	switch mode {
	case domain.PlaybackModeImmediate:
		// For immediate mode with wait_for_end, stop current and play synchronously
		s.Stop()
		return s.synthesizeAndPlayStreamSync(ctx, request, getOutputFormat(request))
	case domain.PlaybackModeQueue:
		// For queue mode with wait_for_end, add to queue and wait for completion
		return s.addToQueueSync(ctx, request)
	case domain.PlaybackModeNoQueue:
		// For no_queue mode with wait_for_end, play without affecting queue but wait
		return s.synthesizeAndPlayStreamSync(ctx, request, getOutputFormat(request))
	default:
		return s.synthesizeAndPlayStreamSync(ctx, request, getOutputFormat(request))
	}
}

// addToQueueSync adds request to queue and waits for completion
func (s *GlobalAudioPlayerService) addToQueueSync(ctx context.Context, request *domain.PlaybackRequest) error {
	// Add to queue normally
	if err := s.addToQueue(ctx, request); err != nil {
		return err
	}
	
	// Wait for playback completion by monitoring status
	for {
		status := s.GetStatus()
		if status.Status == domain.PlaybackStatusIdle || 
		   status.Status == domain.PlaybackStatusStopped {
			break
		}
		
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
			continue
		}
	}
	
	return nil
}

// synthesizeAndPlayStreamSync performs synchronous streaming synthesis and playback
func (s *GlobalAudioPlayerService) synthesizeAndPlayStreamSync(ctx context.Context, request *domain.PlaybackRequest, format domain.OutputFormat) error {
	// Use the existing streaming synthesis but wait for completion
	return s.synthesizeAndPlayStream(ctx, request, format)
	// Note: The existing method already waits for both synthesis and playback completion
}

// Helper functions

func getOutputFormat(request *domain.PlaybackRequest) domain.OutputFormat {
	format := domain.OutputFormatWAV // default
	if request.TTSRequest.OutputFormat != nil {
		format = *request.TTSRequest.OutputFormat
	}
	return format
}
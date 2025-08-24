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

// AudioPlayerService manages audio playback with queue support
type AudioPlayerService struct {
	ttsService *TTSSynthesizer
	player     domain.AudioPlayer
	config     *domain.PlaybackConfig
	queue      []queueItem
	logger     logger.Logger
	mu         sync.RWMutex
	processing bool
}

// queueItem represents an item in the playback queue
type queueItem struct {
	request *domain.PlaybackRequest
	ctx     context.Context
	done    chan error
}

// NewAudioPlayerService creates a new audio player service
func NewAudioPlayerService(ttsService *TTSSynthesizer, player domain.AudioPlayer, config *domain.PlaybackConfig) *AudioPlayerService {
	return NewAudioPlayerServiceWithLogger(ttsService, player, config, logger.NewNoop())
}

// NewAudioPlayerServiceWithLogger creates a new audio player service with logger
func NewAudioPlayerServiceWithLogger(ttsService *TTSSynthesizer, player domain.AudioPlayer, config *domain.PlaybackConfig, log logger.Logger) *AudioPlayerService {
	if config == nil {
		config = domain.DefaultPlaybackConfig()
	}
	if log == nil {
		log = logger.NewNoop()
	}
	
	return &AudioPlayerService{
		ttsService: ttsService,
		player:     player,
		config:     config,
		queue:      make([]queueItem, 0),
		logger:     log,
	}
}

// PlayText plays the given text using TTS synthesis and audio playback
func (s *AudioPlayerService) PlayText(ctx context.Context, text, modelUUID string, options *domain.PlaybackRequest) error {
	// Create TTS request if not provided in options
	var ttsRequest *domain.TTSRequest
	if options != nil && options.TTSRequest != nil {
		ttsRequest = options.TTSRequest
	} else {
		ttsRequest = domain.NewTTSRequestBuilder(modelUUID, text).Build()
	}
	
	// Create playback request
	playbackRequest := &domain.PlaybackRequest{
		TTSRequest: ttsRequest,
	}
	
	// Apply options
	if options != nil {
		if options.Mode != nil {
			playbackRequest.Mode = options.Mode
		}
		if options.Volume != nil {
			playbackRequest.Volume = options.Volume
		}
		if options.StartOffset != nil {
			playbackRequest.StartOffset = options.StartOffset
		}
		if options.FadeInDuration != nil {
			playbackRequest.FadeInDuration = options.FadeInDuration
		}
		if options.FadeOutDuration != nil {
			playbackRequest.FadeOutDuration = options.FadeOutDuration
		}
	}
	
	// Set default mode if not specified
	if playbackRequest.Mode == nil {
		mode := s.config.DefaultMode
		playbackRequest.Mode = &mode
	}
	
	return s.PlayRequest(ctx, playbackRequest)
}

// PlayRequest plays audio based on the playback request
func (s *AudioPlayerService) PlayRequest(ctx context.Context, request *domain.PlaybackRequest) error {
	mode := s.config.DefaultMode
	if request.Mode != nil {
		mode = *request.Mode
	}
	
	switch mode {
	case domain.PlaybackModeImmediate:
		return s.playImmediate(ctx, request)
	case domain.PlaybackModeQueue:
		return s.addToQueue(ctx, request)
	case domain.PlaybackModeNoQueue:
		return s.playWithoutQueue(ctx, request)
	default:
		return fmt.Errorf("unsupported playback mode: %s", mode)
	}
}

// playImmediate stops current playback and plays new audio immediately
func (s *AudioPlayerService) playImmediate(ctx context.Context, request *domain.PlaybackRequest) error {
	// Stop current playback
	s.player.Stop()
	
	// Clear queue
	s.mu.Lock()
	s.queue = make([]queueItem, 0)
	s.mu.Unlock()
	
	return s.synthesizeAndPlay(ctx, request)
}

// addToQueue adds request to queue for sequential playback
func (s *AudioPlayerService) addToQueue(ctx context.Context, request *domain.PlaybackRequest) error {
	done := make(chan error, 1)
	
	// Add item to queue with lock
	func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		
		// Check queue size limit
		if len(s.queue) >= s.config.MaxQueueSize {
			done <- fmt.Errorf("queue is full (max size: %d)", s.config.MaxQueueSize)
			return
		}
		
		item := queueItem{
			request: request,
			ctx:     ctx,
			done:    done,
		}
		
		s.queue = append(s.queue, item)
		
		// Update queue length if player supports it
		if setter, ok := s.player.(interface{ SetQueueLength(int) }); ok {
			setter.SetQueueLength(len(s.queue))
		}
		
		// Start processing queue if not already processing
		if !s.processing && len(s.queue) == 1 {
			go s.processQueue()
		}
	}()
	
	// Wait for this item to be processed (outside the lock)
	return <-done
}

// processQueue processes items in the queue sequentially
func (s *AudioPlayerService) processQueue() {
	s.mu.Lock()
	s.processing = true
	s.mu.Unlock()
	
	defer func() {
		s.mu.Lock()
		s.processing = false
		s.mu.Unlock()
	}()
	
	for {
		s.mu.Lock()
		if len(s.queue) == 0 {
			s.mu.Unlock()
			break
		}
		
		item := s.queue[0]
		s.queue = s.queue[1:]
		queueLen := len(s.queue)
		s.mu.Unlock()
		
		// Update queue length if player supports it
		if setter, ok := s.player.(interface{ SetQueueLength(int) }); ok {
			setter.SetQueueLength(queueLen)
		}
		
		// Play the item
		err := s.synthesizeAndPlay(item.ctx, item.request)
		item.done <- err
		
		if err != nil {
			// Continue processing other items even if one fails
			continue
		}
		
		// Wait for current playback to complete
		for s.player.IsPlaying() {
			select {
			case <-item.ctx.Done():
				s.player.Stop()
				return
			default:
				// Small delay to avoid busy waiting
				// time.Sleep(10 * time.Millisecond)
			}
		}
	}
}

// playWithoutQueue plays audio without queue management
func (s *AudioPlayerService) playWithoutQueue(ctx context.Context, request *domain.PlaybackRequest) error {
	// Create a separate player instance for concurrent playback
	// For now, we'll use the same player but allow overlapping
	return s.synthesizeAndPlay(ctx, request)
}

// synthesizeAndPlay performs TTS synthesis and plays the audio
func (s *AudioPlayerService) synthesizeAndPlay(ctx context.Context, request *domain.PlaybackRequest) error {
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
	
	// Determine output format
	format := domain.OutputFormatWAV // default
	if request.TTSRequest.OutputFormat != nil {
		format = *request.TTSRequest.OutputFormat
	}
	
	// Use streaming synthesis with progressive playback
	return s.synthesizeAndPlayStream(ctx, request, format)
}

// Stop stops current playback and clears queue
func (s *AudioPlayerService) Stop() error {
	s.mu.Lock()
	s.queue = make([]queueItem, 0)
	s.mu.Unlock()
	
	return s.player.Stop()
}

// Pause pauses current playback
func (s *AudioPlayerService) Pause() error {
	return s.player.Pause()
}

// Resume resumes paused playback
func (s *AudioPlayerService) Resume() error {
	return s.player.Resume()
}

// SetVolume sets playback volume
func (s *AudioPlayerService) SetVolume(volume float64) error {
	return s.player.SetVolume(volume)
}

// GetStatus returns current playback status
func (s *AudioPlayerService) GetStatus() domain.PlaybackInfo {
	s.mu.RLock()
	queueLen := len(s.queue)
	s.mu.RUnlock()
	
	info := s.player.GetStatus()
	info.QueueLength = queueLen
	
	return info
}

// GetQueueLength returns the current queue length
func (s *AudioPlayerService) GetQueueLength() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.queue)
}

// ClearQueue clears all items from the queue
func (s *AudioPlayerService) ClearQueue() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.queue = make([]queueItem, 0)
}

// synthesizeAndPlayStream performs streaming TTS synthesis with progressive playback
func (s *AudioPlayerService) synthesizeAndPlayStream(ctx context.Context, request *domain.PlaybackRequest, format domain.OutputFormat) error {
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
	
	// Start streaming synthesis
	go func() {
		err := s.ttsService.SynthesizeStream(ctx, request.TTSRequest, handler)
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

// streamingPlaybackHandler handles streaming TTS for progressive playback
type streamingPlaybackHandler struct {
	writer     io.WriteCloser
	firstChunk bool
	startTime  time.Time
	logger     logger.Logger
}

func (h *streamingPlaybackHandler) OnChunk(chunk *domain.TTSStreamChunk) error {
	if h.firstChunk {
		h.firstChunk = false
		h.logger.Info("First audio chunk received - starting playback", 
			logger.Duration("elapsed", time.Since(h.startTime)),
		)
	}
	
	// Write chunk to pipe for immediate playback
	_, err := h.writer.Write(chunk.Data)
	return err
}

func (h *streamingPlaybackHandler) OnComplete() error {
	h.logger.Info("Streaming synthesis completed", 
		logger.Duration("elapsed", time.Since(h.startTime)),
	)
	return nil
}

func (h *streamingPlaybackHandler) OnError(err error) {
	h.logger.Error("Streaming error occurred", logger.Error(err))
}

// Close closes the audio player service and releases resources
func (s *AudioPlayerService) Close() error {
	s.Stop()
	return s.player.Close()
}
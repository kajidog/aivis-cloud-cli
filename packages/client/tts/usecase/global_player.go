package usecase

import (
    "context"
    "fmt"
    "io"
    "os"
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

    // Factory to create new independent players for no_queue concurrent playback
    newPlayerFactory func() domain.AudioPlayer
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

// SetNewPlayerFactory sets a factory function to create independent players for no_queue playback
func (s *GlobalAudioPlayerService) SetNewPlayerFactory(factory func() domain.AudioPlayer) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.newPlayerFactory = factory
}

// PlayRequest plays audio with the specified playback mode (asynchronous or synchronous)
func (s *GlobalAudioPlayerService) PlayRequest(ctx context.Context, request *domain.PlaybackRequest) error {
	if s.ttsService == nil || s.player == nil {
		return fmt.Errorf("audio player service not initialized")
	}
	
	// Debug logging for queue behavior investigation
	mode := "nil"
	if request.Mode != nil {
		mode = string(*request.Mode)
	}
	waitForEnd := false
	if request.WaitForEnd != nil {
		waitForEnd = *request.WaitForEnd
	}
	s.logger.Info("PlayRequest called", 
		logger.String("mode", mode),
		logger.Bool("wait_for_end", waitForEnd),
		logger.Bool("is_playing", s.player.IsPlaying()),
		logger.Bool("processing", s.processing),
		logger.Int("queue_length", len(s.queue)))
	
	// Handle wait_for_end flag - if true, use synchronous playback regardless of mode
	if request.WaitForEnd != nil && *request.WaitForEnd {
		s.logger.Info("Using synchronous playback (wait_for_end=true)")
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

// PlayRequestWithHistory plays audio respecting mode and wait_for_end while saving to history.
func (s *GlobalAudioPlayerService) PlayRequestWithHistory(ctx context.Context, request *domain.PlaybackRequest, historyFilePath string) error {
    if request == nil || request.TTSRequest == nil {
        return fmt.Errorf("invalid request")
    }

    // Handle wait_for_end flag - synchronous path
    if request.WaitForEnd != nil && *request.WaitForEnd {
        // Determine mode (default: queue for safety to avoid interrupting current playback)
        mode := domain.PlaybackModeQueue
        if request.Mode != nil {
            mode = *request.Mode
        }
        switch mode {
        case domain.PlaybackModeImmediate:
            // Immediate: stop current playback and clear queue, then play synchronously
            s.logger.Info("Synchronous immediate mode - stopping current playback (with history)")
            s.Stop()
            return s.synthesizeAndPlayStreamSyncWithHistory(ctx, request, getOutputFormat(request), historyFilePath)
        case domain.PlaybackModeQueue:
            return s.addToQueueSyncWithHistory(ctx, request, historyFilePath)
        case domain.PlaybackModeNoQueue:
            // Use independent player and wait synchronously
            var tempPlayer domain.AudioPlayer
            s.mu.RLock(); factory := s.newPlayerFactory; s.mu.RUnlock()
            if factory != nil { tempPlayer = factory() } else { tempPlayer = s.player }
            defer tempPlayer.Close()
            if request.Volume != nil { _ = tempPlayer.SetVolume(*request.Volume) }
            return s.streamingSynthesisAndPlayWithPlayerAndHistory(ctx, request, getOutputFormat(request), tempPlayer, historyFilePath)
        default:
            return s.addToQueueSyncWithHistory(ctx, request, historyFilePath)
        }
    }

    // Asynchronous path
    if request.Mode == nil {
        // Default to queue for history-safe behavior (do not interrupt current playback)
        return s.addToQueueWithHistory(ctx, request, historyFilePath)
    }
    switch *request.Mode {
    case domain.PlaybackModeImmediate:
        // Immediate: stop current playback and clear queue, then play asynchronously
        s.Stop()
        go func() { _ = s.streamingSynthesisAndPlayWithHistory(ctx, request, getOutputFormat(request), historyFilePath) }()
        return nil
    case domain.PlaybackModeQueue:
        return s.addToQueueWithHistory(ctx, request, historyFilePath)
    case domain.PlaybackModeNoQueue:
        return s.playWithoutQueueWithHistory(ctx, request, historyFilePath)
    default:
        return s.addToQueueWithHistory(ctx, request, historyFilePath)
    }
}

func (s *GlobalAudioPlayerService) addToQueueWithHistory(ctx context.Context, request *domain.PlaybackRequest, historyFilePath string) error {
    s.mu.Lock(); defer s.mu.Unlock()
    if len(s.queue) >= s.config.MaxQueueSize {
        return fmt.Errorf("queue is full (max size: %d)", s.config.MaxQueueSize)
    }
    item := queueItem{ request: request, ctx: ctx, done: nil, historyFilePath: historyFilePath }
    s.queue = append(s.queue, item)
    if setter, ok := s.player.(interface{ SetQueueLength(int) }); ok { setter.SetQueueLength(len(s.queue)) }
    shouldProcess := !s.processing && !s.player.IsPlaying()
    s.mu.Unlock()
    if shouldProcess { s.processNextQueueItem() }
    s.mu.Lock()
    return nil
}

func (s *GlobalAudioPlayerService) addToQueueSyncWithHistory(ctx context.Context, request *domain.PlaybackRequest, historyFilePath string) error {
    s.mu.Lock()
    if len(s.queue) >= s.config.MaxQueueSize {
        s.mu.Unlock()
        return fmt.Errorf("queue is full (max size: %d)", s.config.MaxQueueSize)
    }
    done := make(chan error, 1)
    item := queueItem{ request: request, ctx: ctx, done: done, historyFilePath: historyFilePath }
    s.queue = append(s.queue, item)
    if setter, ok := s.player.(interface{ SetQueueLength(int) }); ok { setter.SetQueueLength(len(s.queue)) }
    shouldProcess := !s.processing && !s.player.IsPlaying()
    s.mu.Unlock()
    if shouldProcess { s.processNextQueueItem() }
    select {
    case err := <-done: return err
    case <-ctx.Done(): return ctx.Err()
    }
}

func (s *GlobalAudioPlayerService) playWithoutQueueWithHistory(ctx context.Context, request *domain.PlaybackRequest, historyFilePath string) error {
    go func() {
        var tempPlayer domain.AudioPlayer
        s.mu.RLock(); factory := s.newPlayerFactory; s.mu.RUnlock()
        if factory != nil { tempPlayer = factory() } else { tempPlayer = s.player }
        defer tempPlayer.Close()
        if request.Volume != nil { _ = tempPlayer.SetVolume(*request.Volume) }
        _ = s.streamingSynthesisAndPlayWithPlayerAndHistory(ctx, request, getOutputFormat(request), tempPlayer, historyFilePath)
    }()
    return nil
}

// playImmediate stops current playback and plays new audio immediately
func (s *GlobalAudioPlayerService) playImmediate(ctx context.Context, request *domain.PlaybackRequest) error {
	// Stop current playback and clear queue
	s.Stop()
	
	// Add small delay to ensure previous playback process is fully stopped
	// This prevents audio overlap when multiple immediate requests come quickly
	time.Sleep(50 * time.Millisecond)
	
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
	
	// Only trigger queue processing if no playback is active
	// This prevents interference with existing playback
	shouldProcess := !s.processing && !s.player.IsPlaying()
	s.mu.Unlock() // Unlock before calling processNextQueueItem to avoid deadlock
	
	if shouldProcess {
		s.processNextQueueItem()
	}
	
	s.mu.Lock() // Re-lock for defer
	
	return nil
}

// playWithoutQueue plays audio without queue management (asynchronous)
func (s *GlobalAudioPlayerService) playWithoutQueue(ctx context.Context, request *domain.PlaybackRequest) error {
    // Play in background using a dedicated, short-lived player instance to allow overlap
    go func() {
        // Create a separate player (if factory is available) to avoid killing current playback
        var tempPlayer domain.AudioPlayer
        s.mu.RLock()
        factory := s.newPlayerFactory
        s.mu.RUnlock()
        if factory != nil {
            tempPlayer = factory()
        } else {
            // Fallback: use shared player (may interrupt current playback)
            tempPlayer = s.player
        }
        defer tempPlayer.Close()

        // Validate request
        if err := s.ttsService.ValidateRequest(request.TTSRequest); err != nil {
            s.logger.Error("Invalid TTS request for no_queue: " + err.Error())
            return
        }

        // Apply volume if specified
        if request.Volume != nil {
            _ = tempPlayer.SetVolume(*request.Volume)
        }

        // Run streaming synthesis with provided temp player
        if err := s.streamingSynthesisAndPlayWithPlayer(ctx, request, getOutputFormat(request), tempPlayer); err != nil {
            s.logger.Error("no_queue playback error: " + err.Error())
        }
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
			
			// CRITICAL: Process next item in queue after completion
			s.processNextQueueItem()
		}()
		
    var err error
    if item.historyFilePath != "" {
        err = s.streamingSynthesisAndPlayWithHistory(item.ctx, item.request, getOutputFormat(item.request), item.historyFilePath)
    } else {
        err = s.synthesizeAndPlayStream(item.ctx, item.request, getOutputFormat(item.request))
    }
		if err != nil {
			s.logger.Error("Queue playback error", logger.Error(err))
		}
		
		// Notify completion via done channel if present (synchronous operation)
		if item.done != nil {
			select {
			case item.done <- err:
			default:
				// Channel full or closed, ignore
			}
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
    return s.streamingSynthesisAndPlayWithHistory(ctx, request, format, "")
}

// streamingSynthesisAndPlayWithHistory performs streaming synthesis and playback with optional history saving
func (s *GlobalAudioPlayerService) streamingSynthesisAndPlayWithHistory(ctx context.Context, request *domain.PlaybackRequest, format domain.OutputFormat, historyFilePath string) error {
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
	
	// Create streaming handler that writes to pipe and optionally to history file
	var handler domain.TTSStreamHandler
	if historyFilePath != "" {
		// Create history file for concurrent writing
		historyFile, err := os.Create(historyFilePath)
		if err != nil {
			return fmt.Errorf("failed to create history file: %w", err)
		}
		defer historyFile.Close()
		
		// Use tee handler for both playback and history
		handler = &streamingPlaybackHistoryHandler{
			writer:       pipeWriter,
			historyFile:  historyFile,
			firstChunk:   true,
			startTime:    time.Now(),
			logger:       s.logger,
		}
	} else {
		// Use regular playback handler
		handler = &streamingPlaybackHandler{
			writer:      pipeWriter,
			firstChunk:  true,
			startTime:   time.Now(),
			logger:      s.logger,
		}
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

// streamingSynthesisAndPlayWithPlayer performs streaming synthesis and plays using the provided player
func (s *GlobalAudioPlayerService) streamingSynthesisAndPlayWithPlayer(ctx context.Context, request *domain.PlaybackRequest, format domain.OutputFormat, player domain.AudioPlayer) error {
    // Create a pipe for streaming audio data
    pipeReader, pipeWriter := io.Pipe()
    defer pipeReader.Close()

    errChan := make(chan error, 2)

    // Start playback on the provided player
    go func() {
        playbackCtx := context.Background()
        err := player.Play(playbackCtx, pipeReader, format)
        if err != nil { errChan <- fmt.Errorf("playback failed: %w", err) } else { errChan <- nil }
    }()

    // Use simple playback handler (no history)
    handler := &streamingPlaybackHandler{
        writer:     pipeWriter,
        firstChunk: true,
        startTime:  time.Now(),
        logger:     s.logger,
    }

    // Start synthesis
    go func() {
        synthesisCtx := context.Background()
        err := s.ttsService.SynthesizeStream(synthesisCtx, request.TTSRequest, handler)
        pipeWriter.Close()
        if err != nil { errChan <- fmt.Errorf("synthesis failed: %w", err) } else { errChan <- nil }
    }()

    // Wait for both
    var synthErr, playErr error
    for i := 0; i < 2; i++ {
        if err := <-errChan; err != nil {
            if synthErr == nil { synthErr = err } else { playErr = err }
        }
    }
    if synthErr != nil { return synthErr }
    return playErr
}

// streamingPlaybackHistoryHandler handles streaming data for both playback and history
type streamingPlaybackHistoryHandler struct {
	writer      io.WriteCloser
	historyFile *os.File
	firstChunk  bool
	startTime   time.Time
	logger      logger.Logger
}

func (h *streamingPlaybackHistoryHandler) OnChunk(chunk *domain.TTSStreamChunk) error {
	if h.firstChunk {
		h.firstChunk = false
		h.logger.Debug("First audio chunk received, starting progressive playback")
	}
	
	// Write to playback pipe
	if _, err := h.writer.Write(chunk.Data); err != nil {
		return fmt.Errorf("failed to write to playback stream: %w", err)
	}
	
	// Write to history file concurrently
	if h.historyFile != nil {
		if _, err := h.historyFile.Write(chunk.Data); err != nil {
			// Don't fail streaming for history write errors
			h.logger.Warn("Failed to write chunk to history file: " + err.Error())
		}
	}
	
	return nil
}

func (h *streamingPlaybackHistoryHandler) OnComplete() error {
	h.logger.Debug("Streaming synthesis completed", 
		logger.String("duration", time.Since(h.startTime).String()),
	)
	return nil
}

func (h *streamingPlaybackHistoryHandler) OnError(err error) {
	h.logger.Error("Streaming synthesis error: " + err.Error())
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
		s.logger.Info("Synchronous immediate mode - stopping current playback")
		s.Stop()
		return s.synthesizeAndPlayStreamSync(ctx, request, getOutputFormat(request))
	case domain.PlaybackModeQueue:
		// For queue mode with wait_for_end, add to queue and wait for completion
		s.logger.Info("Synchronous queue mode - adding to queue and waiting")
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
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Check queue size limit
	if len(s.queue) >= s.config.MaxQueueSize {
		return fmt.Errorf("queue is full (max size: %d)", s.config.MaxQueueSize)
	}
	
	// Create done channel for synchronous operation
	done := make(chan error, 1)
	
	// Add to queue with done channel for synchronous operation
	item := queueItem{
		request: request,
		ctx:     ctx,
		done:    done, // Synchronous - will wait for completion
	}
	
	s.queue = append(s.queue, item)
	
	// Update queue length if player supports it
	if setter, ok := s.player.(interface{ SetQueueLength(int) }); ok {
		setter.SetQueueLength(len(s.queue))
	}
	
	// Only trigger queue processing if no playback is active
	// This prevents interference with existing playback
	shouldProcess := !s.processing && !s.player.IsPlaying()
	s.mu.Unlock() // Unlock for waiting
	
	if shouldProcess {
		s.processNextQueueItem()
	}
	
	// Wait for completion via done channel
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// synthesizeAndPlayStreamSync performs synchronous streaming synthesis and playback
func (s *GlobalAudioPlayerService) synthesizeAndPlayStreamSync(ctx context.Context, request *domain.PlaybackRequest, format domain.OutputFormat) error {
    // Use the existing streaming synthesis but wait for completion
    return s.synthesizeAndPlayStream(ctx, request, format)
    // Note: The existing method already waits for both synthesis and playback completion
}

func (s *GlobalAudioPlayerService) synthesizeAndPlayStreamSyncWithHistory(ctx context.Context, request *domain.PlaybackRequest, format domain.OutputFormat, historyFilePath string) error {
    return s.streamingSynthesisAndPlayWithHistory(ctx, request, format, historyFilePath)
}

// Helper functions

func getOutputFormat(request *domain.PlaybackRequest) domain.OutputFormat {
	format := domain.OutputFormatWAV // default
	if request.TTSRequest.OutputFormat != nil {
		format = *request.TTSRequest.OutputFormat
	}
	return format
}

// streamingSynthesisAndPlayWithPlayerAndHistory performs streaming synthesis with concurrent playback via provided player and writes to history file
func (s *GlobalAudioPlayerService) streamingSynthesisAndPlayWithPlayerAndHistory(ctx context.Context, request *domain.PlaybackRequest, format domain.OutputFormat, player domain.AudioPlayer, historyFilePath string) error {
    pipeReader, pipeWriter := io.Pipe(); defer pipeReader.Close()
    errChan := make(chan error, 2)

    go func() {
        playbackCtx := context.Background()
        err := player.Play(playbackCtx, pipeReader, format)
        if err != nil { errChan <- fmt.Errorf("playback failed: %w", err) } else { errChan <- nil }
    }()

    // Open history file
    historyFile, err := os.Create(historyFilePath)
    if err != nil { return fmt.Errorf("failed to create history file: %w", err) }
    defer historyFile.Close()

    handler := &streamingPlaybackHistoryHandler{
        writer:      pipeWriter,
        historyFile: historyFile,
        firstChunk:  true,
        startTime:   time.Now(),
        logger:      s.logger,
    }

    go func() {
        synthesisCtx := context.Background()
        err := s.ttsService.SynthesizeStream(synthesisCtx, request.TTSRequest, handler)
        pipeWriter.Close()
        if err != nil { errChan <- fmt.Errorf("synthesis failed: %w", err) } else { errChan <- nil }
    }()

    var synthErr, playErr error
    for i := 0; i < 2; i++ {
        if err := <-errChan; err != nil {
            if synthErr == nil { synthErr = err } else { playErr = err }
        }
    }
    if synthErr != nil { return synthErr }
    return playErr
}

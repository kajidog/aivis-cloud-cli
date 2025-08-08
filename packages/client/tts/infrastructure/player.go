package infrastructure

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/kajidog/aiviscloud-mcp/client/tts/domain"
)

// OSCommandAudioPlayer implements AudioPlayer interface using OS commands
type OSCommandAudioPlayer struct {
	config        *domain.PlaybackConfig
	mu            sync.RWMutex
	status        domain.PlaybackStatus
	currentProc   *os.Process
	queueLength   int
	currentText   string
	volume        float64
	startTime     time.Time
	estimatedDuration time.Duration
}

// NewOSCommandAudioPlayer creates a new OS command-based audio player
func NewOSCommandAudioPlayer(config *domain.PlaybackConfig) *OSCommandAudioPlayer {
	if config == nil {
		config = domain.DefaultPlaybackConfig()
	}
	
	return &OSCommandAudioPlayer{
		config: config,
		status: domain.PlaybackStatusIdle,
		volume: config.DefaultVolume,
	}
}

// getAudioCommand returns the appropriate audio playback command for the current OS
func (p *OSCommandAudioPlayer) getAudioCommand(audioFile string) (string, []string, error) {
	switch runtime.GOOS {
	case "darwin": // macOS
		return "afplay", []string{audioFile}, nil
		
	case "windows":
		// Use PowerShell for Windows - simplified version
		escapedPath := strings.ReplaceAll(audioFile, `'`, `''`) // Escape single quotes for PowerShell
		psCommand := fmt.Sprintf("(New-Object Media.SoundPlayer '%s').PlaySync()", escapedPath)
		return "powershell", []string{"-c", psCommand}, nil
		
	case "linux":
		// Try to find available Linux audio players
		linuxPlayers := []string{"aplay", "paplay", "play", "ffplay"}
		
		for _, player := range linuxPlayers {
			if _, err := exec.LookPath(player); err == nil {
				args := []string{audioFile}
				if player == "ffplay" {
					args = []string{"-nodisp", "-autoexit", audioFile}
				}
				return player, args, nil
			}
		}
		
		// Default to aplay if nothing else is found
		return "aplay", []string{audioFile}, nil
		
	default:
		return "", nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// createTempAudioFile creates a temporary audio file from the reader
func (p *OSCommandAudioPlayer) createTempAudioFile(audioData io.Reader, format domain.OutputFormat) (string, error) {
	// Determine file extension
	var ext string
	switch format {
	case domain.OutputFormatWAV:
		ext = ".wav"
	case domain.OutputFormatMP3:
		ext = ".mp3"
	case domain.OutputFormatFLAC:
		ext = ".flac"
	case domain.OutputFormatAAC:
		ext = ".aac"
	case domain.OutputFormatOpus:
		ext = ".opus"
	default:
		ext = ".wav"
	}
	
	// Create temporary file
	tmpFile, err := os.CreateTemp("", "tts_audio_*"+ext)
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer tmpFile.Close()
	
	// Copy audio data to file
	_, err = io.Copy(tmpFile, audioData)
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to write audio data: %w", err)
	}
	
	return tmpFile.Name(), nil
}

// Play starts playback of the given audio stream
func (p *OSCommandAudioPlayer) Play(ctx context.Context, audioData io.Reader, format domain.OutputFormat) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// Stop current playback if any
	if p.currentProc != nil {
		p.currentProc.Kill()
		p.currentProc = nil
	}
	
	// Create temporary audio file
	audioFile, err := p.createTempAudioFile(audioData, format)
	if err != nil {
		return err
	}
	
	// Get command for current OS
	command, args, err := p.getAudioCommand(audioFile)
	if err != nil {
		os.Remove(audioFile)
		return err
	}
	
	// Start audio playback process
	cmd := exec.CommandContext(ctx, command, args...)
	
	// Configure process options (ignore output)
	cmd.Stdout = nil
	cmd.Stderr = nil
	
	// Start the process
	err = cmd.Start()
	if err != nil {
		os.Remove(audioFile)
		return fmt.Errorf("failed to start audio playback: %w", err)
	}
	
	p.currentProc = cmd.Process
	p.status = domain.PlaybackStatusPlaying
	p.startTime = time.Now()
	
	// 音声時間長の簡易推定
	p.estimatedDuration = p.estimateAudioDuration(audioFile, format)
	
	// Handle process completion in goroutine with position tracking
	go func() {
		defer os.Remove(audioFile) // Clean up temp file
		
		// Wait for process to complete - コマンドが終わるまで待機して終了検知
		err := cmd.Wait()
		actualDuration := time.Since(p.startTime)
		
		p.mu.Lock()
		p.currentProc = nil
		if err != nil && ctx.Err() == nil {
			// Process failed (not due to cancellation)
			p.status = domain.PlaybackStatusStopped
		} else {
			// Process completed successfully or was cancelled
			p.status = domain.PlaybackStatusIdle
		}
		p.mu.Unlock()
		
		// 実際の再生時間をログ出力
		fmt.Printf("Audio playback completed. Estimated: %v, Actual: %v\n", 
			p.estimatedDuration, actualDuration)
	}()
	
	return nil
}

// Stop stops current playback
func (p *OSCommandAudioPlayer) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if p.currentProc != nil {
		err := p.currentProc.Kill()
		p.currentProc = nil
		p.status = domain.PlaybackStatusStopped
		return err
	}
	
	p.status = domain.PlaybackStatusStopped
	return nil
}

// Pause pauses current playback
// Note: OS command-based playback doesn't support pause/resume
// This is a limitation of using external commands
func (p *OSCommandAudioPlayer) Pause() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if p.status == domain.PlaybackStatusPlaying {
		// We can't actually pause external processes, so we'll mark as paused
		// but the audio will continue playing
		p.status = domain.PlaybackStatusPaused
	}
	
	return nil
}

// Resume resumes paused playback  
// Note: OS command-based playback doesn't support pause/resume
func (p *OSCommandAudioPlayer) Resume() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if p.status == domain.PlaybackStatusPaused {
		p.status = domain.PlaybackStatusPlaying
	}
	
	return nil
}

// SetVolume sets the playback volume (0.0 to 1.0)
func (p *OSCommandAudioPlayer) SetVolume(volume float64) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// Clamp volume to valid range
	if volume < 0.0 {
		volume = 0.0
	}
	if volume > 1.0 {
		volume = 1.0
	}
	
	p.volume = volume
	return nil
}

// estimateAudioDuration estimates audio duration based on file size
func (p *OSCommandAudioPlayer) estimateAudioDuration(audioFile string, format domain.OutputFormat) time.Duration {
	stat, err := os.Stat(audioFile)
	if err != nil {
		return 5 * time.Second // Default fallback
	}
	
	// Simple estimation based on file size and format
	fileSize := stat.Size()
	switch format {
	case domain.OutputFormatWAV:
		// WAV: ~176400 bytes per second (44.1kHz, 16bit, stereo)
		return time.Duration(fileSize/176400) * time.Second
	case domain.OutputFormatMP3:
		// MP3: varies with bitrate, rough estimate for 128kbps
		return time.Duration(fileSize/16000) * time.Second
	case domain.OutputFormatFLAC:
		// FLAC: compressed but lossless, roughly 50% of WAV
		return time.Duration(fileSize/88000) * time.Second
	case domain.OutputFormatAAC:
		// AAC: similar to MP3
		return time.Duration(fileSize/16000) * time.Second
	case domain.OutputFormatOpus:
		// Opus: very efficient compression
		return time.Duration(fileSize/8000) * time.Second
	default:
		return time.Duration(fileSize/100000) * time.Second
	}
}

// GetStatus returns current playback status
func (p *OSCommandAudioPlayer) GetStatus() domain.PlaybackInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	// Calculate current position if playing
	var position time.Duration
	if p.status == domain.PlaybackStatusPlaying && !p.startTime.IsZero() {
		position = time.Since(p.startTime)
		if position > p.estimatedDuration {
			position = p.estimatedDuration
		}
	}
	
	return domain.PlaybackInfo{
		Status:      p.status,
		QueueLength: p.queueLength,
		CurrentText: p.currentText,
		Volume:      p.volume,
		Duration:    p.estimatedDuration,  // 推定時間長
		Position:    position,             // 現在位置（推定）
	}
}

// IsPlaying returns true if audio is currently playing
func (p *OSCommandAudioPlayer) IsPlaying() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	return p.status == domain.PlaybackStatusPlaying
}

// Close closes the audio player and releases resources
func (p *OSCommandAudioPlayer) Close() error {
	return p.Stop()
}

// SetCurrentText sets the current text being played (for status reporting)
func (p *OSCommandAudioPlayer) SetCurrentText(text string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.currentText = text
}

// SetQueueLength sets the current queue length (for status reporting)
func (p *OSCommandAudioPlayer) SetQueueLength(length int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.queueLength = length
}
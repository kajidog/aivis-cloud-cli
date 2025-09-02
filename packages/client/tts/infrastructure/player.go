package infrastructure

import (
    "context"
    "fmt"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
    "strings"
    "sync"
    "time"

    "github.com/kajidog/aivis-cloud-cli/client/common/logger"
    "github.com/kajidog/aivis-cloud-cli/client/tts/domain"
)

// OSCommandAudioPlayer implements AudioPlayer interface using OS commands
type OSCommandAudioPlayer struct {
	config        *domain.PlaybackConfig
	logger        logger.Logger
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
	return NewOSCommandAudioPlayerWithLogger(config, logger.NewNoop())
}

// NewOSCommandAudioPlayerWithLogger creates a new OS command-based audio player with logger
func NewOSCommandAudioPlayerWithLogger(config *domain.PlaybackConfig, log logger.Logger) *OSCommandAudioPlayer {
	if config == nil {
		config = domain.DefaultPlaybackConfig()
	}
	if log == nil {
		log = logger.NewNoop()
	}
	
	return &OSCommandAudioPlayer{
		config: config,
		logger: log,
		status: domain.PlaybackStatusIdle,
		volume: config.DefaultVolume,
	}
}

// getAudioCommand returns the appropriate audio playback command for the current OS
func (p *OSCommandAudioPlayer) getAudioCommand(audioFile string) (string, []string, error) {
    // Decide based on OS and file extension for codec compatibility
    ext := strings.ToLower(filepath.Ext(audioFile))
    isCompressed := ext == ".mp3" || ext == ".aac" || ext == ".opus" || ext == ".flac"
    switch runtime.GOOS {
    case "darwin": // macOS (afplay supports common formats)
        return "afplay", []string{audioFile}, nil
		
    case "windows":
        // Prefer ffplay on Windows if available
        if _, err := exec.LookPath("ffplay"); err == nil {
            return "ffplay", []string{"-loglevel", "error", "-nodisp", "-autoexit", "-i", audioFile}, nil
        }
        // Fallback: Use WPF MediaPlayer (presentationCore) as previously working implementation
        escapedPath := strings.ReplaceAll(audioFile, "\\", "\\\\")
        escapedPath = strings.ReplaceAll(escapedPath, "'", "''")
        psCommand := fmt.Sprintf(`Add-Type -AssemblyName presentationCore; $player = New-Object system.windows.media.mediaplayer; $player.open('%s'); $player.Volume = 0.5; $player.Play(); Start-Sleep 1; if ($player.NaturalDuration.HasTimeSpan) { Start-Sleep -s $player.NaturalDuration.TimeSpan.TotalSeconds; } else { Start-Sleep 5; }`, escapedPath)
        return "powershell", []string{"-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", psCommand}, nil
		
    case "linux":
        // Choose player based on codec needs
        var linuxPlayers []string
        if isCompressed {
            // Prefer ffplay (broad codec support), then SoX 'play'
            linuxPlayers = []string{"ffplay", "play"}
        } else {
            // WAV and simple PCM formats can use aplay/paplay efficiently
            linuxPlayers = []string{"aplay", "paplay", "play", "ffplay"}
        }

        for _, player := range linuxPlayers {
            if _, err := exec.LookPath(player); err == nil {
                args := []string{audioFile}
                if player == "ffplay" {
                    args = []string{"-nodisp", "-autoexit", audioFile}
                }
                return player, args, nil
            }
        }

        // Default to ffplay for compressed, aplay otherwise
        if isCompressed {
            return "ffplay", []string{"-nodisp", "-autoexit", audioFile}, nil
        }
        return "aplay", []string{audioFile}, nil
    
    default:
        return "", nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
    }
}

// getStreamingCommand tries to return a command that can consume audio from stdin
// The args returned DO NOT include the stdin placeholder; caller wires stdin via cmd.Stdin
func (p *OSCommandAudioPlayer) getStreamingCommand(format domain.OutputFormat) (string, []string, bool) {
    switch runtime.GOOS {
    case "darwin": // macOS: afplay can read from /dev/stdin
        if _, err := exec.LookPath("afplay"); err == nil {
            // Map volume to afplay's -v [0.0-1.0]
            vol := p.volume
            if vol < 0.0 {
                vol = 0.0
            }
            if vol > 1.0 {
                vol = 1.0
            }
            // Use /dev/stdin so the command consumes from stdin progressively
            return "afplay", []string{"-v", fmt.Sprintf("%.3f", vol), "/dev/stdin"}, true
        }
        return "", nil, false
    case "linux":
        // Prefer ffplay for broad codec support and stdin streaming
        if _, err := exec.LookPath("ffplay"); err == nil {
            // ffplay -nodisp -autoexit -i -  (volume: 0-100)
            vol := int(p.volume * 100)
            if vol < 0 {
                vol = 0
            }
            if vol > 100 {
                vol = 100
            }
            return "ffplay", []string{"-nodisp", "-autoexit", "-volume", fmt.Sprintf("%d", vol), "-i", "-"}, true
        }
        // Fallback to SoX 'play' if available; supports '-' as stdin
        if _, err := exec.LookPath("play"); err == nil {
            // play -q -v <vol> -   (SoX volume is linear; keep within 0.0-1.0)
            vol := p.volume
            if vol <= 0 {
                vol = 1.0
            }
            return "play", []string{"-q", "-v", fmt.Sprintf("%.3f", vol), "-"}, true
        }
        return "", nil, false
    case "windows":
        // Use ffplay for stdin streaming on Windows if available
        if _, err := exec.LookPath("ffplay"); err == nil {
            // ffplay -nodisp -autoexit -i -  (volume: 0-100)
            vol := int(p.volume * 100)
            if vol < 0 { vol = 0 }
            if vol > 100 { vol = 100 }
            return "ffplay", []string{"-loglevel", "error", "-nodisp", "-autoexit", "-volume", fmt.Sprintf("%d", vol), "-i", "-"}, true
        }
        return "", nil, false
    default:
        return "", nil, false
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
	bytesWritten, err := io.Copy(tmpFile, audioData)
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to write audio data: %w", err)
	}
	
	// Ensure file is completely written before closing
	if err := tmpFile.Sync(); err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to sync file: %w", err)
	}
	
	// Debug: Log file size
	p.logger.Debug("Created temporary audio file", 
		logger.String("file_path", tmpFile.Name()), 
		logger.Int64("bytes", bytesWritten),
	)
	
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

    // Prefer stdin streaming when supported by platform/player to avoid growing-file truncation
    if cmdName, args, ok := p.getStreamingCommand(format); ok {
        cmd := exec.CommandContext(ctx, cmdName, args...)
        cmd.Stdin = audioData
        cmd.Stdout = nil
        cmd.Stderr = os.Stderr

        p.logger.Debug("Executing streaming audio playback",
            logger.String("command", cmdName),
            logger.String("args", fmt.Sprintf("%v", args)),
        )

        if err := cmd.Start(); err != nil {
            return fmt.Errorf("failed to start streaming playback: %w", err)
        }

        p.currentProc = cmd.Process
        p.status = domain.PlaybackStatusPlaying
        p.startTime = time.Now()
        p.estimatedDuration = 0

        go func() {
            err := cmd.Wait()
            actualDuration := time.Since(p.startTime)
            p.mu.Lock()
            p.currentProc = nil
            if err != nil && ctx.Err() == nil {
                p.status = domain.PlaybackStatusStopped
            } else {
                p.status = domain.PlaybackStatusIdle
            }
            p.mu.Unlock()
            p.logger.Info("Audio playback completed (stream)", logger.Duration("actual_duration", actualDuration))
        }()

        return nil
    }

    // For formats friendly to progressive playback from a growing file, start playback
    // after the first chunk is written. For formats like WAV/FLAC, write fully first.
    supportsProg := (format == domain.OutputFormatMP3 || format == domain.OutputFormatAAC || format == domain.OutputFormatOpus)
    // On Windows without ffplay, progressive playback from a growing file is unreliable; disable it
    if runtime.GOOS == "windows" {
        if _, err := exec.LookPath("ffplay"); err != nil {
            supportsProg = false
        }
    }

    // Create temporary audio file
    tmpFile, err := os.CreateTemp("", "tts_audio_*")
    if err != nil {
        return fmt.Errorf("failed to create temporary file: %w", err)
    }
    audioFile := tmpFile.Name()
    // Ensure correct extension for players that rely on file extension
    // Close and rename to add extension
    tmpFile.Close()
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
    newPath := audioFile + ext
    if err := os.Rename(audioFile, newPath); err == nil {
        audioFile = newPath
    }

    // Open file for writing
    wf, err := os.OpenFile(audioFile, os.O_CREATE|os.O_WRONLY, 0600)
    if err != nil {
        os.Remove(audioFile)
        return fmt.Errorf("failed to open temp file for writing: %w", err)
    }

    // Channel to signal when initial bytes are flushed
    ready := make(chan struct{}, 1)
    copyDone := make(chan error, 1)

    // Start writer goroutine
    go func() {
        defer wf.Close()
        buf := make([]byte, 32*1024)
        var wrote int64
        for {
            n, rerr := audioData.Read(buf)
            if n > 0 {
                if _, werr := wf.Write(buf[:n]); werr != nil {
                    copyDone <- werr
                    return
                }
                // Signal readiness after first ~32KB for progressive formats
                if supportsProg {
                    wrote += int64(n)
                    if wrote >= 32*1024 {
                        select { case ready <- struct{}{}: default: }
                        supportsProg = false // signal only once
                    }
                }
            }
            if rerr == io.EOF {
                // Final signal if nothing sent yet
                select { case ready <- struct{}{}: default: }
                copyDone <- nil
                return
            }
            if rerr != nil {
                copyDone <- rerr
                return
            }
        }
    }()

    // Determine command now
    command, args, err := p.getAudioCommand(audioFile)
    if err != nil {
        os.Remove(audioFile)
        return err
    }
    // Windows: inject volume into PowerShell MediaPlayer if used
    if runtime.GOOS == "windows" && strings.EqualFold(command, "powershell") {
        if len(args) >= 2 {
            idx := len(args) - 1
            ps := args[idx]
            vol := fmt.Sprintf("%.3f", p.volume)
            ps = strings.Replace(ps, "$player.Volume = 0.5", "$player.Volume = "+vol, 1)
            args[idx] = ps
        }
    }

    // For progressive-friendly formats, wait for initial bytes then start player concurrently
    if format == domain.OutputFormatMP3 || format == domain.OutputFormatAAC || format == domain.OutputFormatOpus {
        // Wait for first chunk to be written (or writer completion)
        select {
        case <-ready:
        case err := <-copyDone:
            if err != nil {
                os.Remove(audioFile)
                return fmt.Errorf("failed while preparing progressive playback: %w", err)
            }
        case <-ctx.Done():
            os.Remove(audioFile)
            return ctx.Err()
        }

        cmd := exec.CommandContext(ctx, command, args...)
        // Avoid writing to parent's stdout (MCP stdio safety)
        cmd.Stdout = nil
        cmd.Stderr = os.Stderr

        p.logger.Debug("Executing progressive file playback",
            logger.String("command", command),
            logger.String("args", fmt.Sprintf("%v", args)),
            logger.String("file", audioFile),
        )

        if err := cmd.Start(); err != nil {
            os.Remove(audioFile)
            return fmt.Errorf("failed to start audio playback: %w", err)
        }

        p.currentProc = cmd.Process
        p.status = domain.PlaybackStatusPlaying
        p.startTime = time.Now()
        p.estimatedDuration = 0 // unknown until file complete

        go func() {
            // Wait for both writer and player
            werr := <-copyDone
            perr := cmd.Wait()
            actualDuration := time.Since(p.startTime)

            p.mu.Lock()
            p.currentProc = nil
            if (werr != nil || perr != nil) && ctx.Err() == nil {
                p.status = domain.PlaybackStatusStopped
            } else {
                p.status = domain.PlaybackStatusIdle
            }
            p.mu.Unlock()

            // Optional: keep playback files for debugging or external resume if env set
            if v := os.Getenv("AIVIS_KEEP_PLAYBACK_FILES"); strings.EqualFold(v, "1") || strings.EqualFold(v, "true") {
                p.logger.Info("Keeping playback file per env flag", logger.String("file", audioFile))
            } else {
                os.Remove(audioFile)
            }
            p.logger.Info("Audio playback completed (progressive)",
                logger.Duration("actual_duration", actualDuration),
            )
        }()

        return nil
    }

    // Non-progressive formats: finish writing then play
    if err := func() error {
        defer func() { select { case <-ready: default: } }()
        if werr := <-copyDone; werr != nil {
            return werr
        }
        return nil
    }(); err != nil {
        os.Remove(audioFile)
        return fmt.Errorf("failed to write audio file: %w", err)
    }

    cmd := exec.CommandContext(ctx, command, args...)
    cmd.Stdout = nil
    cmd.Stderr = os.Stderr

    p.logger.Debug("Executing audio playback command",
        logger.String("command", command),
        logger.String("args", fmt.Sprintf("%v", args)),
    )

    if err := cmd.Start(); err != nil {
        os.Remove(audioFile)
        return fmt.Errorf("failed to start audio playback: %w", err)
    }

    p.currentProc = cmd.Process
    p.status = domain.PlaybackStatusPlaying
    p.startTime = time.Now()
    p.estimatedDuration = p.estimateAudioDuration(audioFile, format)

    go func() {
        err := cmd.Wait()
        actualDuration := time.Since(p.startTime)

        p.mu.Lock()
        p.currentProc = nil
        if err != nil && ctx.Err() == nil {
            p.status = domain.PlaybackStatusStopped
        } else {
            p.status = domain.PlaybackStatusIdle
        }
        p.mu.Unlock()

        if v := os.Getenv("AIVIS_KEEP_PLAYBACK_FILES"); strings.EqualFold(v, "1") || strings.EqualFold(v, "true") {
            p.logger.Info("Keeping playback file per env flag", logger.String("file", audioFile))
        } else {
            os.Remove(audioFile)
        }
        p.logger.Info("Audio playback completed",
            logger.Duration("estimated_duration", p.estimatedDuration),
            logger.Duration("actual_duration", actualDuration),
        )
    }()

    return nil
}

// Stop stops current playback
func (p *OSCommandAudioPlayer) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if p.currentProc != nil {
		err := p.currentProc.Kill()
		// Wait a moment for process to actually terminate
		// This helps prevent audio overlap in immediate mode
		if err == nil {
			time.Sleep(10 * time.Millisecond)
		}
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
		position = min(time.Since(p.startTime), p.estimatedDuration)
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

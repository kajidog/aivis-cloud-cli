package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	ttsDomain "github.com/kajidog/aivis-cloud-cli/client/tts/domain"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/viper"
)

// SynthesizeSpeechParams parameters for synthesize_speech tool
type SynthesizeSpeechParams struct {
	Text               string  `json:"text"`
	ModelUUID          string  `json:"model_uuid,omitempty"`          // optional, uses config default
	Format             string  `json:"format,omitempty"`              // wav, mp3, flac, aac, opus
	Volume             float64 `json:"volume,omitempty"`              // 0.0-2.0
	Rate               float64 `json:"rate,omitempty"`                // 0.5-2.0
	Pitch              float64 `json:"pitch,omitempty"`               // -1.0 to 1.0
	SSML               bool    `json:"ssml,omitempty"`                // enable SSML processing
	EmotionalIntensity float64 `json:"emotional_intensity,omitempty"` // 0.0-2.0
	TempoDynamics      float64 `json:"tempo_dynamics,omitempty"`      // 0.0-2.0
	LeadingSilence     float64 `json:"leading_silence,omitempty"`     // seconds of silence before audio
	TrailingSilence    float64 `json:"trailing_silence,omitempty"`    // seconds of silence after audio
	Channels           string  `json:"channels,omitempty"`            // mono, stereo
	PlaybackMode       string  `json:"playback_mode,omitempty"`       // immediate, queue, no_queue
	WaitForEnd         bool    `json:"wait_for_end,omitempty"`        // wait until playback completes
}

// PlayTextParams parameters for play_text tool (simplified version)
type PlayTextParams struct {
	Text         string `json:"text"`
	PlaybackMode string `json:"playback_mode,omitempty"` // immediate, queue, no_queue
	WaitForEnd   bool   `json:"wait_for_end,omitempty"`  // wait until playback completes
}

// RegisterTTSTools registers all TTS-related MCP tools
func RegisterTTSTools(server *mcp.Server) {
	// Check if simplified mode is enabled (default settings are configured)
	defaultModelUUID := viper.GetString("default_model_uuid")
	useSimplifiedTools := viper.GetBool("use_simplified_tts_tools") && defaultModelUUID != ""

	if useSimplifiedTools {
		// Register simplified text-only tool
		mcp.AddTool(server, &mcp.Tool{
			Name:        "play_text",
			Description: "Play text as speech using default configuration (only text parameter required)",
		}, handlePlayText)
	} else {
		// Register full-featured tool
		mcp.AddTool(server, &mcp.Tool{
			Name:        "synthesize_speech",
			Description: "Convert text to speech and play it locally on the server",
		}, handleSynthesizeSpeech)
	}
}

func handleSynthesizeSpeech(ctx context.Context, req *mcp.CallToolRequest, args SynthesizeSpeechParams) (*mcp.CallToolResult, any, error) {

	if args.Text == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Text is required"}},
			IsError: true,
		}, nil, nil
	}

	// Use default model UUID from config if not provided
	modelUUID := args.ModelUUID
	if modelUUID == "" {
		modelUUID = viper.GetString("default_model_uuid")
		if modelUUID == "" {
			// Use hardcoded default model UUID if not configured
			modelUUID = defaultModelUUID
		}
	}

	// Build TTS request
	request := aivisClient.NewTTSRequest(modelUUID, args.Text)

	// Apply SSML setting
	if args.SSML {
		request = request.WithSSML(true)
	}

	// Apply optional parameters with config defaults
	volume := args.Volume
	if volume == 0 {
		volume = viper.GetFloat64("default_volume")
	}
	if volume > 0 {
		request = request.WithVolume(volume)
	}

	rate := args.Rate
	if rate == 0 {
		rate = viper.GetFloat64("default_rate")
	}
	if rate > 0 {
		request = request.WithSpeakingRate(rate)
	}

	pitch := args.Pitch
	if pitch == 0 {
		pitch = viper.GetFloat64("default_pitch")
	}
	if pitch != 0 {
		request = request.WithPitch(pitch)
	}

	// Apply advanced TTS parameters
	if args.EmotionalIntensity > 0 {
		request = request.WithEmotionalIntensity(args.EmotionalIntensity)
	}

	if args.TempoDynamics > 0 {
		request = request.WithTempoDynamics(args.TempoDynamics)
	}

	if args.LeadingSilence > 0 {
		request = request.WithLeadingSilence(args.LeadingSilence)
	}

	if args.TrailingSilence > 0 {
		request = request.WithTrailingSilence(args.TrailingSilence)
	}

	// Set audio channels
	if args.Channels != "" {
		switch args.Channels {
		case "mono":
			request = request.WithOutputChannels(ttsDomain.AudioChannelsMono)
		case "stereo":
			request = request.WithOutputChannels(ttsDomain.AudioChannelsStereo)
		}
	}

	// Set output format with config default
	format := args.Format
	if format == "" {
		format = viper.GetString("default_format")
	}
	if format != "" {
		switch format {
		case "wav":
			request = request.WithOutputFormat(ttsDomain.OutputFormatWAV)
		case "mp3":
			request = request.WithOutputFormat(ttsDomain.OutputFormatMP3)
		case "flac":
			request = request.WithOutputFormat(ttsDomain.OutputFormatFLAC)
		case "aac":
			request = request.WithOutputFormat(ttsDomain.OutputFormatAAC)
		case "opus":
			request = request.WithOutputFormat(ttsDomain.OutputFormatOpus)
		}
	}

	// Create playback request with options
	playbackReq := aivisClient.NewPlaybackRequest(request.Build())
	
	// Set playback mode (default to immediate for MCP)
	playbackMode := args.PlaybackMode
	if playbackMode == "" {
		playbackMode = viper.GetString("default_playback_mode")
		if playbackMode == "" {
			playbackMode = "immediate" // MCP default
		}
	}
	switch playbackMode {
	case "immediate":
		playbackReq = playbackReq.WithMode(ttsDomain.PlaybackModeImmediate)
	case "queue":
		playbackReq = playbackReq.WithMode(ttsDomain.PlaybackModeQueue)
	case "no_queue":
		playbackReq = playbackReq.WithMode(ttsDomain.PlaybackModeNoQueue)
	default:
		playbackReq = playbackReq.WithMode(ttsDomain.PlaybackModeImmediate) // fallback
	}
	
	// Set wait for end flag
	waitForEnd := args.WaitForEnd
	if !waitForEnd {
		waitForEnd = viper.GetBool("default_wait_for_end")
	}
	playbackReq = playbackReq.WithWaitForEnd(waitForEnd)
	
	// Generate filename and save to history directory (absolute path)
	timestamp := time.Now().Format("20060102_150405")
	if format == "" {
		format = "wav"
	}
	
	// Get history directory from config or use default
	historyDir := viper.GetString("history_store_path")
	if historyDir == "" {
		homeDir, _ := os.UserHomeDir()
		historyDir = filepath.Join(homeDir, ".aivis-cli", "history", "audio")
	} else {
		historyDir = filepath.Join(historyDir, "audio")
	}
	
	// Ensure directory exists
	os.MkdirAll(historyDir, 0755)
	
	// Create absolute path for the audio file
	tempFile := filepath.Join(historyDir, fmt.Sprintf("mcp_%s.%s", timestamp, format))
	
	// Use streaming synthesis with history and playback
	response, err := aivisClient.PlayStreamWithHistory(ctx, playbackReq.Build(), tempFile)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Streaming synthesis and playback failed: %v", err)}},
			IsError: true,
		}, nil, nil
	}
	
	// Wait for playback to complete if requested  
	if waitForEnd {
		for {
			status := aivisClient.GetPlaybackStatus()
			if status.Status == ttsDomain.PlaybackStatusIdle || 
			   status.Status == ttsDomain.PlaybackStatusStopped {
				break
			}
			// Short sleep to avoid busy waiting
			select {
			case <-ctx.Done():
				return &mcp.CallToolResult{
					Content: []mcp.Content{&mcp.TextContent{Text: "Context cancelled while waiting for playback completion"}},
					IsError: true,
				}, nil, nil
			case <-time.After(100 * time.Millisecond):
				continue
			}
		}
	}

	// Format response with playback info including history ID
	resultText := "Audio synthesized and played successfully\n"
	if response.HistoryID > 0 {
		resultText += fmt.Sprintf("History ID: %d\n", response.HistoryID)
	} else {
		resultText += "History ID: (not saved - check history configuration)\n"
	}
	resultText += fmt.Sprintf("Text: %s\n", args.Text)
	resultText += fmt.Sprintf("Model: %s\n", modelUUID)
	if args.SSML {
		resultText += "SSML: enabled\n"
	}
	if format != "" {
		resultText += fmt.Sprintf("Format: %s\n", format)
	}
	if args.Channels != "" {
		resultText += fmt.Sprintf("Channels: %s\n", args.Channels)
	}
	if volume > 0 {
		resultText += fmt.Sprintf("Volume: %.2f\n", volume)
	}
	if rate > 0 {
		resultText += fmt.Sprintf("Speaking Rate: %.2f\n", rate)
	}
	if pitch != 0 {
		resultText += fmt.Sprintf("Pitch: %.2f\n", pitch)
	}
	if args.EmotionalIntensity > 0 {
		resultText += fmt.Sprintf("Emotional Intensity: %.2f\n", args.EmotionalIntensity)
	}
	if args.TempoDynamics > 0 {
		resultText += fmt.Sprintf("Tempo Dynamics: %.2f\n", args.TempoDynamics)
	}
	if args.LeadingSilence > 0 {
		resultText += fmt.Sprintf("Leading Silence: %.2fs\n", args.LeadingSilence)
	}
	if args.TrailingSilence > 0 {
		resultText += fmt.Sprintf("Trailing Silence: %.2fs\n", args.TrailingSilence)
	}
	if playbackMode != "" {
		resultText += fmt.Sprintf("Playback Mode: %s\n", playbackMode)
	}
	if waitForEnd {
		resultText += "Waited for playback completion\n"
		resultText += "Audio playback completed"
	} else {
		resultText += "Audio is now playing on the server"
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: resultText}},
	}, nil, nil
}

func handlePlayText(ctx context.Context, req *mcp.CallToolRequest, args PlayTextParams) (*mcp.CallToolResult, any, error) {

	if args.Text == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Text is required"}},
			IsError: true,
		}, nil, nil
	}

	// Use default model UUID from config
	modelUUID := viper.GetString("default_model_uuid")
	if modelUUID == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Default model UUID is not configured"}},
			IsError: true,
		}, nil, nil
	}

	// Create TTS request with defaults
	request := aivisClient.NewTTSRequest(modelUUID, args.Text)
	
	// Apply default settings from config
	if volume := viper.GetFloat64("default_volume"); volume > 0 {
		request = request.WithVolume(volume)
	}
	if rate := viper.GetFloat64("default_rate"); rate > 0 {
		request = request.WithSpeakingRate(rate)
	}
	if pitch := viper.GetFloat64("default_pitch"); pitch != 0 {
		request = request.WithPitch(pitch)
	}
	if format := viper.GetString("default_format"); format != "" {
		switch format {
		case "wav":
			request = request.WithOutputFormat(ttsDomain.OutputFormatWAV)
		case "mp3":
			request = request.WithOutputFormat(ttsDomain.OutputFormatMP3)
		case "flac":
			request = request.WithOutputFormat(ttsDomain.OutputFormatFLAC)
		case "aac":
			request = request.WithOutputFormat(ttsDomain.OutputFormatAAC)
		case "opus":
			request = request.WithOutputFormat(ttsDomain.OutputFormatOpus)
		}
	}
	
	// Create playback request with options
	playbackReq := aivisClient.NewPlaybackRequest(request.Build())
	
	// Set playback mode (default to immediate for MCP)
	playbackMode := args.PlaybackMode
	if playbackMode == "" {
		playbackMode = viper.GetString("default_playback_mode")
		if playbackMode == "" {
			playbackMode = "immediate" // MCP default
		}
	}
	switch playbackMode {
	case "immediate":
		playbackReq = playbackReq.WithMode(ttsDomain.PlaybackModeImmediate)
	case "queue":
		playbackReq = playbackReq.WithMode(ttsDomain.PlaybackModeQueue)
	case "no_queue":
		playbackReq = playbackReq.WithMode(ttsDomain.PlaybackModeNoQueue)
	default:
		playbackReq = playbackReq.WithMode(ttsDomain.PlaybackModeImmediate) // fallback
	}
	
	// Set wait for end flag
	waitForEnd := args.WaitForEnd
	if !waitForEnd {
		waitForEnd = viper.GetBool("default_wait_for_end")
	}
	playbackReq = playbackReq.WithWaitForEnd(waitForEnd)
	
	// Get format from config for file naming
	format := viper.GetString("default_format")
	if format == "" {
		format = "wav"
	}
	
	// Generate filename and save to history directory (absolute path)
	timestamp := time.Now().Format("20060102_150405")
	
	// Get history directory from config or use default
	historyDir := viper.GetString("history_store_path")
	if historyDir == "" {
		homeDir, _ := os.UserHomeDir()
		historyDir = filepath.Join(homeDir, ".aivis-cli", "history", "audio")
	} else {
		historyDir = filepath.Join(historyDir, "audio")
	}
	
	// Ensure directory exists
	os.MkdirAll(historyDir, 0755)
	
	// Create absolute path for the audio file
	tempFile := filepath.Join(historyDir, fmt.Sprintf("mcp_%s.%s", timestamp, format))
	
	// Use streaming synthesis with history and playback
	response, err := aivisClient.PlayStreamWithHistory(ctx, playbackReq.Build(), tempFile)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Streaming synthesis and playback failed: %v", err)}},
			IsError: true,
		}, nil, nil
	}
	
	// Wait for playback to complete if requested  
	if waitForEnd {
		for {
			status := aivisClient.GetPlaybackStatus()
			if status.Status == ttsDomain.PlaybackStatusIdle || 
			   status.Status == ttsDomain.PlaybackStatusStopped {
				break
			}
			// Short sleep to avoid busy waiting
			select {
			case <-ctx.Done():
				return &mcp.CallToolResult{
					Content: []mcp.Content{&mcp.TextContent{Text: "Context cancelled while waiting for playback completion"}},
					IsError: true,
				}, nil, nil
			case <-time.After(100 * time.Millisecond):
				continue
			}
		}
	}

	// Format response including history ID
	var resultText string
	if response.HistoryID > 0 {
		resultText = fmt.Sprintf("History ID: %d\n", response.HistoryID)
	} else {
		resultText = "History ID: (not saved - check history configuration)\n"
	}
	resultText += fmt.Sprintf("Text: %s\n", args.Text)
	if playbackMode != "" {
		resultText += fmt.Sprintf("Playback Mode: %s\n", playbackMode)
	}
	if args.WaitForEnd {
		resultText += "Audio playback completed"
	} else {
		resultText += "Audio is now playing on the server"
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: resultText}},
	}, nil, nil
}

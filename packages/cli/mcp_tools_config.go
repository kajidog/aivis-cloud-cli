package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/viper"
)

// GetMCPSettingsParams parameters for get_mcp_settings tool
type GetMCPSettingsParams struct {
	// No parameters needed - returns all safe settings
}

// UpdateMCPSettingsParams parameters for update_mcp_settings tool
type UpdateMCPSettingsParams struct {
	BaseURL                     string  `json:"base_url,omitempty"`
	DefaultModelUUID            string  `json:"default_model_uuid,omitempty"`
	DefaultPlaybackMode         string  `json:"default_playback_mode,omitempty"`
	DefaultVolume               float64 `json:"default_volume,omitempty"`
	DefaultRate                 float64 `json:"default_rate,omitempty"`
	DefaultPitch                float64 `json:"default_pitch,omitempty"`
	DefaultFormat               string  `json:"default_format,omitempty"`
	DefaultSSML                 bool    `json:"default_ssml,omitempty"`
	DefaultEmotionalIntensity   float64 `json:"default_emotional_intensity,omitempty"`
	DefaultTempoDynamics        float64 `json:"default_tempo_dynamics,omitempty"`
	DefaultLeadingSilence       float64 `json:"default_leading_silence,omitempty"`
	DefaultTrailingSilence      float64 `json:"default_trailing_silence,omitempty"`
	DefaultChannels             string  `json:"default_channels,omitempty"`
}

// MCPSettings represents the safe configuration settings
type MCPSettings struct {
	BaseURL                   string  `json:"base_url"`
	DefaultModelUUID          string  `json:"default_model_uuid"`
	DefaultPlaybackMode       string  `json:"default_playback_mode"`
	DefaultVolume             float64 `json:"default_volume"`
	DefaultRate               float64 `json:"default_rate"`
	DefaultPitch              float64 `json:"default_pitch"`
	DefaultFormat             string  `json:"default_format"`
	DefaultSSML               bool    `json:"default_ssml"`
	DefaultEmotionalIntensity float64 `json:"default_emotional_intensity"`
	DefaultTempoDynamics      float64 `json:"default_tempo_dynamics"`
	DefaultLeadingSilence     float64 `json:"default_leading_silence"`
	DefaultTrailingSilence    float64 `json:"default_trailing_silence"`
	DefaultChannels           string  `json:"default_channels"`
}

// RegisterConfigTools registers all configuration-related MCP tools
func RegisterConfigTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_mcp_settings",
		Description: "Get current MCP configuration settings (API key excluded for security)",
	}, handleGetMCPSettings)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_mcp_settings",
		Description: "Update MCP configuration settings (API key and system settings cannot be modified)",
	}, handleUpdateMCPSettings)
}

func handleGetMCPSettings(ctx context.Context, req *mcp.CallToolRequest, args GetMCPSettingsParams) (*mcp.CallToolResult, any, error) {
	// Get current settings from viper (excluding sensitive information)
	settings := MCPSettings{
		BaseURL:                   viper.GetString("base_url"),
		DefaultModelUUID:          viper.GetString("default_model_uuid"),
		DefaultPlaybackMode:       viper.GetString("default_playback_mode"),
		DefaultVolume:             viper.GetFloat64("default_volume"),
		DefaultRate:               viper.GetFloat64("default_rate"),
		DefaultPitch:              viper.GetFloat64("default_pitch"),
		DefaultFormat:             viper.GetString("default_format"),
		DefaultSSML:               viper.GetBool("default_ssml"),
		DefaultEmotionalIntensity: viper.GetFloat64("default_emotional_intensity"),
		DefaultTempoDynamics:      viper.GetFloat64("default_tempo_dynamics"),
		DefaultLeadingSilence:     viper.GetFloat64("default_leading_silence"),
		DefaultTrailingSilence:    viper.GetFloat64("default_trailing_silence"),
		DefaultChannels:           viper.GetString("default_channels"),
	}

	// Format response as readable text
	var result strings.Builder
	result.WriteString("Current MCP Settings:\n\n")
	result.WriteString(fmt.Sprintf("Base URL: %s\n", settings.BaseURL))
	result.WriteString(fmt.Sprintf("Default Model UUID: %s\n", settings.DefaultModelUUID))
	result.WriteString(fmt.Sprintf("Default Playback Mode: %s\n", settings.DefaultPlaybackMode))
	result.WriteString(fmt.Sprintf("Default Volume: %.2f\n", settings.DefaultVolume))
	result.WriteString(fmt.Sprintf("Default Rate: %.2f\n", settings.DefaultRate))
	result.WriteString(fmt.Sprintf("Default Pitch: %.2f\n", settings.DefaultPitch))
	result.WriteString(fmt.Sprintf("Default Format: %s\n", settings.DefaultFormat))
	result.WriteString(fmt.Sprintf("Default SSML: %t\n", settings.DefaultSSML))
	result.WriteString(fmt.Sprintf("Default Emotional Intensity: %.2f\n", settings.DefaultEmotionalIntensity))
	result.WriteString(fmt.Sprintf("Default Tempo Dynamics: %.2f\n", settings.DefaultTempoDynamics))
	result.WriteString(fmt.Sprintf("Default Leading Silence: %.2fs\n", settings.DefaultLeadingSilence))
	result.WriteString(fmt.Sprintf("Default Trailing Silence: %.2fs\n", settings.DefaultTrailingSilence))
	result.WriteString(fmt.Sprintf("Default Channels: %s\n", settings.DefaultChannels))
	result.WriteString("\nNote: API key and system settings are not displayed for security reasons.")

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: result.String()}},
	}, nil, nil
}

func handleUpdateMCPSettings(ctx context.Context, req *mcp.CallToolRequest, args UpdateMCPSettingsParams) (*mcp.CallToolResult, any, error) {
	var updates []string
	var errors []string

	// Validate and update base_url
	if args.BaseURL != "" {
		if !strings.HasPrefix(args.BaseURL, "http://") && !strings.HasPrefix(args.BaseURL, "https://") {
			errors = append(errors, "base_url must start with http:// or https://")
		} else {
			viper.Set("base_url", args.BaseURL)
			updates = append(updates, fmt.Sprintf("Base URL: %s", args.BaseURL))
		}
	}

	// Update default_model_uuid
	if args.DefaultModelUUID != "" {
		viper.Set("default_model_uuid", args.DefaultModelUUID)
		updates = append(updates, fmt.Sprintf("Default Model UUID: %s", args.DefaultModelUUID))
	}

	// Validate and update default_playback_mode
	if args.DefaultPlaybackMode != "" {
		validModes := []string{"immediate", "queue", "no_queue"}
		isValid := false
		for _, mode := range validModes {
			if args.DefaultPlaybackMode == mode {
				isValid = true
				break
			}
		}
		if !isValid {
			errors = append(errors, "default_playback_mode must be one of: immediate, queue, no_queue")
		} else {
			viper.Set("default_playback_mode", args.DefaultPlaybackMode)
			updates = append(updates, fmt.Sprintf("Default Playback Mode: %s", args.DefaultPlaybackMode))
		}
	}

	// Validate and update default_volume
	if args.DefaultVolume > 0 {
		if args.DefaultVolume < 0.0 || args.DefaultVolume > 2.0 {
			errors = append(errors, "default_volume must be between 0.0 and 2.0")
		} else {
			viper.Set("default_volume", args.DefaultVolume)
			updates = append(updates, fmt.Sprintf("Default Volume: %.2f", args.DefaultVolume))
		}
	}

	// Validate and update default_rate
	if args.DefaultRate > 0 {
		if args.DefaultRate < 0.5 || args.DefaultRate > 2.0 {
			errors = append(errors, "default_rate must be between 0.5 and 2.0")
		} else {
			viper.Set("default_rate", args.DefaultRate)
			updates = append(updates, fmt.Sprintf("Default Rate: %.2f", args.DefaultRate))
		}
	}

	// Validate and update default_pitch
	if args.DefaultPitch != 0 {
		if args.DefaultPitch < -1.0 || args.DefaultPitch > 1.0 {
			errors = append(errors, "default_pitch must be between -1.0 and 1.0")
		} else {
			viper.Set("default_pitch", args.DefaultPitch)
			updates = append(updates, fmt.Sprintf("Default Pitch: %.2f", args.DefaultPitch))
		}
	}

	// Validate and update default_format
	if args.DefaultFormat != "" {
		validFormats := []string{"wav", "mp3", "flac", "aac", "opus"}
		isValid := false
		for _, format := range validFormats {
			if args.DefaultFormat == format {
				isValid = true
				break
			}
		}
		if !isValid {
			errors = append(errors, "default_format must be one of: wav, mp3, flac, aac, opus")
		} else {
			viper.Set("default_format", args.DefaultFormat)
			updates = append(updates, fmt.Sprintf("Default Format: %s", args.DefaultFormat))
		}
	}

	// Update default_ssml
	if args.DefaultSSML {
		viper.Set("default_ssml", args.DefaultSSML)
		updates = append(updates, fmt.Sprintf("Default SSML: %t", args.DefaultSSML))
	}

	// Validate and update default_emotional_intensity
	if args.DefaultEmotionalIntensity > 0 {
		if args.DefaultEmotionalIntensity < 0.0 || args.DefaultEmotionalIntensity > 2.0 {
			errors = append(errors, "default_emotional_intensity must be between 0.0 and 2.0")
		} else {
			viper.Set("default_emotional_intensity", args.DefaultEmotionalIntensity)
			updates = append(updates, fmt.Sprintf("Default Emotional Intensity: %.2f", args.DefaultEmotionalIntensity))
		}
	}

	// Validate and update default_tempo_dynamics
	if args.DefaultTempoDynamics > 0 {
		if args.DefaultTempoDynamics < 0.0 || args.DefaultTempoDynamics > 2.0 {
			errors = append(errors, "default_tempo_dynamics must be between 0.0 and 2.0")
		} else {
			viper.Set("default_tempo_dynamics", args.DefaultTempoDynamics)
			updates = append(updates, fmt.Sprintf("Default Tempo Dynamics: %.2f", args.DefaultTempoDynamics))
		}
	}

	// Validate and update default_leading_silence
	if args.DefaultLeadingSilence > 0 {
		if args.DefaultLeadingSilence < 0.0 || args.DefaultLeadingSilence > 10.0 {
			errors = append(errors, "default_leading_silence must be between 0.0 and 10.0 seconds")
		} else {
			viper.Set("default_leading_silence", args.DefaultLeadingSilence)
			updates = append(updates, fmt.Sprintf("Default Leading Silence: %.2fs", args.DefaultLeadingSilence))
		}
	}

	// Validate and update default_trailing_silence
	if args.DefaultTrailingSilence > 0 {
		if args.DefaultTrailingSilence < 0.0 || args.DefaultTrailingSilence > 10.0 {
			errors = append(errors, "default_trailing_silence must be between 0.0 and 10.0 seconds")
		} else {
			viper.Set("default_trailing_silence", args.DefaultTrailingSilence)
			updates = append(updates, fmt.Sprintf("Default Trailing Silence: %.2fs", args.DefaultTrailingSilence))
		}
	}

	// Validate and update default_channels
	if args.DefaultChannels != "" {
		validChannels := []string{"mono", "stereo"}
		isValid := false
		for _, channel := range validChannels {
			if args.DefaultChannels == channel {
				isValid = true
				break
			}
		}
		if !isValid {
			errors = append(errors, "default_channels must be one of: mono, stereo")
		} else {
			viper.Set("default_channels", args.DefaultChannels)
			updates = append(updates, fmt.Sprintf("Default Channels: %s", args.DefaultChannels))
		}
	}

	// If there are errors, return them
	if len(errors) > 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Validation errors:\n%s", strings.Join(errors, "\n"))}},
			IsError: true,
		}, nil, nil
	}

	// Save configuration to file
	if len(updates) > 0 {
		if err := viper.WriteConfig(); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to save configuration: %v", err)}},
				IsError: true,
			}, nil, nil
		}

		resultText := fmt.Sprintf("Successfully updated %d setting(s):\n%s\n\nConfiguration saved to file.", 
			len(updates), strings.Join(updates, "\n"))
		
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: resultText}},
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: "No settings were updated (no valid parameters provided)"}},
	}, nil, nil
}
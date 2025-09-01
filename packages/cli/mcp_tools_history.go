package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	ttsDomain "github.com/kajidog/aivis-cloud-cli/client/tts/domain"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListTTSHistoryParams parameters for list_tts_history tool
type ListTTSHistoryParams struct {
	Limit        int    `json:"limit,omitempty"`         // Maximum number of records (default: 10, max: 100)
	Offset       int    `json:"offset,omitempty"`        // Number of records to skip (default: 0)
	ModelUUID    string `json:"model_uuid,omitempty"`    // Filter by model UUID
	TextContains string `json:"text_contains,omitempty"` // Filter by text content
	SortBy       string `json:"sort_by,omitempty"`       // Sort field: id, created_at, text (default: id)
	SortOrder    string `json:"sort_order,omitempty"`    // Sort order: asc, desc (default: desc)
}

// GetTTSHistoryParams parameters for get_tts_history tool
type GetTTSHistoryParams struct {
	ID int `json:"id"` // History record ID (required)
}

// PlayTTSHistoryParams parameters for play_tts_history tool  
type PlayTTSHistoryParams struct {
	ID           int     `json:"id"`                      // History record ID (required)
	Volume       float64 `json:"volume,omitempty"`        // Playback volume (0.0-1.0)
	PlaybackMode string  `json:"playback_mode,omitempty"` // immediate, queue, no_queue
	WaitForEnd   bool    `json:"wait_for_end,omitempty"`  // Wait for playback completion
}

// DeleteTTSHistoryParams parameters for delete_tts_history tool
type DeleteTTSHistoryParams struct {
	ID int `json:"id"` // History record ID (required)
}

// GetTTSHistoryStatsParams parameters for get_tts_history_stats tool
type GetTTSHistoryStatsParams struct {
	// No parameters needed - returns overall statistics
}

// RegisterHistoryTools registers all TTS history-related MCP tools
func RegisterHistoryTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_tts_history",
		Description: "List TTS synthesis history records with pagination and filtering options",
	}, handleListTTSHistory)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_tts_history",
		Description: "Get detailed information about a specific TTS history record",
	}, handleGetTTSHistory)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "play_tts_history",
		Description: "Resume/replay audio from TTS history record (main resume functionality)",
	}, handlePlayTTSHistory)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_tts_history",
		Description: "Delete a specific TTS history record and its associated audio file",
	}, handleDeleteTTSHistory)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_tts_history_stats",
		Description: "Get statistics about TTS history including record count and storage usage",
	}, handleGetTTSHistoryStats)
}

func handleListTTSHistory(ctx context.Context, req *mcp.CallToolRequest, args ListTTSHistoryParams) (*mcp.CallToolResult, any, error) {
	// Validate and set defaults
	if args.Limit <= 0 {
		args.Limit = 10
	}
	if args.Limit > 100 {
		args.Limit = 100 // Prevent excessive results
	}
	if args.Offset < 0 {
		args.Offset = 0
	}
	if args.SortBy == "" {
		args.SortBy = "id"
	}
	if args.SortOrder == "" {
		args.SortOrder = "desc"
	}

	// Validate sort parameters
	validSortFields := []string{"id", "created_at", "text"}
	validSortField := false
	for _, field := range validSortFields {
		if args.SortBy == field {
			validSortField = true
			break
		}
	}
	if !validSortField {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Invalid sort_by field. Must be one of: id, created_at, text"}},
			IsError: true,
		}, nil, nil
	}

	if args.SortOrder != "asc" && args.SortOrder != "desc" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Invalid sort_order. Must be 'asc' or 'desc'"}},
			IsError: true,
		}, nil, nil
	}

	// Build search request
	searchBuilder := aivisClient.NewTTSHistorySearchRequest().
		WithLimit(args.Limit).
		WithOffset(args.Offset).
		WithSorting(args.SortBy, args.SortOrder)

	if args.ModelUUID != "" {
		searchBuilder = searchBuilder.WithModelUUID(args.ModelUUID)
	}
	if args.TextContains != "" {
		searchBuilder = searchBuilder.WithTextContains(args.TextContains)
	}

	searchRequest := searchBuilder.Build()

	// Get history list
	response, err := aivisClient.ListTTSHistory(ctx, searchRequest)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to list TTS history: %v", err)}},
			IsError: true,
		}, nil, nil
	}

	if len(response.Histories) == 0 {
		resultText := fmt.Sprintf("No TTS history records found (total: %d)", response.Total)
		if args.Offset > 0 {
			resultText += fmt.Sprintf(" with offset %d", args.Offset)
		}
		if args.ModelUUID != "" || args.TextContains != "" {
			resultText += " matching the specified filters"
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: resultText}},
		}, nil, nil
	}

	// Format results in compact, token-optimized format
	var result strings.Builder
	result.WriteString(fmt.Sprintf("TTS History (%d-%d of %d total)\n\n",
		args.Offset+1, args.Offset+len(response.Histories), response.Total))

	for _, history := range response.Histories {
		text := history.Text
		if len(text) > 50 {
			text = text[:47] + "..."
		}

		model := history.ModelUUID
		if len(model) > 12 {
			model = model[:12] + "..."
		}

		// Format file size compactly
		size := formatFileSize(history.FileSizeBytes)

		// Format creation time compactly
		created := history.CreatedAt.Format("01/02 15:04")

		result.WriteString(fmt.Sprintf("ID %d: %s\n", history.ID, text))
		result.WriteString(fmt.Sprintf("  Model: %s | Format: %s | Size: %s | Created: %s\n\n",
			model, history.FileFormat, size, created))
	}

	// Add pagination info
	if response.HasMore {
		result.WriteString(fmt.Sprintf("Use offset=%d to see more records", args.Offset+args.Limit))
	} else if args.Offset > 0 {
		result.WriteString("Showing last page of results")
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: result.String()}},
	}, nil, nil
}

func handleGetTTSHistory(ctx context.Context, req *mcp.CallToolRequest, args GetTTSHistoryParams) (*mcp.CallToolResult, any, error) {
	if args.ID <= 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "History ID is required and must be positive"}},
			IsError: true,
		}, nil, nil
	}

	// Get history record
	history, err := aivisClient.GetTTSHistory(ctx, args.ID)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to get TTS history record #%d: %v", args.ID, err)}},
			IsError: true,
		}, nil, nil
	}

	// Format detailed information in compact format
	var result strings.Builder
	result.WriteString(fmt.Sprintf("TTS History Record #%d\n\n", history.ID))

	result.WriteString(fmt.Sprintf("Text: %s\n", history.Text))
	result.WriteString(fmt.Sprintf("Model UUID: %s\n", history.ModelUUID))
	result.WriteString(fmt.Sprintf("Created: %s\n", history.CreatedAt.Format("2006-01-02 15:04:05")))
	result.WriteString(fmt.Sprintf("File: %s (%s, %s)\n", 
		history.FileFormat, formatFileSize(history.FileSizeBytes), history.FilePath))

	if history.Credits != nil {
		result.WriteString(fmt.Sprintf("Credits: %.4f\n", *history.Credits))
	}

	// Show TTS request details if available
	if history.Request != nil {
		result.WriteString("\nRequest Details:\n")
		
		details := make([]string, 0)
		if history.Request.SpeakerUUID != nil {
			details = append(details, fmt.Sprintf("Speaker: %s", *history.Request.SpeakerUUID))
		}
		if history.Request.StyleID != nil {
			details = append(details, fmt.Sprintf("Style ID: %d", *history.Request.StyleID))
		}
		if history.Request.StyleName != nil {
			details = append(details, fmt.Sprintf("Style: %s", *history.Request.StyleName))
		}
		if history.Request.UseSSML != nil && *history.Request.UseSSML {
			details = append(details, "SSML: enabled")
		}
		if history.Request.SpeakingRate != nil {
			details = append(details, fmt.Sprintf("Rate: %.2f", *history.Request.SpeakingRate))
		}
		if history.Request.Pitch != nil {
			details = append(details, fmt.Sprintf("Pitch: %.2f", *history.Request.Pitch))
		}
		if history.Request.Volume != nil {
			details = append(details, fmt.Sprintf("Volume: %.2f", *history.Request.Volume))
		}
		if history.Request.OutputAudioChannels != nil {
			details = append(details, fmt.Sprintf("Channels: %s", string(*history.Request.OutputAudioChannels)))
		}
		if history.Request.LeadingSilenceSeconds != nil {
			details = append(details, fmt.Sprintf("Leading silence: %.2fs", *history.Request.LeadingSilenceSeconds))
		}
		if history.Request.TrailingSilenceSeconds != nil {
			details = append(details, fmt.Sprintf("Trailing silence: %.2fs", *history.Request.TrailingSilenceSeconds))
		}

		result.WriteString(strings.Join(details, " | "))
	}

	result.WriteString(fmt.Sprintf("\n\nUse play_tts_history with ID %d to replay this audio", history.ID))

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: result.String()}},
	}, nil, nil
}

func handlePlayTTSHistory(ctx context.Context, req *mcp.CallToolRequest, args PlayTTSHistoryParams) (*mcp.CallToolResult, any, error) {
	if args.ID <= 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "History ID is required and must be positive"}},
			IsError: true,
		}, nil, nil
	}

	// Validate volume if provided
	if args.Volume > 0 && (args.Volume < 0.0 || args.Volume > 1.0) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Volume must be between 0.0 and 1.0"}},
			IsError: true,
		}, nil, nil
	}

	// Validate playback mode if provided
	if args.PlaybackMode != "" {
		validModes := []string{"immediate", "queue", "no_queue"}
		validMode := false
		for _, mode := range validModes {
			if args.PlaybackMode == mode {
				validMode = true
				break
			}
		}
		if !validMode {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "Invalid playback_mode. Must be one of: immediate, queue, no_queue"}},
				IsError: true,
			}, nil, nil
		}
	}

	// Create playback request with default values
	playbackBuilder := aivisClient.NewPlaybackRequest(nil). // Request will be retrieved from history
		WithWaitForEnd(args.WaitForEnd)

	if args.Volume > 0 {
		playbackBuilder = playbackBuilder.WithVolume(args.Volume)
	}

	// Set playback mode with sensible default for MCP
	playbackMode := args.PlaybackMode
	if playbackMode == "" {
		playbackMode = "immediate" // Default to immediate for MCP
	}
	
	switch playbackMode {
	case "immediate":
		playbackBuilder = playbackBuilder.WithMode(ttsDomain.PlaybackModeImmediate)
	case "queue":
		playbackBuilder = playbackBuilder.WithMode(ttsDomain.PlaybackModeQueue)
	case "no_queue":
		playbackBuilder = playbackBuilder.WithMode(ttsDomain.PlaybackModeNoQueue)
	}

	playbackOptions := playbackBuilder.Build()

	// Get the history record first to show what we're playing
	history, err := aivisClient.GetTTSHistory(ctx, args.ID)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to get TTS history record #%d: %v", args.ID, err)}},
			IsError: true,
		}, nil, nil
	}

	// Play history
	if err := aivisClient.PlayTTSHistory(ctx, args.ID, playbackOptions); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to play TTS history #%d: %v", args.ID, err)}},
			IsError: true,
		}, nil, nil
	}

	// Format response with playback info
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Playing TTS history record #%d\n\n", args.ID))
	result.WriteString(fmt.Sprintf("Text: %s\n", history.Text))
	result.WriteString(fmt.Sprintf("Model: %s\n", history.ModelUUID))
	result.WriteString(fmt.Sprintf("Format: %s (%s)\n", history.FileFormat, formatFileSize(history.FileSizeBytes)))
	
	if args.Volume > 0 {
		result.WriteString(fmt.Sprintf("Volume: %.2f\n", args.Volume))
	}
	result.WriteString(fmt.Sprintf("Playback Mode: %s\n", playbackMode))
	
	if args.WaitForEnd {
		result.WriteString("\nWaiting for playback completion...")
		// Wait for playback to complete
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
		result.WriteString("\nAudio playback completed")
	} else {
		result.WriteString("\nAudio is now playing on the server")
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: result.String()}},
	}, nil, nil
}

func handleDeleteTTSHistory(ctx context.Context, req *mcp.CallToolRequest, args DeleteTTSHistoryParams) (*mcp.CallToolResult, any, error) {
	if args.ID <= 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "History ID is required and must be positive"}},
			IsError: true,
		}, nil, nil
	}

	// Get the record info before deleting (for confirmation message)
	history, err := aivisClient.GetTTSHistory(ctx, args.ID)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to get TTS history record #%d: %v", args.ID, err)}},
			IsError: true,
		}, nil, nil
	}

	// Delete history
	if err := aivisClient.DeleteTTSHistory(ctx, args.ID); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to delete TTS history #%d: %v", args.ID, err)}},
			IsError: true,
		}, nil, nil
	}

	// Format confirmation message
	text := history.Text
	if len(text) > 50 {
		text = text[:47] + "..."
	}

	resultText := fmt.Sprintf("Successfully deleted TTS history record #%d\n\n", args.ID)
	resultText += fmt.Sprintf("Deleted record contained: %s\n", text)
	resultText += fmt.Sprintf("File size freed: %s", formatFileSize(history.FileSizeBytes))

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: resultText}},
	}, nil, nil
}

func handleGetTTSHistoryStats(ctx context.Context, req *mcp.CallToolRequest, args GetTTSHistoryStatsParams) (*mcp.CallToolResult, any, error) {
	// Get history statistics
	stats, err := aivisClient.GetTTSHistoryStats(ctx)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to get TTS history statistics: %v", err)}},
			IsError: true,
		}, nil, nil
	}

	// Format statistics in compact format
	var result strings.Builder
	result.WriteString("TTS History Statistics\n\n")
	result.WriteString(fmt.Sprintf("Total Records: %d\n", stats.TotalRecords))
	result.WriteString(fmt.Sprintf("Total Storage: %s\n", formatFileSize(stats.TotalFileSize)))
	result.WriteString(fmt.Sprintf("Total Credits: %.4f\n", stats.TotalCredits))

	if stats.TotalRecords > 0 {
		avgFileSize := stats.TotalFileSize / int64(stats.TotalRecords)
		avgCredits := stats.TotalCredits / float64(stats.TotalRecords)
		result.WriteString(fmt.Sprintf("\nAverage File Size: %s\n", formatFileSize(avgFileSize)))
		result.WriteString(fmt.Sprintf("Average Credits: %.4f\n", avgCredits))
	}

	if stats.TotalRecords == 0 {
		result.WriteString("\nNo TTS history records found. Start using TTS synthesis to build your history!")
	} else {
		result.WriteString(fmt.Sprintf("\nUse list_tts_history to browse your %d records", stats.TotalRecords))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: result.String()}},
	}, nil, nil
}


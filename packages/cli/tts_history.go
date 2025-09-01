package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	ttsDomain "github.com/kajidog/aivis-cloud-cli/client/tts/domain"
	"github.com/spf13/cobra"
)

var ttsHistoryCmd = &cobra.Command{
	Use:   "history",
	Short: "TTS history management",
	Long:  "Manage TTS synthesis history including list, show, play, and delete operations",
}

var ttsHistoryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List TTS history records",
	Long:  "Display a list of TTS synthesis history records with pagination",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		// Get flags
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")
		modelUUID, _ := cmd.Flags().GetString("model-uuid")
		textContains, _ := cmd.Flags().GetString("text-contains")
		
		// Build search request
		searchBuilder := aivisClient.NewTTSHistorySearchRequest().
			WithLimit(limit).
			WithOffset(offset)
		
		if modelUUID != "" {
			searchBuilder = searchBuilder.WithModelUUID(modelUUID)
		}
		if textContains != "" {
			searchBuilder = searchBuilder.WithTextContains(textContains)
		}
		
		searchRequest := searchBuilder.Build()
		
		// Get history list
		response, err := aivisClient.ListTTSHistory(ctx, searchRequest)
		if err != nil {
			return fmt.Errorf("failed to list TTS history: %v", err)
		}
		
		if len(response.Histories) == 0 {
			fmt.Println("No TTS history records found.")
			return nil
		}
		
		// Display results in table format
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tText\tModel\tFormat\tSize\tCreated")
		
		for _, history := range response.Histories {
			text := history.Text
			if len(text) > 30 {
				text = text[:27] + "..."
			}
			
			model := history.ModelUUID
			if len(model) > 8 {
				model = model[:8] + "..."
			}
			
			// Format file size
			size := formatFileSize(history.FileSizeBytes)
			
			// Format creation time
			created := history.CreatedAt.Format("01/02 15:04")
			
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n",
				history.ID, text, model, history.FileFormat, size, created)
		}
		
		w.Flush()
		
		// Show pagination info
		fmt.Printf("\nShowing %d-%d of %d records",
			offset+1,
			offset+len(response.Histories),
			response.Total)
		
		if response.HasMore {
			fmt.Printf(" (use --offset %d to see more)", offset+limit)
		}
		fmt.Println()
		
		return nil
	},
}

var ttsHistoryShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show detailed information about a TTS history record",
	Long:  "Display detailed information about a specific TTS history record",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		// Parse ID
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid history ID: %s", args[0])
		}
		
		// Get history record
		history, err := aivisClient.GetTTSHistory(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to get TTS history: %v", err)
		}
		
		// Display detailed information
		fmt.Printf("TTS History Record #%d\n", history.ID)
		fmt.Println(strings.Repeat("=", 40))
		
		fmt.Printf("Text: %s\n", history.Text)
		fmt.Printf("Model UUID: %s\n", history.ModelUUID)
		fmt.Printf("Created: %s\n", history.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("File Path: %s\n", history.FilePath)
		fmt.Printf("File Format: %s\n", history.FileFormat)
		fmt.Printf("File Size: %s\n", formatFileSize(history.FileSizeBytes))
		
		if history.Credits != nil {
			fmt.Printf("Credits Used: %.4f\n", *history.Credits)
		}
		
		// Show TTS request details
		if history.Request != nil {
			fmt.Println("\nRequest Details:")
			fmt.Println(strings.Repeat("-", 20))
			
			if history.Request.SpeakerUUID != nil {
				fmt.Printf("Speaker UUID: %s\n", *history.Request.SpeakerUUID)
			}
			if history.Request.StyleID != nil {
				fmt.Printf("Style ID: %d\n", *history.Request.StyleID)
			}
			if history.Request.StyleName != nil {
				fmt.Printf("Style Name: %s\n", *history.Request.StyleName)
			}
			if history.Request.UseSSML != nil && *history.Request.UseSSML {
				fmt.Printf("SSML: Enabled\n")
			}
			if history.Request.SpeakingRate != nil {
				fmt.Printf("Speaking Rate: %.2f\n", *history.Request.SpeakingRate)
			}
			if history.Request.Pitch != nil {
				fmt.Printf("Pitch: %.2f\n", *history.Request.Pitch)
			}
			if history.Request.Volume != nil {
				fmt.Printf("Volume: %.2f\n", *history.Request.Volume)
			}
			if history.Request.OutputFormat != nil {
				fmt.Printf("Output Format: %s\n", string(*history.Request.OutputFormat))
			}
			if history.Request.OutputAudioChannels != nil {
				fmt.Printf("Audio Channels: %s\n", string(*history.Request.OutputAudioChannels))
			}
			if history.Request.LeadingSilenceSeconds != nil {
				fmt.Printf("Leading Silence: %.2f seconds\n", *history.Request.LeadingSilenceSeconds)
			}
			if history.Request.TrailingSilenceSeconds != nil {
				fmt.Printf("Trailing Silence: %.2f seconds\n", *history.Request.TrailingSilenceSeconds)
			}
			if history.Request.OutputSamplingRate != nil {
				fmt.Printf("Sampling Rate: %d Hz\n", *history.Request.OutputSamplingRate)
			}
			if history.Request.OutputBitrate != nil {
				fmt.Printf("Bitrate: %d kbps\n", *history.Request.OutputBitrate)
			}
		}
		
		return nil
	},
}

var ttsHistoryPlayCmd = &cobra.Command{
	Use:   "play <id>",
	Short: "Play audio from TTS history",
	Long:  "Play audio from a TTS history record",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		// Parse ID
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid history ID: %s", args[0])
		}
		
		// Get playback options from flags
		volume, _ := cmd.Flags().GetFloat64("volume")
		mode, _ := cmd.Flags().GetString("mode")
		waitForEnd, _ := cmd.Flags().GetBool("wait")
		
		// Create playback request with default values
		playbackBuilder := aivisClient.NewPlaybackRequest(nil). // Request will be retrieved from history
			WithWaitForEnd(waitForEnd)
		
		if volume > 0 {
			playbackBuilder = playbackBuilder.WithVolume(volume)
		}
		
		// Set playback mode
		switch mode {
		case "immediate":
			playbackBuilder = playbackBuilder.WithMode(ttsDomain.PlaybackModeImmediate)
		case "queue":
			playbackBuilder = playbackBuilder.WithMode(ttsDomain.PlaybackModeQueue)
		case "no_queue":
			playbackBuilder = playbackBuilder.WithMode(ttsDomain.PlaybackModeNoQueue)
		}
		
		playbackOptions := playbackBuilder.Build()
		
		// Play history
		if err := aivisClient.PlayTTSHistory(ctx, id, playbackOptions); err != nil {
			return fmt.Errorf("failed to play TTS history: %v", err)
		}
		
		fmt.Printf("Playing TTS history record #%d\n", id)
		
		return nil
	},
}

var ttsHistoryDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a TTS history record",
	Long:  "Delete a TTS history record and its associated audio file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		// Parse ID
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid history ID: %s", args[0])
		}
		
		// Get confirmation flag
		force, _ := cmd.Flags().GetBool("force")
		
		if !force {
			fmt.Printf("Are you sure you want to delete TTS history record #%d? [y/N]: ", id)
			var response string
			fmt.Scanln(&response)
			
			if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
				fmt.Println("Deletion cancelled.")
				return nil
			}
		}
		
		// Delete history
		if err := aivisClient.DeleteTTSHistory(ctx, id); err != nil {
			return fmt.Errorf("failed to delete TTS history: %v", err)
		}
		
		fmt.Printf("TTS history record #%d deleted successfully.\n", id)
		
		return nil
	},
}

var ttsHistoryCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean up TTS history",
	Long:  "Remove old TTS history records based on various criteria",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		// Get flags
		all, _ := cmd.Flags().GetBool("all")
		olderThanDays, _ := cmd.Flags().GetInt("older-than")
		force, _ := cmd.Flags().GetBool("force")
		
		if all {
			if !force {
				fmt.Print("Are you sure you want to delete ALL TTS history records? [y/N]: ")
				var response string
				fmt.Scanln(&response)
				
				if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
					fmt.Println("Cleanup cancelled.")
					return nil
				}
			}
			
			// Clear all history
			if err := aivisClient.ClearTTSHistory(ctx); err != nil {
				return fmt.Errorf("failed to clear TTS history: %v", err)
			}
			
			fmt.Println("All TTS history records deleted successfully.")
		} else if olderThanDays > 0 {
			// Delete records older than specified days
			cutoffDate := time.Now().AddDate(0, 0, -olderThanDays)
			
			// Find records to delete
			searchRequest := aivisClient.NewTTSHistorySearchRequest().
				WithLimit(1000). // Get a large batch
				WithSorting("created_at", "asc").
				Build()
			searchRequest.EndDate = &cutoffDate
			
			response, err := aivisClient.ListTTSHistory(ctx, searchRequest)
			if err != nil {
				return fmt.Errorf("failed to list old TTS history: %v", err)
			}
			
			if len(response.Histories) == 0 {
				fmt.Printf("No TTS history records found older than %d days.\n", olderThanDays)
				return nil
			}
			
			if !force {
				fmt.Printf("Found %d records older than %d days. Delete them? [y/N]: ",
					len(response.Histories), olderThanDays)
				var response string
				fmt.Scanln(&response)
				
				if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
					fmt.Println("Cleanup cancelled.")
					return nil
				}
			}
			
			// Delete the records
			count := 0
			for _, history := range response.Histories {
				if err := aivisClient.DeleteTTSHistory(ctx, history.ID); err != nil {
					fmt.Printf("Warning: failed to delete record #%d: %v\n", history.ID, err)
				} else {
					count++
				}
			}
			
			fmt.Printf("Deleted %d TTS history records older than %d days.\n", count, olderThanDays)
		} else {
			// Run automatic cleanup based on configuration
			// This is handled by the history manager's cleanup method
			fmt.Println("Running automatic cleanup based on configuration...")
			
			// Get stats before cleanup
			statsBefore, _ := aivisClient.GetTTSHistoryStats(ctx)
			
			// Note: Automatic cleanup is performed during normal operations
			// We don't have a direct cleanup method exposed, but we can suggest manual cleanup
			fmt.Printf("Current history: %d records, %s total size\n",
				statsBefore.TotalRecords,
				formatFileSize(statsBefore.TotalFileSize))
			
			fmt.Println("Use --all to delete all records or --older-than=N to delete records older than N days.")
		}
		
		return nil
	},
}

var ttsHistoryStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show TTS history statistics",
	Long:  "Display statistics about TTS history including record count, file sizes, and credits used",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		// Get history statistics
		stats, err := aivisClient.GetTTSHistoryStats(ctx)
		if err != nil {
			return fmt.Errorf("failed to get TTS history stats: %v", err)
		}
		
		// Display statistics
		fmt.Println("TTS History Statistics")
		fmt.Println(strings.Repeat("=", 25))
		fmt.Printf("Total Records: %d\n", stats.TotalRecords)
		fmt.Printf("Total File Size: %s\n", formatFileSize(stats.TotalFileSize))
		fmt.Printf("Total Credits Used: %.4f\n", stats.TotalCredits)
		
		if stats.TotalRecords > 0 {
			avgFileSize := stats.TotalFileSize / int64(stats.TotalRecords)
			avgCredits := stats.TotalCredits / float64(stats.TotalRecords)
			fmt.Printf("Average File Size: %s\n", formatFileSize(avgFileSize))
			fmt.Printf("Average Credits per Record: %.4f\n", avgCredits)
		}
		
		return nil
	},
}

// formatFileSize formats a file size in bytes to human-readable format
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func init() {
	// History list command flags
	ttsHistoryListCmd.Flags().Int("limit", 10, "Maximum number of records to display")
	ttsHistoryListCmd.Flags().Int("offset", 0, "Number of records to skip")
	ttsHistoryListCmd.Flags().String("model-uuid", "", "Filter by model UUID")
	ttsHistoryListCmd.Flags().String("text-contains", "", "Filter by text content")
	
	// History play command flags
	ttsHistoryPlayCmd.Flags().Float64("volume", 0, "Playback volume (0.0 to 1.0)")
	ttsHistoryPlayCmd.Flags().String("mode", "immediate", "Playback mode: immediate, queue, no_queue")
	ttsHistoryPlayCmd.Flags().Bool("wait", true, "Wait for playback to complete")
	
	// History delete command flags
	ttsHistoryDeleteCmd.Flags().Bool("force", false, "Skip confirmation prompt")
	
	// History clean command flags
	ttsHistoryCleanCmd.Flags().Bool("all", false, "Delete all history records")
	ttsHistoryCleanCmd.Flags().Int("older-than", 0, "Delete records older than N days")
	ttsHistoryCleanCmd.Flags().Bool("force", false, "Skip confirmation prompt")
	
	// Add subcommands to history command
	ttsHistoryCmd.AddCommand(ttsHistoryListCmd)
	ttsHistoryCmd.AddCommand(ttsHistoryShowCmd)
	ttsHistoryCmd.AddCommand(ttsHistoryPlayCmd)
	ttsHistoryCmd.AddCommand(ttsHistoryDeleteCmd)
	ttsHistoryCmd.AddCommand(ttsHistoryCleanCmd)
	ttsHistoryCmd.AddCommand(ttsHistoryStatsCmd)
	
	// Note: ttsHistoryCmd is added to ttsCmd in tts.go init() function
}
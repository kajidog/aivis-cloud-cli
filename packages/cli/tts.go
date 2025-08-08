package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	ttsDomain "github.com/kajidog/aiviscloud-mcp/client/tts/domain"
	"github.com/spf13/cobra"
)

// streamHandler implements TTSStreamHandler for CLI output
type streamHandler struct {
	verbose bool
}

func (h *streamHandler) OnChunk(chunk *ttsDomain.TTSStreamChunk) error {
	if h.verbose {
		fmt.Fprintf(os.Stderr, "Received chunk: %d bytes\n", len(chunk.Data))
	}
	
	// Write chunk to stdout
	_, err := os.Stdout.Write(chunk.Data)
	return err
}

func (h *streamHandler) OnComplete() error {
	if h.verbose {
		fmt.Fprintf(os.Stderr, "Streaming completed\n")
	}
	return nil
}

func (h *streamHandler) OnError(err error) {
	fmt.Fprintf(os.Stderr, "Streaming error: %v\n", err)
}

var ttsCmd = &cobra.Command{
	Use:   "tts",
	Short: "Text-to-speech operations",
	Long:  "Perform text-to-speech synthesis, audio playback, and related operations",
}

var ttsPlayCmd = &cobra.Command{
	Use:   "play [text] [model-uuid]",
	Short: "Synthesize text and play audio",
	Long:  "Convert text to speech using specified model and play the audio",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		text := args[0]
		modelUUID := args[1]

		// Get flags
		volume, _ := cmd.Flags().GetFloat64("volume")
		rate, _ := cmd.Flags().GetFloat64("rate")
		pitch, _ := cmd.Flags().GetFloat64("pitch")
		ssml, _ := cmd.Flags().GetBool("ssml")
		mode, _ := cmd.Flags().GetString("mode")

		// Build TTS request
		request := aivisClient.NewTTSRequest(modelUUID, text)
		
		if volume > 0 {
			request = request.WithVolume(volume)
		}
		if rate > 0 {
			request = request.WithSpeakingRate(rate)
		}
		if pitch != 0 {
			request = request.WithPitch(pitch)
		}
		if ssml {
			request = request.WithSSML(true)
		}

		ttsReq := request.Build()

		// Build playback request
		playbackBuilder := aivisClient.NewPlaybackRequest(ttsReq)
		
		switch mode {
		case "immediate":
			playbackBuilder = playbackBuilder.WithMode(ttsDomain.PlaybackModeImmediate)
		case "queue":
			playbackBuilder = playbackBuilder.WithMode(ttsDomain.PlaybackModeQueue)
		case "no_queue":
			playbackBuilder = playbackBuilder.WithMode(ttsDomain.PlaybackModeNoQueue)
		}

		playbackReq := playbackBuilder.Build()

		ctx := context.Background()
		if err := aivisClient.PlayRequest(ctx, playbackReq); err != nil {
			return fmt.Errorf("failed to play text: %v", err)
		}

		if verbose {
			fmt.Fprintf(os.Stderr, "Successfully played text: %s\n", text)
		}

		return nil
	},
}

var ttsSynthesizeCmd = &cobra.Command{
	Use:   "synthesize [text] [model-uuid] [output-file]",
	Short: "Synthesize text to audio file",
	Long:  "Convert text to speech and save to audio file",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		text := args[0]
		modelUUID := args[1]
		outputFile := args[2]

		// Get flags
		volume, _ := cmd.Flags().GetFloat64("volume")
		rate, _ := cmd.Flags().GetFloat64("rate")
		pitch, _ := cmd.Flags().GetFloat64("pitch")
		ssml, _ := cmd.Flags().GetBool("ssml")
		format, _ := cmd.Flags().GetString("format")

		// Build TTS request
		request := aivisClient.NewTTSRequest(modelUUID, text)
		
		if volume > 0 {
			request = request.WithVolume(volume)
		}
		if rate > 0 {
			request = request.WithSpeakingRate(rate)
		}
		if pitch != 0 {
			request = request.WithPitch(pitch)
		}
		if ssml {
			request = request.WithSSML(true)
		}

		// Set output format
		switch format {
		case "wav":
			request = request.WithOutputFormat(ttsDomain.OutputFormatWAV)
		case "flac":
			request = request.WithOutputFormat(ttsDomain.OutputFormatFLAC)
		case "mp3":
			request = request.WithOutputFormat(ttsDomain.OutputFormatMP3)
		case "aac":
			request = request.WithOutputFormat(ttsDomain.OutputFormatAAC)
		case "opus":
			request = request.WithOutputFormat(ttsDomain.OutputFormatOpus)
		}

		ttsReq := request.Build()

		// Create output file
		file, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %v", err)
		}
		defer file.Close()

		ctx := context.Background()
		if err := aivisClient.SynthesizeToFile(ctx, ttsReq, file); err != nil {
			return fmt.Errorf("failed to synthesize to file: %v", err)
		}

		fmt.Printf("Audio saved to: %s\n", outputFile)
		return nil
	},
}

var ttsStreamCmd = &cobra.Command{
	Use:   "stream [text] [model-uuid]",
	Short: "Stream synthesis with real-time output",
	Long:  "Convert text to speech with streaming synthesis",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		text := args[0]
		modelUUID := args[1]

		// Build basic TTS request
		request := aivisClient.NewTTSRequest(modelUUID, text).Build()

		ctx := context.Background()
		
		// Create streaming handler
		handler := &streamHandler{verbose: verbose}

		if err := aivisClient.SynthesizeStream(ctx, request, handler); err != nil {
			return fmt.Errorf("failed to stream synthesis: %v", err)
		}

		return nil
	},
}

var ttsControlCmd = &cobra.Command{
	Use:   "control [action]",
	Short: "Control audio playback",
	Long:  "Control ongoing audio playback (stop, pause, resume, status)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		action := args[0]

		switch action {
		case "stop":
			if err := aivisClient.StopPlayback(); err != nil {
				return fmt.Errorf("failed to stop playback: %v", err)
			}
			fmt.Println("Playback stopped")

		case "pause":
			if err := aivisClient.PausePlayback(); err != nil {
				return fmt.Errorf("failed to pause playback: %v", err)
			}
			fmt.Println("Playback paused")

		case "resume":
			if err := aivisClient.ResumePlayback(); err != nil {
				return fmt.Errorf("failed to resume playback: %v", err)
			}
			fmt.Println("Playback resumed")

		case "status":
			status := aivisClient.GetPlaybackStatus()
			fmt.Printf("Status: %+v\n", status)

		case "clear":
			aivisClient.ClearPlaybackQueue()
			fmt.Println("Playback queue cleared")

		default:
			return fmt.Errorf("unknown action: %s. Available actions: stop, pause, resume, status, clear", action)
		}

		return nil
	},
}

var ttsVolumeCmd = &cobra.Command{
	Use:   "volume [level]",
	Short: "Set playback volume",
	Long:  "Set audio playback volume (0.0 to 1.0)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		volume, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return fmt.Errorf("invalid volume level: %v", err)
		}

		if volume < 0.0 || volume > 1.0 {
			return fmt.Errorf("volume must be between 0.0 and 1.0")
		}

		if err := aivisClient.SetPlaybackVolume(volume); err != nil {
			return fmt.Errorf("failed to set volume: %v", err)
		}

		fmt.Printf("Volume set to: %.2f\n", volume)
		return nil
	},
}

func init() {
	// TTS play command flags
	ttsPlayCmd.Flags().Float64("volume", 0, "Audio volume (0.0 to 1.0)")
	ttsPlayCmd.Flags().Float64("rate", 0, "Speaking rate")
	ttsPlayCmd.Flags().Float64("pitch", 0, "Pitch adjustment")
	ttsPlayCmd.Flags().Bool("ssml", false, "Enable SSML parsing")
	ttsPlayCmd.Flags().String("mode", "", "Playback mode: immediate, queue, no_queue")

	// TTS synthesize command flags
	ttsSynthesizeCmd.Flags().Float64("volume", 0, "Audio volume (0.0 to 1.0)")
	ttsSynthesizeCmd.Flags().Float64("rate", 0, "Speaking rate")
	ttsSynthesizeCmd.Flags().Float64("pitch", 0, "Pitch adjustment")
	ttsSynthesizeCmd.Flags().Bool("ssml", false, "Enable SSML parsing")
	ttsSynthesizeCmd.Flags().String("format", "wav", "Output format: wav, flac, mp3, aac, opus")

	// Add subcommands to tts command
	ttsCmd.AddCommand(ttsPlayCmd)
	ttsCmd.AddCommand(ttsSynthesizeCmd)
	ttsCmd.AddCommand(ttsStreamCmd)
	ttsCmd.AddCommand(ttsControlCmd)
	ttsCmd.AddCommand(ttsVolumeCmd)
}
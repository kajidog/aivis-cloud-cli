package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	ttsDomain "github.com/kajidog/aivis-cloud-cli/client/tts/domain"
	"github.com/spf13/cobra"
)

// Default model UUID when not specified
const defaultModelUUID = "a59cb814-0083-4369-8542-f51a29e72af7"

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
	Args:  cobra.RangeArgs(0, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		text := ""
		modelUUID := defaultModelUUID
		
		// Get text from args or flag
		if len(args) > 0 {
			text = args[0]
			if len(args) > 1 {
				modelUUID = args[1]
			}
		}
		if flagText, _ := cmd.Flags().GetString("text"); flagText != "" {
			text = flagText
		}
		
		if text == "" {
			return fmt.Errorf("text is required (provide as argument or --text flag)")
		}

		// Check for model-uuid flag
		if flagModelUUID, _ := cmd.Flags().GetString("model-uuid"); flagModelUUID != "" {
			modelUUID = flagModelUUID
		}

		// Get flags
		volume, _ := cmd.Flags().GetFloat64("volume")
		rate, _ := cmd.Flags().GetFloat64("rate")
		pitch, _ := cmd.Flags().GetFloat64("pitch")
		ssml, _ := cmd.Flags().GetBool("ssml")
		channels, _ := cmd.Flags().GetString("channels")
		leadingSilence, _ := cmd.Flags().GetFloat64("leading-silence")
		trailingSilence, _ := cmd.Flags().GetFloat64("trailing-silence")

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
		if channels != "" {
			switch channels {
			case "mono":
				request = request.WithOutputChannels(ttsDomain.AudioChannelsMono)
			case "stereo":
				request = request.WithOutputChannels(ttsDomain.AudioChannelsStereo)
			}
		}
		if leadingSilence > 0 {
			request = request.WithLeadingSilence(leadingSilence)
		}
		if trailingSilence > 0 {
			request = request.WithTrailingSilence(trailingSilence)
		}

		ttsReq := request.Build()

		// Build playback request with WaitForEnd flag for synchronous playback
		playbackBuilder := aivisClient.NewPlaybackRequest(ttsReq).
			WithMode(ttsDomain.PlaybackModeImmediate).
			WithWaitForEnd(true)
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
	Use:   "synthesize [text] [output-file] [model-uuid]",
	Short: "Synthesize text to audio file",
	Long:  "Convert text to speech and save to audio file. If output file is not specified, it will be auto-generated.",
	Args:  cobra.RangeArgs(0, 3),
	RunE: func(cmd *cobra.Command, args []string) error {
		text := ""
		outputFile := ""
		modelUUID := defaultModelUUID
		
		// Get text and output from args or flags
		if len(args) > 0 {
			text = args[0]
			if len(args) > 1 {
				outputFile = args[1]
				if len(args) > 2 {
					modelUUID = args[2]
				}
			}
		}
		if flagText, _ := cmd.Flags().GetString("text"); flagText != "" {
			text = flagText
		}
		if flagOutput, _ := cmd.Flags().GetString("output"); flagOutput != "" {
			outputFile = flagOutput
		}
		
		if text == "" {
			return fmt.Errorf("text is required (provide as argument or --text flag)")
		}
		
		// Get format flag for filename generation
		format, _ := cmd.Flags().GetString("format")
		
		// Auto-generate output filename if not specified
		if outputFile == "" {
			// Generate filename based on current timestamp and format
			timestamp := time.Now().Format("20060102_150405")
			
			// Determine file extension based on format
			var extension string
			switch format {
			case "wav":
				extension = ".wav"
			case "flac":
				extension = ".flac"
			case "mp3":
				extension = ".mp3"
			case "aac":
				extension = ".aac"
			case "opus":
				extension = ".opus"
			default:
				extension = ".wav" // default
			}
			
			outputFile = fmt.Sprintf("tts_%s%s", timestamp, extension)
			
			if verbose {
				fmt.Fprintf(os.Stderr, "Auto-generated output filename: %s\n", outputFile)
			}
		}

		// Check for model-uuid flag
		if flagModelUUID, _ := cmd.Flags().GetString("model-uuid"); flagModelUUID != "" {
			modelUUID = flagModelUUID
		}

		// Get flags
		volume, _ := cmd.Flags().GetFloat64("volume")
		rate, _ := cmd.Flags().GetFloat64("rate")
		pitch, _ := cmd.Flags().GetFloat64("pitch")
		ssml, _ := cmd.Flags().GetBool("ssml")
		channels, _ := cmd.Flags().GetString("channels")
		leadingSilence, _ := cmd.Flags().GetFloat64("leading-silence")
		trailingSilence, _ := cmd.Flags().GetFloat64("trailing-silence")
		samplingRate, _ := cmd.Flags().GetInt("sampling-rate")
		bitrate, _ := cmd.Flags().GetInt("bitrate")

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

		// Set additional options
		if channels != "" {
			switch channels {
			case "mono":
				request = request.WithOutputChannels(ttsDomain.AudioChannelsMono)
			case "stereo":
				request = request.WithOutputChannels(ttsDomain.AudioChannelsStereo)
			}
		}
		if leadingSilence > 0 {
			request = request.WithLeadingSilence(leadingSilence)
		}
		if trailingSilence > 0 {
			request = request.WithTrailingSilence(trailingSilence)
		}
		if samplingRate > 0 {
			request = request.WithOutputSamplingRate(samplingRate)
		}
		if bitrate > 0 {
			request = request.WithOutputBitrate(bitrate)
		}

		ttsReq := request.Build()

		ctx := context.Background()
		
		// Use the new history-aware method
		response, err := aivisClient.SynthesizeToFileWithHistory(ctx, ttsReq, outputFile)
		if err != nil {
			return fmt.Errorf("failed to synthesize to file: %v", err)
		}

		fmt.Printf("Audio saved to: %s\n", outputFile)
		
		// Show history ID if available
		if response.HistoryID > 0 {
			fmt.Printf("History saved with ID: %d\n", response.HistoryID)
		}
		return nil
	},
}

var ttsStreamCmd = &cobra.Command{
	Use:   "stream [text] [model-uuid]",
	Short: "Stream synthesis with real-time output",
	Long:  "Convert text to speech with streaming synthesis",
	Args:  cobra.RangeArgs(0, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		text := ""
		modelUUID := defaultModelUUID
		
		// Get text from args or flag
		if len(args) > 0 {
			text = args[0]
			if len(args) > 1 {
				modelUUID = args[1]
			}
		}
		if flagText, _ := cmd.Flags().GetString("text"); flagText != "" {
			text = flagText
		}
		
		if text == "" {
			return fmt.Errorf("text is required (provide as argument or --text flag)")
		}

		// Check for model-uuid flag
		if flagModelUUID, _ := cmd.Flags().GetString("model-uuid"); flagModelUUID != "" {
			modelUUID = flagModelUUID
		}

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
	ttsPlayCmd.Flags().String("text", "", "Text to synthesize")
	ttsPlayCmd.Flags().String("model-uuid", "", "Voice model UUID (uses default if not specified)")
	ttsPlayCmd.Flags().Float64("volume", 0, "Audio volume (0.0 to 2.0)")
	ttsPlayCmd.Flags().Float64("rate", 0, "Speaking rate (0.5 to 2.0)")
	ttsPlayCmd.Flags().Float64("pitch", 0, "Pitch adjustment (-1.0 to 1.0)")
	ttsPlayCmd.Flags().Bool("ssml", false, "Enable SSML parsing")
	ttsPlayCmd.Flags().String("channels", "", "Audio channels: mono, stereo")
	ttsPlayCmd.Flags().Float64("leading-silence", 0, "Leading silence duration in seconds (0.0 to 60.0)")
	ttsPlayCmd.Flags().Float64("trailing-silence", 0, "Trailing silence duration in seconds (0.0 to 60.0)")

	// TTS synthesize command flags
	ttsSynthesizeCmd.Flags().String("text", "", "Text to synthesize")
	ttsSynthesizeCmd.Flags().String("output", "", "Output file path (auto-generated if not specified)")
	ttsSynthesizeCmd.Flags().String("model-uuid", "", "Voice model UUID (uses default if not specified)")
	ttsSynthesizeCmd.Flags().Float64("volume", 0, "Audio volume (0.0 to 2.0)")
	ttsSynthesizeCmd.Flags().Float64("rate", 0, "Speaking rate (0.5 to 2.0)")
	ttsSynthesizeCmd.Flags().Float64("pitch", 0, "Pitch adjustment (-1.0 to 1.0)")
	ttsSynthesizeCmd.Flags().Bool("ssml", false, "Enable SSML parsing")
	ttsSynthesizeCmd.Flags().String("format", "wav", "Output format: wav, flac, mp3, aac, opus")
	ttsSynthesizeCmd.Flags().String("channels", "", "Audio channels: mono, stereo")
	ttsSynthesizeCmd.Flags().Float64("leading-silence", 0, "Leading silence duration in seconds (0.0 to 60.0)")
	ttsSynthesizeCmd.Flags().Float64("trailing-silence", 0, "Trailing silence duration in seconds (0.0 to 60.0)")
	ttsSynthesizeCmd.Flags().Int("sampling-rate", 0, "Output sampling rate (8000, 11025, 12000, 16000, 22050, 24000, 44100, 48000)")
	ttsSynthesizeCmd.Flags().Int("bitrate", 0, "Output bitrate in kbps (8 to 320, not applicable for wav/flac)")

	// TTS stream command flags
	ttsStreamCmd.Flags().String("text", "", "Text to synthesize")
	ttsStreamCmd.Flags().String("model-uuid", "", "Voice model UUID (uses default if not specified)")

	// Add subcommands to tts command
	ttsCmd.AddCommand(ttsPlayCmd)
	ttsCmd.AddCommand(ttsSynthesizeCmd)
	ttsCmd.AddCommand(ttsStreamCmd)
	ttsCmd.AddCommand(ttsControlCmd)
	ttsCmd.AddCommand(ttsVolumeCmd)
	ttsCmd.AddCommand(ttsHistoryCmd) // Add history command
}
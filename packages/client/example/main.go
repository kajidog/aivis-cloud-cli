package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kajidog/aivis-cloud-cli/client"
	"github.com/kajidog/aivis-cloud-cli/client/tts/domain"
)

func main() {
	// Note: This is a demonstration of the API. 
	// For actual use, you need a valid API key and model UUID from AivisCloud.
	fmt.Println("AivisCloud Audio Playback API Demo")
	fmt.Println("===================================")
	
	// Create client with API key
	apiKey := "your-api-key-here"
	client, err := client.New(apiKey)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Example 1: Basic audio playback
	fmt.Println("Example 1: Basic audio playback")
	fmt.Println("client.PlayText(ctx, \"こんにちは、世界！\", \"model-uuid\")")
	err = client.PlayText(ctx, "こんにちは、世界！", "model-uuid")
	if err != nil {
		fmt.Printf("Expected error (no valid API key/model): %v\n", err)
	}

	// Wait for playback to complete
	time.Sleep(2 * time.Second)

	// Example 2: Audio playback with custom options including new fields
	fmt.Println("Example 2: Custom playback options")
	ttsRequest := client.NewTTSRequest("model-uuid", "カスタム音声設定のテストです").
		WithVolume(0.8).
		WithSpeakingRate(1.2).
		WithLeadingSilence(0.2).
		WithTrailingSilence(0.3).
		WithOutputChannels(domain.AudioChannelsStereo).
		Build()

	playbackRequest := client.NewPlaybackRequest(ttsRequest).
		WithMode(domain.PlaybackModeQueue).
		WithVolume(0.7).
		Build()

	err = client.PlayRequest(ctx, playbackRequest)
	if err != nil {
		client.GetLogger().Errorf("Failed to play with options: %v", err)
	}

	// Example 3: Playback control
	fmt.Println("Example 3: Playback control")
	
	// Queue multiple texts
	client.PlayTextWithOptions(ctx, "最初のメッセージ", "model-uuid", 
		&domain.PlaybackRequest{Mode: &[]domain.PlaybackMode{domain.PlaybackModeQueue}[0]})
	client.PlayTextWithOptions(ctx, "2番目のメッセージ", "model-uuid", 
		&domain.PlaybackRequest{Mode: &[]domain.PlaybackMode{domain.PlaybackModeQueue}[0]})

	// Get status
	status := client.GetPlaybackStatus()
	fmt.Printf("Playback status: %+v\n", status)

	// Pause after 1 second
	go func() {
		time.Sleep(1 * time.Second)
		fmt.Println("Pausing playback...")
		client.PausePlayback()
		
		time.Sleep(2 * time.Second)
		fmt.Println("Resuming playback...")
		client.ResumePlayback()
	}()

	// Wait for all playback to complete
	time.Sleep(10 * time.Second)

	// Example 4: Volume control
	fmt.Println("Example 4: Volume control")
	client.SetPlaybackVolume(0.5)
	client.PlayText(ctx, "音量を下げてテストしています", "model-uuid")

	time.Sleep(3 * time.Second)

	// Stop all playback and clear queue
	client.StopPlayback()
	client.ClearPlaybackQueue()

	fmt.Println("Audio playback examples completed!")
}
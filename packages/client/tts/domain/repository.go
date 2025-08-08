package domain

import (
	"context"
	"io"
)

// TTSRepository defines the interface for text-to-speech operations
type TTSRepository interface {
	// Synthesize performs text-to-speech synthesis
	Synthesize(ctx context.Context, request *TTSRequest) (*TTSResponse, error)

	// SynthesizeStream performs streaming text-to-speech synthesis
	SynthesizeStream(ctx context.Context, request *TTSRequest) (io.ReadCloser, error)
}

// TTSStreamHandler handles streaming TTS responses
type TTSStreamHandler interface {
	// OnChunk is called when a new audio chunk is received
	OnChunk(chunk *TTSStreamChunk) error

	// OnComplete is called when streaming is complete
	OnComplete() error

	// OnError is called when an error occurs during streaming
	OnError(err error)
}

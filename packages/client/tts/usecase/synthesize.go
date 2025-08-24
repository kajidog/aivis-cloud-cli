package usecase

import (
	"bufio"
	"context"
	"io"
	"time"

	"github.com/kajidog/aivis-cloud-cli/client/tts/domain"
)

// TTSSynthesizer handles text-to-speech synthesis use cases
type TTSSynthesizer struct {
	repository domain.TTSRepository
}

// NewTTSSynthesizer creates a new TTS synthesizer
func NewTTSSynthesizer(repository domain.TTSRepository) *TTSSynthesizer {
	return &TTSSynthesizer{
		repository: repository,
	}
}

// Synthesize performs text-to-speech synthesis
func (s *TTSSynthesizer) Synthesize(ctx context.Context, request *domain.TTSRequest) (*domain.TTSResponse, error) {
	return s.repository.Synthesize(ctx, request)
}

// SynthesizeToFile performs text-to-speech synthesis and writes to a file
func (s *TTSSynthesizer) SynthesizeToFile(ctx context.Context, request *domain.TTSRequest, writer io.Writer) error {
	response, err := s.repository.Synthesize(ctx, request)
	if err != nil {
		return err
	}
	defer response.AudioData.Close()

	_, err = io.Copy(writer, response.AudioData)
	return err
}

// SynthesizeStream performs streaming text-to-speech synthesis
func (s *TTSSynthesizer) SynthesizeStream(ctx context.Context, request *domain.TTSRequest, handler domain.TTSStreamHandler) error {
	stream, err := s.repository.SynthesizeStream(ctx, request)
	if err != nil {
		handler.OnError(err)
		return err
	}
	defer stream.Close()

	return s.processStream(stream, handler)
}

// processStream processes the streaming response
func (s *TTSSynthesizer) processStream(stream io.ReadCloser, handler domain.TTSStreamHandler) error {
	reader := bufio.NewReader(stream)
	buffer := make([]byte, 4096)

	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			chunk := &domain.TTSStreamChunk{
				Data:      buffer[:n],
				Timestamp: time.Now(),
				IsLast:    false,
			}

			if handlerErr := handler.OnChunk(chunk); handlerErr != nil {
				handler.OnError(handlerErr)
				return handlerErr
			}
		}

		if err == io.EOF {
			// Mark the last chunk
			if n > 0 {
				chunk := &domain.TTSStreamChunk{
					Data:      buffer[:n],
					Timestamp: time.Now(),
					IsLast:    true,
				}
				if handlerErr := handler.OnChunk(chunk); handlerErr != nil {
					handler.OnError(handlerErr)
					return handlerErr
				}
			}
			break
		}

		if err != nil {
			handler.OnError(err)
			return err
		}
	}

	return handler.OnComplete()
}

// ValidateRequest validates a TTS request
func (s *TTSSynthesizer) ValidateRequest(request *domain.TTSRequest) error {
	if request.ModelUUID == "" {
		return &ValidationError{Field: "ModelUUID", Message: "Model UUID is required"}
	}

	if request.Text == "" {
		return &ValidationError{Field: "Text", Message: "Text is required"}
	}

	if len(request.Text) > 3000 {
		return &ValidationError{Field: "Text", Message: "Text must not exceed 3000 characters"}
	}

	// Validate style configuration
	if request.StyleID != nil && request.StyleName != nil {
		return &ValidationError{Field: "Style", Message: "StyleID and StyleName cannot be specified simultaneously"}
	}

	if request.StyleID != nil && (*request.StyleID < 0 || *request.StyleID > 31) {
		return &ValidationError{Field: "StyleID", Message: "StyleID must be between 0 and 31"}
	}

	// Validate voice parameters
	if request.SpeakingRate != nil && (*request.SpeakingRate < 0.5 || *request.SpeakingRate > 2.0) {
		return &ValidationError{Field: "SpeakingRate", Message: "SpeakingRate must be between 0.5 and 2.0"}
	}

	if request.Pitch != nil && (*request.Pitch < -1.0 || *request.Pitch > 1.0) {
		return &ValidationError{Field: "Pitch", Message: "Pitch must be between -1.0 and 1.0"}
	}

	if request.Volume != nil && (*request.Volume < 0.0 || *request.Volume > 2.0) {
		return &ValidationError{Field: "Volume", Message: "Volume must be between 0.0 and 2.0"}
	}

	if request.EmotionalIntensity != nil && (*request.EmotionalIntensity < 0.0 || *request.EmotionalIntensity > 2.0) {
		return &ValidationError{Field: "EmotionalIntensity", Message: "EmotionalIntensity must be between 0.0 and 2.0"}
	}

	if request.TempoDynamics != nil && (*request.TempoDynamics < 0.0 || *request.TempoDynamics > 2.0) {
		return &ValidationError{Field: "TempoDynamics", Message: "TempoDynamics must be between 0.0 and 2.0"}
	}

	// Validate output configuration
	if request.OutputSamplingRate != nil && *request.OutputSamplingRate <= 0 {
		return &ValidationError{Field: "OutputSamplingRate", Message: "OutputSamplingRate must be positive"}
	}

	if request.OutputAudioChannels != nil {
		if *request.OutputAudioChannels != domain.AudioChannelsMono && *request.OutputAudioChannels != domain.AudioChannelsStereo {
			return &ValidationError{Field: "OutputAudioChannels", Message: "OutputAudioChannels must be 'mono' or 'stereo'"}
		}
	}

	if request.OutputBitrate != nil && *request.OutputBitrate <= 0 {
		return &ValidationError{Field: "OutputBitrate", Message: "OutputBitrate must be positive"}
	}

	// Validate Opus-specific sampling rates
	if request.OutputFormat != nil && *request.OutputFormat == domain.OutputFormatOpus {
		if request.OutputSamplingRate != nil {
			validRates := []int{8000, 12000, 16000, 24000, 48000}
			valid := false
			for _, rate := range validRates {
				if *request.OutputSamplingRate == rate {
					valid = true
					break
				}
			}
			if !valid {
				return &ValidationError{Field: "OutputSamplingRate", Message: "For Opus format, OutputSamplingRate must be 8000, 12000, 16000, 24000, or 48000 Hz"}
			}
		}
	}

	return nil
}

// ValidationError represents a request validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

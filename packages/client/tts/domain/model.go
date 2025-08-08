package domain

import (
	"io"
	"time"

	"github.com/kajidog/aiviscloud-mcp/client/common/http"
)

// OutputFormat represents the audio output format
type OutputFormat string

const (
	OutputFormatWAV  OutputFormat = "wav"
	OutputFormatFLAC OutputFormat = "flac"
	OutputFormatMP3  OutputFormat = "mp3"
	OutputFormatAAC  OutputFormat = "aac"
	OutputFormatOpus OutputFormat = "opus"
)

// Language represents the supported language
type Language string

const (
	LanguageJapanese Language = "ja"
)

// TTSRequest represents a text-to-speech synthesis request
type TTSRequest struct {
	// Required fields
	ModelUUID string `json:"model_uuid"`
	Text      string `json:"text"`

	// Optional speaker configuration
	SpeakerUUID *string `json:"speaker_uuid,omitempty"`
	StyleID     *int    `json:"style_id,omitempty"`
	StyleName   *string `json:"style_name,omitempty"`

	// Optional user dictionary
	UserDictionaryUUID *string `json:"user_dictionary_uuid,omitempty"`

	// Text processing options
	UseSSML  *bool     `json:"use_ssml,omitempty"`
	Language *Language `json:"language,omitempty"`

	// Audio output configuration
	OutputFormat            *OutputFormat `json:"output_format,omitempty"`
	OutputSamplingRate      *int          `json:"output_sampling_rate,omitempty"`
	OutputAudioChannels     *int          `json:"output_audio_channels,omitempty"`
	OutputBitrate           *int          `json:"output_bitrate,omitempty"`
	LineBreakSilenceSeconds *float64      `json:"line_break_silence_seconds,omitempty"`

	// Voice parameters
	SpeakingRate       *float64 `json:"speaking_rate,omitempty"`
	Pitch              *float64 `json:"pitch,omitempty"`
	Volume             *float64 `json:"volume,omitempty"`
	EmotionalIntensity *float64 `json:"emotional_intensity,omitempty"`
	TempoDynamics      *float64 `json:"tempo_dynamics,omitempty"`
	PrePhonemeLength   *float64 `json:"pre_phoneme_length,omitempty"`
	PostPhonemeLength  *float64 `json:"post_phoneme_length,omitempty"`
}

// TTSResponse represents a text-to-speech synthesis response
type TTSResponse struct {
	AudioData   io.ReadCloser     `json:"-"`
	BillingInfo *http.BillingInfo `json:"billing_info,omitempty"`
	FileName    string            `json:"filename,omitempty"`
}

// TTSStreamChunk represents a chunk of streaming audio data
type TTSStreamChunk struct {
	Data      []byte    `json:"data"`
	Timestamp time.Time `json:"timestamp"`
	IsLast    bool      `json:"is_last"`
}

// TTSRequestBuilder helps build TTS requests with method chaining
type TTSRequestBuilder struct {
	request *TTSRequest
}

// NewTTSRequestBuilder creates a new TTS request builder
func NewTTSRequestBuilder(modelUUID, text string) *TTSRequestBuilder {
	return &TTSRequestBuilder{
		request: &TTSRequest{
			ModelUUID: modelUUID,
			Text:      text,
		},
	}
}

// WithSpeaker sets the speaker UUID
func (b *TTSRequestBuilder) WithSpeaker(speakerUUID string) *TTSRequestBuilder {
	b.request.SpeakerUUID = &speakerUUID
	return b
}

// WithStyleID sets the style ID
func (b *TTSRequestBuilder) WithStyleID(styleID int) *TTSRequestBuilder {
	b.request.StyleID = &styleID
	return b
}

// WithStyleName sets the style name
func (b *TTSRequestBuilder) WithStyleName(styleName string) *TTSRequestBuilder {
	b.request.StyleName = &styleName
	return b
}

// WithUserDictionary sets the user dictionary UUID
func (b *TTSRequestBuilder) WithUserDictionary(userDictionaryUUID string) *TTSRequestBuilder {
	b.request.UserDictionaryUUID = &userDictionaryUUID
	return b
}

// WithSSML enables or disables SSML processing
func (b *TTSRequestBuilder) WithSSML(useSSML bool) *TTSRequestBuilder {
	b.request.UseSSML = &useSSML
	return b
}

// WithLanguage sets the language
func (b *TTSRequestBuilder) WithLanguage(language Language) *TTSRequestBuilder {
	b.request.Language = &language
	return b
}

// WithOutputFormat sets the output format
func (b *TTSRequestBuilder) WithOutputFormat(format OutputFormat) *TTSRequestBuilder {
	b.request.OutputFormat = &format
	return b
}

// WithOutputSamplingRate sets the output sampling rate
func (b *TTSRequestBuilder) WithOutputSamplingRate(rate int) *TTSRequestBuilder {
	b.request.OutputSamplingRate = &rate
	return b
}

// WithOutputChannels sets the output audio channels
func (b *TTSRequestBuilder) WithOutputChannels(channels int) *TTSRequestBuilder {
	b.request.OutputAudioChannels = &channels
	return b
}

// WithOutputBitrate sets the output bitrate
func (b *TTSRequestBuilder) WithOutputBitrate(bitrate int) *TTSRequestBuilder {
	b.request.OutputBitrate = &bitrate
	return b
}

// WithLineBreakSilence sets the line break silence duration
func (b *TTSRequestBuilder) WithLineBreakSilence(seconds float64) *TTSRequestBuilder {
	b.request.LineBreakSilenceSeconds = &seconds
	return b
}

// WithSpeakingRate sets the speaking rate
func (b *TTSRequestBuilder) WithSpeakingRate(rate float64) *TTSRequestBuilder {
	b.request.SpeakingRate = &rate
	return b
}

// WithPitch sets the pitch
func (b *TTSRequestBuilder) WithPitch(pitch float64) *TTSRequestBuilder {
	b.request.Pitch = &pitch
	return b
}

// WithVolume sets the volume
func (b *TTSRequestBuilder) WithVolume(volume float64) *TTSRequestBuilder {
	b.request.Volume = &volume
	return b
}

// WithEmotionalIntensity sets the emotional intensity
func (b *TTSRequestBuilder) WithEmotionalIntensity(intensity float64) *TTSRequestBuilder {
	b.request.EmotionalIntensity = &intensity
	return b
}

// WithTempoDynamics sets the tempo dynamics
func (b *TTSRequestBuilder) WithTempoDynamics(dynamics float64) *TTSRequestBuilder {
	b.request.TempoDynamics = &dynamics
	return b
}

// WithPrePhonemeLength sets the pre-phoneme length
func (b *TTSRequestBuilder) WithPrePhonemeLength(length float64) *TTSRequestBuilder {
	b.request.PrePhonemeLength = &length
	return b
}

// WithPostPhonemeLength sets the post-phoneme length
func (b *TTSRequestBuilder) WithPostPhonemeLength(length float64) *TTSRequestBuilder {
	b.request.PostPhonemeLength = &length
	return b
}

// Build returns the built TTS request
func (b *TTSRequestBuilder) Build() *TTSRequest {
	return b.request
}

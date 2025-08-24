package infrastructure

import (
	"context"
	"io"

	"github.com/kajidog/aivis-cloud-cli/client/common/http"
	"github.com/kajidog/aivis-cloud-cli/client/tts/domain"
)

// TTSAPIRepository implements the TTSRepository interface using HTTP API calls
type TTSAPIRepository struct {
	httpClient *http.Client
}

// NewTTSAPIRepository creates a new TTS API repository
func NewTTSAPIRepository(httpClient *http.Client) *TTSAPIRepository {
	return &TTSAPIRepository{
		httpClient: httpClient,
	}
}

// Synthesize performs text-to-speech synthesis
func (r *TTSAPIRepository) Synthesize(ctx context.Context, request *domain.TTSRequest) (*domain.TTSResponse, error) {
	httpReq := &http.Request{
		Method: "POST",
		Path:   "/v1/tts/synthesize",
		Body:   request,
	}

	resp, err := r.httpClient.DoStream(ctx, httpReq)
	if err != nil {
		return nil, err
	}

	billingInfo := http.GetBillingInfo(resp.Headers)
	fileName := http.GetFileName(resp.Headers)

	return &domain.TTSResponse{
		AudioData:   resp.Body,
		BillingInfo: billingInfo,
		FileName:    fileName,
	}, nil
}

// SynthesizeStream performs streaming text-to-speech synthesis
func (r *TTSAPIRepository) SynthesizeStream(ctx context.Context, request *domain.TTSRequest) (io.ReadCloser, error) {
	httpReq := &http.Request{
		Method: "POST",
		Path:   "/v1/tts/synthesize",
		Body:   request,
	}

	resp, err := r.httpClient.DoStream(ctx, httpReq)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

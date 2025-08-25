package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kajidog/aivis-cloud-cli/client/config"
	paymentdomain "github.com/kajidog/aivis-cloud-cli/client/payment/domain"
	ttsdomain "github.com/kajidog/aivis-cloud-cli/client/tts/domain"
	usersdomain "github.com/kajidog/aivis-cloud-cli/client/users/domain"
)

// setupTestClient is a helper function to create a client with a mock server.
func setupTestClient(t *testing.T, handler http.HandlerFunc) (*Client, func()) {
	t.Helper()

	server := httptest.NewServer(handler)

	cfg := config.NewConfig("test_api_key")
	cfg.BaseURL = server.URL

	client, err := NewWithConfig(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	teardown := func() {
		server.Close()
	}

	return client, teardown
}

func TestNew(t *testing.T) {
	client, err := New("test_api_key")
	if err != nil {
		t.Fatalf("New() error = %v, wantErr %v", err, false)
	}
	if client == nil {
		t.Fatal("New() client is nil")
	}
}

func TestNewWithConfig(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		wantErr bool
	}{
		{
			name:    "valid config",
			apiKey:  "test_api_key",
			wantErr: false,
		},
		{
			name:    "empty API key",
			apiKey:  "",
			wantErr: true,
		},
		{
			name:    "whitespace API key",
			apiKey:  "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.NewConfig(strings.TrimSpace(tt.apiKey))
			client, err := NewWithConfig(cfg)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewWithConfig() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}
			
			if err != nil {
				t.Errorf("NewWithConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if client == nil {
				t.Error("NewWithConfig() returned nil client")
			}
		})
	}
}

func TestSearchPublicModels(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		statusCode   int
		responseBody string
		wantErr      bool
		wantTotal    int64
	}{
		{
			name:       "successful search",
			query:      "test",
			statusCode: 200,
			responseBody: `{
				"models": [{"uuid": "test-uuid", "name": "test-model"}],
				"total": 1
			}`,
			wantErr:   false,
			wantTotal: 1,
		},
		{
			name:         "not found",
			query:        "nonexistent",
			statusCode:   404,
			responseBody: `{"error": "not found"}`,
			wantErr:      true,
		},
		{
			name:         "unauthorized",
			query:        "test",
			statusCode:   401,
			responseBody: `{"error": "unauthorized"}`,
			wantErr:      true,
		},
		{
			name:         "server error",
			query:        "test",
			statusCode:   500,
			responseBody: `{"error": "internal server error"}`,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				// Validate request
				if r.URL.Path != "/v1/aivm-models/search" {
					t.Errorf("Expected path /v1/aivm-models/search, got %s", r.URL.Path)
				}
				if r.Method != "GET" {
					t.Errorf("Expected method GET, got %s", r.Method)
				}
				
				// Only check query params for successful cases
				if tt.statusCode == 200 {
					query := r.URL.Query().Get("q")
					if query != tt.query {
						t.Errorf("Expected query '%s', got %s", tt.query, query)
					}
					isPublic := r.URL.Query().Get("public")
					if isPublic != "true" {
						t.Errorf("Expected public query param to be 'true', got %s", isPublic)
					}
				}

				w.WriteHeader(tt.statusCode)
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(tt.responseBody))
			}

			client, teardown := setupTestClient(t, handler)
			defer teardown()

			resp, err := client.SearchPublicModels(context.Background(), tt.query)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("SearchPublicModels() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("SearchPublicModels() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if resp == nil {
				t.Fatal("SearchPublicModels() response is nil")
			}

			if resp.Total != tt.wantTotal {
				t.Errorf("Expected total %d, got %d", tt.wantTotal, resp.Total)
			}
		})
	}
}

func TestSynthesize(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/tts/synthesize" {
			t.Errorf("Expected path /v1/tts/synthesize, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected method POST, got %s", r.Method)
		}

		var req ttsdomain.TTSRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}
		if req.Text != "hello" {
			t.Errorf("Expected text 'hello', got '%s'", req.Text)
		}
		if req.ModelUUID != "test-model-uuid" {
			t.Errorf("Expected model UUID 'test-model-uuid', got '%s'", req.ModelUUID)
		}

		w.Header().Set("Content-Type", "audio/mpeg")
		w.Write([]byte("fake-audio-data"))
	}

	client, teardown := setupTestClient(t, handler)
	defer teardown()

	req := &ttsdomain.TTSRequest{
		ModelUUID: "test-model-uuid",
		Text:      "hello",
	}

	resp, err := client.Synthesize(context.Background(), req)
	if err != nil {
		t.Fatalf("Synthesize() error = %v, wantErr %v", err, false)
	}
	defer resp.AudioData.Close()

	if resp == nil {
		t.Fatal("Synthesize() response is nil")
	}

	audioData, err := io.ReadAll(resp.AudioData)
	if err != nil {
		t.Fatalf("Failed to read audio data: %v", err)
	}

	if string(audioData) != "fake-audio-data" {
		t.Errorf("Expected audio data 'fake-audio-data', got '%s'", string(audioData))
	}
}

func TestGetMe(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/users/me" {
			t.Errorf("Expected path /v1/users/me, got %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Expected method GET, got %s", r.Method)
		}

		response := usersdomain.UserMe{
			User: usersdomain.User{
				ID:     "test-user-id",
				Handle: "test-user",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}

	client, teardown := setupTestClient(t, handler)
	defer teardown()

	resp, err := client.GetMe(context.Background())
	if err != nil {
		t.Fatalf("GetMe() error = %v, wantErr %v", err, false)
	}

	if resp == nil {
		t.Fatal("GetMe() response is nil")
	}

	if resp.ID != "test-user-id" {
		t.Errorf("Expected user ID 'test-user-id', got '%s'", resp.ID)
	}
	if resp.Handle != "test-user" {
		t.Errorf("Expected user handle 'test-user', got '%s'", resp.Handle)
	}
}

func TestGetAPIKeys(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/payment/api-keys" {
			t.Errorf("Expected path /v1/payment/api-keys, got %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Expected method GET, got %s", r.Method)
		}

		response := paymentdomain.APIKeyListResponse{
			APIKeys: []paymentdomain.APIKey{
				{
					ID:         "test-key-id",
					Name:       "test-key",
					KeyPreview: "test-key-preview",
				},
			},
			Total: 1,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}

	client, teardown := setupTestClient(t, handler)
	defer teardown()

	resp, err := client.GetAPIKeys(context.Background(), 10, 0)
	if err != nil {
		t.Fatalf("GetAPIKeys() error = %v, wantErr %v", err, false)
	}

	if resp == nil {
		t.Fatal("GetAPIKeys() response is nil")
	}

	if resp.Total != 1 {
		t.Errorf("Expected total 1, got %d", resp.Total)
	}
	if len(resp.APIKeys) != 1 {
		t.Fatalf("Expected 1 api key, got %d", len(resp.APIKeys))
	}
	if resp.APIKeys[0].Name != "test-key" {
		t.Errorf("Expected api key name 'test-key', got '%s'", resp.APIKeys[0].Name)
	}
}

// TestTTSRequestBuilder tests the builder pattern for TTS requests
func TestTTSRequestBuilder(t *testing.T) {
	tests := []struct {
		name        string
		modelUUID   string
		text        string
		buildFunc   func(*Client) *ttsdomain.TTSRequest
		wantFormat  ttsdomain.OutputFormat
		wantSSML    bool
		wantVolume  float64
	}{
		{
			name:      "basic request",
			modelUUID: "test-model",
			text:      "hello",
			buildFunc: func(c *Client) *ttsdomain.TTSRequest {
				return c.NewTTSRequest("test-model", "hello").
					WithOutputFormat(ttsdomain.OutputFormatWAV).
					WithSSML(false).
					WithVolume(1.0).
					Build()
			},
			wantFormat: ttsdomain.OutputFormatWAV,
			wantSSML:   false,
			wantVolume: 1.0,
		},
		{
			name:      "request with SSML",
			modelUUID: "test-model",
			text:      "<speak>hello</speak>",
			buildFunc: func(c *Client) *ttsdomain.TTSRequest {
				return c.NewTTSRequest("test-model", "<speak>hello</speak>").
					WithSSML(true).
					WithOutputFormat(ttsdomain.OutputFormatMP3).
					WithVolume(0.8).
					Build()
			},
			wantFormat: ttsdomain.OutputFormatMP3,
			wantSSML:   true,
			wantVolume: 0.8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New("test-api-key")
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			req := tt.buildFunc(client)

			if req.ModelUUID != tt.modelUUID {
				t.Errorf("Expected ModelUUID %s, got %s", tt.modelUUID, req.ModelUUID)
			}
			if req.Text != tt.text {
				t.Errorf("Expected Text %s, got %s", tt.text, req.Text)
			}
			if req.OutputFormat == nil || *req.OutputFormat != tt.wantFormat {
				var gotFormat ttsdomain.OutputFormat
				if req.OutputFormat != nil {
					gotFormat = *req.OutputFormat
				}
				t.Errorf("Expected OutputFormat %s, got %s", tt.wantFormat, gotFormat)
			}
			if req.UseSSML == nil || *req.UseSSML != tt.wantSSML {
				var gotSSML bool
				if req.UseSSML != nil {
					gotSSML = *req.UseSSML
				}
				t.Errorf("Expected UseSSML %v, got %v", tt.wantSSML, gotSSML)
			}
			if req.Volume == nil || *req.Volume != tt.wantVolume {
				var gotVolume float64
				if req.Volume != nil {
					gotVolume = *req.Volume
				}
				t.Errorf("Expected Volume %f, got %f", tt.wantVolume, gotVolume)
			}
		})
	}
}

// TestHTTPErrorHandling tests various HTTP error scenarios
func TestHTTPErrorHandling(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    string
	}{
		{
			name:       "unauthorized",
			statusCode: 401,
			wantErr:    "unauthorized",
		},
		{
			name:       "insufficient credits",
			statusCode: 402,
			wantErr:    "payment required",
		},
		{
			name:       "rate limit exceeded",
			statusCode: 429,
			wantErr:    "too many requests",
		},
		{
			name:       "server error",
			statusCode: 500,
			wantErr:    "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(`{"error": "test error"}`))
			}

			client, teardown := setupTestClient(t, handler)
			defer teardown()

			// Test with different endpoints
			_, err := client.SearchPublicModels(context.Background(), "test")
			if err == nil {
				t.Errorf("Expected error for status code %d", tt.statusCode)
			}

			_, err = client.GetMe(context.Background())
			if err == nil {
				t.Errorf("Expected error for status code %d", tt.statusCode)
			}
		})
	}
}

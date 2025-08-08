package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/kajidog/aiviscloud-mcp/client/common/errors"
	"github.com/kajidog/aiviscloud-mcp/client/config"
)

// Client represents an HTTP client for Aivis Cloud API
type Client struct {
	config     *config.Config
	httpClient *http.Client
}

// NewClient creates a new HTTP client
func NewClient(cfg *config.Config) *Client {
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: cfg.HTTPTimeout,
		},
	}
}

// Request represents an HTTP request
type Request struct {
	Method  string
	Path    string
	Body    interface{}
	Query   url.Values
	Headers map[string]string
}

// Response represents an HTTP response
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       io.ReadCloser
}

// Do executes an HTTP request
func (c *Client) Do(ctx context.Context, req *Request) (*Response, error) {
	httpReq, err := c.buildHTTPRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to build HTTP request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	// Check for API errors based on status code
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		var message string
		switch resp.StatusCode {
		case 401:
			message = "API key is required or invalid"
		case 402:
			message = "Credit balance is insufficient"
		case 404:
			message = "Specified model UUID not found"
		case 422:
			message = "Request parameter format is incorrect"
		case 429:
			message = "Rate limit exceeded"
		case 500:
			message = "Unknown error occurred during synthesis server connection"
		case 502:
			message = "Failed to connect to synthesis server"
		case 503:
			message = "Synthesis server is experiencing issues"
		case 504:
			message = "Connection to synthesis server timed out"
		default:
			message = string(body)
		}

		return nil, errors.NewAPIErrorFromHTTP(resp, message)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       resp.Body,
	}, nil
}

// DoJSON executes an HTTP request and unmarshals JSON response
func (c *Client) DoJSON(ctx context.Context, req *Request, result interface{}) error {
	resp, err := c.Do(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if len(body) > 0 {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("failed to unmarshal JSON response: %w", err)
		}
	}

	return nil
}

// DoStream executes an HTTP request and returns a streaming response
func (c *Client) DoStream(ctx context.Context, req *Request) (*Response, error) {
	return c.Do(ctx, req)
}

// buildHTTPRequest builds an HTTP request from the request object
func (c *Client) buildHTTPRequest(ctx context.Context, req *Request) (*http.Request, error) {
	// Build URL
	u, err := url.Parse(c.config.BaseURL + req.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Add query parameters
	if req.Query != nil {
		u.RawQuery = req.Query.Encode()
	}

	// Prepare request body
	var body io.Reader
	contentType := "application/json"

	if req.Body != nil {
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		body = bytes.NewReader(bodyBytes)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, u.String(), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", contentType)
	httpReq.Header.Set("User-Agent", c.config.UserAgent)
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	// Add custom headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	return httpReq, nil
}

// GetBillingInfo extracts billing information from response headers
func GetBillingInfo(headers http.Header) *BillingInfo {
	return &BillingInfo{
		BillingMode:        headers.Get("X-Aivis-Billing-Mode"),
		CharacterCount:     headers.Get("X-Aivis-Character-Count"),
		CreditsUsed:        headers.Get("X-Aivis-Credits-Used"),
		CreditsRemaining:   headers.Get("X-Aivis-Credits-Remaining"),
		RateLimitRequests:  headers.Get("X-Aivis-RateLimit-Requests-Limit"),
		RateLimitRemaining: headers.Get("X-Aivis-RateLimit-Requests-Remaining"),
		RateLimitReset:     headers.Get("X-Aivis-RateLimit-Requests-Reset"),
		ContentDisposition: headers.Get("Content-Disposition"),
	}
}

// GetFileName extracts filename from Content-Disposition header
func GetFileName(headers http.Header) string {
	contentDisp := headers.Get("Content-Disposition")
	if contentDisp == "" {
		return ""
	}

	// Parse Content-Disposition header to extract filename
	parts := strings.Split(contentDisp, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "filename=") {
			filename := strings.TrimPrefix(part, "filename=")
			filename = strings.Trim(filename, "\"")
			return filename
		}
	}

	return ""
}

// BillingInfo contains billing and rate limit information from API response headers
type BillingInfo struct {
	BillingMode        string `json:"billing_mode,omitempty"`
	CharacterCount     string `json:"character_count,omitempty"`
	CreditsUsed        string `json:"credits_used,omitempty"`
	CreditsRemaining   string `json:"credits_remaining,omitempty"`
	RateLimitRequests  string `json:"rate_limit_requests,omitempty"`
	RateLimitRemaining string `json:"rate_limit_remaining,omitempty"`
	RateLimitReset     string `json:"rate_limit_reset,omitempty"`
	ContentDisposition string `json:"content_disposition,omitempty"`
}

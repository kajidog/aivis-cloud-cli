package config

import (
	"io"
	"os"
	"strings"
	"time"
)

// Config holds the configuration for the Aivis Cloud client
type Config struct {
	// APIKey is the authentication key for Aivis Cloud API
	APIKey string

	// BaseURL is the base URL for Aivis Cloud API
	BaseURL string

	// HTTPTimeout is the timeout for HTTP requests
	HTTPTimeout time.Duration

	// UserAgent is the user agent string to send with requests
	UserAgent string
	
	// DefaultPlaybackMode sets the default playback mode for audio
	DefaultPlaybackMode string
	
	// LogLevel sets the logging level (DEBUG, INFO, WARN, ERROR)
	LogLevel string
	
	// LogOutput sets the log output destination (stdout, stderr, or file path)
	LogOutput string
	
	// LogFormat sets the log output format (text, json)
	LogFormat string
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		BaseURL:             "https://api.aivis-project.com",
		HTTPTimeout:         60 * time.Second,
		UserAgent:           "aiviscloud-go-client/1.0.0",
		DefaultPlaybackMode: "immediate",
		LogLevel:            "INFO",
		LogOutput:           "stdout",
		LogFormat:           "text",
	}
}

// NewConfig creates a new configuration with the provided API key
func NewConfig(apiKey string) *Config {
	config := DefaultConfig()
	config.APIKey = apiKey
	return config
}

// WithBaseURL sets a custom base URL
func (c *Config) WithBaseURL(baseURL string) *Config {
	c.BaseURL = baseURL
	return c
}

// WithTimeout sets a custom HTTP timeout
func (c *Config) WithTimeout(timeout time.Duration) *Config {
	c.HTTPTimeout = timeout
	return c
}

// WithUserAgent sets a custom user agent
func (c *Config) WithUserAgent(userAgent string) *Config {
	c.UserAgent = userAgent
	return c
}

// WithDefaultPlaybackMode sets the default playback mode
func (c *Config) WithDefaultPlaybackMode(mode string) *Config {
	c.DefaultPlaybackMode = mode
	return c
}

// WithLogLevel sets the logging level
func (c *Config) WithLogLevel(level string) *Config {
	c.LogLevel = level
	return c
}

// WithLogOutput sets the log output destination
func (c *Config) WithLogOutput(output string) *Config {
	c.LogOutput = output
	return c
}

// WithLogFormat sets the log output format
func (c *Config) WithLogFormat(format string) *Config {
	c.LogFormat = format
	return c
}

// GetLogWriter returns the appropriate writer for log output
func (c *Config) GetLogWriter() (io.Writer, error) {
	switch strings.ToLower(c.LogOutput) {
	case "stdout":
		return os.Stdout, nil
	case "stderr":
		return os.Stderr, nil
	default:
		// Assume it's a file path
		file, err := os.OpenFile(c.LogOutput, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, err
		}
		return file, nil
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.APIKey == "" {
		return &ValidationError{Field: "APIKey", Message: "API key is required"}
	}
	if c.BaseURL == "" {
		return &ValidationError{Field: "BaseURL", Message: "Base URL is required"}
	}
	if c.HTTPTimeout <= 0 {
		return &ValidationError{Field: "HTTPTimeout", Message: "HTTP timeout must be positive"}
	}
	return nil
}

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

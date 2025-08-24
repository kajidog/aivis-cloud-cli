package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ParseLogLevel parses a string log level
func ParseLogLevel(level string) (LogLevel, error) {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return DEBUG, nil
	case "INFO":
		return INFO, nil
	case "WARN":
		return WARN, nil
	case "ERROR":
		return ERROR, nil
	default:
		return INFO, fmt.Errorf("invalid log level: %s", level)
	}
}

// Logger interface defines the logging contract
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	
	WithFields(fields ...Field) Logger
	SetLevel(level LogLevel)
}

// Field represents a structured logging field
type Field struct {
	Key   string
	Value interface{}
}

// Format represents the output format
type Format string

const (
	TextFormat Format = "text"
	JSONFormat Format = "json"
)

// Config holds logger configuration
type Config struct {
	Level  LogLevel
	Output io.Writer
	Format Format
}

// DefaultConfig returns a default logger configuration
func DefaultConfig() *Config {
	return &Config{
		Level:  INFO,
		Output: os.Stdout,
		Format: TextFormat,
	}
}

// structuredLogger is the default implementation of Logger
type structuredLogger struct {
	config *Config
	fields []Field
}

// New creates a new logger with the given configuration
func New(config *Config) Logger {
	if config == nil {
		config = DefaultConfig()
	}
	return &structuredLogger{
		config: config,
		fields: make([]Field, 0),
	}
}

// NewWithWriter creates a logger with a specific writer
func NewWithWriter(writer io.Writer) Logger {
	config := DefaultConfig()
	config.Output = writer
	return New(config)
}

// Debug logs a debug message
func (l *structuredLogger) Debug(msg string, fields ...Field) {
	l.log(DEBUG, msg, fields...)
}

// Info logs an info message
func (l *structuredLogger) Info(msg string, fields ...Field) {
	l.log(INFO, msg, fields...)
}

// Warn logs a warning message
func (l *structuredLogger) Warn(msg string, fields ...Field) {
	l.log(WARN, msg, fields...)
}

// Error logs an error message
func (l *structuredLogger) Error(msg string, fields ...Field) {
	l.log(ERROR, msg, fields...)
}

// Debugf logs a debug message with formatting
func (l *structuredLogger) Debugf(format string, args ...interface{}) {
	l.log(DEBUG, fmt.Sprintf(format, args...))
}

// Infof logs an info message with formatting
func (l *structuredLogger) Infof(format string, args ...interface{}) {
	l.log(INFO, fmt.Sprintf(format, args...))
}

// Warnf logs a warning message with formatting
func (l *structuredLogger) Warnf(format string, args ...interface{}) {
	l.log(WARN, fmt.Sprintf(format, args...))
}

// Errorf logs an error message with formatting
func (l *structuredLogger) Errorf(format string, args ...interface{}) {
	l.log(ERROR, fmt.Sprintf(format, args...))
}

// WithFields returns a new logger with additional fields
func (l *structuredLogger) WithFields(fields ...Field) Logger {
	newFields := make([]Field, len(l.fields)+len(fields))
	copy(newFields, l.fields)
	copy(newFields[len(l.fields):], fields)
	
	return &structuredLogger{
		config: l.config,
		fields: newFields,
	}
}

// SetLevel sets the logging level
func (l *structuredLogger) SetLevel(level LogLevel) {
	l.config.Level = level
}

// log is the internal logging method
func (l *structuredLogger) log(level LogLevel, msg string, fields ...Field) {
	if level < l.config.Level {
		return
	}
	
	allFields := make([]Field, len(l.fields)+len(fields))
	copy(allFields, l.fields)
	copy(allFields[len(l.fields):], fields)
	
	entry := logEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   msg,
		Fields:    allFields,
	}
	
	var output string
	switch l.config.Format {
	case JSONFormat:
		output = l.formatJSON(entry)
	default:
		output = l.formatText(entry)
	}
	
	fmt.Fprintln(l.config.Output, output)
}

// logEntry represents a single log entry
type logEntry struct {
	Timestamp time.Time
	Level     LogLevel
	Message   string
	Fields    []Field
}

// formatText formats the log entry as text
func (l *structuredLogger) formatText(entry logEntry) string {
	var sb strings.Builder
	
	// Timestamp
	sb.WriteString(entry.Timestamp.Format("2006-01-02 15:04:05"))
	sb.WriteString(" ")
	
	// Level
	sb.WriteString("[")
	sb.WriteString(entry.Level.String())
	sb.WriteString("] ")
	
	// Message
	sb.WriteString(entry.Message)
	
	// Fields
	if len(entry.Fields) > 0 {
		sb.WriteString(" ")
		for i, field := range entry.Fields {
			if i > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(field.Key)
			sb.WriteString("=")
			sb.WriteString(fmt.Sprintf("%v", field.Value))
		}
	}
	
	return sb.String()
}

// formatJSON formats the log entry as JSON
func (l *structuredLogger) formatJSON(entry logEntry) string {
	data := map[string]interface{}{
		"timestamp": entry.Timestamp.Format(time.RFC3339),
		"level":     entry.Level.String(),
		"message":   entry.Message,
	}
	
	// Add fields
	for _, field := range entry.Fields {
		data[field.Key] = field.Value
	}
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		// Fallback to text format if JSON marshaling fails
		return l.formatText(entry)
	}
	
	return string(jsonData)
}

// Helper functions for creating fields
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value.String()}
}

func Error(err error) Field {
	return Field{Key: "error", Value: err.Error()}
}

// NoopLogger is a logger that does nothing
type NoopLogger struct{}

// NewNoop returns a logger that discards all output
func NewNoop() Logger {
	return &NoopLogger{}
}

func (n *NoopLogger) Debug(msg string, fields ...Field)          {}
func (n *NoopLogger) Info(msg string, fields ...Field)           {}
func (n *NoopLogger) Warn(msg string, fields ...Field)           {}
func (n *NoopLogger) Error(msg string, fields ...Field)          {}
func (n *NoopLogger) Debugf(format string, args ...interface{})  {}
func (n *NoopLogger) Infof(format string, args ...interface{})   {}
func (n *NoopLogger) Warnf(format string, args ...interface{})   {}
func (n *NoopLogger) Errorf(format string, args ...interface{})  {}
func (n *NoopLogger) WithFields(fields ...Field) Logger          { return n }
func (n *NoopLogger) SetLevel(level LogLevel)                    {}
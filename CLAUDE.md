# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a monorepo containing an Aivis Cloud API Golang client library and CLI tool. The project implements a clean architecture pattern with feature-based organization rather than traditional layered structure.

### Components

- **Client Library** (`packages/client/`): Go SDK for integrating Aivis Cloud API functionality into applications
- **CLI Tool** (`packages/cli/`): Command-line interface built on top of the client library, including MCP server support
- **NPM Package** (`packages/npm/`): NPM wrapper that distributes pre-built CLI binaries

## Architecture

### Clean Architecture Implementation

The codebase follows clean architecture principles with three main layers organized by feature:

- **Domain Layer** (`*/domain/`): Contains business logic, models, and repository interfaces
- **Usecase Layer** (`*/usecase/`): Implements application-specific business rules and validation
- **Infrastructure Layer** (`*/infrastructure/`): Handles external API communication

### Feature Organization

Instead of grouping by technical layers, code is organized by business features:

- **TTS** (`tts/`): Text-to-speech synthesis and audio playback functionality
- **Models** (`models/`): Model search and discovery functionality
- **Users** (`users/`): User account and profile management
- **Payment** (`payment/`): Billing and payment information
- **Common** (`common/`): Shared utilities (HTTP client, error handling, logging)
- **Config** (`config/`): Configuration management

### CLI Architecture

The CLI (`packages/cli/`) is built using Cobra and provides:

- **Command Structure**: Organized into subcommands (`tts`, `models`, `config`, `users`, `payment`, `mcp`)
- **Configuration Management**: Viper-based config with YAML files and environment variables
- **Global Client**: Shared `aivisClient` instance across all commands
- **MCP Server**: Model Context Protocol server with stdio and HTTP transports

### Key Design Patterns

**Builder Pattern**: Both TTS requests and model search requests use builder pattern for easy construction:

```go
request := client.NewTTSRequest("model-uuid", "text").
    WithSSML(true).
    WithOutputFormat(domain.OutputFormatMP3).
    WithOutputChannels(domain.AudioChannelsStereo).
    WithLeadingSilence(0.1).
    WithTrailingSilence(0.2).
    Build()
```

**Repository Pattern**: Each feature defines repository interfaces in domain layer, implemented in infrastructure layer

**Dependency Injection**: Main client assembles all dependencies and provides a unified API

## Common Commands

### Repository Structure

This is a monorepo with the following structure:
- `packages/client/` - Go client library
- `packages/cli/` - Go CLI application  
- `packages/npm/` - NPM package that wraps the CLI binary

### Building Client Library

```bash
cd packages/client
go mod tidy
go build -v ./...
```

### Building CLI

```bash
cd packages/cli
go mod tidy
go build -o aivis-cli
```

### Cross-platform CLI Build

```bash
cd packages/cli
# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o aivis-cli-darwin-amd64
# macOS (Apple Silicon) 
GOOS=darwin GOARCH=arm64 go build -o aivis-cli-darwin-arm64
# Windows
GOOS=windows GOARCH=amd64 go build -o aivis-cli-windows-amd64.exe
# Linux
GOOS=linux GOARCH=amd64 go build -o aivis-cli-linux-amd64
```

### Testing

```bash
# Test client library
cd packages/client && go test -v
cd packages/client && go test -cover

# Test specific functionality
cd packages/client && go test -v -run TestSearchPublicModels

# Test CLI (includes client library tests)
cd packages/cli && go test -v
```

### Running Examples

```bash
cd packages/client/example
go run main.go
```

### Module Paths

```bash
# Client library module:
github.com/kajidog/aivis-cloud-cli/client

# CLI module:
github.com/kajidog/aivis-cloud-cli/cli
```

## API Integration

### Aivis Cloud API

The client integrates with Aivis Cloud API (`https://api.aivis-project.com`) providing:

- **Text-to-Speech Synthesis** (`/v1/tts/synthesize`): Converts text to audio with streaming support
- **Model Search** (`/v1/aivm-models/search`): Searches available voice models
- **Model Details** (`/v1/aivm-models/{uuid}`): Retrieves specific model information

### Authentication

All API calls require Bearer token authentication via `Authorization` header.

### Error Handling

The HTTP client automatically maps API status codes to structured errors:

- 401: Invalid API key
- 402: Insufficient credits
- 404: Model not found
- 422: Invalid parameters
- 429: Rate limit exceeded
- 5xx: Server errors

### Response Headers

The client extracts billing and rate limit information from custom headers:

- `X-Aivis-Billing-Mode`: Billing mode (PayAsYouGo, Subscription)
- `X-Aivis-Credits-*`: Credit usage and remaining balance
- `X-Aivis-RateLimit-*`: Rate limiting information

## Key Components

### Main Client (`client.go`)

Central facade that combines all functionality. Provides unified API for TTS and model operations while managing configuration and HTTP client lifecycle.

### HTTP Client (`common/http/client.go`)

Handles all HTTP communication with features:

- Automatic error mapping from status codes
- Streaming response support for audio data
- Billing information extraction from headers
- Request/response logging and retry logic

### TTS Domain (`tts/domain/model.go`)

Comprehensive model supporting all TTS parameters from OpenAPI spec:

- Audio formats: WAV, FLAC, MP3, AAC, Opus
- Voice parameters: speaking rate, pitch, volume, emotional intensity
- SSML support for rich text markup
- Streaming response handling

### Configuration (`config/config.go`)

Manages client settings with validation:

- API key (required)
- Base URL (default: api.aivis-project.com)
- HTTP timeout (default: 60s)
- User agent string
- Default playback mode for audio playback

### Audio Playback System (`tts/domain/player.go`, `tts/infrastructure/player.go`, `tts/usecase/player.go`)

Cross-platform audio playback system using OS commands:

- **OS Command Integration**: Uses native audio commands (afplay on macOS, PowerShell Media.SoundPlayer on Windows, aplay/paplay/ffplay on Linux)
- **Progressive Streaming Playback**: Starts playback as soon as first audio chunk arrives (~500ms) using io.Pipe
- **Queue Management**: Three playback modes (immediate, queue, no_queue) - Note: Queue modes only work in persistent services like MCP server
- **Position Tracking**: File-size-based duration estimation with real-time position calculation
- **Process Management**: Uses `exec.Command` with proper process lifecycle management
- **Temporary File Handling**: Creates temporary audio files that are automatically cleaned up

## Important Implementation Details

### Streaming Audio

TTS synthesis supports streaming responses where audio is generated and delivered in chunks as synthesis progresses. The streaming handler interface allows custom processing of audio chunks.

### Request Validation

Each usecase layer validates requests before sending to API:

- TTS: Text length limits, parameter ranges, format compatibility
- Models: Pagination bounds, sort field validation

### Memory Management

Audio responses use `io.ReadCloser` to avoid loading entire audio files into memory. Callers must properly close response streams.

### Builder Pattern Usage

Request builders are the preferred way to construct API requests as they provide type safety and clear documentation of available parameters.

### Audio Playback Architecture

The audio playback system follows the same clean architecture pattern:

- **Domain**: Defines `AudioPlayer` interface, playback modes, and configuration models
- **Infrastructure**: Implements OS-specific audio commands using `OSCommandAudioPlayer`
- **Usecase**: Manages playback queue, coordinates TTS synthesis with audio playback
- **Client Integration**: Provides unified API methods like `PlayText()`, `PlayRequest()`, playback controls

**Playback Modes (MCP Server Only):**

- `PlaybackModeImmediate`: Stop current audio and play new audio immediately (default for CLI)
- `PlaybackModeQueue`: Add to queue and play sequentially after current audio completes (MCP server only)
- `PlaybackModeNoQueue`: Allow simultaneous audio playback without queue management (MCP server only)

**OS Command Detection:**
The system automatically detects available audio players on each platform:

- **macOS**: `afplay` (built-in)
- **Windows**: PowerShell `Media.SoundPlayer` class
- **Linux**: Tries `aplay`, `paplay`, `play`, `ffplay` in order of preference

**Position Tracking:**
Uses file size and format-specific bitrate estimates to calculate audio duration, combined with `time.Since(startTime)` for real-time position tracking. Process completion is detected using `cmd.Wait()`.

## Integration Notes

This client is designed for integration into:

- CLI applications
- MCP (Model Context Protocol) servers
- Desktop applications requiring text-to-speech with audio playback
- Server applications that need to generate and play audio
- Other Golang applications requiring AivisCloud functionality

The unified client interface makes it easy to add both TTS synthesis and audio playback capabilities to any Go application with minimal configuration.

## Audio Playback API Usage

### Basic Audio Playback

```go
// Direct text-to-audio playback
err := client.PlayText(ctx, "こんにちは", "model-uuid")

// With custom playback options
request := client.NewTTSRequest("model-uuid", "text").Build()
options := client.NewPlaybackRequest(request).
    WithMode(domain.PlaybackModeQueue).
    WithVolume(0.8).
    Build()
err := client.PlayRequest(ctx, options)
```

### Playback Control

```go
// Control playback
client.PausePlayback()    // Note: Limited support with OS commands
client.ResumePlayback()
client.StopPlayback()     // Terminates current audio process
client.ClearPlaybackQueue()

// Get playback status with position tracking
status := client.GetPlaybackStatus()
fmt.Printf("Playing: %s, Position: %.1fs/%.1fs\n",
    status.CurrentText,
    status.Position.Seconds(),
    status.Duration.Seconds())
```

### Configuration

```go
// Set default playback behavior
config := config.NewConfig(apiKey).
    WithDefaultPlaybackMode("queue")  // "immediate", "queue", "no_queue"
client := client.NewWithConfig(config)
```

### Logging Configuration

The client library includes a comprehensive structured logging system:

```go
// Configure logging
config := config.NewConfig(apiKey).
    WithLogLevel("DEBUG").           // DEBUG, INFO, WARN, ERROR
    WithLogOutput("/var/log/app.log"). // stdout, stderr, or file path
    WithLogFormat("json")            // text or json

client := client.NewWithConfig(config)

// Access logger directly
logger := client.GetLogger()
logger.Info("Custom message",
    logger.String("key", "value"),
    logger.Int("count", 42),
)
```

**Logging Features:**

- **Structured Logging**: Key-value pairs with type safety
- **Multiple Formats**: Human-readable text or machine-parseable JSON
- **Configurable Output**: Console (stdout/stderr) or file output
- **Log Levels**: DEBUG (detailed), INFO (general), WARN, ERROR
- **Performance**: NoOp logger for production when logging is disabled

## CLI Tool and MCP Support

### Aivis Cloud CLI

The CLI provides command-line access to AivisCloud functionality:

```bash
# Basic usage
export AIVIS_API_KEY=your_api_key
./aivis-cli models search --limit 5
./aivis-cli tts synthesize "こんにちは" model-uuid output.wav

# MCP server (stdio default)
./aivis-cli mcp

# MCP server with HTTP (for remote access)
./aivis-cli mcp --transport http --port 8080

# With logging configuration
./aivis-cli --log-level=DEBUG --log-format=json models search "voice" --limit 3
./aivis-cli -v --log-output=/tmp/aivis.log tts synthesize "テスト" model-uuid test.wav

# Configuration file logging setup (~/.aivis-cli.yaml)
# log_level: "DEBUG"  
# log_output: "/var/log/aivis.log"
# log_format: "json"
```

### MCP (Model Context Protocol) Server

The CLI includes an MCP server that provides AI assistants access to AivisCloud voice models:

#### Starting MCP Server

```bash
# Stdio mode (default and recommended)
export AIVIS_API_KEY=your_api_key
./aivis-cli mcp

# HTTP mode (for testing or remote access)
./aivis-cli mcp --transport http --port 8080

# Custom port for HTTP
./aivis-cli mcp --transport http --port 3000
```

#### Claude Desktop Integration

Configure Claude Desktop (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "aivis-cloud-api": {
      "command": "/path/to/aivis-cli",
      "args": ["mcp"],
      "env": {
        "AIVIS_API_KEY": "your_api_key_here"
      }
    }
  }
}
```

For HTTP mode (if needed):

```json
{
  "mcpServers": {
    "aivis-cloud-api": {
      "command": "npx",
      "args": ["-y", "mcp-remote", "http://localhost:8080"]
    }
  }
}
```

#### Available MCP Tools

- **search_models**: Search voice models (returns compact summary, default 5 results)

  - Parameters: `query`, `author`, `tags`, `limit`, `sort`, `public_only`
  - Token-optimized: Shows essential info only (name, UUID, author, downloads)

- **get_model**: Get essential model information (minimal token usage)

  - Parameters: `uuid` (required)
  - Compact format: Basic details, brief description (<100 chars), speaker count

- **get_model_speakers**: Get speaker list (compact format with language and style counts)
  - Parameters: `uuid` (required)
  - Shows speaker names, languages, and style counts (not individual styles)

#### MCP Implementation Details

- **Architecture**: Split into `mcp.go` (server/transport) and `mcp_tools.go` (tool implementations)
- **API Integration**: Uses existing `aivisClient` for all operations
- **Authentication**: Leverages existing API key management (environment variables, config files)
- **Token Efficiency**: Optimized output format to minimize LLM token consumption
- **Transport Support**: Both stdio (default) and HTTP transports supported

#### MCP Tool Examples

```bash
# Search models (returns 5 results by default)
search_models({"query": "female voice", "limit": 3})

# Get specific model info
get_model({"uuid": "a59cb814-0083-4369-8542-f51a29e72af7"})

# Get model speakers
get_model_speakers({"uuid": "a59cb814-0083-4369-8542-f51a29e72af7"})
```

## Development Workflow

### Making Changes

When working in this codebase:

1. **Client Library Changes**: Make changes in `packages/client/`, run tests with `go test -v`
2. **CLI Changes**: Make changes in `packages/cli/`, test by building and running commands
3. **MCP Changes**: Update MCP tools in `mcp_tools_*.go`, test with `./aivis-cli mcp` (stdio default) or `./aivis-cli mcp --transport http --port 8080` for HTTP testing 
4. **Cross-package Changes**: Remember that CLI depends on client library via `replace` directive in `go.mod`

### Important Files

- `packages/cli/main.go`: CLI entry point with Cobra command setup
- `packages/cli/mcp_server.go`: MCP server implementation with stdio/HTTP transports  
- `packages/cli/mcp_tools_*.go`: Individual MCP tool implementations
- `packages/client/client.go`: Main client facade combining all functionality
- `packages/client/config/config.go`: Configuration management with validation

### Go Version Requirements

- Client library: Go 1.21+
- CLI: Go 1.23+ (required for MCP SDK)

### Testing Strategy

The client library includes comprehensive tests with mock HTTP servers. CLI testing is primarily integration-based through the built binary.

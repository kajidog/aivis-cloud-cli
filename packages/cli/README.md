# AivisCloud CLI

Command-line interface for AivisCloud API, providing text-to-speech synthesis, audio playback, and model management capabilities.

## Installation

```bash
cd packages/cli
go build -o aivis-cli
```

## Configuration

### Initialize Configuration
```bash
./aivis-cli config init
```

### Set API Key
```bash
./aivis-cli config set api_key YOUR_API_KEY
```

### Configuration Options
- `api_key`: Your AivisCloud API key (required)
- `base_url`: API base URL (default: https://api.aivis-project.com)
- `timeout`: Request timeout (default: 60s)
- `default_playback_mode`: Default audio playback mode (immediate, queue, no_queue)

## Usage

### Text-to-Speech

#### Play Text
```bash
# Basic usage
./aivis-cli tts play "こんにちは" model-uuid

# With options
./aivis-cli tts play "Hello world" model-uuid --volume 0.8 --rate 1.2 --mode queue
```

#### Synthesize to File
```bash
# Basic synthesis
./aivis-cli tts synthesize "Hello" model-uuid output.wav

# With custom format and options
./aivis-cli tts synthesize "Hello" model-uuid output.mp3 --format mp3 --volume 0.9
```

#### Stream Synthesis
```bash
./aivis-cli tts stream "Hello world" model-uuid > output.wav
```

#### Playback Control
```bash
# Control playback
./aivis-cli tts control stop
./aivis-cli tts control pause
./aivis-cli tts control resume
./aivis-cli tts control status
./aivis-cli tts control clear

# Set volume
./aivis-cli tts volume 0.7
```

### Model Management

#### Search Models
```bash
# Basic search
./aivis-cli models search "voice"

# Search with filters
./aivis-cli models search --author "author-name" --tags "tag1,tag2" --limit 5

# Public models only
./aivis-cli models search "japanese" --public
```

#### Get Model Details
```bash
# Model information
./aivis-cli models get model-uuid

# Model speakers
./aivis-cli models get model-uuid --speakers
```

#### List Models
```bash
# Popular models
./aivis-cli models popular --limit 10

# Recent models
./aivis-cli models recent --limit 5

# Top-rated models
./aivis-cli models top-rated
```

### Configuration Management

```bash
# Show current config
./aivis-cli config show

# Set/unset values
./aivis-cli config set key value
./aivis-cli config unset key

# Validate configuration
./aivis-cli config validate
```

## Output Formats

Most commands support multiple output formats:
- `--output table` (default): Human-readable table format
- `--output json`: Machine-readable JSON format

## TTS Parameters

Available TTS parameters:
- `--volume`: Audio volume (0.0 to 1.0)
- `--rate`: Speaking rate multiplier
- `--pitch`: Pitch adjustment
- `--ssml`: Enable SSML markup processing
- `--format`: Audio format (wav, flac, mp3, aac, opus)

## Playback Modes

- `immediate`: Stop current audio and play immediately
- `queue`: Add to queue, play after current audio
- `no_queue`: Allow simultaneous playback

## Environment Variables

- `AIVIS_API_KEY`: API key
- `AIVIS_BASE_URL`: API base URL
- `AIVIS_TIMEOUT`: Request timeout

## Examples

```bash
# Complete workflow
./aivis-cli config init
./aivis-cli config set api_key your-key-here
./aivis-cli models search "japanese" --limit 3
./aivis-cli tts play "こんにちは、世界！" selected-model-uuid --mode queue

# Batch synthesis
./aivis-cli tts synthesize "Hello" model-uuid hello.wav --format wav
./aivis-cli tts synthesize "Goodbye" model-uuid goodbye.mp3 --format mp3

# Model exploration
./aivis-cli models popular --output json | jq '.data[].name'
./aivis-cli models get model-uuid --speakers --output table
```
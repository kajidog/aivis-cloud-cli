# @kajidog/aivis-cloud-cli

## 0.5.0

### Minor Changes

- 3ff1502: feat: add TTS history management (resume) functionality

  Add comprehensive TTS synthesis history management with sequential ID system:

  - Auto-save TTS synthesis history with all request parameters
  - Resume functionality to replay past synthesis results
  - History management commands: list, show, play, delete, clean, stats
  - Configurable history settings (enable/disable, max count, storage path)
  - Sequential ID system (1, 2, 3...) for user-friendly CLI operations
  - Automatic filename generation with timestamp format
  - Full parameter preservation including volume, pitch, format, channels, etc.

  This enables users to easily manage and replay their TTS synthesis history.

### Patch Changes

- 3ff1502: feat(cli): add `config keys` command and strict schema validation for `config set/validate`; improve error messages.

  docs(npm): fold advanced sections, clarify API error codes are server-side, and add FFplay PATH refresh notes under a collapsible section.

- 3ff1502: docs: add FFplay installation guide and clarify playback policy;
  fix: stabilize MCP playback on Windows (stdin streaming when ffplay is available, non‑progressive fallback otherwise); ensure history ID availability with short wait.

## 0.4.1

### Patch Changes

- 74fbf3a: Improve MCP documentation and default behavior:
  - Add streaming audio synthesis documentation for MCP tools
  - Fix MCP default playback mode to 'immediate' when not configured
  - Update README with streaming synthesis details for synthesize_speech tool

## 0.4.0

### Minor Changes

- bf68788: Add MCP configuration management tools and enhance search_models output

  **New MCP Tools:**

  - `get_mcp_settings`: Get current MCP configuration (API key excluded for security)
  - `update_mcp_settings`: Update MCP configuration with validation and security restrictions

  **Enhanced MCP Tools:**

  - `search_models`: Now includes supported languages, speaker count, and creation date

  **Security Features:**

  - API key protection (cannot be read or modified via MCP)
  - System setting restrictions (logging and simplified TTS settings are read-only)
  - Parameter validation for all configuration updates

## 0.3.0

### Minor Changes

- Update MCP Go SDK to v0.3.0

  - Updated github.com/modelcontextprotocol/go-sdk from v0.2.0 to v0.3.0
  - Updated all MCP tool handler signatures to match new SDK API
  - Tool handlers now use simplified signature: `func(ctx context.Context, req *mcp.CallToolRequest, args T) (*mcp.CallToolResult, any, error)`
  - Removed usage of deprecated `mcp.CallToolParamsFor[T]` and `mcp.CallToolResultFor[T]` types

  This update brings compatibility with the latest MCP Go SDK and includes performance improvements and bug fixes.

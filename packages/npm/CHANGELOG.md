# @kajidog/aivis-cloud-cli

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

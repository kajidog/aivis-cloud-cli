# @kajidog/aivis-cloud-cli

## 0.3.0

### Minor Changes

- Update MCP Go SDK to v0.3.0

  - Updated github.com/modelcontextprotocol/go-sdk from v0.2.0 to v0.3.0
  - Updated all MCP tool handler signatures to match new SDK API
  - Tool handlers now use simplified signature: `func(ctx context.Context, req *mcp.CallToolRequest, args T) (*mcp.CallToolResult, any, error)`
  - Removed usage of deprecated `mcp.CallToolParamsFor[T]` and `mcp.CallToolResultFor[T]` types

  This update brings compatibility with the latest MCP Go SDK and includes performance improvements and bug fixes.

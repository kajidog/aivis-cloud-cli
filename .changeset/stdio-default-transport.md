---
"@kajidog/aivis-cloud-cli": minor
---

Update MCP server to use stdio as default transport

- Change default MCP transport from HTTP to stdio for better Claude Desktop integration
- Update documentation in CLAUDE.md and packages/npm/README.md to reflect stdio as default
- Maintain full backward compatibility with HTTP transport via --transport http flag
- Improve user experience by eliminating need for port configuration in most cases
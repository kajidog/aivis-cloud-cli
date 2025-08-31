---
"@kajidog/aivis-cloud-cli": minor
---

Add MCP configuration management tools and enhance search_models output

**New MCP Tools:**
- `get_mcp_settings`: Get current MCP configuration (API key excluded for security)
- `update_mcp_settings`: Update MCP configuration with validation and security restrictions

**Enhanced MCP Tools:**
- `search_models`: Now includes supported languages, speaker count, and creation date

**Security Features:**
- API key protection (cannot be read or modified via MCP)
- System setting restrictions (logging and simplified TTS settings are read-only)
- Parameter validation for all configuration updates
---
"@kajidog/aivis-cloud-cli": patch
---

fix(mcp): improve history playback reliability and path handling.

- Enforce `no_queue` mode for MCP `play_tts_history`, and let `wait_for_end` follow the request argument.
- Expand `history_store_path` (`~` and env vars) in MCP tools and client config; ensure absolute path usage.
- Reduce zero-byte history records by waiting until the history file has non-zero size before saving metadata.
- Propagate history settings (`history_enabled`, `history_max_count`, `history_store_path`) from config into the client.

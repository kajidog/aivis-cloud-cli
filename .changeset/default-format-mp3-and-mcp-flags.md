---
"@kajidog/aivis-cloud-cli": minor
---

feat: change default audio format to mp3 and expose authoritative playback characteristics in MCP responses.

- Default format: switch from wav to mp3 across CLI/client defaults and examples.
- MCP (tts/history): include playback mode and streaming flags using client‑computed values (no guessing in MCP layer).
- Client: add `streaming_synthesis`, `streaming_playback`, and `effective_mode` to TTS response; expose `DetectStreamingPlayback()` and compute values based on OS/player/format.
- Docs: update configuration table and examples to reflect mp3 as default.

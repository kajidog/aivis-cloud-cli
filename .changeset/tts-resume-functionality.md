---
"@kajidog/aivis-cloud-cli": minor
---

feat: add TTS history management (resume) functionality

Add comprehensive TTS synthesis history management with sequential ID system:

- Auto-save TTS synthesis history with all request parameters
- Resume functionality to replay past synthesis results  
- History management commands: list, show, play, delete, clean, stats
- Configurable history settings (enable/disable, max count, storage path)
- Sequential ID system (1, 2, 3...) for user-friendly CLI operations
- Automatic filename generation with timestamp format
- Full parameter preservation including volume, pitch, format, channels, etc.

This enables users to easily manage and replay their TTS synthesis history.
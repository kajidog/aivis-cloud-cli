---
"@kajidog/aivis-cloud-cli": patch
---

Fix Go dependency resolution in npm package build script

- Add go mod download step before building binaries
- Include fallback to go mod tidy if download fails
- Resolve missing go.sum entries that caused build failures in CI
- Ensure all Go dependencies are available before cross-compilation
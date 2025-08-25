---
"@kajidog/aivis-cloud-cli": patch
---

Fix CLI build test in GitHub Actions

- Explicitly specify output binary name (-o aivis-cli) in go build command
- Add binary verification step to debug build output
- Ensure CLI binary is created with correct name for testing
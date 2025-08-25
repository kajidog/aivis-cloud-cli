---
"@kajidog/aivis-cloud-cli": patch
---

Update GitHub Actions to use supported versions

- Update actions/upload-artifact from v3 to v4 (v3 deprecated)
- Update actions/cache from v3 to v4 for better performance
- Remove npm cache configuration that can cause lockfile issues in CI
- Ensure all GitHub Actions use currently supported versions
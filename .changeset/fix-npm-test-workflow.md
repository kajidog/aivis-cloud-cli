---
"@kajidog/aivis-cloud-cli": patch
---

Fix npm package test workflow and avoid registry access issues

- Simplify npm test to focus on binary build verification only
- Remove npm dependency installation that caused 403 registry errors
- Test cross-platform binary compilation directly using build script
- Add comprehensive binary validation with size checking
- Resolve CI failures related to npm registry access permissions
---
"@kajidog/aivis-cloud-cli": patch
---

Fix npm version error in release workflow

- Remove package-lock.json in CI to avoid "Invalid Version" errors  
- Use fresh npm install for changesets dependencies in release workflow
- Apply same CI fix from test workflow to release workflow
- Ensure reliable release automation without npm registry conflicts
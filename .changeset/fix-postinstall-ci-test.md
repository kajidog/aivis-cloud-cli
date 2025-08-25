---
"@kajidog/aivis-cloud-cli": patch
---

Fix npm postinstall script execution in CI environment

- Skip postinstall script during npm install using --ignore-scripts flag
- Execute postinstall validation after binaries are built
- Prevent postinstall failure when binaries don't exist yet during dependency installation
- Add proper test sequence: install dependencies → build binaries → validate postinstall
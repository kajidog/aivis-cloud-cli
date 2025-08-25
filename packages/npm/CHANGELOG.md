# @kajidog/aivis-cloud-cli

## 0.2.0

### Minor Changes

- b61eb08: Add automated release management with changesets

  - Setup changesets for version management and automated releases
  - Configure GitHub Actions workflow for automatic npm publishing
  - Add changeset scripts to package.json for easy version management

### Patch Changes

- 2b7af71: Improve project documentation and CI/CD pipeline

  - Add automated testing workflow for pull requests
  - Simplify root README with better navigation to detailed documentation
  - Restructure documentation to focus on user-facing features in main README
  - Move technical details to individual package READMEs for better organization

- 2b7af71: Fix CLI build test in GitHub Actions

  - Explicitly specify output binary name (-o aivis-cli) in go build command
  - Add binary verification step to debug build output
  - Ensure CLI binary is created with correct name for testing

- 2b7af71: Update GitHub Actions to use supported versions

  - Update actions/upload-artifact from v3 to v4 (v3 deprecated)
  - Update actions/cache from v3 to v4 for better performance
  - Remove npm cache configuration that can cause lockfile issues in CI
  - Ensure all GitHub Actions use currently supported versions

- 2b7af71: Fix Go dependency resolution in npm package build script

  - Add go mod download step before building binaries
  - Include fallback to go mod tidy if download fails
  - Resolve missing go.sum entries that caused build failures in CI
  - Ensure all Go dependencies are available before cross-compilation

- 2b7af71: Fix npm package test workflow and avoid registry access issues

  - Simplify npm test to focus on binary build verification only
  - Remove npm dependency installation that caused 403 registry errors
  - Test cross-platform binary compilation directly using build script
  - Add comprehensive binary validation with size checking
  - Resolve CI failures related to npm registry access permissions

- 2b7af71: Fix npm postinstall script execution in CI environment

  - Skip postinstall script during npm install using --ignore-scripts flag
  - Execute postinstall validation after binaries are built
  - Prevent postinstall failure when binaries don't exist yet during dependency installation
  - Add proper test sequence: install dependencies → build binaries → validate postinstall

- cedab33: Fix npm version error in release workflow

  - Remove package-lock.json in CI to avoid "Invalid Version" errors
  - Use fresh npm install for changesets dependencies in release workflow
  - Apply same CI fix from test workflow to release workflow
  - Ensure reliable release automation without npm registry conflicts

- 2b7af71: Improve client library test coverage and documentation

  - Add comprehensive table-driven tests for better organization and maintainability
  - Include error handling tests for HTTP status codes (401, 402, 429, 500)
  - Add TTS request builder pattern validation tests
  - Fix type safety issues in test assertions for pointer fields
  - Update README with detailed test documentation and examples
  - Improve overall code quality and test coverage

- 2b7af71: Optimize CI/CD performance and reduce unnecessary testing

  - Replace slow cross-platform build test with fast single-platform validation
  - Simplify release workflow by removing complex retry logic
  - Focus PR tests on essential functionality: Go client tests + CLI build verification
  - Reduce CI execution time while maintaining code quality assurance

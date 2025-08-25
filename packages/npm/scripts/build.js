#!/usr/bin/env node

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const platforms = [
  { os: 'linux', arch: 'amd64' },
  { os: 'linux', arch: 'arm64' },
  { os: 'darwin', arch: 'amd64' },
  { os: 'darwin', arch: 'arm64' },
  { os: 'windows', arch: 'amd64' },
  { os: 'windows', arch: 'arm64' }
];

const cliDir = path.resolve(__dirname, '../../cli');
const binariesDir = path.resolve(__dirname, '../binaries');

function ensureDirectoryExists(dir) {
  if (!fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true });
  }
}

function buildBinary(platform) {
  const { os, arch } = platform;
  const binaryName = os === 'windows' ? 'aivis-cli.exe' : 'aivis-cli';
  const outputDir = path.join(binariesDir, `${os}-${arch}`);
  const outputPath = path.join(outputDir, binaryName);

  ensureDirectoryExists(outputDir);

  console.log(`Building for ${os}/${arch}...`);

  try {
    // Set environment variables for cross-compilation
    const env = {
      ...process.env,
      GOOS: os,
      GOARCH: arch,
      CGO_ENABLED: '0'
    };

    // Build the Go binary
    execSync(`go build -o "${outputPath}" .`, {
      cwd: cliDir,
      env: env,
      stdio: 'inherit'
    });

    console.log(`✓ Built ${os}/${arch} binary: ${outputPath}`);
  } catch (error) {
    console.error(`✗ Failed to build ${os}/${arch} binary:`, error.message);
    process.exit(1);
  }
}

function main() {
  console.log('Building AivisCloud CLI binaries for all platforms...');
  
  // Ensure the CLI directory exists and has a Go module
  if (!fs.existsSync(path.join(cliDir, 'go.mod'))) {
    console.error('Error: CLI directory does not contain a Go module');
    console.error('Expected go.mod at:', path.join(cliDir, 'go.mod'));
    process.exit(1);
  }

  // Resolve dependencies once at the start
  console.log('Resolving Go dependencies...');
  try {
    // First, ensure go.sum is up to date
    execSync('go mod tidy', {
      cwd: cliDir,
      stdio: 'inherit'
    });
    console.log('✓ go mod tidy completed');
    
    // Then download all dependencies
    execSync('go mod download', {
      cwd: cliDir,
      stdio: 'inherit'
    });
    console.log('✓ Dependencies downloaded successfully');
  } catch (error) {
    console.error('✗ Failed to resolve Go dependencies:', error.message);
    console.error('This might be due to network issues or missing dependencies.');
    console.error('Please ensure you have internet access and all Go modules are accessible.');
    process.exit(1);
  }

  // Clean existing binaries directory
  if (fs.existsSync(binariesDir)) {
    console.log('Cleaning existing binaries...');
    fs.rmSync(binariesDir, { recursive: true, force: true });
  }

  // Build for each platform
  platforms.forEach(buildBinary);

  console.log('\n✓ All binaries built successfully!');
  console.log('\nBinaries location:', binariesDir);
  
  // List built binaries
  console.log('\nBuilt binaries:');
  platforms.forEach(({ os, arch }) => {
    const binaryName = os === 'windows' ? 'aivis-cli.exe' : 'aivis-cli';
    const binaryPath = path.join(binariesDir, `${os}-${arch}`, binaryName);
    if (fs.existsSync(binaryPath)) {
      const stats = fs.statSync(binaryPath);
      console.log(`  ${os}/${arch}: ${binaryPath} (${(stats.size / 1024 / 1024).toFixed(2)} MB)`);
    }
  });
}

if (require.main === module) {
  main();
}
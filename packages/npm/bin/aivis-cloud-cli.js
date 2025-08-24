#!/usr/bin/env node

const { spawn } = require('child_process');
const path = require('path');
const fs = require('fs');

function getPlatform() {
  const platform = process.platform;
  const arch = process.arch;

  let osName, archName;

  // Map Node.js platform names to Go platform names
  switch (platform) {
    case 'darwin':
      osName = 'darwin';
      break;
    case 'linux':
      osName = 'linux';
      break;
    case 'win32':
      osName = 'windows';
      break;
    default:
      throw new Error(`Unsupported platform: ${platform}`);
  }

  // Map Node.js arch names to Go arch names
  switch (arch) {
    case 'x64':
      archName = 'amd64';
      break;
    case 'arm64':
      archName = 'arm64';
      break;
    default:
      throw new Error(`Unsupported architecture: ${arch}`);
  }

  return { os: osName, arch: archName };
}

function getBinaryPath() {
  const { os, arch } = getPlatform();
  const binaryName = os === 'windows' ? 'aivis-cli.exe' : 'aivis-cli';
  const binaryDir = path.join(__dirname, '..', 'binaries', `${os}-${arch}`);
  const binaryPath = path.join(binaryDir, binaryName);

  if (!fs.existsSync(binaryPath)) {
    console.error(`Binary not found for ${os}-${arch}: ${binaryPath}`);
    console.error('Please ensure the binary is properly installed.');
    process.exit(1);
  }

  return binaryPath;
}

function runBinary() {
  try {
    const binaryPath = getBinaryPath();
    const args = process.argv.slice(2);

    // Spawn the binary with the same arguments
    const child = spawn(binaryPath, args, {
      stdio: 'inherit',
      env: process.env
    });

    // Handle child process events
    child.on('error', (error) => {
      console.error('Failed to start aivis-cli:', error.message);
      process.exit(1);
    });

    child.on('exit', (code, signal) => {
      if (signal) {
        console.error(`aivis-cli was killed by signal ${signal}`);
        process.exit(1);
      } else {
        process.exit(code || 0);
      }
    });

    // Handle process termination
    process.on('SIGINT', () => {
      child.kill('SIGINT');
    });

    process.on('SIGTERM', () => {
      child.kill('SIGTERM');
    });

  } catch (error) {
    console.error('Error running aivis-cli:', error.message);
    process.exit(1);
  }
}

// Main execution
if (require.main === module) {
  runBinary();
}

module.exports = { getPlatform, getBinaryPath };
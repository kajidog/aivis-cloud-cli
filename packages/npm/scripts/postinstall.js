#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

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

function validateBinaryInstallation() {
  try {
    const { os, arch } = getPlatform();
    const binaryName = os === 'windows' ? 'aivis-cli.exe' : 'aivis-cli';
    const binaryDir = path.join(__dirname, '..', 'binaries', `${os}-${arch}`);
    const binaryPath = path.join(binaryDir, binaryName);

    console.log(`Checking for aivis-cli binary for ${os}/${arch}...`);

    if (!fs.existsSync(binaryPath)) {
      console.error(`❌ Binary not found: ${binaryPath}`);
      console.error('The binary for your platform was not included in this package.');
      console.error('Please report this issue at: https://github.com/kajidog/aivis-cloud-cli/issues');
      process.exit(1);
    }

    // Check if binary is executable (Unix-like systems)
    if (os !== 'windows') {
      try {
        fs.accessSync(binaryPath, fs.constants.X_OK);
      } catch (error) {
        console.log(`Setting executable permissions for ${binaryPath}...`);
        fs.chmodSync(binaryPath, 0o755);
      }
    }

    // Get binary file size for validation
    const stats = fs.statSync(binaryPath);
    if (stats.size === 0) {
      console.error(`❌ Binary file is empty: ${binaryPath}`);
      process.exit(1);
    }

    console.log(`✅ Binary verified: ${binaryPath} (${(stats.size / 1024 / 1024).toFixed(2)} MB)`);
    console.log('');
    console.log('🎉 AivisCloud CLI has been successfully installed!');
    console.log('');
    console.log('Usage:');
    console.log('  npx @kajidog/aivis-cloud-cli --help');
    console.log('  npx @kajidog/aivis-cloud-cli tts synthesize --text "Hello, world!" --model-uuid <model-id>');
    console.log('');
    console.log('For more information, visit: https://github.com/kajidog/aivis-cloud-cli');

  } catch (error) {
    console.error('❌ Post-install validation failed:', error.message);
    console.error('');
    console.error('Platform:', process.platform);
    console.error('Architecture:', process.arch);
    console.error('');
    console.error('Please report this issue at: https://github.com/kajidog/aivis-cloud-cli/issues');
    process.exit(1);
  }
}

function main() {
  console.log('Running post-install validation...');
  validateBinaryInstallation();
}

if (require.main === module) {
  main();
}

module.exports = { getPlatform, validateBinaryInstallation };
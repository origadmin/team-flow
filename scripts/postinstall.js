#!/usr/bin/env node

/**
 * postinstall script for team-flow npm package
 * 
 * This script runs after `npm install` or `npx skills add` completes.
 * It ensures the flow CLI binary is available in the bin/ directory.
 */

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

const packageRoot = path.join(__dirname, '..');
const binDir = path.join(packageRoot, 'bin');
const flowSource = path.join(packageRoot, 'cmd', 'flow');

// Ensure bin directory exists
if (!fs.existsSync(binDir)) {
  fs.mkdirSync(binDir, { recursive: true });
}

// Check if flow binary already exists
const isWindows = process.platform === 'win32';
const flowBinary = path.join(binDir, isWindows ? 'flow.exe' : 'flow');

if (fs.existsSync(flowBinary)) {
  console.log('✅ flow CLI already exists in bin/');
  process.exit(0);
}

// Try to build flow CLI from source
console.log('🔨 Building flow CLI from source...');

try {
  // Check if Go is available
  execSync('go version', { stdio: 'ignore' });
  
  const buildCmd = isWindows 
    ? 'go build -o flow.exe ./cmd/flow/'
    : 'go build -o flow ./cmd/flow/';
  
  execSync(buildCmd, { 
    cwd: packageRoot,
    stdio: 'inherit'
  });
  
  // Move to bin/
  const builtBinary = path.join(packageRoot, isWindows ? 'flow.exe' : 'flow');
  if (fs.existsSync(builtBinary)) {
    fs.renameSync(builtBinary, flowBinary);
    console.log('✅ flow CLI built and installed to bin/');
  }
} catch (error) {
  console.warn('⚠️  Go not available or build failed.');
  console.warn('📦 flow CLI will be downloaded on first use.');
  console.warn('   To pre-build: npm run build:flow');
}

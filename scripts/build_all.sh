#!/usr/bin/env bash
set -e

# ==============================================================================
# Aldea Cross-Platform Single-Binary Build & Release Automation (Phase 4)
# ==============================================================================

DIST_DIR="dist"
mkdir -p "${DIST_DIR}"

echo "[+] Compiling Aldea Single-Binary CLI and Headless Daemons..."

# 1. Linux AMD64 Release Binaries
echo "[+] Building Linux AMD64 binaries..."
GOOS=linux GOARCH=amd64 go build -o "${DIST_DIR}/aldea-linux-amd64" ./cmd/aldea
GOOS=linux GOARCH=amd64 go build -o "${DIST_DIR}/noded-linux-amd64" ./cmd/noded
GOOS=linux GOARCH=amd64 go build -o "${DIST_DIR}/trackerd-linux-amd64" ./cmd/trackerd

# 2. Windows AMD64 Release Binaries
echo "[+] Building Windows AMD64 binaries..."
GOOS=windows GOARCH=amd64 go build -o "${DIST_DIR}/aldea-windows-amd64.exe" ./cmd/aldea
GOOS=windows GOARCH=amd64 go build -o "${DIST_DIR}/noded-windows-amd64.exe" ./cmd/noded
GOOS=windows GOARCH=amd64 go build -o "${DIST_DIR}/trackerd-windows-amd64.exe" ./cmd/trackerd

# 3. macOS ARM64 Release Binaries
echo "[+] Building macOS ARM64 binaries..."
GOOS=darwin GOARCH=arm64 go build -o "${DIST_DIR}/aldea-darwin-arm64" ./cmd/aldea
GOOS=darwin GOARCH=arm64 go build -o "${DIST_DIR}/noded-darwin-arm64" ./cmd/noded
GOOS=darwin GOARCH=arm64 go build -o "${DIST_DIR}/trackerd-darwin-arm64" ./cmd/trackerd

echo "[✓] Single-Binary Compilation Complete. Binaries saved in '${DIST_DIR}/':"
ls -lh "${DIST_DIR}/"

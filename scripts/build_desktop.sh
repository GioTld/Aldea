#!/usr/bin/env bash
set -e

# ==============================================================================
# Aldea Desktop App Cross-Platform Packaging Script (Phase 3 Final Item)
# ==============================================================================

OUTPUT_DIR="build/bin"
mkdir -p "${OUTPUT_DIR}"

echo "[+] Building Aldea Phase 3 Desktop Application..."

# 1. Native Build (Linux AMD64)
echo "[+] Compiling Native Linux Desktop Binary..."
go build -tags "desktop,production" -o "${OUTPUT_DIR}/aldea-desktop-linux-amd64" ./gui

# 2. Windows AMD64 Target Binary
echo "[+] Compiling Windows Target Binary..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -tags "desktop,production" -o "${OUTPUT_DIR}/aldea-desktop-windows-amd64.exe" ./gui

# 3. macOS ARM64 Target Binary (Requires macOS SDK when using CGO)
echo "[+] Compiling macOS Target Binary..."
GOOS=darwin GOARCH=arm64 go build -tags "desktop,production" -o "${OUTPUT_DIR}/aldea-desktop-darwin-arm64" ./gui || echo "[!] macOS cross-compilation skipped (requires macOS SDK toolchain)"

echo "[✓] Build Complete. Executables generated in '${OUTPUT_DIR}/':"
ls -lh "${OUTPUT_DIR}/"

#!/usr/bin/env bash
set -e

echo "=== Aldea P2P Storage Network Demo ==="
echo "1. Building binaries..."
go build -o bin/aldea ./cmd/aldea
go build -o bin/noded ./cmd/noded
go build -o bin/trackerd ./cmd/trackerd

echo "2. Starting Docker Compose topology (1 tracker + 8 nodes)..."
docker compose up -d --build

echo "3. Waiting for network nodes to register with tracker..."
sleep 3

echo "4. Initializing local client configuration..."
./bin/aldea init --tracker "http://localhost:8080" --key "aldea-p2p-mesh-secret-key-32b!"

echo "5. Creating sample file for upload..."
echo "Hello from Aldea P2P storage pool! Data resilience test payload." > /tmp/aldea_demo_sample.txt

echo "6. Uploading file using 'aldea put'..."
UPLOAD_OUTPUT=$(./bin/aldea put /tmp/aldea_demo_sample.txt)
echo "$UPLOAD_OUTPUT"

FILE_ID=$(echo "$UPLOAD_OUTPUT" | grep "uploaded:" | awk -F ' → ' '{print $2}')
echo "Generated File ID: $FILE_ID"

echo "7. Checking initial network status..."
./bin/aldea status

echo "8. Simulating node failure: stopping node3..."
docker compose stop node3

echo "9. Downloading file despite node failure..."
./bin/aldea get "$FILE_ID" /tmp/aldea_demo_recovered.txt

echo "10. Verifying file content integrity..."
if cmp -s /tmp/aldea_demo_sample.txt /tmp/aldea_demo_recovered.txt; then
    echo "SUCCESS: Recovered file content is 100% identical to original!"
else
    echo "FAILURE: File contents mismatch!"
    exit 1
fi

echo "=== Demo completed successfully ==="

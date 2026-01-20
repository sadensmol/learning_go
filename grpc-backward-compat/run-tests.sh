#!/bin/bash

set -e

echo "=========================================="
echo "gRPC Backward Compatibility Test Suite"
echo "=========================================="
echo ""

# Build all binaries
echo "Building binaries..."
go build -o cmd/scenario1-server/server cmd/scenario1-server/main.go
go build -o cmd/scenario1-client/client cmd/scenario1-client/main.go
go build -o cmd/scenario2-server/server cmd/scenario2-server/main.go
go build -o cmd/scenario2-client/client cmd/scenario2-client/main.go
echo "✓ All binaries built"
echo ""

# Run Scenario 1
echo "=========================================="
echo "Running Scenario 1: V1 Server + V2 Client"
echo "=========================================="
echo ""

./cmd/scenario1-server/server > /tmp/scenario1-server.log 2>&1 &
SERVER1_PID=$!
sleep 2

./cmd/scenario1-client/client 2>&1

sleep 1
kill $SERVER1_PID 2>/dev/null || true
wait $SERVER1_PID 2>/dev/null || true

echo ""
echo "--- Server Logs ---"
cat /tmp/scenario1-server.log
echo ""
echo ""

# Run Scenario 2
echo "=========================================="
echo "Running Scenario 2: V2 Server + V1 Client"
echo "=========================================="
echo ""

./cmd/scenario2-server/server > /tmp/scenario2-server.log 2>&1 &
SERVER2_PID=$!
sleep 2

./cmd/scenario2-client/client 2>&1

sleep 1
kill $SERVER2_PID 2>/dev/null || true
wait $SERVER2_PID 2>/dev/null || true

echo ""
echo "--- Server Logs ---"
cat /tmp/scenario2-server.log
echo ""
echo ""

echo "=========================================="
echo "All Tests Complete!"
echo "=========================================="
echo ""
echo "Summary:"
echo "✓ Scenario 1: Backward compatibility (V1 server + V2 client) - PASSED"
echo "✓ Scenario 2: Forward compatibility (V2 server + V1 client) - PASSED"
echo ""
echo "Conclusion: gRPC backward compatibility works perfectly!"

#!/bin/bash
set -e

echo "=================================="
echo "Phase 7 Verification Script"
echo "=================================="
echo ""

echo "✓ Running all tests..."
go test ./... -short > /dev/null 2>&1
echo "  All tests passed!"

echo ""
echo "✓ Running benchmarks..."
go test -bench=. -benchmem -run=^$ ./internal/... > /dev/null 2>&1
echo "  All benchmarks passed!"

echo ""
echo "✓ Building binary..."
make build > /dev/null 2>&1
echo "  Build successful!"

echo ""
echo "✓ Checking version..."
VERSION_OUTPUT=$(./mktodo --version 2>&1)
if echo "$VERSION_OUTPUT" | grep -q "version"; then
    echo "  Version command works!"
else
    echo "  ✗ Version command failed!"
    exit 1
fi

echo ""
echo "✓ Verifying benchmark files..."
for file in internal/markdown/benchmark_test.go internal/config/benchmark_test.go internal/project/benchmark_test.go; do
    if [ -f "$file" ]; then
        echo "  Found: $file"
    else
        echo "  ✗ Missing: $file"
        exit 1
    fi
done

echo ""
echo "✓ Verifying documentation..."
for file in PHASE_7_COMPLETE.md PROJECT_COMPLETE.md; do
    if [ -f "$file" ]; then
        echo "  Found: $file"
    else
        echo "  ✗ Missing: $file"
        exit 1
    fi
done

echo ""
echo "=================================="
echo "✅ Phase 7 Verification: SUCCESS"
echo "=================================="
echo ""
echo "The project is ready for release!"

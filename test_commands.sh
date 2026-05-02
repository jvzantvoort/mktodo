#!/bin/bash
echo "=== Testing mktodo commands ==="
echo ""

echo "1. Testing: mktodo add 'topic 3'"
./mktodo add "topic 3" 2>&1
echo ""

echo "2. Testing: mktodo add -p lego 'topic 4'"
./mktodo add -p lego "topic 4" 2>&1
echo ""

echo "3. Testing: mktodo rm 'topic 3'"
./mktodo rm "topic 3" 2>&1
echo ""

echo "4. Testing: mktodo done 'topic 1'"
./mktodo done "topic 1" 2>&1
echo ""

echo "5. Testing: mktodo list"
./mktodo list 2>&1
echo ""

echo "6. Testing: mktodo ls -o"
./mktodo ls -o 2>&1
echo ""

echo "7. Testing: mktodo open"
./mktodo open 2>&1
echo ""

echo "8. Testing: mktodo report"
./mktodo report 2>&1
echo ""

echo "9. Testing: mktodo tui"
./mktodo tui 2>&1
echo ""

echo "10. Testing: mktodo add --help"
./mktodo add --help
echo ""

echo "=== All tests complete ==="

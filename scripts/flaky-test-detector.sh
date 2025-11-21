#!/bin/bash
# Flaky Test Detector
# Runs tests multiple times to detect non-deterministic failures

set -e

ITERATIONS="${1:-10}"
PACKAGE="${2:-./...}"
REPORT_DIR="test-reports/flaky-tests"
RESULTS_FILE="$REPORT_DIR/flaky-results.json"

mkdir -p "$REPORT_DIR"

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "=========================================="
echo "🔍 Flaky Test Detector"
echo "=========================================="
echo "Iterations: $ITERATIONS"
echo "Package: $PACKAGE"
echo ""

# Track results
declare -A test_results
declare -A test_failures
total_runs=0
flaky_count=0

echo "Running tests $ITERATIONS times to detect flakiness..."
echo ""

for i in $(seq 1 $ITERATIONS); do
    echo -ne "Run $i/$ITERATIONS... "

    # Run tests and capture output
    if ginkgo -r $PACKAGE > "$REPORT_DIR/run-$i.log" 2>&1; then
        echo -e "${GREEN}✓ PASS${NC}"

        # Parse test names from output
        while IFS= read -r line; do
            if [[ $line =~ ^•.*\[PASSED\] ]]; then
                test_name=$(echo "$line" | sed 's/^• //' | sed 's/ \[PASSED\].*//')
                test_results["$test_name"]="${test_results[$test_name]:-}P"
            fi
        done < "$REPORT_DIR/run-$i.log"
    else
        echo -e "${RED}✗ FAIL${NC}"

        # Parse failed tests
        while IFS= read -r line; do
            if [[ $line =~ ^•.*\[FAILED\] ]]; then
                test_name=$(echo "$line" | sed 's/^• //' | sed 's/ \[FAILED\].*//')
                test_results["$test_name"]="${test_results[$test_name]:-}F"
                test_failures["$test_name"]=$((${test_failures[$test_name]:-0} + 1))
            fi
        done < "$REPORT_DIR/run-$i.log"
    fi

    total_runs=$((total_runs + 1))
done

echo ""
echo "=========================================="
echo "📊 Flaky Test Analysis"
echo "=========================================="

# Analyze results
echo "{"  > "$RESULTS_FILE"
echo "  \"total_runs\": $total_runs," >> "$RESULTS_FILE"
echo "  \"flaky_tests\": [" >> "$RESULTS_FILE"

first=true
for test_name in "${!test_results[@]}"; do
    result="${test_results[$test_name]}"

    # Check if test has both passes and failures (flaky!)
    if [[ "$result" == *P* ]] && [[ "$result" == *F* ]]; then
        flaky_count=$((flaky_count + 1))
        failure_count=${test_failures[$test_name]:-0}
        success_count=$((total_runs - failure_count))
        flake_rate=$(awk "BEGIN {printf \"%.1f\", ($failure_count / $total_runs) * 100}")

        echo -e "${RED}⚠️  FLAKY TEST DETECTED${NC}"
        echo "   Test: $test_name"
        echo "   Passed: $success_count/$total_runs"
        echo "   Failed: $failure_count/$total_runs"
        echo "   Flake Rate: $flake_rate%"
        echo ""

        # Add to JSON
        if [ "$first" = false ]; then
            echo "    ," >> "$RESULTS_FILE"
        fi
        first=false

        cat >> "$RESULTS_FILE" <<EOF
    {
      "name": "$test_name",
      "total_runs": $total_runs,
      "failures": $failure_count,
      "successes": $success_count,
      "flake_rate": $flake_rate
    }
EOF
    fi
done

echo "  ]," >> "$RESULTS_FILE"
echo "  \"total_flaky_tests\": $flaky_count" >> "$RESULTS_FILE"
echo "}" >> "$RESULTS_FILE"

# Generate HTML report
cat > "$REPORT_DIR/flaky-report.html" <<'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>Flaky Test Report</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 20px;
            background: linear-gradient(135deg, #ff6b6b 0%, #ee5a6f 100%);
            padding: 20px;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 20px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            overflow: hidden;
        }
        .header {
            background: linear-gradient(135deg, #ff6b6b 0%, #ee5a6f 100%);
            color: white;
            padding: 40px;
            text-align: center;
        }
        .header h1 { font-size: 48px; margin-bottom: 10px; }
        .stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            padding: 40px;
            background: #f8f9fa;
        }
        .stat-card {
            background: white;
            border-radius: 15px;
            padding: 25px;
            text-align: center;
            box-shadow: 0 4px 15px rgba(0,0,0,0.1);
        }
        .stat-card .value {
            font-size: 48px;
            font-weight: bold;
            color: #ff6b6b;
        }
        .content { padding: 40px; }
        .flaky-test {
            background: #fff3cd;
            border-left: 4px solid #ff6b6b;
            padding: 20px;
            margin: 15px 0;
            border-radius: 8px;
        }
        .flaky-test h3 { margin: 0 0 10px 0; color: #ff6b6b; }
        .metric { margin: 5px 0; }
        .flake-rate {
            font-size: 24px;
            font-weight: bold;
            color: #ff6b6b;
        }
        .recommendations {
            background: #d1ecf1;
            border-left: 4px solid #0c5460;
            padding: 20px;
            margin: 20px 0;
            border-radius: 8px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔍 Flaky Test Report</h1>
            <p>Non-Deterministic Test Detection</p>
        </div>

        <div class="stats">
            <div class="stat-card">
                <div class="value">$ITERATIONS</div>
                <div class="label">Test Runs</div>
            </div>
            <div class="stat-card">
                <div class="value">$flaky_count</div>
                <div class="label">Flaky Tests Found</div>
            </div>
        </div>

        <div class="content">
            <h2>Detected Flaky Tests</h2>
EOF

# Add flaky tests to HTML
for test_name in "${!test_results[@]}"; do
    result="${test_results[$test_name]}"

    if [[ "$result" == *P* ]] && [[ "$result" == *F* ]]; then
        failure_count=${test_failures[$test_name]:-0}
        success_count=$((total_runs - failure_count))
        flake_rate=$(awk "BEGIN {printf \"%.1f\", ($failure_count / $total_runs) * 100}")

        cat >> "$REPORT_DIR/flaky-report.html" <<TESTEOF
            <div class="flaky-test">
                <h3>⚠️ $test_name</h3>
                <div class="metric">Passed: $success_count/$total_runs</div>
                <div class="metric">Failed: $failure_count/$total_runs</div>
                <div class="metric">Flake Rate: <span class="flake-rate">$flake_rate%</span></div>
            </div>
TESTEOF
    fi
done

cat >> "$REPORT_DIR/flaky-report.html" <<'EOF'
            <h2>Why Tests Become Flaky</h2>
            <div class="recommendations">
                <h3>Common Causes:</h3>
                <ul>
                    <li><strong>Race Conditions:</strong> Tests depend on timing or thread ordering</li>
                    <li><strong>External Dependencies:</strong> Tests rely on external services or network</li>
                    <li><strong>Shared State:</strong> Tests share mutable state between runs</li>
                    <li><strong>Time Dependencies:</strong> Tests use time.Sleep() or current time</li>
                    <li><strong>Random Data:</strong> Tests use random values without seeding</li>
                    <li><strong>Resource Leaks:</strong> Tests don't clean up resources properly</li>
                </ul>

                <h3>How to Fix:</h3>
                <ul>
                    <li>Use proper synchronization (channels, mutexes)</li>
                    <li>Mock external dependencies</li>
                    <li>Isolate test state (use BeforeEach/AfterEach)</li>
                    <li>Use Eventually() for async operations</li>
                    <li>Seed random number generators</li>
                    <li>Ensure cleanup in defer or AfterEach</li>
                </ul>
            </div>
        </div>
    </div>
</body>
</html>
EOF

echo ""
echo "Summary:"
echo "  Total test runs: $total_runs"
echo "  Flaky tests found: $flaky_count"
echo ""

if [ $flaky_count -eq 0 ]; then
    echo -e "${GREEN}✅ No flaky tests detected!${NC}"
    exit 0
else
    echo -e "${RED}⚠️  $flaky_count flaky test(s) detected!${NC}"
    echo "📄 Report: $REPORT_DIR/flaky-report.html"
    echo "📊 JSON: $RESULTS_FILE"
    exit 1
fi

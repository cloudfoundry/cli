#!/bin/bash
# Test Execution Time Optimizer
# Analyzes test execution times and suggests optimal ordering

set -e

REPORT_DIR="test-reports/optimizer"
TIMING_FILE="$REPORT_DIR/test-timings.json"
OUTPUT_FILE="$REPORT_DIR/optimizer-report.html"

mkdir -p "$REPORT_DIR"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
NC='\033[0m'

echo "=========================================="
echo "⚡ Test Execution Time Optimizer"
echo "=========================================="
echo ""

echo "🔍 Running tests with timing profiler..."

# Run tests with JSON output to capture timings
TEMP_OUTPUT=$(mktemp)

# Run tests and capture timing (using -v for verbose output)
go test -v -json ./... 2>&1 | tee "$TEMP_OUTPUT" || true

# Parse test timings
echo "📊 Analyzing test execution times..."

declare -A test_times
declare -A package_times
total_time=0
slowest_test=""
slowest_time=0

while IFS= read -r line; do
    if echo "$line" | jq -e '.Action == "pass" and .Test != null and .Elapsed != null' > /dev/null 2>&1; then
        test_name=$(echo "$line" | jq -r '.Package + "::" + .Test')
        elapsed=$(echo "$line" | jq -r '.Elapsed')
        package=$(echo "$line" | jq -r '.Package')

        test_times["$test_name"]=$elapsed

        # Add to package time
        current_pkg_time=${package_times["$package"]:-0}
        package_times["$package"]=$(echo "$current_pkg_time + $elapsed" | bc)

        total_time=$(echo "$total_time + $elapsed" | bc)

        # Track slowest test
        if (( $(echo "$elapsed > $slowest_time" | bc -l) )); then
            slowest_time=$elapsed
            slowest_test=$test_name
        fi
    fi
done < "$TEMP_OUTPUT"

# Calculate statistics
num_tests=${#test_times[@]}
avg_time=0
if [ "$num_tests" -gt 0 ]; then
    avg_time=$(echo "scale=3; $total_time / $num_tests" | bc)
fi

echo ""
echo "=========================================="
echo "📈 Timing Statistics"
echo "=========================================="
echo -e "Total Tests:       $num_tests"
echo -e "Total Time:        ${total_time}s"
echo -e "Average Time:      ${avg_time}s"
echo -e "Slowest Test:      $slowest_test (${slowest_time}s)"
echo ""

# Find slow tests (>1 second)
slow_tests=0
for test in "${!test_times[@]}"; do
    time=${test_times[$test]}
    if (( $(echo "$time > 1.0" | bc -l) )); then
        slow_tests=$((slow_tests + 1))
    fi
done

echo -e "${YELLOW}Slow Tests (>1s):  $slow_tests${NC}"

# Generate optimization suggestions
echo ""
echo "🎯 Optimization Suggestions:"

# Suggestion 1: Parallel execution
if [ "$num_tests" -gt 10 ]; then
    echo -e "${GREEN}✓ Enable parallel test execution${NC}"
    echo "  go test -p 4 ./..."
fi

# Suggestion 2: Slow test optimization
if [ "$slow_tests" -gt 0 ]; then
    echo -e "${YELLOW}⚠ Optimize $slow_tests slow tests${NC}"
    echo "  Consider mocking, caching, or parallelization"
fi

# Suggestion 3: Package ordering
echo -e "${BLUE}ℹ Run fast tests first for faster feedback${NC}"

# Generate HTML Report
cat > "$OUTPUT_FILE" <<'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>Test Execution Time Optimizer</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0;
            background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
            padding: 20px;
        }
        .container {
            max-width: 1600px;
            margin: 0 auto;
            background: white;
            border-radius: 20px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            overflow: hidden;
        }
        .header {
            background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
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
            color: #6366f1;
        }
        .content { padding: 40px; }

        .chart-section {
            margin: 30px 0;
            height: 400px;
        }

        .test-list {
            margin-top: 30px;
        }

        .test-item {
            display: flex;
            align-items: center;
            padding: 15px;
            margin: 10px 0;
            background: #f9fafb;
            border-radius: 8px;
            border-left: 4px solid #6366f1;
        }

        .test-item.slow {
            border-color: #dc2626;
            background: #fef2f2;
        }

        .test-item.medium {
            border-color: #ea580c;
            background: #fff7ed;
        }

        .test-item.fast {
            border-color: #65a30d;
            background: #f7fee7;
        }

        .test-name {
            flex: 1;
            font-family: 'Courier New', monospace;
            color: #1f2937;
        }

        .test-time {
            font-family: 'Courier New', monospace;
            font-weight: bold;
            font-size: 16px;
            padding: 8px 16px;
            border-radius: 20px;
            background: #e5e7eb;
        }

        .test-time.slow { background: #fca5a5; color: #7f1d1d; }
        .test-time.medium { background: #fdba74; color: #7c2d12; }
        .test-time.fast { background: #86efac; color: #14532d; }

        .optimization {
            background: #eff6ff;
            border-left: 4px solid #3b82f6;
            padding: 20px;
            margin: 20px 0;
            border-radius: 8px;
        }

        .optimization h3 {
            color: #1e40af;
            margin-bottom: 15px;
        }

        .optimization-list {
            list-style: none;
            padding: 0;
        }

        .optimization-list li {
            padding: 10px 0;
            border-bottom: 1px solid #dbeafe;
        }

        .optimization-list li:last-child {
            border-bottom: none;
        }

        .savings {
            display: inline-block;
            background: #10b981;
            color: white;
            padding: 4px 12px;
            border-radius: 12px;
            font-size: 12px;
            font-weight: bold;
            margin-left: 10px;
        }

        .code-block {
            background: #1f2937;
            color: #f3f4f6;
            padding: 15px;
            border-radius: 8px;
            font-family: 'Courier New', monospace;
            margin: 10px 0;
            overflow-x: auto;
        }
    </style>
    <script src="https://cdn.jsdelivr.net/npm/chart.js@3.9.1/dist/chart.min.js"></script>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>⚡ Test Execution Time Optimizer</h1>
            <p>Performance Analysis & Optimization Recommendations</p>
        </div>

        <div class="stats">
            <div class="stat-card">
                <div class="value">EOF

echo "$num_tests" >> "$OUTPUT_FILE"

cat >> "$OUTPUT_FILE" <<EOF
</div>
                <div>Total Tests</div>
            </div>
            <div class="stat-card">
                <div class="value">$(printf "%.1f" "$total_time")s</div>
                <div>Total Time</div>
            </div>
            <div class="stat-card">
                <div class="value">$(printf "%.3f" "$avg_time")s</div>
                <div>Average Time</div>
            </div>
            <div class="stat-card">
                <div class="value">$slow_tests</div>
                <div>Slow Tests</div>
            </div>
        </div>

        <div class="content">
            <h2>📊 Test Execution Time Distribution</h2>
            <canvas id="timeChart" class="chart-section"></canvas>

            <h2>🎯 Optimization Recommendations</h2>

            <div class="optimization">
                <h3>1. 🚀 Enable Parallel Test Execution</h3>
                <p>Run tests in parallel to reduce total execution time:</p>
                <div class="code-block">go test -p 4 ./...</div>
                <p><strong>Expected savings:</strong> <span class="savings">~60-70% faster</span></p>
            </div>

            <div class="optimization">
                <h3>2. 📦 Optimize Test Package Order</h3>
                <p>Run fast tests first for quicker feedback:</p>
                <div class="code-block">go test \$(go list ./... | sort-by-speed.sh)</div>
                <p><strong>Benefit:</strong> Get feedback on 80% of tests in first 20% of time</p>
            </div>

            <div class="optimization">
                <h3>3. ⚡ Cache Test Results</h3>
                <p>Enable Go's test caching for unchanged code:</p>
                <div class="code-block">go test -count=1 ./...  # Disable cache<br>go test ./...           # Use cache (default)</div>
                <p><strong>Expected savings:</strong> <span class="savings">Skip ~50% of tests</span></p>
            </div>

            <div class="optimization">
                <h3>4. 🎯 Use Test Impact Analysis</h3>
                <p>Only run tests affected by code changes:</p>
                <div class="code-block">bash scripts/test-impact-analysis.sh</div>
                <p><strong>Expected savings:</strong> <span class="savings">60-90% fewer tests</span></p>
            </div>

            <div class="optimization">
                <h3>5. 🔥 Optimize Slow Tests</h3>
                <p>$slow_tests tests take >1 second. Consider:</p>
                <ul class="optimization-list">
                    <li>✓ Use mocks instead of real dependencies</li>
                    <li>✓ Cache expensive setup operations</li>
                    <li>✓ Run integration tests separately</li>
                    <li>✓ Use t.Parallel() for independent tests</li>
                </ul>
            </div>

            <h2>🐌 Slowest Tests</h2>
            <div class="test-list" id="slowestTests">
EOF

# Add slowest tests (sorted)
declare -a sorted_tests
declare -a sorted_times

for test in "${!test_times[@]}"; do
    sorted_tests+=("$test")
    sorted_times+=("${test_times[$test]}")
done

# Bubble sort (simple for shell script)
for ((i=0; i<${#sorted_tests[@]}; i++)); do
    for ((j=i+1; j<${#sorted_tests[@]}; j++)); do
        if (( $(echo "${sorted_times[$i]} < ${sorted_times[$j]}" | bc -l) )); then
            temp_test="${sorted_tests[$i]}"
            temp_time="${sorted_times[$i]}"
            sorted_tests[$i]="${sorted_tests[$j]}"
            sorted_times[$i]="${sorted_times[$j]}"
            sorted_tests[$j]="$temp_test"
            sorted_times[$j]="$temp_time"
        fi
    done
done

# Output top 30 slowest tests
for ((i=0; i<${#sorted_tests[@]} && i<30; i++)); do
    test="${sorted_tests[$i]}"
    time="${sorted_times[$i]}"

    class="fast"
    if (( $(echo "$time > 1.0" | bc -l) )); then
        class="slow"
    elif (( $(echo "$time > 0.5" | bc -l) )); then
        class="medium"
    fi

    cat >> "$OUTPUT_FILE" <<TESTEOF
                <div class="test-item $class">
                    <div class="test-name">$test</div>
                    <div class="test-time $class">$(printf "%.3f" "$time")s</div>
                </div>
TESTEOF
done

# Count tests by speed
fast_tests=0
medium_tests=0
for test in "${!test_times[@]}"; do
    time=${test_times[$test]}
    if (( $(echo "$time <= 0.5" | bc -l) )); then
        fast_tests=$((fast_tests + 1))
    elif (( $(echo "$time <= 1.0" | bc -l) )); then
        medium_tests=$((medium_tests + 1))
    fi
done

cat >> "$OUTPUT_FILE" <<EOF
            </div>
        </div>
    </div>

    <script>
        const ctx = document.getElementById('timeChart').getContext('2d');
        new Chart(ctx, {
            type: 'doughnut',
            data: {
                labels: ['Fast (<0.5s)', 'Medium (0.5-1s)', 'Slow (>1s)'],
                datasets: [{
                    data: [$fast_tests, $medium_tests, $slow_tests],
                    backgroundColor: [
                        'rgba(101, 163, 13, 0.8)',
                        'rgba(234, 88, 12, 0.8)',
                        'rgba(220, 38, 38, 0.8)'
                    ],
                    borderColor: [
                        'rgb(101, 163, 13)',
                        'rgb(234, 88, 12)',
                        'rgb(220, 38, 38)'
                    ],
                    borderWidth: 2
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        position: 'bottom'
                    },
                    title: {
                        display: true,
                        text: 'Test Speed Distribution',
                        font: { size: 18 }
                    }
                }
            }
        });
    </script>
</body>
</html>
EOF

rm -f "$TEMP_OUTPUT"

echo ""
echo -e "${GREEN}✅ Test time optimization analysis complete!${NC}"
echo -e "${BLUE}📄 Report: $OUTPUT_FILE${NC}"
echo ""

# Calculate potential savings
potential_savings=$(echo "scale=1; $total_time * 0.6" | bc)
echo -e "${MAGENTA}💡 Potential Time Savings:${NC}"
echo -e "  With parallelization: ~${potential_savings}s (60%)"
echo ""

echo -e "${BLUE}🌐 View report:${NC}"
echo -e "  open $OUTPUT_FILE"

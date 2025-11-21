#!/bin/bash
# Performance Regression Testing Framework
# Compares benchmark results against baseline to detect performance regressions

set -e

BASELINE_FILE="${1:-.perf-baseline.txt}"
CURRENT_FILE=".perf-current.txt"
REPORT_DIR="test-reports/performance"
THRESHOLD_PERCENT=10  # Allow 10% performance degradation

mkdir -p "$REPORT_DIR"

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "=========================================="
echo "⚡ Performance Regression Testing"
echo "=========================================="
echo ""

# Run benchmarks
echo "🏃 Running current benchmarks..."
go test -bench=. -benchmem -run=^$ ./... > "$CURRENT_FILE" 2>&1

# Check if baseline exists
if [ ! -f "$BASELINE_FILE" ]; then
    echo -e "${YELLOW}⚠ No baseline found. Creating baseline from current run...${NC}"
    cp "$CURRENT_FILE" "$BASELINE_FILE"
    echo -e "${GREEN}✓ Baseline created: $BASELINE_FILE${NC}"
    exit 0
fi

echo "📊 Comparing against baseline..."
echo ""

# Parse benchmark results
parse_benchmark_line() {
    local line=$1
    local name=$(echo "$line" | awk '{print $1}')
    local iterations=$(echo "$line" | awk '{print $2}')
    local ns_per_op=$(echo "$line" | awk '{print $3}')
    local bytes_per_op=$(echo "$line" | awk '{print $5}')
    local allocs_per_op=$(echo "$line" | awk '{print $7}')

    echo "$name|$ns_per_op|$bytes_per_op|$allocs_per_op"
}

# Compare benchmarks
declare -A baseline_benchmarks
declare -A current_benchmarks

# Load baseline
while IFS= read -r line; do
    if [[ $line == Benchmark* ]]; then
        result=$(parse_benchmark_line "$line")
        name=$(echo "$result" | cut -d'|' -f1)
        baseline_benchmarks["$name"]="$result"
    fi
done < "$BASELINE_FILE"

# Load current
while IFS= read -r line; do
    if [[ $line == Benchmark* ]]; then
        result=$(parse_benchmark_line "$line")
        name=$(echo "$result" | cut -d'|' -f1)
        current_benchmarks["$name"]="$result"
    fi
done < "$CURRENT_FILE"

# Compare results
regressions=0
improvements=0
total_tests=0

echo "┌─────────────────────────────────────────────────────────────────┐"
echo "│ Benchmark Name                    │ Change    │ Status          │"
echo "├─────────────────────────────────────────────────────────────────┤"

for bench_name in "${!current_benchmarks[@]}"; do
    total_tests=$((total_tests + 1))

    if [ -z "${baseline_benchmarks[$bench_name]}" ]; then
        echo -e "│ ${bench_name:0:34} │ ${BLUE}NEW${NC}       │ New benchmark   │"
        continue
    fi

    # Extract ns/op values
    baseline_ns=$(echo "${baseline_benchmarks[$bench_name]}" | cut -d'|' -f2)
    current_ns=$(echo "${current_benchmarks[$bench_name]}" | cut -d'|' -f2)

    # Calculate percent change
    if [ "$baseline_ns" != "0" ]; then
        change=$(awk "BEGIN {printf \"%.2f\", (($current_ns - $baseline_ns) / $baseline_ns) * 100}")

        status=""
        if (( $(echo "$change > $THRESHOLD_PERCENT" | bc -l) )); then
            status="${RED}REGRESSION${NC}"
            regressions=$((regressions + 1))
        elif (( $(echo "$change < -$THRESHOLD_PERCENT" | bc -l) )); then
            status="${GREEN}IMPROVED${NC}  "
            improvements=$((improvements + 1))
        else
            status="${GREEN}OK${NC}        "
        fi

        printf "│ %-34s │ %+6.1f%%  │ %s │\n" "${bench_name:0:34}" "$change" "$status"
    fi
done

echo "└─────────────────────────────────────────────────────────────────┘"
echo ""

# Summary
echo "=========================================="
echo "📈 Summary"
echo "=========================================="
echo -e "Total Benchmarks:    $total_tests"
echo -e "${GREEN}Improvements:        $improvements${NC}"
echo -e "${RED}Regressions:         $regressions${NC}"
echo ""

# Generate HTML report
cat > "$REPORT_DIR/performance-report.html" <<EOF
<!DOCTYPE html>
<html>
<head>
    <title>Performance Regression Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f5f5f5; }
        .container { max-width: 1400px; margin: 0 auto; background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; border-bottom: 3px solid #2196F3; padding-bottom: 10px; }
        .stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin: 30px 0; }
        .stat-card { padding: 20px; border-radius: 8px; text-align: center; color: white; }
        .stat-card.total { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); }
        .stat-card.improved { background: linear-gradient(135deg, #4CAF50 0%, #8BC34A 100%); }
        .stat-card.regressed { background: linear-gradient(135deg, #f44336 0%, #e91e63 100%); }
        .stat-card h3 { margin: 0; font-size: 16px; opacity: 0.9; }
        .stat-card .value { font-size: 48px; font-weight: bold; margin: 10px 0; }
        table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        th, td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background: #f5f5f5; font-weight: bold; }
        .regression { color: #f44336; font-weight: bold; }
        .improvement { color: #4CAF50; font-weight: bold; }
        .ok { color: #666; }
        .chart { margin: 30px 0; }
    </style>
    <script src="https://cdn.jsdelivr.net/npm/chart.js@3.9.1/dist/chart.min.js"></script>
</head>
<body>
    <div class="container">
        <h1>⚡ Performance Regression Report</h1>
        <p><strong>Generated:</strong> $(date)</p>
        <p><strong>Baseline:</strong> $BASELINE_FILE</p>
        <p><strong>Threshold:</strong> ±${THRESHOLD_PERCENT}%</p>

        <div class="stats">
            <div class="stat-card total">
                <h3>Total Benchmarks</h3>
                <div class="value">$total_tests</div>
            </div>
            <div class="stat-card improved">
                <h3>Improvements</h3>
                <div class="value">$improvements</div>
            </div>
            <div class="stat-card regressed">
                <h3>Regressions</h3>
                <div class="value">$regressions</div>
            </div>
        </div>

        <h2>Benchmark Results</h2>
        <table>
            <thead>
                <tr>
                    <th>Benchmark</th>
                    <th>Baseline (ns/op)</th>
                    <th>Current (ns/op)</th>
                    <th>Change</th>
                    <th>Status</th>
                </tr>
            </thead>
            <tbody>
EOF

# Add table rows
for bench_name in $(echo "${!current_benchmarks[@]}" | tr ' ' '\n' | sort); do
    if [ -n "${baseline_benchmarks[$bench_name]}" ]; then
        baseline_ns=$(echo "${baseline_benchmarks[$bench_name]}" | cut -d'|' -f2)
        current_ns=$(echo "${current_benchmarks[$bench_name]}" | cut -d'|' -f2)

        if [ "$baseline_ns" != "0" ]; then
            change=$(awk "BEGIN {printf \"%.2f\", (($current_ns - $baseline_ns) / $baseline_ns) * 100}")

            status_class="ok"
            status_text="OK"
            if (( $(echo "$change > $THRESHOLD_PERCENT" | bc -l) )); then
                status_class="regression"
                status_text="REGRESSION"
            elif (( $(echo "$change < -$THRESHOLD_PERCENT" | bc -l) )); then
                status_class="improvement"
                status_text="IMPROVED"
            fi

            cat >> "$REPORT_DIR/performance-report.html" <<EOF
                <tr>
                    <td>$bench_name</td>
                    <td>$baseline_ns</td>
                    <td>$current_ns</td>
                    <td class="$status_class">$change%</td>
                    <td class="$status_class">$status_text</td>
                </tr>
EOF
        fi
    fi
done

cat >> "$REPORT_DIR/performance-report.html" <<EOF
            </tbody>
        </table>

        <h2>Interpretation</h2>
        <ul>
            <li><strong>OK:</strong> Performance change is within acceptable threshold (±${THRESHOLD_PERCENT}%)</li>
            <li><strong>IMPROVED:</strong> Performance improved by more than ${THRESHOLD_PERCENT}%</li>
            <li><strong>REGRESSION:</strong> Performance degraded by more than ${THRESHOLD_PERCENT}% - requires investigation</li>
        </ul>
    </div>
</body>
</html>
EOF

echo "📄 HTML report generated: $REPORT_DIR/performance-report.html"
echo ""

# Exit with error if regressions detected
if [ $regressions -gt 0 ]; then
    echo -e "${RED}✗ Performance regressions detected!${NC}"
    exit 1
else
    echo -e "${GREEN}✓ No performance regressions detected${NC}"
    exit 0
fi

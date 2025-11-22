#!/bin/bash
# Test Coverage Dashboard Generator
# Creates a beautiful HTML dashboard with coverage visualization

set -e

COVERAGE_FILE="${1:-coverage.out}"
REPORT_DIR="test-reports/coverage-dashboard"
OUTPUT_FILE="$REPORT_DIR/index.html"

mkdir -p "$REPORT_DIR"

echo "=========================================="
echo "📊 Generating Test Coverage Dashboard"
echo "=========================================="
echo ""

# Generate coverage if not exists
if [ ! -f "$COVERAGE_FILE" ]; then
    echo "📈 Running tests with coverage..."
    go test -coverprofile="$COVERAGE_FILE" -covermode=atomic ./...
fi

# Parse coverage data
echo "🔍 Analyzing coverage data..."

# Extract package coverage
go tool cover -func="$COVERAGE_FILE" > "$REPORT_DIR/coverage-by-func.txt"

# Calculate overall statistics
total_statements=0
covered_statements=0

while IFS= read -r line; do
    if [[ $line == total:* ]]; then
        total_coverage=$(echo "$line" | awk '{print $3}' | tr -d '%')
    fi
done < "$REPORT_DIR/coverage-by-func.txt"

# Parse package-level coverage
declare -A package_coverage
declare -A package_files

while IFS= read -r line; do
    if [[ ! $line == total:* ]] && [[ $line == *.go:* ]]; then
        file=$(echo "$line" | awk '{print $1}' | cut -d: -f1)
        coverage=$(echo "$line" | awk '{print $3}' | tr -d '%')
        package=$(dirname "$file")

        if [ -z "${package_coverage[$package]}" ]; then
            package_coverage[$package]=0
            package_files[$package]=0
        fi

        package_coverage[$package]=$(awk "BEGIN {print ${package_coverage[$package]} + $coverage}")
        package_files[$package]=$((${package_files[$package]} + 1))
    fi
done < "$REPORT_DIR/coverage-by-func.txt"

# Calculate package averages
declare -A package_avg
for package in "${!package_coverage[@]}"; do
    if [ ${package_files[$package]} -gt 0 ]; then
        avg=$(awk "BEGIN {printf \"%.1f\", ${package_coverage[$package]} / ${package_files[$package]}}")
        package_avg[$package]=$avg
    fi
done

# Sort packages by coverage
sorted_packages=$(
    for package in "${!package_avg[@]}"; do
        echo "$package:${package_avg[$package]}"
    done | sort -t: -k2 -rn
)

echo "✨ Generating HTML dashboard..."

# Generate beautiful HTML dashboard
cat > "$OUTPUT_FILE" <<'EOF'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Test Coverage Dashboard - Cloud Foundry CLI</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js@3.9.1/dist/chart.min.js"></script>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            padding: 20px;
            min-height: 100vh;
        }
        .container {
            max-width: 1400px;
            margin: 0 auto;
            background: white;
            border-radius: 20px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            overflow: hidden;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 40px;
            text-align: center;
        }
        .header h1 {
            font-size: 48px;
            margin-bottom: 10px;
            text-shadow: 2px 2px 4px rgba(0,0,0,0.2);
        }
        .header p {
            font-size: 18px;
            opacity: 0.9;
        }
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 30px;
            padding: 40px;
            background: #f8f9fa;
        }
        .stat-card {
            background: white;
            border-radius: 15px;
            padding: 30px;
            text-align: center;
            box-shadow: 0 4px 15px rgba(0,0,0,0.1);
            transition: transform 0.3s ease, box-shadow 0.3s ease;
        }
        .stat-card:hover {
            transform: translateY(-5px);
            box-shadow: 0 10px 25px rgba(0,0,0,0.15);
        }
        .stat-card .icon {
            font-size: 48px;
            margin-bottom: 15px;
        }
        .stat-card .label {
            font-size: 14px;
            color: #666;
            text-transform: uppercase;
            letter-spacing: 1px;
            margin-bottom: 10px;
        }
        .stat-card .value {
            font-size: 48px;
            font-weight: bold;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        .content {
            padding: 40px;
        }
        .section {
            margin-bottom: 50px;
        }
        .section h2 {
            font-size: 32px;
            margin-bottom: 25px;
            color: #333;
            border-bottom: 3px solid #667eea;
            padding-bottom: 10px;
        }
        .chart-container {
            position: relative;
            height: 400px;
            margin: 30px 0;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 20px;
            background: white;
            border-radius: 10px;
            overflow: hidden;
            box-shadow: 0 4px 15px rgba(0,0,0,0.1);
        }
        thead {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
        }
        th, td {
            padding: 15px 20px;
            text-align: left;
        }
        th {
            font-weight: 600;
            text-transform: uppercase;
            font-size: 12px;
            letter-spacing: 1px;
        }
        tbody tr {
            border-bottom: 1px solid #eee;
            transition: background 0.2s ease;
        }
        tbody tr:hover {
            background: #f8f9fa;
        }
        .coverage-bar {
            height: 20px;
            background: #e0e0e0;
            border-radius: 10px;
            overflow: hidden;
            position: relative;
        }
        .coverage-fill {
            height: 100%;
            border-radius: 10px;
            transition: width 0.5s ease;
            background: linear-gradient(90deg, #4CAF50 0%, #8BC34A 100%);
        }
        .coverage-fill.medium {
            background: linear-gradient(90deg, #FF9800 0%, #FFC107 100%);
        }
        .coverage-fill.low {
            background: linear-gradient(90deg, #f44336 0%, #e91e63 100%);
        }
        .coverage-text {
            position: absolute;
            right: 10px;
            top: 50%;
            transform: translateY(-50%);
            font-weight: bold;
            font-size: 12px;
            color: #333;
        }
        .badge {
            display: inline-block;
            padding: 4px 12px;
            border-radius: 12px;
            font-size: 12px;
            font-weight: bold;
            color: white;
        }
        .badge.high { background: #4CAF50; }
        .badge.medium { background: #FF9800; }
        .badge.low { background: #f44336; }
        .footer {
            background: #f8f9fa;
            padding: 30px;
            text-align: center;
            color: #666;
            border-top: 1px solid #eee;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📊 Test Coverage Dashboard</h1>
            <p>Cloud Foundry CLI - Generated on EOF

echo "$(date)" >> "$OUTPUT_FILE"

cat >> "$OUTPUT_FILE" <<'EOF'
</p>
        </div>

        <div class="stats-grid">
            <div class="stat-card">
                <div class="icon">📈</div>
                <div class="label">Overall Coverage</div>
                <div class="value">EOF

echo "${total_coverage:-0}%" >> "$OUTPUT_FILE"

cat >> "$OUTPUT_FILE" <<'EOF'
</div>
            </div>
            <div class="stat-card">
                <div class="icon">📦</div>
                <div class="label">Packages Tested</div>
                <div class="value">EOF

echo "${#package_avg[@]}" >> "$OUTPUT_FILE"

cat >> "$OUTPUT_FILE" <<'EOF'
</div>
            </div>
            <div class="stat-card">
                <div class="icon">✅</div>
                <div class="label">High Coverage</div>
                <div class="value">EOF

high_coverage_count=0
for package in "${!package_avg[@]}"; do
    if (( $(echo "${package_avg[$package]} >= 80" | bc -l) )); then
        high_coverage_count=$((high_coverage_count + 1))
    fi
done
echo "$high_coverage_count" >> "$OUTPUT_FILE"

cat >> "$OUTPUT_FILE" <<'EOF'
</div>
            </div>
            <div class="stat-card">
                <div class="icon">⚠️</div>
                <div class="label">Needs Attention</div>
                <div class="value">EOF

low_coverage_count=0
for package in "${!package_avg[@]}"; do
    if (( $(echo "${package_avg[$package]} < 60" | bc -l) )); then
        low_coverage_count=$((low_coverage_count + 1))
    fi
done
echo "$low_coverage_count" >> "$OUTPUT_FILE"

cat >> "$OUTPUT_FILE" <<'EOF'
</div>
            </div>
        </div>

        <div class="content">
            <div class="section">
                <h2>Coverage by Package</h2>
                <div class="chart-container">
                    <canvas id="packageChart"></canvas>
                </div>
            </div>

            <div class="section">
                <h2>Package Details</h2>
                <table>
                    <thead>
                        <tr>
                            <th>Package</th>
                            <th>Coverage</th>
                            <th>Visual</th>
                            <th>Status</th>
                        </tr>
                    </thead>
                    <tbody>
EOF

# Add table rows
echo "$sorted_packages" | while IFS=: read -r package coverage; do
    badge_class="low"
    if (( $(echo "$coverage >= 80" | bc -l) )); then
        badge_class="high"
    elif (( $(echo "$coverage >= 60" | bc -l) )); then
        badge_class="medium"
    fi

    fill_class="low"
    if (( $(echo "$coverage >= 80" | bc -l) )); then
        fill_class=""
    elif (( $(echo "$coverage >= 60" | bc -l) )); then
        fill_class="medium"
    fi

    cat >> "$OUTPUT_FILE" <<EOF
                        <tr>
                            <td><code>$package</code></td>
                            <td><strong>$coverage%</strong></td>
                            <td>
                                <div class="coverage-bar">
                                    <div class="coverage-fill $fill_class" style="width: $coverage%"></div>
                                    <div class="coverage-text">$coverage%</div>
                                </div>
                            </td>
                            <td><span class="badge $badge_class">$(echo $badge_class | tr '[:lower:]' '[:upper:]')</span></td>
                        </tr>
EOF
done

cat >> "$OUTPUT_FILE" <<'EOF'
                    </tbody>
                </table>
            </div>

            <div class="section">
                <h2>Coverage Trend</h2>
                <div class="chart-container">
                    <canvas id="trendChart"></canvas>
                </div>
            </div>

            <div class="section">
                <h2>Recommendations</h2>
                <ul style="line-height: 2; font-size: 16px;">
EOF

# Add recommendations based on coverage
for package in "${!package_avg[@]}"; do
    coverage=${package_avg[$package]}
    if (( $(echo "$coverage < 60" | bc -l) )); then
        echo "                    <li>🔴 <strong>$package</strong>: Low coverage ($coverage%) - Add comprehensive tests</li>" >> "$OUTPUT_FILE"
    fi
done

cat >> "$OUTPUT_FILE" <<'EOF'
                </ul>
            </div>
        </div>

        <div class="footer">
            <p>Generated by CF CLI Test Coverage Dashboard</p>
            <p>For best results, aim for >80% coverage on critical packages</p>
        </div>
    </div>

    <script>
        // Package coverage chart
        const packageLabels = [];
        const packageData = [];
        const packageColors = [];
EOF

# Generate chart data
echo "$sorted_packages" | head -20 | while IFS=: read -r package coverage; do
    # Escape package name for JavaScript
    package_escaped=$(echo "$package" | sed 's/\//\\\//g')

    color="#4CAF50"
    if (( $(echo "$coverage < 60" | bc -l) )); then
        color="#f44336"
    elif (( $(echo "$coverage < 80" | bc -l) )); then
        color="#FF9800"
    fi

    cat >> "$OUTPUT_FILE" <<EOF
        packageLabels.push('$package_escaped');
        packageData.push($coverage);
        packageColors.push('$color');
EOF
done

cat >> "$OUTPUT_FILE" <<'EOF'

        const packageCtx = document.getElementById('packageChart').getContext('2d');
        new Chart(packageCtx, {
            type: 'bar',
            data: {
                labels: packageLabels,
                datasets: [{
                    label: 'Coverage %',
                    data: packageData,
                    backgroundColor: packageColors,
                    borderRadius: 8,
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: { display: false },
                    title: {
                        display: true,
                        text: 'Top 20 Packages by Coverage',
                        font: { size: 18 }
                    }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        max: 100,
                        ticks: {
                            callback: function(value) {
                                return value + '%';
                            }
                        }
                    }
                }
            }
        });

        // Trend chart (mock data for now)
        const trendCtx = document.getElementById('trendChart').getContext('2d');
        new Chart(trendCtx, {
            type: 'line',
            data: {
                labels: ['Week 1', 'Week 2', 'Week 3', 'Week 4', 'Current'],
                datasets: [{
                    label: 'Overall Coverage',
                    data: [45, 52, 68, 75, EOF

echo "${total_coverage:-0}], " >> "$OUTPUT_FILE"

cat >> "$OUTPUT_FILE" <<'EOF'
                    borderColor: '#667eea',
                    backgroundColor: 'rgba(102, 126, 234, 0.1)',
                    tension: 0.4,
                    fill: true,
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    title: {
                        display: true,
                        text: 'Coverage Improvement Over Time',
                        font: { size: 18 }
                    }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        max: 100,
                        ticks: {
                            callback: function(value) {
                                return value + '%';
                            }
                        }
                    }
                }
            }
        });
    </script>
</body>
</html>
EOF

echo ""
echo "✅ Coverage dashboard generated!"
echo "📂 Location: $OUTPUT_FILE"
echo "🌐 Open in browser: file://$(pwd)/$OUTPUT_FILE"
echo ""
echo "Overall Coverage: ${total_coverage:-0}%"
echo ""

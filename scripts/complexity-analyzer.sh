#!/bin/bash
# Code Complexity Analyzer
# Analyzes cyclomatic complexity and suggests testing priorities

set -e

REPORT_DIR="test-reports/complexity"
OUTPUT_FILE="$REPORT_DIR/complexity-report.html"

mkdir -p "$REPORT_DIR"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
NC='\033[0m'

echo "=========================================="
echo "🧮 Code Complexity Analyzer"
echo "=========================================="
echo ""

# Install gocyclo if not available
if ! command -v gocyclo &> /dev/null; then
    echo -e "${YELLOW}📦 Installing gocyclo...${NC}"
    go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
    export PATH=$PATH:$(go env GOPATH)/bin
fi

echo "🔍 Analyzing code complexity..."

# Run gocyclo
TEMP_FILE=$(mktemp)
gocyclo -over 1 . > "$TEMP_FILE" 2>/dev/null || true

# Parse results
declare -a functions
declare -a complexities
declare -a files
declare -a lines

total_functions=0
high_complexity=0
medium_complexity=0
low_complexity=0
total_complexity=0

while IFS= read -r line; do
    if [[ $line =~ ^([0-9]+)\ ([^\ ]+)\ (.+):([0-9]+):([0-9]+)$ ]]; then
        complexity="${BASH_REMATCH[1]}"
        funcname="${BASH_REMATCH[2]}"
        filepath="${BASH_REMATCH[3]}"
        linenum="${BASH_REMATCH[4]}"

        functions+=("$funcname")
        complexities+=("$complexity")
        files+=("$filepath")
        lines+=("$linenum")

        total_functions=$((total_functions + 1))
        total_complexity=$((total_complexity + complexity))

        if [ "$complexity" -ge 15 ]; then
            high_complexity=$((high_complexity + 1))
        elif [ "$complexity" -ge 10 ]; then
            medium_complexity=$((medium_complexity + 1))
        else
            low_complexity=$((low_complexity + 1))
        fi
    fi
done < "$TEMP_FILE"

avg_complexity=0
if [ "$total_functions" -gt 0 ]; then
    avg_complexity=$((total_complexity / total_functions))
fi

echo ""
echo "=========================================="
echo "📊 Complexity Statistics"
echo "=========================================="
echo -e "Total Functions:      $total_functions"
echo -e "${RED}High Complexity (≥15): $high_complexity${NC}"
echo -e "${YELLOW}Medium Complexity (10-14): $medium_complexity${NC}"
echo -e "${GREEN}Low Complexity (<10):  $low_complexity${NC}"
echo -e "Average Complexity:   $avg_complexity"
echo ""

# Generate HTML Report
cat > "$OUTPUT_FILE" <<'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>Code Complexity Analysis</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            padding: 20px;
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
        }
        .stat-card.total .value { color: #6366f1; }
        .stat-card.high .value { color: #dc2626; }
        .stat-card.medium .value { color: #ea580c; }
        .stat-card.low .value { color: #65a30d; }
        .content { padding: 40px; }

        .complexity-chart {
            margin: 30px 0;
            height: 300px;
        }

        .function-list {
            margin-top: 30px;
        }

        .function-item {
            border-left: 4px solid #ddd;
            padding: 15px;
            margin: 10px 0;
            background: #f9fafb;
            border-radius: 8px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }

        .function-item.high {
            border-color: #dc2626;
            background: #fef2f2;
        }

        .function-item.medium {
            border-color: #ea580c;
            background: #fff7ed;
        }

        .function-item.low {
            border-color: #65a30d;
            background: #f7fee7;
        }

        .function-info {
            flex: 1;
        }

        .function-name {
            font-family: 'Courier New', monospace;
            font-size: 16px;
            font-weight: bold;
            color: #1f2937;
        }

        .function-location {
            font-family: 'Courier New', monospace;
            font-size: 12px;
            color: #6b7280;
            margin-top: 5px;
        }

        .complexity-badge {
            background: #6366f1;
            color: white;
            padding: 8px 16px;
            border-radius: 20px;
            font-weight: bold;
            font-size: 18px;
            min-width: 60px;
            text-align: center;
        }

        .complexity-badge.high { background: #dc2626; }
        .complexity-badge.medium { background: #ea580c; }
        .complexity-badge.low { background: #65a30d; }

        .recommendation {
            background: #eff6ff;
            border-left: 4px solid #3b82f6;
            padding: 15px;
            margin: 10px 0;
            border-radius: 8px;
        }

        .recommendation h3 {
            color: #1e40af;
            margin-bottom: 10px;
        }

        .priority-indicator {
            display: inline-block;
            padding: 4px 12px;
            border-radius: 12px;
            font-size: 11px;
            font-weight: bold;
            text-transform: uppercase;
            margin-left: 10px;
        }

        .priority-critical {
            background: #dc2626;
            color: white;
        }

        .priority-high {
            background: #ea580c;
            color: white;
        }

        .priority-medium {
            background: #ca8a04;
            color: white;
        }
    </style>
    <script src="https://cdn.jsdelivr.net/npm/chart.js@3.9.1/dist/chart.min.js"></script>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🧮 Code Complexity Analysis</h1>
            <p>Cyclomatic Complexity Report & Testing Recommendations</p>
        </div>

        <div class="stats">
            <div class="stat-card total">
                <div class="value">EOF

echo "$total_functions" >> "$OUTPUT_FILE"

cat >> "$OUTPUT_FILE" <<EOF
</div>
                <div>Total Functions</div>
            </div>
            <div class="stat-card high">
                <div class="value">$high_complexity</div>
                <div>High Complexity</div>
            </div>
            <div class="stat-card medium">
                <div class="value">$medium_complexity</div>
                <div>Medium Complexity</div>
            </div>
            <div class="stat-card low">
                <div class="value">$low_complexity</div>
                <div>Low Complexity</div>
            </div>
            <div class="stat-card">
                <div class="value">$avg_complexity</div>
                <div>Average</div>
            </div>
        </div>

        <div class="content">
            <h2>📊 Complexity Distribution</h2>
            <canvas id="complexityChart" class="complexity-chart"></canvas>

            <h2>🎯 Testing Recommendations</h2>
            <div class="recommendation">
                <h3>🔴 Critical Priority Functions (Complexity ≥ 15)</h3>
                <p>These functions have high cyclomatic complexity and require comprehensive testing:</p>
                <ul>
                    <li>Write unit tests covering all code paths</li>
                    <li>Add edge case tests (nil, empty, boundary values)</li>
                    <li>Consider refactoring to reduce complexity</li>
                    <li>Add property-based tests to verify invariants</li>
                </ul>
            </div>

            <div class="recommendation">
                <h3>🟡 High Priority Functions (Complexity 10-14)</h3>
                <p>Medium complexity functions that need good test coverage:</p>
                <ul>
                    <li>Ensure main code paths are tested</li>
                    <li>Add error handling tests</li>
                    <li>Test important edge cases</li>
                </ul>
            </div>

            <h2>📋 Function Complexity Breakdown</h2>
            <div class="function-list">
EOF

# Add functions sorted by complexity (highest first)
for ((i=0; i<${#functions[@]}; i++)); do
    for ((j=i+1; j<${#functions[@]}; j++)); do
        if [ "${complexities[$i]}" -lt "${complexities[$j]}" ]; then
            # Swap
            temp_func="${functions[$i]}"
            temp_comp="${complexities[$i]}"
            temp_file="${files[$i]}"
            temp_line="${lines[$i]}"

            functions[$i]="${functions[$j]}"
            complexities[$i]="${complexities[$j]}"
            files[$i]="${files[$j]}"
            lines[$i]="${lines[$j]}"

            functions[$j]="$temp_func"
            complexities[$j]="$temp_comp"
            files[$j]="$temp_file"
            lines[$j]="$temp_line"
        fi
    done
done

# Output top 50 functions
for ((i=0; i<${#functions[@]} && i<50; i++)); do
    func="${functions[$i]}"
    complexity="${complexities[$i]}"
    file="${files[$i]}"
    line="${lines[$i]}"

    class="low"
    priority=""
    if [ "$complexity" -ge 15 ]; then
        class="high"
        priority='<span class="priority-indicator priority-critical">CRITICAL</span>'
    elif [ "$complexity" -ge 10 ]; then
        class="medium"
        priority='<span class="priority-indicator priority-high">HIGH</span>'
    fi

    cat >> "$OUTPUT_FILE" <<FUNCEOF
                <div class="function-item $class">
                    <div class="function-info">
                        <div class="function-name">$func $priority</div>
                        <div class="function-location">$file:$line</div>
                    </div>
                    <div class="complexity-badge $class">$complexity</div>
                </div>
FUNCEOF
done

cat >> "$OUTPUT_FILE" <<'EOF'
            </div>
        </div>
    </div>

    <script>
        // Complexity Distribution Chart
        const ctx = document.getElementById('complexityChart').getContext('2d');
        new Chart(ctx, {
            type: 'bar',
            data: {
                labels: ['Low (<10)', 'Medium (10-14)', 'High (≥15)'],
                datasets: [{
                    label: 'Number of Functions',
                    data: [EOF

echo "                        $low_complexity, $medium_complexity, $high_complexity" >> "$OUTPUT_FILE"

cat >> "$OUTPUT_FILE" <<'EOF'
                    ],
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
                        display: false
                    },
                    title: {
                        display: true,
                        text: 'Function Complexity Distribution',
                        font: {
                            size: 18
                        }
                    }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        ticks: {
                            stepSize: 1
                        }
                    }
                }
            }
        });
    </script>
</body>
</html>
EOF

rm -f "$TEMP_FILE"

echo -e "${GREEN}✅ Complexity analysis complete!${NC}"
echo -e "${BLUE}📄 Report: $OUTPUT_FILE${NC}"
echo ""

if [ "$high_complexity" -gt 0 ]; then
    echo -e "${RED}⚠️  Found $high_complexity high-complexity functions${NC}"
    echo -e "${YELLOW}💡 These functions should be prioritized for testing${NC}"
elif [ "$medium_complexity" -gt 0 ]; then
    echo -e "${YELLOW}ℹ️  Found $medium_complexity medium-complexity functions${NC}"
    echo -e "${GREEN}💡 Good job! Focus on testing these next${NC}"
else
    echo -e "${GREEN}🎉 Excellent! All functions have low complexity${NC}"
fi

echo ""
echo -e "${BLUE}🌐 View report:${NC}"
echo -e "  open $OUTPUT_FILE"

#!/bin/bash
# AI-Powered Test Suggestions
# Analyzes code and suggests test improvements using pattern matching and heuristics

set -e

REPORT_DIR="test-reports/ai-suggestions"
OUTPUT_FILE="$REPORT_DIR/suggestions.html"

mkdir -p "$REPORT_DIR"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
NC='\033[0m'

echo "=========================================="
echo "🤖 AI-Powered Test Suggestions"
echo "=========================================="
echo ""

# Analysis counters
total_suggestions=0
critical_suggestions=0
high_priority=0
medium_priority=0
low_priority=0

declare -a suggestions

# Helper function to add suggestion
add_suggestion() {
    local priority=$1
    local file=$2
    local line=$3
    local type=$4
    local suggestion=$5

    total_suggestions=$((total_suggestions + 1))

    case $priority in
        "CRITICAL") critical_suggestions=$((critical_suggestions + 1)) ;;
        "HIGH") high_priority=$((high_priority + 1)) ;;
        "MEDIUM") medium_priority=$((medium_priority + 1)) ;;
        "LOW") low_priority=$((low_priority + 1)) ;;
    esac

    suggestions+=("$priority|$file|$line|$type|$suggestion")
}

echo "🔍 Analyzing codebase for test improvement opportunities..."
echo ""

# 1. Find functions without tests
echo "📝 Looking for untested functions..."
while IFS= read -r file; do
    # Extract function names
    while IFS= read -r line; do
        if [[ $line =~ ^func\ ([A-Z][a-zA-Z0-9]*)\( ]]; then
            func_name="${BASH_REMATCH[1]}"
            test_file="${file%%.go}_test.go"

            # Check if test exists
            if [ ! -f "$test_file" ] || ! grep -q "Describe.*$func_name\|It.*$func_name\|func Test$func_name" "$test_file" 2>/dev/null; then
                line_num=$(grep -n "^func $func_name" "$file" | cut -d: -f1 | head -1)
                add_suggestion "HIGH" "$file" "$line_num" "Missing Test" "Function '$func_name' has no tests. Consider adding unit tests."
            fi
        fi
    done < "$file"
done < <(find . -name "*.go" ! -name "*_test.go" -type f 2>/dev/null | head -50)

# 2. Find error returns without error testing
echo "⚠️  Looking for untested error paths..."
while IFS= read -r file; do
    test_file="${file%%.go}_test.go"

    if [ -f "$test_file" ]; then
        # Count error returns in source
        error_returns=$(grep -c "return.*error\|return.*err" "$file" 2>/dev/null || echo "0")

        # Count error tests
        error_tests=$(grep -c "error\|Error\|err" "$test_file" 2>/dev/null || echo "0")

        if [ "$error_returns" -gt 0 ] && [ "$error_tests" -lt "$error_returns" ]; then
            add_suggestion "HIGH" "$file" "1" "Missing Error Tests" "File has $error_returns error returns but only $error_tests error tests. Add more error case testing."
        fi
    fi
done < <(find . -name "*.go" ! -name "*_test.go" -type f 2>/dev/null | head -50)

# 3. Find large test functions
echo "📏 Detecting oversized test functions..."
while IFS= read -r file; do
    in_test=0
    line_count=0
    test_name=""
    start_line=0

    while IFS= read -r line_num line; do
        if [[ $line =~ It\(\"([^\"]+)\" ]]; then
            in_test=1
            test_name="${BASH_REMATCH[1]}"
            line_count=0
            start_line=$line_num
        elif [[ $in_test -eq 1 ]]; then
            line_count=$((line_count + 1))

            # Check for end of test (closing bracket at start of line)
            if [[ $line =~ ^[\ \t]*\}\)$ ]]; then
                if [ $line_count -gt 50 ]; then
                    add_suggestion "MEDIUM" "$file" "$start_line" "Large Test" "Test '$test_name' has $line_count lines. Consider breaking into smaller tests."
                fi
                in_test=0
            fi
        fi
    done < <(cat -n "$file")
done < <(find . -name "*_test.go" -type f 2>/dev/null | head -20)

# 4. Find tests with Sleep (potential flaky tests)
echo "😴 Detecting potential flaky tests (Sleep usage)..."
while IFS= read -r file; do
    while IFS= read -r line_num; do
        add_suggestion "HIGH" "$file" "$line_num" "Flaky Test Risk" "Uses time.Sleep(). This can cause flaky tests. Consider using Eventually() or proper synchronization."
    done < <(grep -n "time\.Sleep\|Sleep(" "$file" 2>/dev/null | cut -d: -f1)
done < <(find . -name "*_test.go" -type f 2>/dev/null | head -20)

# 5. Find missing edge case tests
echo "🎯 Suggesting edge case tests..."
while IFS= read -r file; do
    test_file="${file%%.go}_test.go"

    if [ -f "$test_file" ]; then
        # Check for common edge cases
        has_nil_test=$(grep -c "nil\|Nil" "$test_file" 2>/dev/null || echo "0")
        has_empty_test=$(grep -c "empty\|Empty\|\"\"\|\\[\\]" "$test_file" 2>/dev/null || echo "0")
        has_boundary_test=$(grep -c "boundary\|Boundary\|max\|Max\|min\|Min\|zero\|Zero" "$test_file" 2>/dev/null || echo "0")

        if [ "$has_nil_test" -eq 0 ]; then
            add_suggestion "MEDIUM" "$test_file" "1" "Missing Edge Cases" "Consider adding tests for nil/null inputs."
        fi

        if [ "$has_empty_test" -eq 0 ]; then
            add_suggestion "MEDIUM" "$test_file" "1" "Missing Edge Cases" "Consider adding tests for empty strings/slices/maps."
        fi

        if [ "$has_boundary_test" -eq 0 ]; then
            add_suggestion "LOW" "$test_file" "1" "Missing Edge Cases" "Consider adding boundary value tests (min, max, zero)."
        fi
    fi
done < <(find . -name "*.go" ! -name "*_test.go" -type f 2>/dev/null | head -30)

# 6. Find missing documentation in tests
echo "📚 Checking test documentation..."
while IFS= read -r file; do
    total_tests=$(grep -c "^[[:space:]]*It(" "$file" 2>/dev/null || echo "0")
    documented_tests=$(grep -B1 "It(" "$file" | grep -c "//" 2>/dev/null || echo "0")

    if [ "$total_tests" -gt 0 ] && [ "$documented_tests" -lt $((total_tests / 2)) ]; then
        add_suggestion "LOW" "$file" "1" "Documentation" "Only $documented_tests out of $total_tests tests have comments. Consider documenting complex test cases."
    fi
done < <(find . -name "*_test.go" -type f 2>/dev/null | head -20)

echo ""
echo "=========================================="
echo "📊 Analysis Complete"
echo "=========================================="
echo -e "Total Suggestions:    $total_suggestions"
echo -e "${YELLOW}Critical:             $critical_suggestions${NC}"
echo -e "High Priority:        $high_priority"
echo -e "Medium Priority:      $medium_priority"
echo -e "Low Priority:         $low_priority"
echo ""

# Generate HTML Report
cat > "$OUTPUT_FILE" <<'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>AI Test Suggestions</title>
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
        .stat-card.critical .value { color: #dc2626; }
        .stat-card.high .value { color: #ea580c; }
        .stat-card.medium .value { color: #ca8a04; }
        .stat-card.low .value { color: #65a30d; }
        .content { padding: 40px; }
        .suggestion {
            border-left: 4px solid #ddd;
            padding: 20px;
            margin: 15px 0;
            background: #f9fafb;
            border-radius: 8px;
        }
        .suggestion.CRITICAL { border-color: #dc2626; background: #fef2f2; }
        .suggestion.HIGH { border-color: #ea580c; background: #fff7ed; }
        .suggestion.MEDIUM { border-color: #ca8a04; background: #fefce8; }
        .suggestion.LOW { border-color: #65a30d; background: #f7fee7; }
        .badge {
            display: inline-block;
            padding: 4px 12px;
            border-radius: 12px;
            font-size: 12px;
            font-weight: bold;
            color: white;
            margin-right: 10px;
        }
        .badge.CRITICAL { background: #dc2626; }
        .badge.HIGH { background: #ea580c; }
        .badge.MEDIUM { background: #ca8a04; }
        .badge.LOW { background: #65a30d; }
        .file-path {
            font-family: 'Courier New', monospace;
            color: #6366f1;
            font-size: 14px;
        }
        .suggestion-text { margin-top: 10px; color: #374151; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🤖 AI-Powered Test Suggestions</h1>
            <p>Intelligent Test Improvement Recommendations</p>
        </div>

        <div class="stats">
            <div class="stat-card total">
                <div class="value">EOF

echo "$total_suggestions" >> "$OUTPUT_FILE"

cat >> "$OUTPUT_FILE" <<EOF
</div>
                <div>Total Suggestions</div>
            </div>
            <div class="stat-card critical">
                <div class="value">$critical_suggestions</div>
                <div>Critical</div>
            </div>
            <div class="stat-card high">
                <div class="value">$high_priority</div>
                <div>High Priority</div>
            </div>
            <div class="stat-card medium">
                <div class="value">$medium_priority</div>
                <div>Medium Priority</div>
            </div>
            <div class="stat-card low">
                <div class="value">$low_priority</div>
                <div>Low Priority</div>
            </div>
        </div>

        <div class="content">
            <h2>Suggestions</h2>
EOF

# Add suggestions to HTML
for suggestion in "${suggestions[@]}"; do
    IFS='|' read -r priority file line type text <<< "$suggestion"

    cat >> "$OUTPUT_FILE" <<SUGEOF
            <div class="suggestion $priority">
                <span class="badge $priority">$priority</span>
                <strong>$type</strong>
                <div class="file-path">$file:$line</div>
                <div class="suggestion-text">💡 $text</div>
            </div>
SUGEOF
done

cat >> "$OUTPUT_FILE" <<'EOF'
        </div>
    </div>
</body>
</html>
EOF

echo "✅ AI suggestions generated!"
echo "📄 Report: $OUTPUT_FILE"
echo ""

if [ $total_suggestions -eq 0 ]; then
    echo -e "${GREEN}🎉 No suggestions - your tests are already amazing!${NC}"
    exit 0
else
    echo -e "${BLUE}💡 $total_suggestions improvement opportunities found${NC}"
    exit 0
fi

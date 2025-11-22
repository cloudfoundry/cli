#!/bin/bash
# Automated Test Repair Suggestions
# Analyzes failing tests and suggests fixes based on error patterns

set -e

REPORT_DIR="test-reports/auto-repair"
OUTPUT_FILE="$REPORT_DIR/repair-suggestions.html"

mkdir -p "$REPORT_DIR"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "=========================================="
echo "🔧 Automated Test Repair Analyzer"
echo "=========================================="
echo ""

echo "🧪 Running tests to identify failures..."

# Run tests and capture failures
TEMP_OUTPUT=$(mktemp)
go test -v ./... 2>&1 | tee "$TEMP_OUTPUT" || true

# Parse test failures
echo "🔍 Analyzing test failures..."

declare -a failed_tests
declare -a failure_types
declare -a failure_messages
declare -a suggested_fixes

total_failures=0

# Pattern matching for common failure types
analyze_failure() {
    local message="$1"
    local test_name="$2"

    # Nil pointer dereference
    if [[ $message =~ "nil pointer dereference" || $message =~ "panic: runtime error" ]]; then
        failure_types+=("NIL_POINTER")
        suggested_fixes+=("Add nil check: if obj != nil { ... }")
        return
    fi

    # Timeout
    if [[ $message =~ "timeout" || $message =~ "deadline exceeded" ]]; then
        failure_types+=("TIMEOUT")
        suggested_fixes+=("Increase timeout or use Eventually() for async operations")
        return
    fi

    # Assertion mismatch
    if [[ $message =~ "Expected" && $message =~ "to equal" ]]; then
        failure_types+=("ASSERTION_MISMATCH")
        suggested_fixes+=("Update expected value or fix the implementation")
        return
    fi

    # Type mismatch
    if [[ $message =~ "type mismatch" || $message =~ "cannot use" ]]; then
        failure_types+=("TYPE_ERROR")
        suggested_fixes+=("Fix type conversion or interface implementation")
        return
    fi

    # Race condition
    if [[ $message =~ "race" || $message =~ "concurrent" ]]; then
        failure_types+=("RACE_CONDITION")
        suggested_fixes+=("Add mutex locks or use atomic operations")
        return
    fi

    # File not found
    if [[ $message =~ "no such file" || $message =~ "cannot find" ]]; then
        failure_types+=("FILE_NOT_FOUND")
        suggested_fixes+=("Check file path or create test fixture")
        return
    fi

    # Network error
    if [[ $message =~ "connection refused" || $message =~ "no route to host" ]]; then
        failure_types+=("NETWORK_ERROR")
        suggested_fixes+=("Use mock server or check network configuration")
        return
    fi

    # Default
    failure_types+=("UNKNOWN")
    suggested_fixes+=("Review test logic and error message")
}

while IFS= read -r line; do
    if [[ $line =~ FAIL:\ (.+) ]]; then
        test_name="${BASH_REMATCH[1]}"
        failed_tests+=("$test_name")

        # Look for error message in next few lines
        error_msg=""
        for ((i=0; i<10; i++)); do
            read -r next_line || break
            if [[ $next_line =~ Error:\ (.+) ]] || [[ $next_line =~ panic:\ (.+) ]]; then
                error_msg="${BASH_REMATCH[1]}"
                break
            fi
        done

        failure_messages+=("$error_msg")
        analyze_failure "$error_msg" "$test_name"
        total_failures=$((total_failures + 1))
    fi
done < "$TEMP_OUTPUT"

echo ""
echo "=========================================="
echo "📊 Failure Analysis"
echo "=========================================="
echo -e "Total Failures:    $total_failures"

# Count by type
declare -A type_counts
for type in "${failure_types[@]}"; do
    type_counts[$type]=$((${type_counts[$type]:-0} + 1))
done

echo ""
echo "Failure Types:"
for type in "${!type_counts[@]}"; do
    echo -e "  $type: ${type_counts[$type]}"
done

# Generate HTML Report
cat > "$OUTPUT_FILE" <<'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>Automated Test Repair Suggestions</title>
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
        .summary {
            padding: 30px;
            background: #f8f9fa;
            text-align: center;
        }
        .summary-value {
            font-size: 72px;
            font-weight: bold;
            color: #dc2626;
        }
        .content { padding: 40px; }

        .failure-item {
            background: #fef2f2;
            border-left: 4px solid #dc2626;
            border-radius: 8px;
            padding: 20px;
            margin: 20px 0;
        }

        .failure-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 15px;
        }

        .test-name {
            font-family: 'Courier New', monospace;
            font-weight: bold;
            color: #1f2937;
        }

        .failure-badge {
            background: #dc2626;
            color: white;
            padding: 6px 12px;
            border-radius: 12px;
            font-size: 11px;
            font-weight: bold;
        }

        .error-message {
            background: #2d1517;
            color: #fca5a5;
            padding: 15px;
            border-radius: 6px;
            font-family: 'Courier New', monospace;
            font-size: 13px;
            margin: 10px 0;
            overflow-x: auto;
        }

        .suggestion-box {
            background: #eff6ff;
            border-left: 4px solid #3b82f6;
            padding: 15px;
            border-radius: 6px;
            margin-top: 15px;
        }

        .suggestion-box h4 {
            color: #1e40af;
            margin: 0 0 10px 0;
        }

        .suggestion-box ul {
            margin: 10px 0;
            padding-left: 20px;
        }

        .suggestion-box li {
            margin: 8px 0;
            color: #1f2937;
        }

        .code-example {
            background: #1f2937;
            color: #10b981;
            padding: 12px;
            border-radius: 4px;
            font-family: 'Courier New', monospace;
            font-size: 12px;
            margin: 8px 0;
        }

        .quick-fix {
            background: #10b981;
            color: white;
            padding: 8px 16px;
            border: none;
            border-radius: 6px;
            cursor: pointer;
            font-weight: bold;
            margin-top: 10px;
        }

        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
            margin: 20px 0;
        }

        .stat-card {
            background: white;
            padding: 20px;
            border-radius: 8px;
            border: 1px solid #e5e7eb;
            text-align: center;
        }

        .stat-value {
            font-size: 32px;
            font-weight: bold;
            color: #6366f1;
        }

        .stat-label {
            color: #6b7280;
            font-size: 14px;
            margin-top: 5px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔧 Automated Test Repair Suggestions</h1>
            <p>Intelligent Failure Analysis & Fix Recommendations</p>
        </div>

        <div class="summary">
            <div class="summary-value">EOF

echo "$total_failures" >> "$OUTPUT_FILE"

cat >> "$OUTPUT_FILE" <<'EOF'
</div>
            <p style="color: #6b7280; font-size: 18px; margin-top: 10px;">
                Failing Tests Detected
            </p>
        </div>

        <div class="content">
            <h2>📊 Failure Type Distribution</h2>
            <div class="stats-grid">
EOF

# Add type counts
for type in "${!type_counts[@]}"; do
    count="${type_counts[$type]}"
    cat >> "$OUTPUT_FILE" <<TYPEEOF
                <div class="stat-card">
                    <div class="stat-value">$count</div>
                    <div class="stat-label">$type</div>
                </div>
TYPEEOF
done

cat >> "$OUTPUT_FILE" <<'EOF'
            </div>

            <h2>🔍 Detailed Failure Analysis & Repair Suggestions</h2>
EOF

# Add each failure with suggestions
for ((i=0; i<${#failed_tests[@]}; i++)); do
    test="${failed_tests[$i]}"
    message="${failure_messages[$i]}"
    type="${failure_types[$i]}"
    fix="${suggested_fixes[$i]}"

    cat >> "$OUTPUT_FILE" <<FAILEOF
            <div class="failure-item">
                <div class="failure-header">
                    <div class="test-name">$test</div>
                    <div class="failure-badge">$type</div>
                </div>

                <div class="error-message">$message</div>

                <div class="suggestion-box">
                    <h4>💡 Suggested Fix</h4>
                    <p>$fix</p>

                    <div class="code-example">// Recommended approach:<br>// $fix</div>

                    <button class="quick-fix" onclick="alert('Copy suggested fix and apply to test')">
                        Apply Quick Fix
                    </button>
                </div>
            </div>
FAILEOF
done

if [ "$total_failures" -eq 0 ]; then
    cat >> "$OUTPUT_FILE" <<'EOF'
            <div style="text-align: center; padding: 60px; color: #10b981;">
                <div style="font-size: 64px; margin-bottom: 20px;">🎉</div>
                <h2>All Tests Passing!</h2>
                <p>No test failures detected. Great job!</p>
            </div>
EOF
fi

cat >> "$OUTPUT_FILE" <<'EOF'
        </div>
    </div>
</body>
</html>
EOF

rm -f "$TEMP_OUTPUT"

echo ""
if [ "$total_failures" -eq 0 ]; then
    echo -e "${GREEN}✅ All tests passing!${NC}"
else
    echo -e "${YELLOW}⚠️  Found $total_failures failing tests${NC}"
    echo -e "${BLUE}📄 Repair suggestions: $OUTPUT_FILE${NC}"
fi

echo ""
echo -e "${BLUE}🌐 View report:${NC}"
echo -e "  open $OUTPUT_FILE"

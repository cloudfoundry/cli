#!/bin/bash
# Test Code Duplication Detector
# Finds duplicated test code and suggests refactoring opportunities

set -e

REPORT_DIR="test-reports/duplication"
OUTPUT_FILE="$REPORT_DIR/duplication-report.html"

mkdir -p "$REPORT_DIR"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "=========================================="
echo "🔍 Test Code Duplication Detector"
echo "=========================================="
echo ""

# Install dupl if not available
if ! command -v dupl &> /dev/null; then
    echo -e "${YELLOW}📦 Installing dupl...${NC}"
    go install github.com/mibk/dupl@latest
    export PATH=$PATH:$(go env GOPATH)/bin
fi

echo "🔍 Scanning for duplicated test code..."

# Run dupl on test files
TEMP_OUTPUT=$(mktemp)
dupl -threshold 15 -files '*_test.go' . > "$TEMP_OUTPUT" 2>&1 || true

# Parse results
total_duplications=0
duplicate_lines=0

while IFS= read -r line; do
    if [[ $line =~ found\ ([0-9]+)\ clones ]]; then
        count="${BASH_REMATCH[1]}"
        total_duplications=$((total_duplications + count))
    elif [[ $line =~ ([0-9]+)\ lines ]]; then
        lines="${BASH_REMATCH[1]}"
        duplicate_lines=$((duplicate_lines + lines))
    fi
done < "$TEMP_OUTPUT"

echo ""
echo "=========================================="
echo "📊 Duplication Statistics"
echo "=========================================="
echo -e "Duplicate Blocks:  $total_duplications"
echo -e "Duplicate Lines:   $duplicate_lines"
echo ""

# Generate HTML Report
cat > "$OUTPUT_FILE" <<'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>Test Code Duplication Report</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0;
            background: linear-gradient(135deg, #fa709a 0%, #fee140 100%);
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
            background: linear-gradient(135deg, #fa709a 0%, #fee140 100%);
            color: white;
            padding: 40px;
            text-align: center;
        }
        .header h1 { font-size: 48px; margin-bottom: 10px; }
        .stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
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

        .duplication-block {
            background: #fef2f2;
            border-left: 4px solid #dc2626;
            border-radius: 8px;
            padding: 20px;
            margin: 20px 0;
        }

        .code-block {
            background: #1f2937;
            color: #e5e7eb;
            padding: 15px;
            border-radius: 6px;
            font-family: 'Courier New', monospace;
            font-size: 13px;
            overflow-x: auto;
            margin: 10px 0;
        }

        .file-ref {
            color: #3b82f6;
            font-family: 'Courier New', monospace;
            font-size: 13px;
            margin: 8px 0;
        }

        .recommendation {
            background: #eff6ff;
            border-left: 4px solid #3b82f6;
            padding: 15px;
            margin: 15px 0;
            border-radius: 6px;
        }

        .recommendation h4 {
            color: #1e40af;
            margin: 0 0 10px 0;
        }

        .severity {
            display: inline-block;
            padding: 4px 12px;
            border-radius: 12px;
            font-size: 11px;
            font-weight: bold;
            text-transform: uppercase;
            margin-left: 10px;
        }

        .severity.high { background: #dc2626; color: white; }
        .severity.medium { background: #ea580c; color: white; }
        .severity.low { background: #65a30d; color: white; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔍 Test Code Duplication Report</h1>
            <p>Duplicated Test Code Detection & Refactoring Suggestions</p>
        </div>

        <div class="stats">
            <div class="stat-card">
                <div class="value">EOF

echo "$total_duplications" >> "$OUTPUT_FILE"

cat >> "$OUTPUT_FILE" <<EOF
</div>
                <div>Duplicate Blocks</div>
            </div>
            <div class="stat-card">
                <div class="value">$duplicate_lines</div>
                <div>Duplicate Lines</div>
            </div>
        </div>

        <div class="content">
            <h2>🎯 Refactoring Recommendations</h2>

            <div class="recommendation">
                <h4>💡 Extract Common Test Setup</h4>
                <p>Create helper functions or use BeforeEach() for repeated setup code:</p>
                <div class="code-block">
BeforeEach(func() {<br>
&nbsp;&nbsp;&nbsp;&nbsp;testData = setupCommonTestData()<br>
&nbsp;&nbsp;&nbsp;&nbsp;mockServer = startMockServer()<br>
})
                </div>
            </div>

            <div class="recommendation">
                <h4>💡 Use Table-Driven Tests</h4>
                <p>Convert repeated test patterns to table-driven tests:</p>
                <div class="code-block">
DescribeTable("user validation",<br>
&nbsp;&nbsp;&nbsp;&nbsp;func(input string, expected bool) {<br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;result := ValidateUser(input)<br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Expect(result).To(Equal(expected))<br>
&nbsp;&nbsp;&nbsp;&nbsp;},<br>
&nbsp;&nbsp;&nbsp;&nbsp;Entry("valid user", "john@example.com", true),<br>
&nbsp;&nbsp;&nbsp;&nbsp;Entry("invalid email", "not-an-email", false),<br>
)
                </div>
            </div>

            <div class="recommendation">
                <h4>💡 Create Test Fixtures</h4>
                <p>Use testhelpers/fixtures for commonly used test data</p>
            </div>

            <div class="recommendation">
                <h4>💡 Build Reusable Matchers</h4>
                <p>Create custom Gomega matchers for repeated assertions</p>
            </div>

            <h2>📋 Detected Duplications</h2>
EOF

# Parse and display duplications
if [ "$total_duplications" -eq 0 ]; then
    cat >> "$OUTPUT_FILE" <<'EOF'
            <div style="text-align: center; padding: 60px; color: #10b981;">
                <div style="font-size: 64px; margin-bottom: 20px;">🎉</div>
                <h2>No Significant Duplication!</h2>
                <p>Your test code is well-refactored. Great job!</p>
            </div>
EOF
else
    # Add duplication details from dupl output
    cat >> "$OUTPUT_FILE" <<'EOF'
            <div class="duplication-block">
                <strong>Duplicated Test Code Found</strong>
                <span class="severity high">HIGH</span>
                <div class="code-block">
EOF

    # Include dupl output (truncated for HTML safety)
    sed 's/&/\&amp;/g; s/</\&lt;/g; s/>/\&gt;/g' "$TEMP_OUTPUT" | head -100 >> "$OUTPUT_FILE"

    cat >> "$OUTPUT_FILE" <<'EOF'
                </div>
                <div class="recommendation">
                    <h4>💡 Refactoring Suggestion</h4>
                    <p>Extract duplicated code into helper functions in testhelpers/ package</p>
                </div>
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
if [ "$total_duplications" -eq 0 ]; then
    echo -e "${GREEN}✅ No significant duplication found!${NC}"
else
    echo -e "${YELLOW}⚠️  Found $total_duplications duplicate blocks${NC}"
    echo -e "${BLUE}💡 Consider refactoring to reduce duplication${NC}"
fi

echo ""
echo -e "${BLUE}📄 Report: $OUTPUT_FILE${NC}"
echo -e "${BLUE}🌐 View report:${NC}"
echo -e "  open $OUTPUT_FILE"

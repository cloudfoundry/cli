#!/bin/bash
# Security Vulnerability Scanner for Tests
# Scans test code for security vulnerabilities and unsafe patterns

set -e

REPORT_DIR="test-reports/security"
OUTPUT_FILE="$REPORT_DIR/security-report.html"

mkdir -p "$REPORT_DIR"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "=========================================="
echo "🔒 Security Vulnerability Scanner"
echo "=========================================="
echo ""

# Install gosec if not available
if ! command -v gosec &> /dev/null; then
    echo -e "${YELLOW}📦 Installing gosec...${NC}"
    go install github.com/securego/gosec/v2/cmd/gosec@latest
    export PATH=$PATH:$(go env GOPATH)/bin
fi

echo "🔍 Scanning for security vulnerabilities..."

# Run gosec
TEMP_JSON=$(mktemp)
TEMP_OUTPUT=$(mktemp)

gosec -fmt=json -out="$TEMP_JSON" ./... 2>&1 | tee "$TEMP_OUTPUT" || true

# Parse results
total_issues=0
high_severity=0
medium_severity=0
low_severity=0

if [ -f "$TEMP_JSON" ] && [ -s "$TEMP_JSON" ]; then
    # Use jq to parse JSON if available, otherwise use grep
    if command -v jq &> /dev/null; then
        total_issues=$(jq -r '.Issues | length' "$TEMP_JSON" 2>/dev/null || echo "0")
        high_severity=$(jq -r '[.Issues[] | select(.severity == "HIGH")] | length' "$TEMP_JSON" 2>/dev/null || echo "0")
        medium_severity=$(jq -r '[.Issues[] | select(.severity == "MEDIUM")] | length' "$TEMP_JSON" 2>/dev/null || echo "0")
        low_severity=$(jq -r '[.Issues[] | select(.severity == "LOW")] | length' "$TEMP_JSON" 2>/dev/null || echo "0")
    else
        total_issues=$(grep -c '"severity"' "$TEMP_JSON" 2>/dev/null || echo "0")
        high_severity=$(grep -c '"severity": "HIGH"' "$TEMP_JSON" 2>/dev/null || echo "0")
        medium_severity=$(grep -c '"severity": "MEDIUM"' "$TEMP_JSON" 2>/dev/null || echo "0")
        low_severity=$(grep -c '"severity": "LOW"' "$TEMP_JSON" 2>/dev/null || echo "0")
    fi
fi

echo ""
echo "=========================================="
echo "📊 Security Scan Results"
echo "=========================================="
echo -e "Total Issues:      $total_issues"
echo -e "${RED}High Severity:     $high_severity${NC}"
echo -e "${YELLOW}Medium Severity:   $medium_severity${NC}"
echo -e "${GREEN}Low Severity:      $low_severity${NC}"
echo ""

# Additional custom checks for test-specific issues
echo "🔍 Running custom test security checks..."

test_specific_issues=0

# Check for hardcoded credentials in tests
cred_issues=$(grep -r "password.*=.*\"" --include="*_test.go" . 2>/dev/null | wc -l || echo "0")
test_specific_issues=$((test_specific_issues + cred_issues))

# Check for SQL injection patterns in tests
sql_issues=$(grep -r "fmt.Sprintf.*SELECT\|fmt.Sprintf.*INSERT" --include="*_test.go" . 2>/dev/null | wc -l || echo "0")
test_specific_issues=$((test_specific_issues + sql_issues))

# Check for insecure random usage
rand_issues=$(grep -r "math/rand" --include="*_test.go" . | grep -v "rand.Seed" | wc -l || echo "0")

echo -e "Test-specific issues: $test_specific_issues"

# Generate HTML Report
cat > "$OUTPUT_FILE" <<'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>Security Vulnerability Report</title>
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

        .vulnerability {
            border-left: 4px solid #dc2626;
            background: #fef2f2;
            padding: 20px;
            margin: 15px 0;
            border-radius: 8px;
        }

        .severity-badge {
            display: inline-block;
            padding: 6px 12px;
            border-radius: 12px;
            font-size: 11px;
            font-weight: bold;
            text-transform: uppercase;
            color: white;
        }

        .severity-badge.high { background: #dc2626; }
        .severity-badge.medium { background: #ea580c; }
        .severity-badge.low { background: #65a30d; }

        .code-snippet {
            background: #1f2937;
            color: #e5e7eb;
            padding: 15px;
            border-radius: 6px;
            font-family: 'Courier New', monospace;
            font-size: 13px;
            margin: 10px 0;
            overflow-x: auto;
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

        .success-message {
            text-align: center;
            padding: 60px;
            color: #10b981;
        }

        .success-message .icon {
            font-size: 64px;
            margin-bottom: 20px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔒 Security Vulnerability Report</h1>
            <p>Test Code Security Analysis</p>
        </div>

        <div class="stats">
            <div class="stat-card total">
                <div class="value">EOF

echo "$total_issues" >> "$OUTPUT_FILE"

cat >> "$OUTPUT_FILE" <<EOF
</div>
                <div>Total Issues</div>
            </div>
            <div class="stat-card high">
                <div class="value">$high_severity</div>
                <div>High Severity</div>
            </div>
            <div class="stat-card medium">
                <div class="value">$medium_severity</div>
                <div>Medium Severity</div>
            </div>
            <div class="stat-card low">
                <div class="value">$low_severity</div>
                <div>Low Severity</div>
            </div>
        </div>

        <div class="content">
            <h2>🛡️ Security Best Practices for Tests</h2>

            <div class="recommendation">
                <h4>1. Never Hardcode Credentials</h4>
                <p>Use environment variables or test fixtures instead:</p>
                <div class="code-snippet">
// Bad<br>
password := "mysecretpassword123"<br>
<br>
// Good<br>
password := os.Getenv("TEST_PASSWORD")
                </div>
            </div>

            <div class="recommendation">
                <h4>2. Use crypto/rand for Security-Critical Tests</h4>
                <p>Don't use math/rand for cryptographic operations:</p>
                <div class="code-snippet">
// Bad<br>
import "math/rand"<br>
token := rand.Int()<br>
<br>
// Good<br>
import "crypto/rand"<br>
token, _ := rand.Int(rand.Reader, big.NewInt(1000000))
                </div>
            </div>

            <div class="recommendation">
                <h4>3. Validate All External Input</h4>
                <p>Even in tests, validate data from external sources</p>
            </div>

            <div class="recommendation">
                <h4>4. Avoid SQL Injection in Test Queries</h4>
                <p>Use parameterized queries even in tests</p>
            </div>

            <h2>🔍 Detected Vulnerabilities</h2>
EOF

if [ "$total_issues" -eq 0 ] && [ "$test_specific_issues" -eq 0 ]; then
    cat >> "$OUTPUT_FILE" <<'EOF'
            <div class="success-message">
                <div class="icon">🎉</div>
                <h2>No Security Issues Found!</h2>
                <p>Your test code follows security best practices.</p>
            </div>
EOF
else
    # Add gosec output
    if [ -f "$TEMP_JSON" ] && [ -s "$TEMP_JSON" ]; then
        cat >> "$OUTPUT_FILE" <<'EOF'
            <div class="vulnerability">
                <strong>GoSec Report</strong>
                <span class="severity-badge high">SECURITY</span>
                <div class="code-snippet">
EOF
        # Safely embed JSON output
        if command -v jq &> /dev/null; then
            jq -r '.Issues[] | "File: \(.file):\(.line)\nSeverity: \(.severity)\nRule: \(.rule_id)\nDetails: \(.details)\n"' "$TEMP_JSON" 2>/dev/null | sed 's/&/\&amp;/g; s/</\&lt;/g; s/>/\&gt;/g' | head -100 >> "$OUTPUT_FILE" || true
        fi

        cat >> "$OUTPUT_FILE" <<'EOF'
                </div>
            </div>
EOF
    fi

    # Add test-specific issues
    if [ "$cred_issues" -gt 0 ]; then
        cat >> "$OUTPUT_FILE" <<EOF
            <div class="vulnerability">
                <strong>Hardcoded Credentials Detected</strong>
                <span class="severity-badge high">HIGH</span>
                <p>Found $cred_issues instances of potential hardcoded credentials in tests.</p>
                <div class="recommendation">
                    <h4>💡 Recommendation</h4>
                    <p>Use environment variables or test fixtures for credentials</p>
                </div>
            </div>
EOF
    fi
fi

cat >> "$OUTPUT_FILE" <<'EOF'
        </div>
    </div>
</body>
</html>
EOF

rm -f "$TEMP_JSON" "$TEMP_OUTPUT"

echo ""
if [ "$total_issues" -eq 0 ] && [ "$test_specific_issues" -eq 0 ]; then
    echo -e "${GREEN}✅ No security issues found!${NC}"
else
    echo -e "${RED}⚠️  Found security issues${NC}"
    echo -e "${YELLOW}Please review and fix the identified vulnerabilities${NC}"
fi

echo ""
echo -e "${BLUE}📄 Report: $OUTPUT_FILE${NC}"
echo -e "${BLUE}🌐 View report:${NC}"
echo -e "  open $OUTPUT_FILE"

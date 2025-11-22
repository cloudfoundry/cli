#!/bin/bash
# Test Impact Analysis
# Determines which tests to run based on code changes

set -e

BASE_BRANCH="${1:-master}"
REPORT_DIR="test-reports/test-impact"
OUTPUT_FILE="$REPORT_DIR/impact-analysis.html"

mkdir -p "$REPORT_DIR"

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "=========================================="
echo "🎯 Test Impact Analysis"
echo "=========================================="
echo "Base branch: $BASE_BRANCH"
echo ""

# Get changed files
echo "📝 Analyzing changed files..."
git fetch origin $BASE_BRANCH > /dev/null 2>&1 || true
CHANGED_FILES=$(git diff --name-only origin/$BASE_BRANCH...HEAD | grep '\.go$' | grep -v '_test\.go$' || true)

if [ -z "$CHANGED_FILES" ]; then
    echo -e "${YELLOW}No Go source files changed${NC}"
    echo "All tests should be run"
    exit 0
fi

echo -e "${BLUE}Changed files:${NC}"
echo "$CHANGED_FILES" | sed 's/^/  - /'
echo ""

# Build dependency graph
echo "🔍 Building test dependency graph..."

declare -A affected_packages
declare -A test_files
total_changed=0
total_tests=0

# For each changed file, find which packages import it
for file in $CHANGED_FILES; do
    total_changed=$((total_changed + 1))
    package=$(dirname "$file")

    # Find direct test file
    base_name=$(basename "$file" .go)
    test_file="${package}/${base_name}_test.go"

    if [ -f "$test_file" ]; then
        test_files["$test_file"]=1
        total_tests=$((total_tests + 1))
        echo "  ✓ Direct test: $test_file"
    fi

    # Find packages that import this package
    affected_packages["$package"]=1

    # Use go list to find importers
    importers=$(go list -f '{{.ImportPath}}' ./... 2>/dev/null | while read pkg; do
        if go list -f '{{range .Imports}}{{.}} {{end}}' "$pkg" 2>/dev/null | grep -q "$package"; then
            echo "$pkg"
        fi
    done)

    for importer in $importers; do
        affected_packages["$importer"]=1

        # Find test files in importing package
        importer_path=$(echo "$importer" | sed "s|github.com/cloudfoundry/cli/||")
        for tfile in $(find "$importer_path" -name "*_test.go" 2>/dev/null || true); do
            test_files["$tfile"]=1
            total_tests=$((total_tests + 1))
        done
    done
done

echo ""
echo "=========================================="
echo "📊 Impact Analysis Results"
echo "=========================================="
echo -e "Changed files:     $total_changed"
echo -e "Affected packages: ${#affected_packages[@]}"
echo -e "Tests to run:      $total_tests"
echo ""

# Calculate test reduction
total_test_files=$(find . -name "*_test.go" | wc -l)
reduction=$(awk "BEGIN {printf \"%.1f\", (1 - ($total_tests / $total_test_files)) * 100}")

echo -e "${GREEN}Test Reduction: $reduction%${NC}"
echo ""

# Generate test command
echo "🚀 Recommended test command:"
echo ""

if [ $total_tests -eq 0 ]; then
    echo "  # No specific tests identified - run all tests"
    echo "  ginkgo -r"
else
    echo "  # Run only affected tests:"
    for pkg in "${!affected_packages[@]}"; do
        echo "  ginkgo $pkg"
    done
fi

echo ""

# Generate HTML report
cat > "$OUTPUT_FILE" <<'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>Test Impact Analysis</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 0;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
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
            font-size: 42px;
            font-weight: bold;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
        }
        .content { padding: 40px; }
        .section { margin: 30px 0; }
        .section h2 {
            font-size: 28px;
            color: #333;
            border-bottom: 3px solid #667eea;
            padding-bottom: 10px;
            margin-bottom: 20px;
        }
        .file-list {
            background: #f8f9fa;
            border-radius: 8px;
            padding: 20px;
            margin: 15px 0;
        }
        .file-item {
            padding: 8px 0;
            border-bottom: 1px solid #dee2e6;
        }
        .file-item:last-child { border-bottom: none; }
        .command-box {
            background: #1e1e1e;
            color: #d4d4d4;
            padding: 20px;
            border-radius: 8px;
            font-family: 'Courier New', monospace;
            overflow-x: auto;
            margin: 15px 0;
        }
        .command-box .comment { color: #6a9955; }
        .reduction {
            font-size: 64px;
            font-weight: bold;
            text-align: center;
            margin: 30px 0;
            background: linear-gradient(135deg, #4CAF50 0%, #8BC34A 100%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
        }
        .benefit {
            background: #d4edda;
            border-left: 4px solid #28a745;
            padding: 15px;
            margin: 10px 0;
            border-radius: 4px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎯 Test Impact Analysis</h1>
            <p>Smart Test Selection Based on Code Changes</p>
        </div>

        <div class="stats">
            <div class="stat-card">
                <div class="value">$total_changed</div>
                <div class="label">Changed Files</div>
            </div>
            <div class="stat-card">
                <div class="value">${#affected_packages[@]}</div>
                <div class="label">Affected Packages</div>
            </div>
            <div class="stat-card">
                <div class="value">$total_tests</div>
                <div class="label">Tests to Run</div>
            </div>
            <div class="stat-card">
                <div class="value">$total_test_files</div>
                <div class="label">Total Tests</div>
            </div>
        </div>

        <div class="reduction">↓ $reduction% ↓</div>
        <p style="text-align: center; font-size: 24px; color: #666;">Test Reduction</p>

        <div class="content">
            <div class="section">
                <h2>Changed Files</h2>
                <div class="file-list">
EOF

echo "$CHANGED_FILES" | while read file; do
    echo "                    <div class=\"file-item\">📝 $file</div>" >> "$OUTPUT_FILE"
done

cat >> "$OUTPUT_FILE" <<'EOF'
                </div>
            </div>

            <div class="section">
                <h2>Affected Packages</h2>
                <div class="file-list">
EOF

for pkg in "${!affected_packages[@]}"; do
    echo "                    <div class=\"file-item\">📦 $pkg</div>" >> "$OUTPUT_FILE"
done

cat >> "$OUTPUT_FILE" <<'EOF'
                </div>
            </div>

            <div class="section">
                <h2>Recommended Test Command</h2>
                <div class="command-box">
EOF

if [ $total_tests -eq 0 ]; then
    echo "                    <span class=\"comment\"># No specific tests identified - run all tests</span><br>" >> "$OUTPUT_FILE"
    echo "                    ginkgo -r" >> "$OUTPUT_FILE"
else
    echo "                    <span class=\"comment\"># Run only affected tests:</span><br>" >> "$OUTPUT_FILE"
    for pkg in "${!affected_packages[@]}"; do
        echo "                    ginkgo $pkg<br>" >> "$OUTPUT_FILE"
    done
fi

cat >> "$OUTPUT_FILE" <<'EOF'
                </div>
            </div>

            <div class="section">
                <h2>Benefits</h2>
                <div class="benefit">
                    <strong>⚡ Faster Feedback:</strong> Run only relevant tests, get results faster
                </div>
                <div class="benefit">
                    <strong>💰 Cost Savings:</strong> Reduce CI/CD compute time and costs
                </div>
                <div class="benefit">
                    <strong>🎯 Focused Testing:</strong> Test what changed, not everything
                </div>
                <div class="benefit">
                    <strong>🔄 Better CI/CD:</strong> Faster iteration cycles for developers
                </div>
            </div>

            <div class="section">
                <h2>How It Works</h2>
                <ol style="line-height: 2;">
                    <li><strong>Detect Changes:</strong> Compare current branch with base branch</li>
                    <li><strong>Build Graph:</strong> Analyze Go import dependencies</li>
                    <li><strong>Find Tests:</strong> Identify direct and indirect test files</li>
                    <li><strong>Optimize:</strong> Run only tests that could be affected</li>
                </ol>
            </div>
        </div>
    </div>
</body>
</html>
EOF

echo "✅ Test impact analysis complete!"
echo "📄 Report: $OUTPUT_FILE"
echo ""
echo "💡 Tip: Use this in CI to run only affected tests on PRs"

#!/bin/bash
# Mutation Testing Framework for Cloud Foundry CLI
# This script performs mutation testing to validate test quality

set -e

MUTATION_DIR=".mutations"
REPORT_DIR="test-reports/mutations"
PACKAGE="${1:-./cf/errors}"

echo "=========================================="
echo "🧬 Mutation Testing Framework"
echo "=========================================="
echo "Package: $PACKAGE"
echo ""

mkdir -p "$MUTATION_DIR"
mkdir -p "$REPORT_DIR"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Mutation operators
declare -A MUTATIONS=(
    ["=="]="!="
    ["!="]="=="
    [">"]="<"
    ["<"]=">"
    [">="]="<="
    ["<="]=">="
    ["&&"]="||"
    ["||"]="&&"
    ["true"]="false"
    ["false"]="true"
    ["+"]="–"
    ["-"]="+"
    ["*"]="/"
    ["/"]="*"
)

total_mutations=0
killed_mutations=0
survived_mutations=0

# Function to apply mutation to a file
apply_mutation() {
    local file=$1
    local from=$2
    local to=$3
    local line=$4

    cp "$file" "$file.backup"

    # Apply mutation at specific line
    sed -i "${line}s/${from}/${to}/1" "$file"
}

# Function to revert mutation
revert_mutation() {
    local file=$1
    mv "$file.backup" "$file"
}

# Function to run tests
run_tests() {
    local package=$1

    # Run tests and capture exit code
    if ginkgo "$package" > /dev/null 2>&1; then
        return 0  # Tests passed
    else
        return 1  # Tests failed
    fi
}

echo "🔍 Analyzing source files..."

# Find all Go source files in package
for source_file in $(find "$PACKAGE" -name "*.go" ! -name "*_test.go"); do
    echo ""
    echo "📄 Mutating: $source_file"

    # Count operators in file
    file_mutations=0

    for operator in "${!MUTATIONS[@]}"; do
        # Find lines with operator
        while IFS= read -r line_num; do
            if [ -n "$line_num" ]; then
                total_mutations=$((total_mutations + 1))
                file_mutations=$((file_mutations + 1))

                mutant="${MUTATIONS[$operator]}"

                echo -n "  Mutation #$total_mutations: Line $line_num: $operator → $mutant ... "

                # Apply mutation
                apply_mutation "$source_file" "$operator" "$mutant" "$line_num"

                # Run tests
                if run_tests "$PACKAGE"; then
                    # Tests still pass - mutation survived!
                    echo -e "${RED}SURVIVED${NC} ❌"
                    survived_mutations=$((survived_mutations + 1))

                    # Log survived mutation
                    echo "$source_file:$line_num: $operator → $mutant (SURVIVED)" >> "$REPORT_DIR/survived.txt"
                else
                    # Tests failed - mutation killed!
                    echo -e "${GREEN}KILLED${NC} ✓"
                    killed_mutations=$((killed_mutations + 1))
                fi

                # Revert mutation
                revert_mutation "$source_file"
            fi
        done < <(grep -n "$operator" "$source_file" | cut -d: -f1)
    done

    echo "  Applied $file_mutations mutations to this file"
done

# Calculate mutation score
if [ $total_mutations -gt 0 ]; then
    mutation_score=$((killed_mutations * 100 / total_mutations))
else
    mutation_score=0
fi

# Generate report
echo ""
echo "=========================================="
echo "📊 Mutation Testing Results"
echo "=========================================="
echo -e "Total Mutations:     $total_mutations"
echo -e "${GREEN}Killed Mutations:    $killed_mutations${NC}"
echo -e "${RED}Survived Mutations:  $survived_mutations${NC}"
echo ""
echo -e "Mutation Score:      ${mutation_score}%"
echo ""

if [ $mutation_score -ge 80 ]; then
    echo -e "${GREEN}✓ Excellent test quality!${NC}"
elif [ $mutation_score -ge 60 ]; then
    echo -e "${YELLOW}⚠ Good test quality, but room for improvement${NC}"
else
    echo -e "${RED}✗ Tests need improvement${NC}"
fi

# Generate HTML report
cat > "$REPORT_DIR/mutation-report.html" <<EOF
<!DOCTYPE html>
<html>
<head>
    <title>Mutation Testing Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; border-bottom: 3px solid #4CAF50; padding-bottom: 10px; }
        .stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin: 30px 0; }
        .stat-card { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 20px; border-radius: 8px; text-align: center; }
        .stat-card h3 { margin: 0; font-size: 16px; opacity: 0.9; }
        .stat-card .value { font-size: 48px; font-weight: bold; margin: 10px 0; }
        .score { font-size: 72px; font-weight: bold; color: #4CAF50; text-align: center; margin: 30px 0; }
        .score.medium { color: #FF9800; }
        .score.low { color: #f44336; }
        .survived { background: #ffebee; padding: 10px; margin: 10px 0; border-left: 4px solid #f44336; border-radius: 4px; }
        .progress-bar { background: #e0e0e0; height: 30px; border-radius: 15px; overflow: hidden; margin: 20px 0; }
        .progress-fill { background: linear-gradient(90deg, #4CAF50 0%, #8BC34A 100%); height: 100%; transition: width 0.3s ease; display: flex; align-items: center; justify-content: center; color: white; font-weight: bold; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🧬 Mutation Testing Report</h1>
        <p><strong>Package:</strong> $PACKAGE</p>
        <p><strong>Generated:</strong> $(date)</p>

        <div class="score $([ $mutation_score -ge 80 ] && echo "" || ([ $mutation_score -ge 60 ] && echo "medium" || echo "low"))">
            ${mutation_score}%
        </div>

        <div class="progress-bar">
            <div class="progress-fill" style="width: ${mutation_score}%">
                Mutation Score
            </div>
        </div>

        <div class="stats">
            <div class="stat-card">
                <h3>Total Mutations</h3>
                <div class="value">$total_mutations</div>
            </div>
            <div class="stat-card" style="background: linear-gradient(135deg, #4CAF50 0%, #8BC34A 100%);">
                <h3>Killed</h3>
                <div class="value">$killed_mutations</div>
            </div>
            <div class="stat-card" style="background: linear-gradient(135deg, #f44336 0%, #e91e63 100%);">
                <h3>Survived</h3>
                <div class="value">$survived_mutations</div>
            </div>
        </div>

        <h2>Analysis</h2>
        <p>
            Mutation testing works by intentionally introducing bugs (mutations) into your code and checking if your tests catch them.
            A high mutation score indicates that your tests are effective at catching bugs.
        </p>

        <h3>Score Interpretation:</h3>
        <ul>
            <li><strong>80-100%:</strong> Excellent - Your tests are very effective</li>
            <li><strong>60-79%:</strong> Good - Tests are effective but can be improved</li>
            <li><strong>Below 60%:</strong> Needs improvement - Many mutations survive</li>
        </ul>

        $(if [ $survived_mutations -gt 0 ]; then
            echo "<h2>⚠️ Survived Mutations</h2>"
            echo "<p>These mutations were not caught by your tests. Consider adding tests to cover these cases:</p>"
            if [ -f "$REPORT_DIR/survived.txt" ]; then
                while IFS= read -r line; do
                    echo "<div class='survived'>$line</div>"
                done < "$REPORT_DIR/survived.txt"
            fi
        fi)
    </div>
</body>
</html>
EOF

echo ""
echo "📄 HTML report generated: $REPORT_DIR/mutation-report.html"
echo ""

# Exit with appropriate code
if [ $mutation_score -ge 80 ]; then
    exit 0
else
    exit 1
fi

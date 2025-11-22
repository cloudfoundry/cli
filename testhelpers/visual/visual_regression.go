package visual

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// VisualTester provides visual regression testing
type VisualTester struct {
	baselineDir string
	currentDir  string
	diffDir     string
}

// NewVisualTester creates a new visual tester
func NewVisualTester() *VisualTester {
	return &VisualTester{
		baselineDir: "testdata/visual/baseline",
		currentDir:  "testdata/visual/current",
		diffDir:     "testdata/visual/diff",
	}
}

// CaptureOutput captures CLI output for visual comparison
func (vt *VisualTester) CaptureOutput(name string, output string) error {
	// Create directories if needed
	os.MkdirAll(vt.currentDir, 0755)
	os.MkdirAll(vt.baselineDir, 0755)
	os.MkdirAll(vt.diffDir, 0755)

	// Save current output
	currentPath := filepath.Join(vt.currentDir, name+".txt")
	if err := ioutil.WriteFile(currentPath, []byte(output), 0644); err != nil {
		return err
	}

	return nil
}

// Compare compares current output with baseline
func (vt *VisualTester) Compare(name string) (*ComparisonResult, error) {
	currentPath := filepath.Join(vt.currentDir, name+".txt")
	baselinePath := filepath.Join(vt.baselineDir, name+".txt")

	// Read current
	current, err := ioutil.ReadFile(currentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read current: %v", err)
	}

	// Read baseline
	baseline, err := ioutil.ReadFile(baselinePath)
	if os.IsNotExist(err) {
		// No baseline - create it
		if err := ioutil.WriteFile(baselinePath, current, 0644); err != nil {
			return nil, err
		}
		return &ComparisonResult{
			Name:         name,
			IsNew:        true,
			Match:        true,
			Message:      "New baseline created",
		}, nil
	}
	if err != nil {
		return nil, err
	}

	// Compare
	currentHash := md5.Sum(current)
	baselineHash := md5.Sum(baseline)

	match := currentHash == baselineHash

	result := &ComparisonResult{
		Name:         name,
		IsNew:        false,
		Match:        match,
		CurrentHash:  fmt.Sprintf("%x", currentHash),
		BaselineHash: fmt.Sprintf("%x", baselineHash),
	}

	if !match {
		result.Message = "Visual regression detected!"
		result.DiffPath = vt.generateDiff(name, string(baseline), string(current))
	} else {
		result.Message = "Output matches baseline"
	}

	return result, nil
}

// generateDiff generates a diff file
func (vt *VisualTester) generateDiff(name, baseline, current string) string {
	diffPath := filepath.Join(vt.diffDir, name+".diff")

	var diff string
	diff += "=== BASELINE ===\n"
	diff += baseline
	diff += "\n\n"
	diff += "=== CURRENT ===\n"
	diff += current
	diff += "\n\n"
	diff += "=== CHANGES ===\n"

	// Simple line-by-line diff
	baselineLines := splitLines(baseline)
	currentLines := splitLines(current)

	maxLines := len(baselineLines)
	if len(currentLines) > maxLines {
		maxLines = len(currentLines)
	}

	for i := 0; i < maxLines; i++ {
		var baseLine, currLine string
		if i < len(baselineLines) {
			baseLine = baselineLines[i]
		}
		if i < len(currentLines) {
			currLine = currentLines[i]
		}

		if baseLine != currLine {
			diff += fmt.Sprintf("Line %d:\n", i+1)
			diff += fmt.Sprintf("- %s\n", baseLine)
			diff += fmt.Sprintf("+ %s\n", currLine)
		}
	}

	ioutil.WriteFile(diffPath, []byte(diff), 0644)

	return diffPath
}

// ComparisonResult contains the result of a visual comparison
type ComparisonResult struct {
	Name         string
	IsNew        bool
	Match        bool
	CurrentHash  string
	BaselineHash string
	Message      string
	DiffPath     string
}

// UpdateBaseline updates the baseline with current output
func (vt *VisualTester) UpdateBaseline(name string) error {
	currentPath := filepath.Join(vt.currentDir, name+".txt")
	baselinePath := filepath.Join(vt.baselineDir, name+".txt")

	current, err := ioutil.ReadFile(currentPath)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(baselinePath, current, 0644)
}

// CleanCurrent removes current outputs
func (vt *VisualTester) CleanCurrent() error {
	return os.RemoveAll(vt.currentDir)
}

// Helper to split text into lines
func splitLines(text string) []string {
	var lines []string
	current := ""

	for _, ch := range text {
		if ch == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(ch)
		}
	}

	if current != "" {
		lines = append(lines, current)
	}

	return lines
}

// VisualReporter generates HTML report for visual regression results
type VisualReporter struct {
	results []*ComparisonResult
}

// NewVisualReporter creates a new reporter
func NewVisualReporter() *VisualReporter {
	return &VisualReporter{
		results: make([]*ComparisonResult, 0),
	}
}

// AddResult adds a comparison result
func (vr *VisualReporter) AddResult(result *ComparisonResult) {
	vr.results = append(vr.results, result)
}

// GenerateReport generates HTML report
func (vr *VisualReporter) GenerateReport(outputPath string) error {
	totalTests := len(vr.results)
	passed := 0
	failed := 0
	new := 0

	for _, r := range vr.results {
		if r.IsNew {
			new++
		} else if r.Match {
			passed++
		} else {
			failed++
		}
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Visual Regression Report</title>
    <style>
        body { font-family: Arial; margin: 20px; background: linear-gradient(135deg, #6366f1 0%%, #8b5cf6 100%%); padding: 20px; }
        .container { max-width: 1200px; margin: 0 auto; background: white; border-radius: 20px; padding: 40px; }
        h1 { color: #333; }
        .stats { display: grid; grid-template-columns: repeat(4, 1fr); gap: 20px; margin: 30px 0; }
        .stat { background: #f8f9fa; padding: 20px; border-radius: 10px; text-align: center; }
        .stat .value { font-size: 48px; font-weight: bold; }
        .stat.total .value { color: #6366f1; }
        .stat.passed .value { color: #10b981; }
        .stat.failed .value { color: #ef4444; }
        .stat.new .value { color: #f59e0b; }
        .result { border: 1px solid #ddd; padding: 20px; margin: 15px 0; border-radius: 8px; }
        .result.pass { border-left: 4px solid #10b981; background: #f0fdf4; }
        .result.fail { border-left: 4px solid #ef4444; background: #fef2f2; }
        .result.new { border-left: 4px solid #f59e0b; background: #fffbeb; }
        .diff-link { color: #6366f1; text-decoration: none; font-weight: bold; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🎨 Visual Regression Report</h1>

        <div class="stats">
            <div class="stat total">
                <div class="value">%d</div>
                <div>Total Tests</div>
            </div>
            <div class="stat passed">
                <div class="value">%d</div>
                <div>Passed</div>
            </div>
            <div class="stat failed">
                <div class="value">%d</div>
                <div>Failed</div>
            </div>
            <div class="stat new">
                <div class="value">%d</div>
                <div>New</div>
            </div>
        </div>

        <h2>Results</h2>
`, totalTests, passed, failed, new)

	for _, r := range vr.results {
		status := "pass"
		icon := "✅"
		if r.IsNew {
			status = "new"
			icon = "🆕"
		} else if !r.Match {
			status = "fail"
			icon = "❌"
		}

		html += fmt.Sprintf(`
        <div class="result %s">
            <h3>%s %s</h3>
            <p>%s</p>
`, status, icon, r.Name, r.Message)

		if !r.Match && !r.IsNew {
			html += fmt.Sprintf(`
            <p><strong>Current Hash:</strong> %s</p>
            <p><strong>Baseline Hash:</strong> %s</p>
            <p><a href="%s" class="diff-link">View Diff</a></p>
`, r.CurrentHash, r.BaselineHash, r.DiffPath)
		}

		html += `        </div>`
	}

	html += `
    </div>
</body>
</html>`

	return ioutil.WriteFile(outputPath, []byte(html), 0644)
}

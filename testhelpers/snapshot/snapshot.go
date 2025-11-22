package snapshot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/onsi/gomega"
)

const snapshotDir = "testdata/snapshots"

// Snapshot provides snapshot testing functionality
type Snapshot struct {
	testName string
	update   bool
}

// New creates a new Snapshot instance
func New(testName string) *Snapshot {
	// Check if UPDATE_SNAPSHOTS environment variable is set
	update := os.Getenv("UPDATE_SNAPSHOTS") == "true"

	return &Snapshot{
		testName: sanitizeTestName(testName),
		update:   update,
	}
}

// MatchSnapshot compares the given data with the stored snapshot
func (s *Snapshot) MatchSnapshot(data interface{}) {
	snapshotPath := s.getSnapshotPath()

	// Convert data to string
	var dataStr string
	switch v := data.(type) {
	case string:
		dataStr = v
	case []byte:
		dataStr = string(v)
	default:
		// Try to marshal as JSON
		jsonBytes, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			panic(fmt.Sprintf("Failed to marshal data to JSON: %v", err))
		}
		dataStr = string(jsonBytes)
	}

	// Ensure snapshot directory exists
	dir := filepath.Dir(snapshotPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create snapshot directory: %v", err))
	}

	// If update mode, write new snapshot
	if s.update {
		if err := ioutil.WriteFile(snapshotPath, []byte(dataStr), 0644); err != nil {
			panic(fmt.Sprintf("Failed to write snapshot: %v", err))
		}
		fmt.Printf("📸 Updated snapshot: %s\n", snapshotPath)
		return
	}

	// Read existing snapshot
	existingSnapshot, err := ioutil.ReadFile(snapshotPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Snapshot doesn't exist, create it
			if err := ioutil.WriteFile(snapshotPath, []byte(dataStr), 0644); err != nil {
				panic(fmt.Sprintf("Failed to create snapshot: %v", err))
			}
			fmt.Printf("📸 Created new snapshot: %s\n", snapshotPath)
			return
		}
		panic(fmt.Sprintf("Failed to read snapshot: %v", err))
	}

	// Compare snapshots
	gomega.Expect(dataStr).To(gomega.Equal(string(existingSnapshot)),
		fmt.Sprintf("Snapshot mismatch for %s\nRun with UPDATE_SNAPSHOTS=true to update", s.testName))
}

// MatchJSONSnapshot compares the given data as formatted JSON
func (s *Snapshot) MatchJSONSnapshot(data interface{}) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		panic(fmt.Sprintf("Failed to marshal JSON: %v", err))
	}

	s.MatchSnapshot(string(jsonBytes))
}

// MatchOutputSnapshot is specifically for command output
func (s *Snapshot) MatchOutputSnapshot(output string) {
	// Normalize line endings
	normalized := strings.ReplaceAll(output, "\r\n", "\n")
	s.MatchSnapshot(normalized)
}

// getSnapshotPath returns the path to the snapshot file
func (s *Snapshot) getSnapshotPath() string {
	return filepath.Join(snapshotDir, fmt.Sprintf("%s.snap", s.testName))
}

// sanitizeTestName converts a test name into a valid filename
func sanitizeTestName(name string) string {
	// Replace invalid characters
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, "*", "_")
	name = strings.ReplaceAll(name, "?", "_")
	name = strings.ReplaceAll(name, "\"", "_")
	name = strings.ReplaceAll(name, "<", "_")
	name = strings.ReplaceAll(name, ">", "_")
	name = strings.ReplaceAll(name, "|", "_")

	// Lowercase and trim
	name = strings.ToLower(name)
	name = strings.Trim(name, "_")

	return name
}

// DiffSnapshots shows a diff between current and stored snapshot
func (s *Snapshot) DiffSnapshots(current, stored string) string {
	var diff bytes.Buffer

	currentLines := strings.Split(current, "\n")
	storedLines := strings.Split(stored, "\n")

	maxLines := len(currentLines)
	if len(storedLines) > maxLines {
		maxLines = len(storedLines)
	}

	diff.WriteString("Snapshot Diff:\n")
	diff.WriteString("─────────────────────────────────────\n")

	for i := 0; i < maxLines; i++ {
		var currentLine, storedLine string

		if i < len(currentLines) {
			currentLine = currentLines[i]
		}
		if i < len(storedLines) {
			storedLine = storedLines[i]
		}

		if currentLine != storedLine {
			diff.WriteString(fmt.Sprintf("Line %d:\n", i+1))
			diff.WriteString(fmt.Sprintf("- %s\n", storedLine))
			diff.WriteString(fmt.Sprintf("+ %s\n", currentLine))
		}
	}

	return diff.String()
}

// CleanSnapshots removes all snapshot files
func CleanSnapshots() error {
	return os.RemoveAll(snapshotDir)
}

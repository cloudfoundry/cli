package glob

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDirGlob(t *testing.T) {
	fixturesDir := globFixturesDir()
	dir := NewDir(fixturesDir)
	patterns := []string{"**/*.stdout.log", "*.stderr.log", "../some.config", "some_directory/**/*"}

	filePaths, err := dir.Glob(patterns...)
	if err != nil {
		t.Fatalf("Expected No Error but got %v", err)
		return
	}

	// regular files
	assertPathIsIncluded(t, filePaths, "app.stdout.log")
	assertPathIsIncluded(t, filePaths, "app.stderr.log")

	// file in a directory
	assertPathIsIncluded(t, filePaths, "other_logs/other_app.stdout.log")

	// file in a sub directory
	assertPathIsIncluded(t, filePaths, "other_logs/more_logs/more.stdout.log")

	// directories
	assertPathIsIncluded(t, filePaths, "some_directory/sub_dir")
	assertPathIsIncluded(t, filePaths, "some_directory/sub_dir/other_sub_dir")

	// file that is not matching filter
	assertPathIsExcluded(t, filePaths, "some_directory")
	assertPathIsExcluded(t, filePaths, "other_logs/other_app.stderr.log")
	assertPathIsExcluded(t, filePaths, "../some.config")
}

func globFixturesDir() string {
	pwd, _ := os.Getwd()
	return filepath.Join(pwd, "..", "fixtures", "test_glob")
}

func assertPathIsIncluded(t *testing.T, filePaths []string, includedPath string) {
	found := findFile(filePaths, includedPath)

	if !found {
		t.Errorf("Expected to find file path `%s`", includedPath)
	}
}

func assertPathIsExcluded(t *testing.T, filePaths []string, excludedPath string) {
	found := findFile(filePaths, excludedPath)

	if found {
		t.Errorf("Expected to NOT find file path `%s`", excludedPath)
	}
}

func findFile(filePaths []string, expectedPath string) bool {
	for _, filePath := range filePaths {
		if filePath == expectedPath {
			return true
		}
	}
	return false
}

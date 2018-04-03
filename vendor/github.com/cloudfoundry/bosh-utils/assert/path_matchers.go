package assert

import (
	"fmt"
	"path/filepath"
	"runtime"
)

const windows = runtime.GOOS == "windows"

// MatchPath is a GomegaMatcher for filepaths, on Unix systems paths are
// compared unmodified.  On Windows, Unix absolute paths (leading '/') are
// converted to Windows absolute paths using the current working directory
// for the volume name.
type MatchPath string

func isSlash(c byte) bool { return c == '\\' || c == '/' }

func (m MatchPath) isAbs(path string) bool {
	return filepath.IsAbs(path) || (windows && len(path) != 0 && isSlash(path[0]))
}

func (m MatchPath) cleanPath(s string) string {
	if !windows || !m.isAbs(s) {
		return s
	}
	a, err := filepath.Abs(s)
	if err != nil {
		return s
	}
	return a
}

func (m MatchPath) Match(actual interface{}) (bool, error) {
	path, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("MatchPath: expects a string got: %T", actual)
	}
	if path == string(m) || (filepath.Clean(path) == filepath.Clean(string(m))) {
		return true, nil
	}
	return m.cleanPath(path) == m.cleanPath(string(m)), nil
}

func (m MatchPath) FailureMessage(actual interface{}) string {
	if windows {
		// show both the provided and cleaned paths
		if s, ok := actual.(string); ok {
			return fmt.Sprintf("Expected\n\t%v\n\t%v (clean)\nto match file\n\t%v\n\t%v (clean)",
				actual, m.cleanPath(s), m, m.cleanPath(string(m)))
		}
	}
	return fmt.Sprintf("Expected\n\t%v\nto match file\n\t%v", actual, m)
}

func (m MatchPath) NegatedFailureMessage(actual interface{}) string {
	if windows {
		// show both the provided and cleaned paths
		if s, ok := actual.(string); ok {
			return fmt.Sprintf("Expected\n\t%v\n\t%v (clean)\nnot to match file\n\t%v\n\t%v (clean)",
				actual, m.cleanPath(s), m, m.cleanPath(string(m)))
		}
	}
	return fmt.Sprintf("Expected\n\t%v\nnot to match file\n\t%v", actual, m)
}

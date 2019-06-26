package matchers

import (
	"fmt"
	"github.com/onsi/gomega"
	"path/filepath"
	"regexp"
	"strings"
)

type PathMatcher struct {
	actualPath   string
	expectedPath string
}

func MatchPath(expectedPath string) gomega.OmegaMatcher {
	return &PathMatcher{expectedPath: expectedPath}
}

func (matcher *PathMatcher) Match(actual interface{}) (success bool, err error) {
	actualPath, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("MatchPath: Actual must be a string, got %T", actual)
	}

	resolvedActualPath, err := resolveSymLinks(actualPath)
	if err != nil {
		return false, err
	}

	resolvedExpectedPath, err := resolveSymLinks(matcher.expectedPath)
	if err != nil {
		return false, err
	}

	matcher.actualPath = resolvedActualPath
	matcher.expectedPath = resolvedExpectedPath

	return strings.EqualFold(regexp.QuoteMeta(matcher.expectedPath), regexp.QuoteMeta(matcher.actualPath)), nil
}

func (matcher *PathMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected:\n %v\n to match actual:\n%v\n", matcher.expectedPath, matcher.actualPath)
}

func (matcher *PathMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected:\n %v\n not to match actual:\n%v\n", matcher.expectedPath, matcher.actualPath)
}

func resolveSymLinks(path string) (string, error) {
	theRealDir, err := filepath.EvalSymlinks(filepath.Dir(path))

	if err != nil {
		return "", err
	}

	return filepath.Join(theRealDir, filepath.Base(path)), nil
}
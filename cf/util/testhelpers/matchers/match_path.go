package matchers

import (
	"fmt"
	"github.com/onsi/gomega"
	"regexp"
	"strings"
)

type PathMatcher struct {
	actualPath   interface{}
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
	matcher.actualPath = actualPath

	return strings.EqualFold(regexp.QuoteMeta(matcher.expectedPath), regexp.QuoteMeta(actualPath)), nil
}

func (matcher *PathMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected:\n %v\n to match actual:\n%v\n", matcher.expectedPath, matcher.actualPath)
}

func (matcher *PathMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected:\n %v\n not to match actual:\n%v\n", matcher.expectedPath, matcher.actualPath)
}

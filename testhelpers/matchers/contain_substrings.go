package matchers

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/onsi/gomega"
)

type SliceMatcher struct {
	expected      [][]string
	failedAtIndex int
}

func ContainSubstrings(substrings ...[]string) gomega.OmegaMatcher {
	return &SliceMatcher{expected: substrings}
}

func (matcher *SliceMatcher) Match(actual interface{}) (success bool, err error) {
	actualStrings, ok := actual.([]string)
	if !ok {
		return false, nil
	}

	matcher.failedAtIndex = 0
	for _, actualValue := range actualStrings {
		allStringsFound := true
		for _, expectedValue := range matcher.expected[matcher.failedAtIndex] {
			allStringsFound = allStringsFound && strings.Contains(terminal.Decolorize(actualValue), expectedValue)
		}

		if allStringsFound {
			matcher.failedAtIndex++
			if matcher.failedAtIndex == len(matcher.expected) {
				return true, nil
			}
		}
	}

	return false, nil
}

func (matcher *SliceMatcher) FailureMessage(actual interface{}) string {
	actualStrings, ok := actual.([]string)
	if !ok {
		return fmt.Sprintf("Expected actual to be a slice of strings, but it's actually a %T", actual)
	}

	return fmt.Sprintf("expected to find \"%s\" in actual:\n'%s'\n", matcher.expected[matcher.failedAtIndex], strings.Join(actualStrings, "\n"))
}

func (matcher *SliceMatcher) NegatedFailureMessage(actual interface{}) string {
	actualStrings, ok := actual.([]string)
	if !ok {
		return fmt.Sprintf("Expected actual to be a slice of strings, but it's actually a %T", actual)
	}
	return fmt.Sprintf("expected to not find \"%s\" in actual:\n'%s'\n", matcher.expected[matcher.failedAtIndex], strings.Join(actualStrings, "\n"))
}

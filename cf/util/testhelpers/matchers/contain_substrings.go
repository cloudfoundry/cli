package matchers

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/cf/terminal"
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

	allStringsMatched := make([]bool, len(matcher.expected))

	for index, expectedArray := range matcher.expected {
		for _, actualValue := range actualStrings {

			allStringsFound := true

			for _, expectedValue := range expectedArray {
				allStringsFound = allStringsFound && strings.Contains(terminal.Decolorize(actualValue), expectedValue)
			}

			if allStringsFound {
				allStringsMatched[index] = true
				break
			}
		}
	}

	for index, value := range allStringsMatched {
		if !value {
			matcher.failedAtIndex = index
			return false, nil
		}
	}

	return true, nil
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

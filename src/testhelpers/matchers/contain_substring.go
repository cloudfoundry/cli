package matchers

import (
	"fmt"
	"github.com/onsi/gomega"
	"strings"
)

type SliceMatcher struct {
	expected [][]string
}

func ContainSubstrings(substrings ...[]string) gomega.OmegaMatcher {
	return SliceMatcher{expected: substrings}
}

func (matcher SliceMatcher) Match(actual interface{}) (success bool, message string, err error) {
	actualStrings, ok := actual.([]string)
	if !ok {
		return false, "must match a slice of strings", nil
	}

	expectedIndex := 0
	for _, actualValue := range actualStrings {
		allStringsFound := true
		for _, expectedValue := range matcher.expected[expectedIndex] {
			allStringsFound = allStringsFound && strings.Contains(strings.ToLower(actualValue), strings.ToLower(expectedValue))
		}

		if allStringsFound {
			expectedIndex++
			if expectedIndex == len(matcher.expected) {
				return true, "", nil
			}
		}
	}

	return false, fmt.Sprintf("\"%s\" not found in actual:\n'%s'\n", matcher.expected[expectedIndex], actual), nil
}

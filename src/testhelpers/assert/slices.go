package assert

import (
	"testing"
	"fmt"
	"strings"
)

type Line []string
func (line Line)String()string{
	return strings.Join(line,", ")
}

type Lines []Line

func SliceContains(t *testing.T, actual []string, expected Lines, msgAndArgs ...interface{}) bool {
	expectedIndex := 0
	for _, actualValue := range actual {
		allStringsFound := true
		for _, expectedValue := range expected[expectedIndex] {
			allStringsFound = allStringsFound && strings.Contains(actualValue, expectedValue)
		}

		if allStringsFound {
			expectedIndex++
			if expectedIndex == len(expected) {
				return true
			}
		}
	}
	return Fail(t, fmt.Sprintf("\"%s\" not found", expected[expectedIndex]), msgAndArgs...)
}

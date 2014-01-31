package assert

import (
	"fmt"
	"strings"
	mr "github.com/tjarratt/mr_t"
)

type Line []string

func (line Line) String() string {
	return strings.Join(line, ", ")
}

type Lines []Line

func SliceContains(t mr.TestingT, actual []string, expected Lines, msgAndArgs ...interface{}) bool {
	expectedIndex := 0
	for _, actualValue := range actual {
		allStringsFound := true
		for _, expectedValue := range expected[expectedIndex] {
			allStringsFound = allStringsFound && strings.Contains(strings.ToLower(actualValue), strings.ToLower(expectedValue))
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

func SliceDoesNotContain(t mr.TestingT, actual []string, expected Lines, msgAndArgs ...interface{}) bool {
	for i, actualValue := range actual {
		for _, expectedLine := range expected {
			allStringsFound := true
			for _, expectedValue := range expectedLine {
				allStringsFound = allStringsFound && strings.Contains(strings.ToLower(actualValue), strings.ToLower(expectedValue))
			}
			if allStringsFound {
				return Fail(t, fmt.Sprintf("\"%s\" found on line %d", expectedLine, i), msgAndArgs...)
			}
		}
	}
	return true
}

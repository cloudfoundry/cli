package assert

import (
	"testing"
	"fmt"
	"strings"
)

func SliceContains(t *testing.T, actual []string, expected []string, msgAndArgs ...interface{}) bool{
	expectedIndex := 0
	for _, actualValue := range actual {
		if strings.Contains(actualValue,expected[expectedIndex]) {
			expectedIndex++
			if expectedIndex == len(expected) {
				return true
			}
		}
	}
	return Fail(t, fmt.Sprintf("\"%s\" not found",expected[expectedIndex]), msgAndArgs...)
}

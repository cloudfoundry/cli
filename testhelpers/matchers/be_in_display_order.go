package matchers

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/cf/terminal"
	"github.com/onsi/gomega"
)

type OrderMatcher struct {
	expected   [][]string
	failedText string
}

func BeInDisplayOrder(substrings ...[]string) gomega.OmegaMatcher {
	return &OrderMatcher{
		expected: substrings,
	}
}

func (matcher *OrderMatcher) Match(actualStr interface{}) (success bool, err error) {
	actual, ok := actualStr.([]string)
	if !ok {
		return false, nil
	}

	//loop and match, stop at last actual[0] line
	for len(actual) > 1 {
		if matched, msg := matchSingleLine(actual[0], matcher.expected[0]); matched {
			if len(matcher.expected) == 1 {
				return true, nil //no more expected to match, all passed
			}
			matcher.expected = matcher.expected[1:]
		} else if msg != "" {
			matcher.failedText = msg
			return false, nil
		}
		actual = actual[1:]
	}

	//match the last actual line with the rest of expected
	matched, msg := matchSingleLine(actual[0], matcher.expected[0])
	if matched && len(matcher.expected) == 1 {
		return true, nil
	} else if msg != "" {
		matcher.failedText = msg
		return false, nil
	} else if matched {
		matcher.failedText = matcher.expected[1][0]
		return false, nil
	}
	matcher.failedText = matcher.expected[0][0]
	return false, nil
}

func matchSingleLine(actual string, expected []string) (bool, string) {
	matched := false
	for i, target := range expected {
		if index := strings.Index(terminal.Decolorize(actual), target); index != -1 {
			if i == len(expected)-1 {
				return true, ""
			}
			matched = true
			actual = actual[index+len(target):]
		} else if matched {
			return false, target
		} else {
			return false, ""
		}
	}
	return false, ""
}

func (matcher *OrderMatcher) FailureMessage(actual interface{}) string {
	actualStrings, ok := actual.([]string)
	if !ok {
		return fmt.Sprintf("Expected actual to be a slice of strings, but it's actually a %T", actual)
	}

	return fmt.Sprintf("expected to find \"%s\" in display order in actual:\n'%s'\n", matcher.failedText, strings.Join(actualStrings, "\n"))
}

func (matcher *OrderMatcher) NegatedFailureMessage(actual interface{}) string {
	actualStrings, ok := actual.([]string)
	if !ok {
		return fmt.Sprintf("Expected actual to be a slice of strings, but it's actually a %T", actual)
	}
	return fmt.Sprintf("expected to not find strings in display order in actual:\n'%s'\n", strings.Join(actualStrings, "\n"))
}

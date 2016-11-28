package matchers

import (
	"fmt"

	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	"github.com/onsi/gomega"
)

type haveSucceededMatcher struct{}

func HaveSucceeded() gomega.OmegaMatcher {
	return haveSucceededMatcher{}
}

func (matcher haveSucceededMatcher) Match(actual interface{}) (bool, error) {
	switch actual.(type) {
	case testcmd.RunCommandResult:
		result := actual.(testcmd.RunCommandResult)
		return result == testcmd.RunCommandResultSuccess, nil
	default:
		return false, fmt.Errorf("Expected actual value to be an enum, but it was a %T", actual)
	}
}

func (matcher haveSucceededMatcher) FailureMessage(_ interface{}) string {
	return "Expected command to have succeeded but it did not"
}

func (matcher haveSucceededMatcher) NegatedFailureMessage(_ interface{}) string {
	return "Expected command to have not succeeded but it did"
}

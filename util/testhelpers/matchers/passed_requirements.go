package matchers

import (
	"fmt"

	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	"github.com/onsi/gomega"
)

type havePassedRequirementsMatcher struct{}

func HavePassedRequirements() gomega.OmegaMatcher {
	return havePassedRequirementsMatcher{}
}

func (matcher havePassedRequirementsMatcher) Match(actual interface{}) (bool, error) {
	switch actual.(type) {
	case bool:
		asBool := actual.(bool)
		return asBool == true, nil
	case testcmd.RunCommandResult:
		result := actual.(testcmd.RunCommandResult)
		return result == testcmd.RunCommandResultSuccess, nil
	default:
		return false, fmt.Errorf("Expected actual value to be a bool or enum, but it was a %T", actual)
	}
}

func (matcher havePassedRequirementsMatcher) FailureMessage(_ interface{}) string {
	return "Expected command to pass requirements but it did not"
}

func (matcher havePassedRequirementsMatcher) NegatedFailureMessage(_ interface{}) string {
	return "Expected command to have not passed requirements but it did"
}

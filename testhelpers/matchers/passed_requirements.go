package matchers

import (
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/onsi/gomega"
)

type havePassedRequirementsMatcher struct{}

func HavePassedRequirements() gomega.OmegaMatcher {
	return havePassedRequirementsMatcher{}
}

func (matcher havePassedRequirementsMatcher) Match(actual interface{}) (bool, error) {
	asBool, ok := actual.(bool)
	if !ok {
		return false, errors.NewWithFmt("Expected actual value to be a bool, but it was a %T", actual)
	}

	return asBool == true, nil
}

func (matcher havePassedRequirementsMatcher) FailureMessage(_ interface{}) string {
	return "Expected command to pass requirements but it did not"
}

func (matcher havePassedRequirementsMatcher) NegatedFailureMessage(_ interface{}) string {
	return "Expected command to have not passed requirements but it did"
}

package matchers

import (
	"fmt"

	"code.cloudfoundry.org/cli/util/testhelpers/net"
	"github.com/onsi/gomega"
)

func HaveAllRequestsCalled() gomega.OmegaMatcher {
	return &allRequestsCalledMatcher{}
}

type allRequestsCalledMatcher struct {
	failureMessage string
}

func (matcher *allRequestsCalledMatcher) Match(actual interface{}) (bool, error) {
	testHandler, ok := actual.(*net.TestHandler)
	if !ok {
		return false, fmt.Errorf("Expected a test handler, got %T", actual)
	}

	if testHandler.AllRequestsCalled() {
		matcher.failureMessage = "Expected all requests to not be called, but they were all called"
		return true, nil
	}

	message := fmt.Sprint("Failed to call requests:\n")
	for i := testHandler.CallCount; i < len(testHandler.Requests); i++ {
		message += fmt.Sprintf("%#v\n", testHandler.Requests[i])
	}
	message += "\n"
	matcher.failureMessage = message
	return false, nil
}

func (matcher *allRequestsCalledMatcher) FailureMessage(actual interface{}) string {
	return matcher.failureMessage
}

func (matcher *allRequestsCalledMatcher) NegatedFailureMessage(actual interface{}) string {
	return matcher.failureMessage
}

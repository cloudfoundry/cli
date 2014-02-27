package net

import (
	"errors"
	"fmt"
	"github.com/onsi/gomega"
)

func HaveAllRequestsCalled() gomega.OmegaMatcher {
	return allRequestsCalledMatcher{}
}

type allRequestsCalledMatcher struct{}

func (matcher allRequestsCalledMatcher) Match(actual interface{}) (bool, string, error) {
	testHandler, ok := actual.(*TestHandler)
	if !ok {
		return false, "", errors.New(fmt.Sprintf("Expected a test handler, got %T", actual))
	}

	if testHandler.AllRequestsCalled() {
		message := "Expected all requests to not be called, but they were all called"
		return true, message, nil
	} else {
		message := fmt.Sprint("Failed to call requests:\n")
		for i := testHandler.CallCount; i < len(testHandler.Requests); i++ {
			message += fmt.Sprintf("%#v\n", testHandler.Requests[i])
		}
		message += "\n"
		return false, message, nil
	}
}

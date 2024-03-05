package translatableerror_test

import (
	. "code.cloudfoundry.org/cli/command/translatableerror"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UnauthorizedError", func() {
	DescribeTable("error",
		func(message string, expected string) {
			Expect(UnauthorizedError{Message: message}).To(MatchError(expected))
		},
		Entry("bad credentials returns custom message", "Bad credentials", "Credentials were rejected, please try again."),
		Entry("invalid origin returns custom message", "The origin provided in the login hint is invalid.", "The origin provided is invalid."),
		Entry("umatched origin returns custom message", "The origin provided in the login_hint does not match an active Identity Provider, that supports password grant.", "The origin provided is invalid."),
		Entry("anything else", "I am a BANANA", "I am a BANANA"),
	)
})

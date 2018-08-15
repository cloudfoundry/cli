package translatableerror_test

import (
	. "code.cloudfoundry.org/cli/command/translatableerror"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("UnauthorizedError", func() {
	DescribeTable("error",
		func(message string, expected string) {
			Expect(UnauthorizedError{Message: message}).To(MatchError(expected))
		},
		Entry("bad credentials returns custom message", "Bad credentials", "Credentials were rejected, please try again."),
		Entry("invalid origin returns custom message", "The origin provided in the login hint is invalid.", "The origin provided is invalid."),
		Entry("anything else", "I am a BANANA", "I am a BANANA"),
	)
})

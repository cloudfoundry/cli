package requirements_test

import (
	. "code.cloudfoundry.org/cli/cf/requirements"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NumberArguments", func() {
	It("returns an error if the number of arguments doesn't match", func() {
		args := []string{"one", "two"}
		numberArgumentsRequirement := NewNumberArguments(args, []string{"SPACE"})

		err := numberArgumentsRequirement.Execute()
		Expect(err).To(MatchError(NumberArgumentsError{ExpectedArgs: []string{"SPACE"}}))
	})

	It("returns nil if the number of arguments matches", func() {
		args := []string{"one"}
		numberArgumentsRequirement := NewNumberArguments(args, []string{"SPACE"})

		err := numberArgumentsRequirement.Execute()
		Expect(err).NotTo(HaveOccurred())
	})
})

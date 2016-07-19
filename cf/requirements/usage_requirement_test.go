package requirements_test

import (
	. "code.cloudfoundry.org/cli/cf/requirements"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UsageRequirement", func() {
	It("doesn't return an error when the predicate returns false", func() {
		err := NewUsageRequirement(nil, "Some error message", func() bool { return false }).Execute()
		Expect(err).NotTo(HaveOccurred())
	})

	It("errors when the predicate returns true", func() {
		usableCmd := usableFunc(func() string { return "Usage text!" })

		err := NewUsageRequirement(usableCmd, "Some error message", func() bool { return true }).Execute()

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Some error message"))

		Expect(err.Error()).To(ContainSubstring("Usage text!"))
	})
})

type usableFunc func() string

func (u usableFunc) Usage() string { return u() }

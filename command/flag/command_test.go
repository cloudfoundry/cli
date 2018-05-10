package flag_test

import (
	. "code.cloudfoundry.org/cli/command/flag"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Command", func() {
	var command Command

	BeforeEach(func() {
		command = Command{}
	})

	Describe("UnmarshalFlag", func() {
		It("unmarshals into a filtered string", func() {
			err := command.UnmarshalFlag("default")
			Expect(err).ToNot(HaveOccurred())
			Expect(command.IsSet).To(BeTrue())
			Expect(command.Value).To(BeEmpty())
		})
	})
})

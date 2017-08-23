package flag_test

import (
	. "code.cloudfoundry.org/cli/command/flag"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Buildpack", func() {
	var buildpack Buildpack

	BeforeEach(func() {
		buildpack = Buildpack{}
	})

	Describe("UnmarshalFlag", func() {
		It("unmarshals into a filtered string", func() {
			err := buildpack.UnmarshalFlag("default")
			Expect(err).ToNot(HaveOccurred())
			Expect(buildpack.IsSet).To(BeTrue())
			Expect(buildpack.Value).To(BeEmpty())
		})
	})
})

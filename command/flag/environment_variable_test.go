package flag_test

import (
	"os"

	. "code.cloudfoundry.org/cli/command/flag"
	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("EnvironmentVariable", func() {
	var envVar EnvironmentVariable

	BeforeEach(func() {
		envVar = EnvironmentVariable("")
		os.Clearenv()

		var err error
		for _, v := range []string{"ABC", "FOO_BAR", "ACKBAR"} {
			err = os.Setenv(v, "")
			Expect(err).NotTo(HaveOccurred())
		}
	})

	AfterEach(func() {
		os.Clearenv()
	})

	Describe("Complete", func() {
		Context("when the prefix is empty", func() {
			It("returns no matches", func() {
				Expect(envVar.Complete("")).To(BeEmpty())
			})
		})

		Context("when the prefix does not start with $", func() {
			It("returns no matches", func() {
				Expect(envVar.Complete("A$A")).To(BeEmpty())
			})
		})

		Context("when the prefix starts with $", func() {
			Context("when only $ is specified", func() {
				It("returns all environment variables", func() {
					matches := envVar.Complete("$")
					Expect(matches).To(HaveLen(3))
					Expect(matches).To(ConsistOf(
						flags.Completion{Item: "$ABC"},
						flags.Completion{Item: "$FOO_BAR"},
						flags.Completion{Item: "$ACKBAR"},
					))
				})
			})

			Context("when additional characters are specified", func() {
				Context("when there are matching environment variables", func() {
					It("returns the matching environment variables", func() {
						matches := envVar.Complete("$A")
						Expect(matches).To(HaveLen(2))
						Expect(matches).To(ConsistOf(
							flags.Completion{Item: "$ABC"},
							flags.Completion{Item: "$ACKBAR"},
						))
					})
				})

				Context("when there are no matching environment variables", func() {
					It("returns no matches", func() {
						Expect(envVar.Complete("$Z")).To(BeEmpty())
					})
				})
			})
		})
	})
})

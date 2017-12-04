// +build !windows

package flag_test

import (
	"fmt"
	"os"
	"strings"

	. "code.cloudfoundry.org/cli/command/flag"
	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("EnvironmentVariable", func() {
	var (
		envVar  EnvironmentVariable
		envList []string
	)

	BeforeEach(func() {
		envList = []string{
			"ENVIRONMENTVARIABLE_TEST_ABC",
			"ENVIRONMENTVARIABLE_TEST_FOO_BAR",
			"ENVIRONMENTVARIABLE_TEST_ACKBAR",
			"ENVIRONMENTVARIABLE_TEST_abc",
		}

		var err error
		for _, v := range envList {
			err = os.Setenv(v, "")
			Expect(err).NotTo(HaveOccurred())
		}
	})

	AfterEach(func() {
		var err error
		for _, v := range envList {
			err = os.Unsetenv(v)
			Expect(err).ToNot(HaveOccurred())
		}
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
					keyValPairs := os.Environ()
					envVars := make([]string, len(keyValPairs))
					for i, keyValPair := range keyValPairs {
						envVars[i] = fmt.Sprintf("$%s", strings.Split(keyValPair, "=")[0])
					}

					matches := envVar.Complete("$")
					Expect(matches).To(HaveLen(len(keyValPairs)))
					for _, v := range envVars {
						Expect(matches).To(ContainElement(flags.Completion{Item: v}))
					}
				})
			})

			Context("when additional characters are specified", func() {
				Context("when there are matching environment variables", func() {
					It("returns the matching environment variables", func() {
						matches := envVar.Complete("$ENVIRONMENTVARIABLE_TEST_A")
						Expect(matches).To(HaveLen(2))
						Expect(matches).To(ConsistOf(
							flags.Completion{Item: "$ENVIRONMENTVARIABLE_TEST_ABC"},
							flags.Completion{Item: "$ENVIRONMENTVARIABLE_TEST_ACKBAR"},
						))
					})

					It("is case sensitive", func() {
						matches := envVar.Complete("$ENVIRONMENTVARIABLE_TEST_a")
						Expect(matches).To(HaveLen(1))
						Expect(matches).To(ConsistOf(
							flags.Completion{Item: "$ENVIRONMENTVARIABLE_TEST_abc"},
						))
					})
				})

				Context("when there are no matching environment variables", func() {
					It("returns no matches", func() {
						Expect(envVar.Complete("$ZZZZ")).To(BeEmpty())
					})
				})
			})
		})
	})

	Describe("UnmarshalFlag", func() {
		var rawFlagValue string

		BeforeEach(func() {
			envVar = ""
			rawFlagValue = "SOME_FLAG"
		})

		JustBeforeEach(func() {
			_ = envVar.UnmarshalFlag(rawFlagValue)
		})

		Context("when the env variable value prefix starts with the WorkAroundPrefix", func() {
			BeforeEach(func() {
				rawFlagValue = WorkAroundPrefix + "SOME_FLAG"
			})

			It("removes the WorkAroundPrefix", func() {
				Expect(string(envVar)).ToNot(ContainSubstring(WorkAroundPrefix))
				Expect(string(envVar)).To(Equal("SOME_FLAG"))
			})
		})

		Context("when the env variable value does not start with the WorkAroundPrefix", func() {
			It("unmarshals the value with no error", func() {
				Expect(string(envVar)).To(Equal("SOME_FLAG"))
			})
		})
	})
})

package flag_test

import (
	"fmt"
	"os"

	. "code.cloudfoundry.org/cli/command/flag"
	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("CredentialsOrJSON", func() {
	var credsOrJSON CredentialsOrJSON

	BeforeEach(func() {
		credsOrJSON = CredentialsOrJSON{}
	})

	Describe("default value", func() {
		It("is not set", func() {
			Expect(credsOrJSON.IsSet).To(BeFalse())
		})

		It("is empty", func() {
			Expect(credsOrJSON.Value).To(BeEmpty())
		})

		It("does not need to prompt the user for credentials", func() {
			Expect(credsOrJSON.UserPromptCredentials).To(BeEmpty())
		})
	})

	// The Complete method is not tested because it shares the same code as
	// Path.Complete().

	Describe("UnmarshalFlag", func() {
		Describe("empty credentials", func() {
			BeforeEach(func() {
				err := credsOrJSON.UnmarshalFlag("")
				Expect(err).NotTo(HaveOccurred())
			})

			It("is set", func() {
				Expect(credsOrJSON.IsSet).To(BeTrue())
			})

			It("is empty", func() {
				Expect(credsOrJSON.Value).To(BeEmpty())
			})

			It("does not need to prompt the user for credentials", func() {
				Expect(credsOrJSON.UserPromptCredentials).To(BeEmpty())
			})
		})

		DescribeTable("when the input is valid JSON",
			func(input string) {
				err := credsOrJSON.UnmarshalFlag(input)
				Expect(err).NotTo(HaveOccurred())

				Expect(credsOrJSON.IsSet).To(BeTrue())
				Expect(credsOrJSON.UserPromptCredentials).To(BeEmpty())
				Expect(credsOrJSON.Value).To(HaveLen(1))
				Expect(credsOrJSON.Value).To(HaveKeyWithValue("some", "json"))
			},
			Entry("valid JSON", `{"some": "json"}`),
			Entry("valid JSON in single quotes", `'{"some": "json"}'`),
			Entry("valid JSON in double quotes", `"{"some": "json"}"`),
		)

		Describe("reading JSON from a file", func() {
			var path string

			AfterEach(func() {
				os.Remove(path)
			})

			When("the file contains valid JSON", func() {
				BeforeEach(func() {
					path = tempFile(`{"some":"json"}`)
				})

				It("reads the JSON from the file", func() {
					err := credsOrJSON.UnmarshalFlag(path)
					Expect(err).NotTo(HaveOccurred())

					Expect(credsOrJSON.IsSet).To(BeTrue())
					Expect(credsOrJSON.Value).To(HaveLen(1))
					Expect(credsOrJSON.Value).To(HaveKeyWithValue("some", "json"))
				})

				It("does not need to prompt the user for credentials", func() {
					Expect(credsOrJSON.UserPromptCredentials).To(BeEmpty())
				})
			})

			When("the file has invalid JSON", func() {
				BeforeEach(func() {
					path = tempFile(`{"this is":"invalid JSON"`)
				})

				It("errors with the invalid configuration error", func() {
					err := credsOrJSON.UnmarshalFlag(path)
					Expect(err).To(Equal(&flags.Error{
						Type:    flags.ErrRequired,
						Message: fmt.Sprintf("The file '%s' contains invalid JSON. Please provide a path to a file containing a valid JSON object.", path),
					}))
				})
			})
		})

		Describe("prompting the user to enter credentials", func() {
			When("there is a credential", func() {
				BeforeEach(func() {
					err := credsOrJSON.UnmarshalFlag("foo")
					Expect(err).NotTo(HaveOccurred())
				})

				It("says the user must be prompted for a credential", func() {
					Expect(credsOrJSON.Value).To(BeEmpty())
					Expect(credsOrJSON.UserPromptCredentials).To(ConsistOf("foo"))
				})
			})

			When("there are many credentials", func() {
				BeforeEach(func() {
					err := credsOrJSON.UnmarshalFlag("foo, bar,baz ,bax moo")
					Expect(err).NotTo(HaveOccurred())
				})

				It("says the user must be prompted for the credential", func() {
					Expect(credsOrJSON.Value).To(BeEmpty())
					Expect(credsOrJSON.UserPromptCredentials).To(ConsistOf("foo", "bar", "baz", "bax moo"))
				})
			})
		})
	})
})

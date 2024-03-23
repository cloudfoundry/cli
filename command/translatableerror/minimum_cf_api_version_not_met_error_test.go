package translatableerror_test

import (
	. "code.cloudfoundry.org/cli/command/translatableerror"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MinimumAPIVersionNotMetError", func() {
	Describe("Error", func() {
		When("the Command field is not empty and CurrentVersion is not empty", func() {
			It("returns a template with the values of the Command and CurrentVersion fields", func() {
				err := MinimumCFAPIVersionNotMetError{
					Command:        "--some-flag",
					CurrentVersion: "1.2.3",
				}

				Expect(err.Error()).To(Equal("{{.Command}} requires CF API version {{.MinimumVersion}} or higher. Your target is {{.CurrentVersion}}."))
			})
		})

		When("the Command field is not empty and CurrentVersion is empty", func() {
			It("returns a template with the value of the Command field", func() {
				err := MinimumCFAPIVersionNotMetError{
					Command: "--some-flag",
				}

				Expect(err.Error()).To(Equal("{{.Command}} requires CF API version {{.MinimumVersion}} or higher."))
			})
		})

		When("the Command field is empty and CurrentVersion is not empty", func() {
			It("returns a template with the value of the CurrentVersion field", func() {
				err := MinimumCFAPIVersionNotMetError{
					CurrentVersion: "1.2.3",
				}

				Expect(err.Error()).To(Equal("This command requires CF API version {{.MinimumVersion}} or higher. Your target is {{.CurrentVersion}}."))
			})
		})

		When("the Command field is empty and CurrentVersion is empty", func() {
			It("returns a template without Command or CurrentVersion values", func() {
				err := MinimumCFAPIVersionNotMetError{}

				Expect(err.Error()).To(Equal("This command requires CF API version {{.MinimumVersion}} or higher."))
			})
		})
	})
})

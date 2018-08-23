package translatableerror_test

import (
	. "code.cloudfoundry.org/cli/command/translatableerror"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = FDescribe("MinimumUAAAPIVersionNotMetError", func() {
	Describe("Error", func() {
		When("the Command field is not empty and CurrentVersion is not empty", func() {
			It("returns a template with the values of the Command and CurrentVersion fields", func() {
				err := MinimumUAAAPIVersionNotMetError{
					Command:        "--some-flag",
					CurrentVersion: "1.2.3",
				}

				Expect(err.Error()).To(Equal("{{.Command}} requires UAA API version {{.MinimumVersion}} or higher, but your current version is {{.CurrentVersion}}"))
			})
		})

		When("the Command field is not empty and CurrentVersion is empty", func() {
			It("returns a template with the value of the Command field", func() {
				err := MinimumUAAAPIVersionNotMetError{
					Command: "--some-flag",
				}

				Expect(err.Error()).To(Equal("{{.Command}} requires UAA API version {{.MinimumVersion}} or higher."))
			})
		})

		When("the Command field is empty and CurrentVersion is not empty", func() {
			It("returns a template with the value of the CurrentVersion field", func() {
				err := MinimumUAAAPIVersionNotMetError{
					CurrentVersion: "1.2.3",
				}

				Expect(err.Error()).To(Equal("This command requires UAA API version {{.MinimumVersion}} or higher, but your current version is {{.CurrentVersion}}"))
			})
		})

		When("the Command field is empty and CurrentVersion is empty", func() {
			It("returns a template without Command or CurrentVersion values", func() {
				err := MinimumUAAAPIVersionNotMetError{}

				Expect(err.Error()).To(Equal("This command requires UAA API version {{.MinimumVersion}} or higher."))
			})
		})
	})
})

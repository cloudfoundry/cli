package translatableerror_test

import (
	. "code.cloudfoundry.org/cli/command/translatableerror"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MinimumAPIVersionNotMetError", func() {
	Describe("Error", func() {
		Context("when the Command field is not empty and CurrentVersion is not empty", func() {
			It("returns a template with the values of the Command and CurrentVersion fields", func() {
				err := MinimumAPIVersionNotMetError{
					Command:        "--some-flag",
					CurrentVersion: "1.2.3",
				}

				Expect(err.Error()).To(Equal("{{.Command}} requires CF API version {{.MinimumVersion}} or higher. Your target is {{.CurrentVersion}}."))
			})
		})

		Context("when the Command field is not empty and CurrentVersion is empty", func() {
			It("returns a template with the value of the Command field", func() {
				err := MinimumAPIVersionNotMetError{
					Command: "--some-flag",
				}

				Expect(err.Error()).To(Equal("{{.Command}} requires CF API version {{.MinimumVersion}} or higher."))
			})
		})

		Context("when the Command field is empty and CurrentVersion is not empty", func() {
			It("returns a template with the value of the CurrentVersion field", func() {
				err := MinimumAPIVersionNotMetError{
					CurrentVersion: "1.2.3",
				}

				Expect(err.Error()).To(Equal("This command requires CF API version {{.MinimumVersion}} or higher. Your target is {{.CurrentVersion}}."))
			})
		})

		Context("when the Command field is empty and CurrentVersion is empty", func() {
			It("returns a template without Command or CurrentVersion values", func() {
				err := MinimumAPIVersionNotMetError{}

				Expect(err.Error()).To(Equal("This command requires CF API version {{.MinimumVersion}} or higher."))
			})
		})
	})
})

package translatableerror_test

import (
	. "code.cloudfoundry.org/cli/command/translatableerror"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MinimumUAAAPIVersionNotMetError", func() {
	Describe("Error", func() {
		When("the Command field is not empty", func() {
			It("returns a template with the value of the Command field", func() {
				err := MinimumUAAAPIVersionNotMetError{
					Command: "--some-flag",
				}

				Expect(err.Error()).To(Equal("{{.Command}} requires UAA API version {{.MinimumVersion}} or higher. Update your Cloud Foundry instance."))
			})
		})

		When("the Command field is empty", func() {
			It("returns a template with the value of the CurrentVersion field", func() {
				err := MinimumUAAAPIVersionNotMetError{}

				Expect(err.Error()).To(Equal("This command requires UAA API version {{.MinimumVersion}} or higher. Update your Cloud Foundry instance."))
			})
		})

		When("the Command field is empty and CurrentVersion is empty", func() {
			It("returns a template without Command or CurrentVersion values", func() {
				err := MinimumUAAAPIVersionNotMetError{}

				Expect(err.Error()).To(Equal("This command requires UAA API version {{.MinimumVersion}} or higher. Update your Cloud Foundry instance."))
			})
		})
	})
})

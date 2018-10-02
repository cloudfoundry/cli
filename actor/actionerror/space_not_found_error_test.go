package actionerror_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SpaceNotFoundError", func() {
	Describe("Error", func() {
		When("the name is specified", func() {
			It("returns an error message with the name of the missing space", func() {
				err := actionerror.SpaceNotFoundError{
					Name: "some-space",
				}
				Expect(err.Error()).To(Equal("Space 'some-space' not found."))
			})
		})

		When("the name is not specified, but the GUID is specified", func() {
			It("returns an error message with the GUID of the missing space", func() {
				err := actionerror.SpaceNotFoundError{
					GUID: "some-space-guid",
				}
				Expect(err.Error()).To(Equal("Space with GUID 'some-space-guid' not found."))
			})
		})

		When("neither the name nor the GUID is specified", func() {
			It("returns a generic error message for the missing space", func() {
				err := actionerror.SpaceNotFoundError{}
				Expect(err.Error()).To(Equal("Space '' not found."))
			})
		})
	})
})

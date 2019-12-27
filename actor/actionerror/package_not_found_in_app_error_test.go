package actionerror_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PackageNotFoundInAppError", func() {
	Describe("Error", func() {
		When("package GUID and app name are specified", func() {
			It("returns an error message with both", func() {
				err := actionerror.PackageNotFoundInAppError{
					GUID:    "some-package-guid",
					AppName: "zora",
				}
				Expect(err.Error()).To(
					Equal("Package with guid 'some-package-guid' not found in app 'zora'."))
			})
		})
		When("the app name is specified and package GUID is not", func() {
			It("returns an error message with the app name that is missing packages", func() {
				err := actionerror.PackageNotFoundInAppError{
					AppName: "dora",
				}
				Expect(err.Error()).To(Equal("Package not found in app 'dora'."))
			})
		})
		When("neither app name nor package GUID is specified", func() {
			It("returns a generic error message for the missing package", func() {
				err := actionerror.PackageNotFoundInAppError{}
				Expect(err.Error()).To(Equal("Package not found."))
			})
		})
	})
})

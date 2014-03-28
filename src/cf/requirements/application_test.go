package requirements_test

import (
	"cf/errors"
	"cf/models"
	. "cf/requirements"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testterm "testhelpers/terminal"
)

var _ = Describe("Testing with ginkgo", func() {

	It("TestApplicationReqExecute", func() {
		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		appRepo := &testapi.FakeApplicationRepository{}
		appRepo.ReadReturns.App = app
		ui := new(testterm.FakeUI)

		appReq := NewApplicationRequirement("foo", ui, appRepo)
		success := appReq.Execute()

		Expect(success).To(BeTrue())
		Expect(appRepo.ReadArgs.Name).To(Equal("foo"))
		Expect(appReq.GetApplication()).To(Equal(app))
	})

	It("TestApplicationReqExecuteWhenApplicationNotFound", func() {
		appRepo := &testapi.FakeApplicationRepository{}
		appRepo.ReadReturns.Error = errors.NewModelNotFoundError("app", "foo")
		ui := new(testterm.FakeUI)

		appReq := NewApplicationRequirement("foo", ui, appRepo)

		testassert.AssertPanic(testterm.FailedWasCalled, func() {
			appReq.Execute()
		})
	})
})

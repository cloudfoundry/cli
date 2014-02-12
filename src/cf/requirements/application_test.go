package requirements_test

import (
	"cf/models"
	. "cf/requirements"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testterm "testhelpers/terminal"
)

var _ = Describe("Testing with ginkgo", func() {

	It("TestApplicationReqExecute", func() {
		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		appRepo := &testapi.FakeApplicationRepository{ReadApp: app}
		ui := new(testterm.FakeUI)

		appReq := NewApplicationRequirement("foo", ui, appRepo)
		success := appReq.Execute()

		Expect(success).To(BeTrue())
		Expect(appRepo.ReadName).To(Equal("foo"))
		Expect(appReq.GetApplication()).To(Equal(app))
	})

	It("TestApplicationReqExecuteWhenApplicationNotFound", func() {
		appRepo := &testapi.FakeApplicationRepository{ReadNotFound: true}
		ui := new(testterm.FakeUI)

		appReq := NewApplicationRequirement("foo", ui, appRepo)

		testassert.AssertPanic(mr.T(), testterm.FailedWasCalled, func() {
			appReq.Execute()
		})
	})
})

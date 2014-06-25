package requirements_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/cloudfoundry/cli/cf/requirements"
	testassert "github.com/cloudfoundry/cli/testhelpers/assert"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ApplicationRequirement", func() {
	var ui *testterm.FakeUI
	var appRepo *testapi.FakeApplicationRepository

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
		appRepo = &testapi.FakeApplicationRepository{}
	})

	It("succeeds when an app with the given name exists", func() {
		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		appRepo.ReadReturns.App = app

		appReq := NewApplicationRequirement("foo", ui, appRepo)

		Expect(appReq.Execute()).To(BeTrue())
		Expect(appRepo.ReadArgs.Name).To(Equal("foo"))
		Expect(appReq.GetApplication()).To(Equal(app))
	})

	It("fails when an app with the given name cannot be found", func() {
		appRepo.ReadReturns.Error = errors.NewModelNotFoundError("app", "foo")

		testassert.AssertPanic(testterm.FailedWasCalled, func() {
			NewApplicationRequirement("foo", ui, appRepo).Execute()
		})
	})
})

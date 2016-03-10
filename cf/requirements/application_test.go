package requirements_test

import (
	testApplication "github.com/cloudfoundry/cli/cf/api/applications/fakes"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/cloudfoundry/cli/cf/requirements"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ApplicationRequirement", func() {
	var appRepo *testApplication.FakeApplicationRepository

	BeforeEach(func() {
		appRepo = &testApplication.FakeApplicationRepository{}
	})

	It("succeeds when an app with the given name exists", func() {
		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		appRepo.ReadReturns(app, nil)

		appReq := NewApplicationRequirement("foo", appRepo)

		err := appReq.Execute()
		Expect(err).NotTo(HaveOccurred())

		Expect(appRepo.ReadArgsForCall(0)).To(Equal("foo"))
		Expect(appReq.GetApplication()).To(Equal(app))
	})

	It("fails when an app with the given name cannot be found", func() {
		appError := errors.NewModelNotFoundError("app", "foo")
		appRepo.ReadReturns(models.Application{}, appError)

		err := NewApplicationRequirement("foo", appRepo).Execute()
		Expect(err).To(HaveOccurred())
		Expect(err).To(Equal(appError))
	})
})

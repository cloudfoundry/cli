package requirements_test

import (
	"code.cloudfoundry.org/cli/cf/api/applications/applicationsfakes"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/models"
	. "code.cloudfoundry.org/cli/cf/requirements"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ApplicationRequirement", func() {
	var appRepo *applicationsfakes.FakeRepository

	BeforeEach(func() {
		appRepo = new(applicationsfakes.FakeRepository)
	})

	It("succeeds when an app with the given name exists", func() {
		app := models.Application{}
		app.Name = "my-app"
		app.GUID = "my-app-guid"
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

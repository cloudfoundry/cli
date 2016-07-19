package requirements_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"

	"code.cloudfoundry.org/cli/cf/api/applications/applicationsfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DeaApplication", func() {
	var (
		req     requirements.DEAApplicationRequirement
		appRepo *applicationsfakes.FakeRepository
		appName string
	)

	BeforeEach(func() {
		appName = "fake-app-name"
		appRepo = new(applicationsfakes.FakeRepository)
		req = requirements.NewDEAApplicationRequirement(appName, appRepo)
	})

	Describe("GetApplication", func() {
		It("returns an empty application", func() {
			Expect(req.GetApplication()).To(Equal(models.Application{}))
		})

		Context("when the requirement has been executed", func() {
			BeforeEach(func() {
				app := models.Application{}
				app.GUID = "fake-app-guid"
				appRepo.ReadReturns(app, nil)

				req.Execute()
			})

			It("returns the application", func() {
				Expect(req.GetApplication().GUID).To(Equal("fake-app-guid"))
			})
		})
	})

	Describe("Execute", func() {
		Context("when the returned application is a Diego application", func() {
			BeforeEach(func() {
				app := models.Application{}
				app.Diego = true
				appRepo.ReadReturns(app, nil)
			})

			It("fails with error", func() {
				err := req.Execute()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("The app is running on the Diego backend, which does not support this command."))
			})
		})

		Context("when the returned application is not a Diego application", func() {
			BeforeEach(func() {
				app := models.Application{}
				app.Diego = false
				appRepo.ReadReturns(app, nil)
			})

			It("succeeds", func() {
				err := req.Execute()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when finding the application results in an error", func() {
			BeforeEach(func() {
				appRepo.ReadReturns(models.Application{}, errors.New("find-err"))
			})

			It("fails with error", func() {
				err := req.Execute()
				Expect(err.Error()).To(ContainSubstring("find-err"))
			})
		})
	})
})

package application_test

import (
	testApplication "github.com/cloudfoundry/cli/cf/api/applications/fakes"
	"github.com/cloudfoundry/cli/cf/manifest"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/generic"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testmanifest "github.com/cloudfoundry/cli/testhelpers/manifest"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/application"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("restart command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		starter             *testcmd.FakeApplicationStarter
		stopper             *testcmd.FakeApplicationStopper
		manifestRepo        *testmanifest.FakeManifestRepository
		appRepo             *testApplication.FakeApplicationRepository
		config              core_config.ReadWriter
		app                 models.Application
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		starter = &testcmd.FakeApplicationStarter{}
		stopper = &testcmd.FakeApplicationStopper{}
		config = testconfig.NewRepositoryWithDefaults()
		manifestRepo = &testmanifest.FakeManifestRepository{}
		appRepo = &testApplication.FakeApplicationRepository{}

		app = models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCommand(NewRestart(ui, config, starter, stopper, manifestRepo, appRepo), args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails when not logged in", func() {
			requirementsFactory.TargetedSpaceSuccess = true

			Expect(runCommand()).To(BeFalse())
		})

		It("fails when a space is not targeted", func() {
			requirementsFactory.LoginSuccess = true

			Expect(runCommand()).To(BeFalse())
		})
	})

	Context("when logged in, targeting a space, and an app name is provided", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = true
			appRepo.ReadReturns.App = app

			stopper.ApplicationStopReturns(app, nil)
		})

		Context("app name is provided", func() {
			It("restarts the given app", func() {
				runCommand("my-app")

				application, orgName, spaceName := stopper.ApplicationStopArgsForCall(0)
				Expect(application).To(Equal(app))
				Expect(orgName).To(Equal(config.OrganizationFields().Name))
				Expect(spaceName).To(Equal(config.SpaceFields().Name))

				application, orgName, spaceName = starter.ApplicationStartArgsForCall(0)
				Expect(application).To(Equal(app))
				Expect(orgName).To(Equal(config.OrganizationFields().Name))
				Expect(spaceName).To(Equal(config.SpaceFields().Name))

				Expect(appRepo.ReadArgs.Name).To(Equal("my-app"))
			})
		})

		Context("app name is not provided", func() {
			BeforeEach(func() {
				manifestRepo.ReadManifestReturns.Manifest = &manifest.Manifest{
					Path: "manifest.yml",
					Data: generic.NewMap(map[interface{}]interface{}{
						"applications": []interface{}{
							generic.NewMap(map[interface{}]interface{}{
								"name":      "my-app",
								"memory":    "128MB",
								"instances": 1,
							}),
						},
					}),
				}
			})

			It("restarts app from manifest", func() {
				runCommand()

				application, orgName, spaceName := stopper.ApplicationStopArgsForCall(0)
				Expect(application).To(Equal(app))
				Expect(orgName).To(Equal(config.OrganizationFields().Name))
				Expect(spaceName).To(Equal(config.SpaceFields().Name))

				application, orgName, spaceName = starter.ApplicationStartArgsForCall(0)
				Expect(application).To(Equal(app))
				Expect(orgName).To(Equal(config.OrganizationFields().Name))
				Expect(spaceName).To(Equal(config.SpaceFields().Name))

				Expect(appRepo.ReadArgs.Name).To(Equal("my-app"))
			})
		})
	})
})

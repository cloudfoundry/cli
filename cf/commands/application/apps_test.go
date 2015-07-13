package application_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/plugin/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("list-apps command", func() {
	var (
		ui                  *testterm.FakeUI
		configRepo          core_config.Repository
		appSummaryRepo      *testapi.FakeAppSummaryRepo
		requirementsFactory *testreq.FakeReqFactory
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.Config = configRepo
		deps.RepoLocator = deps.RepoLocator.SetAppSummaryRepository(appSummaryRepo)
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("apps").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		appSummaryRepo = &testapi.FakeAppSummaryRepo{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{
			LoginSuccess:         true,
			TargetedSpaceSuccess: true,
		}

		app1Routes := []models.RouteSummary{
			models.RouteSummary{
				Host: "app1",
				Domain: models.DomainFields{
					Name:                   "cfapps.io",
					Shared:                 true,
					OwningOrganizationGuid: "org-123",
					Guid: "domain-guid",
				},
			},
			models.RouteSummary{
				Host: "app1",
				Domain: models.DomainFields{
					Name: "example.com",
				},
			}}

		app2Routes := []models.RouteSummary{
			models.RouteSummary{
				Host:   "app2",
				Domain: models.DomainFields{Name: "cfapps.io"},
			}}

		app := models.Application{}
		app.Name = "Application-1"
		app.Guid = "Application-1-guid"
		app.State = "started"
		app.RunningInstances = 1
		app.InstanceCount = 1
		app.Memory = 512
		app.DiskQuota = 1024
		app.Routes = app1Routes

		app2 := models.Application{}
		app2.Name = "Application-2"
		app2.Guid = "Application-2-guid"
		app2.State = "started"
		app2.RunningInstances = 1
		app2.InstanceCount = 2
		app2.Memory = 256
		app2.DiskQuota = 1024
		app2.Routes = app2Routes

		appSummaryRepo.GetSummariesInCurrentSpaceApps = []models.Application{app, app2}

		deps = command_registry.NewDependency()
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("apps", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("requirements", func() {
		It("requires the user to be logged in", func() {
			requirementsFactory.LoginSuccess = false

			Expect(runCommand()).To(BeFalse())
		})

		It("requires the user to have a space targeted", func() {
			requirementsFactory.TargetedSpaceSuccess = false

			Expect(runCommand()).To(BeFalse())
		})
		It("should fail with usage when provided any arguments", func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = true
			Expect(runCommand("blahblah")).To(BeFalse())
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "No argument required"},
			))
		})
	})

	Describe("when invoked by a plugin", func() {
		var (
			pluginAppModels []plugin_models.GetAppsModel
		)

		BeforeEach(func() {
			pluginAppModels = []plugin_models.GetAppsModel{}
			deps.PluginModels.AppsSummary = &pluginAppModels
		})

		It("populates the plugin models upon execution", func() {
			testcmd.RunCliCommand("apps", []string{}, requirementsFactory, updateCommandDependency, true)

			Ω(pluginAppModels[0].Name).To(Equal("Application-1"))
			Ω(pluginAppModels[0].Guid).To(Equal("Application-1-guid"))
			Ω(pluginAppModels[1].Name).To(Equal("Application-2"))
			Ω(pluginAppModels[1].Guid).To(Equal("Application-2-guid"))
			Ω(pluginAppModels[0].State).To(Equal("started"))
			Ω(pluginAppModels[0].TotalInstances).To(Equal(1))
			Ω(pluginAppModels[0].RunningInstances).To(Equal(1))
			Ω(pluginAppModels[0].Memory).To(Equal(int64(512)))
			Ω(pluginAppModels[0].DiskQuota).To(Equal(int64(1024)))
			Ω(pluginAppModels[0].Routes[0].Host).To(Equal("app1"))
			Ω(pluginAppModels[0].Routes[1].Host).To(Equal("app1"))
			Ω(pluginAppModels[0].Routes[0].Domain.Name).To(Equal("cfapps.io"))
			Ω(pluginAppModels[0].Routes[0].Domain.Shared).To(BeTrue())
			Ω(pluginAppModels[0].Routes[0].Domain.OwningOrganizationGuid).To(Equal("org-123"))
			Ω(pluginAppModels[0].Routes[0].Domain.Guid).To(Equal("domain-guid"))
		})
	})

	Context("when the user is logged in and a space is targeted", func() {
		It("lists apps in a table", func() {
			runCommand()

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting apps in", "my-org", "my-space", "my-user"},
				[]string{"OK"},
				[]string{"Application-1", "started", "1/1", "512M", "1G", "app1.cfapps.io", "app1.example.com"},
				[]string{"Application-2", "started", "1/2", "256M", "1G", "app2.cfapps.io"},
			))
		})

		Context("when an app's running instances is unknown", func() {
			It("dipslays a '?' for running instances", func() {
				appRoutes := []models.RouteSummary{
					models.RouteSummary{
						Host:   "app1",
						Domain: models.DomainFields{Name: "cfapps.io"},
					}}
				app := models.Application{}
				app.Name = "Application-1"
				app.Guid = "Application-1-guid"
				app.State = "started"
				app.RunningInstances = -1
				app.InstanceCount = 2
				app.Memory = 512
				app.DiskQuota = 1024
				app.Routes = appRoutes

				appSummaryRepo.GetSummariesInCurrentSpaceApps = []models.Application{app}

				runCommand()

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Getting apps in", "my-org", "my-space", "my-user"},
					[]string{"OK"},
					[]string{"Application-1", "started", "?/2", "512M", "1G", "app1.cfapps.io"},
				))
			})
		})

		Context("when there are no apps", func() {
			It("tells the user that there are no apps", func() {
				appSummaryRepo.GetSummariesInCurrentSpaceApps = []models.Application{}

				runCommand()
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Getting apps in", "my-org", "my-space", "my-user"},
					[]string{"OK"},
					[]string{"No apps found"},
				))
			})
		})
	})
})

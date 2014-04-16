package application_test

import (
	. "cf/commands/application"
	"cf/configuration"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

var _ = Describe("list-apps command", func() {
	var (
		ui                  *testterm.FakeUI
		configRepo          configuration.ReadWriter
		appSummaryRepo      *testapi.FakeAppSummaryRepo
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		appSummaryRepo = &testapi.FakeAppSummaryRepo{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{
			LoginSuccess:         true,
			TargetedSpaceSuccess: true,
		}
	})

	runCommand := func() {
		cmd := NewListApps(ui, configRepo, appSummaryRepo)
		testcmd.RunCommand(cmd, testcmd.NewContext("apps", []string{}), requirementsFactory)
	}

	Describe("requirements", func() {
		It("requires the user to be logged in", func() {
			requirementsFactory.LoginSuccess = false

			runCommand()
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("requires the user to have a space targeted", func() {
			requirementsFactory.TargetedSpaceSuccess = false

			runCommand()
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	Context("when the user is logged in and a space is targeted", func() {
		It("lists apps in a table", func() {
			app1Routes := []models.RouteSummary{
				models.RouteSummary{
					Host: "app1",
					Domain: models.DomainFields{
						Name: "cfapps.io",
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
			app.State = "started"
			app.RunningInstances = 1
			app.InstanceCount = 1
			app.Memory = 512
			app.DiskQuota = 1024
			app.Routes = app1Routes

			app2 := models.Application{}
			app2.Name = "Application-2"
			app2.State = "started"
			app2.RunningInstances = 1
			app2.InstanceCount = 2
			app2.Memory = 256
			app2.DiskQuota = 1024
			app2.Routes = app2Routes

			appSummaryRepo.GetSummariesInCurrentSpaceApps = []models.Application{app, app2}

			runCommand()

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Getting apps in", "my-org", "my-space", "my-user"},
				{"OK"},
				{"Application-1", "started", "1/1", "512M", "1G", "app1.cfapps.io", "app1.example.com"},
				{"Application-2", "started", "1/2", "256M", "1G", "app2.cfapps.io"},
			})
		})

		Context("when there are no apps", func() {
			It("tells the user that there are no apps", func() {
				runCommand()
				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"Getting apps in", "my-org", "my-space", "my-user"},
					{"OK"},
					{"No apps found"},
				})
			})
		})
	})
})

package application_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	. "github.com/cloudfoundry/cli/cf/commands/application"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
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
		configRepo          core_config.ReadWriter
		appSummaryRepo      *testapi.FakeAppSummaryRepo
		requirementsFactory *testreq.FakeReqFactory
		spaceRepo           *testapi.FakeSpaceRepository
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		appSummaryRepo = &testapi.FakeAppSummaryRepo{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		spaceRepo = new(testapi.FakeSpaceRepository)
		requirementsFactory = &testreq.FakeReqFactory{
			LoginSuccess:         true,
			TargetedSpaceSuccess: true,
		}
	})

	runCommand := func(args ...string) bool {
		cmd := NewListApps(ui, configRepo, appSummaryRepo, spaceRepo)
		return testcmd.RunCommand(cmd, args, requirementsFactory)
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
			Expect(ui.FailedWithUsage).To(BeTrue())
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
				runCommand()
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Getting apps in", "my-org", "my-space", "my-user"},
					[]string{"OK"},
					[]string{"No apps found"},
				))
			})
		})

		Context("when space flag is provided", func() {
			BeforeEach(func() {
				space := models.Space{}
				space.Name = "my-space"
				space.Guid = "my-space-guid"
				spaceRepo.Spaces = []models.Space{space}
			})

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

				app := models.Application{
					ApplicationFields: models.ApplicationFields{
						Name:             "Application-1",
						State:            "started",
						RunningInstances: 1,
						InstanceCount:    1,
						Memory:           512,
						DiskQuota:        1024,
					},
					Routes: app1Routes,
				}

				app2 := models.Application{
					ApplicationFields: models.ApplicationFields{
						Name:             "Application-2",
						State:            "started",
						RunningInstances: 1,
						InstanceCount:    2,
						Memory:           256,
						DiskQuota:        1024,
					},
					Routes: app2Routes,
				}

				appSummaryRepo.GetSpaceSummariesApps = []models.Application{app, app2}

				runCommand("-s", "my-space")
				Expect(spaceRepo.FindByNameName).To(Equal("my-space"))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Getting apps in", "my-org", "my-space", "my-user"},
					[]string{"OK"},
					[]string{"Application-1", "started", "1/1", "512M", "1G", "app1.cfapps.io", "app1.example.com"},
					[]string{"Application-2", "started", "1/2", "256M", "1G", "app2.cfapps.io"},
				))
			})

			It("fails when the space is not found", func() {
				spaceRepo.FindByNameNotFound = true

				runCommand("-s", "my-space")
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"my-space", "not found"},
				))
			})

			It("should not list apps in the current space", func() {
				appRoutes1 := []models.RouteSummary{
					models.RouteSummary{
						Host:   "app1",
						Domain: models.DomainFields{Name: "cfapps.io"},
					}}

				appRoutes2 := []models.RouteSummary{
					models.RouteSummary{
						Host:   "app2",
						Domain: models.DomainFields{Name: "cfapps.io"},
					}}

				app1 := models.Application{
					ApplicationFields: models.ApplicationFields{
						Name:             "Application-1",
						State:            "started",
						RunningInstances: 1,
						InstanceCount:    2,
						Memory:           512,
						DiskQuota:        1024,
					},
					Routes: appRoutes1,
				}

				app2 := models.Application{
					ApplicationFields: models.ApplicationFields{
						Name:             "Application-2",
						State:            "started",
						RunningInstances: 1,
						InstanceCount:    2,
						Memory:           512,
						DiskQuota:        1024,
					},
					Routes: appRoutes2,
				}

				appSummaryRepo.GetSummariesInCurrentSpaceApps = []models.Application{app1}

				appSummaryRepo.GetSpaceSummariesApps = []models.Application{app2}

				runCommand("-s", "my-space")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Getting apps in", "my-org", "my-space", "my-user"},
					[]string{"OK"},
					[]string{"Application-2", "started", "1/2", "512M", "1G", "app2.cfapps.io"},
				))
			})

			It("should list apps even if the space name is current space", func() {
				appRoutes1 := []models.RouteSummary{
					models.RouteSummary{
						Host:   "app1",
						Domain: models.DomainFields{Name: "cfapps.io"},
					}}

				app1 := models.Application{
					ApplicationFields: models.ApplicationFields{
						Name:             "Application-1",
						State:            "started",
						RunningInstances: 1,
						InstanceCount:    2,
						Memory:           512,
						DiskQuota:        1024,
					},
					Routes: appRoutes1,
				}

				appSummaryRepo.GetSummariesInCurrentSpaceApps = []models.Application{app1}

				appSummaryRepo.GetSpaceSummariesApps = []models.Application{app1}

				runCommand("-s", "my-space")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Getting apps in", "my-org", "my-space", "my-user"},
					[]string{"OK"},
					[]string{"Application-1", "started", "1/2", "512M", "1G", "app1.cfapps.io"},
				))
			})
		})
	})
})

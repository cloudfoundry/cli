package application_test

import (
	"time"

	testAppInstanaces "github.com/cloudfoundry/cli/cf/api/app_instances/fakes"
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/formatters"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/plugin/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	testtime "github.com/cloudfoundry/cli/testhelpers/time"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("app Command", func() {
	var (
		ui                  *testterm.FakeUI
		configRepo          core_config.Repository
		appSummaryRepo      *testapi.FakeAppSummaryRepo
		appInstancesRepo    *testAppInstanaces.FakeAppInstancesRepository
		appLogsNoaaRepo     *testapi.FakeLogsNoaaRepository
		requirementsFactory *testreq.FakeReqFactory
		app                 models.Application
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetLogsNoaaRepository(appLogsNoaaRepo)
		deps.Config = configRepo
		deps.RepoLocator = deps.RepoLocator.SetAppSummaryRepository(appSummaryRepo)
		deps.RepoLocator = deps.RepoLocator.SetAppInstancesRepository(appInstancesRepo)
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("app").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		appSummaryRepo = &testapi.FakeAppSummaryRepo{}
		appLogsNoaaRepo = &testapi.FakeLogsNoaaRepository{}
		appInstancesRepo = &testAppInstanaces.FakeAppInstancesRepository{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{
			LoginSuccess:         true,
			TargetedSpaceSuccess: true,
		}
		app = makeAppWithRoute("my-app")
		appSummaryRepo.GetSummarySummary = app

		deps = command_registry.NewDependency()
		updateCommandDependency(false)
	})

	runCommand := func(args ...string) bool {
		cmd := command_registry.Commands.FindCommand("app")
		return testcmd.RunCliCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails if not logged in", func() {
			requirementsFactory.LoginSuccess = false
			Expect(runCommand("my-app")).To(BeFalse())
		})

		It("fails if a space is not targeted", func() {
			requirementsFactory.TargetedSpaceSuccess = false
			Expect(runCommand("my-app")).To(BeFalse())
		})

		It("fails with usage when not provided exactly one arg", func() {
			passed := runCommand()
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires an argument"},
			))
			Expect(passed).To(BeFalse())
		})
	})

	Describe("when invoked by a plugin", func() {
		var (
			pluginAppModel *plugin_models.GetAppModel
		)

		BeforeEach(func() {
			app = makeAppWithRoute("my-app")
			appInstance := models.AppInstanceFields{
				State:     models.InstanceRunning,
				Since:     testtime.MustParse("Mon Jan 2 15:04:05 -0700 MST 2006", "Mon Jan 2 15:04:05 -0700 MST 2012"),
				Details:   "normal",
				CpuUsage:  1.0,
				DiskQuota: 1 * formatters.GIGABYTE,
				DiskUsage: 32 * formatters.MEGABYTE,
				MemQuota:  64 * formatters.MEGABYTE,
				MemUsage:  13 * formatters.MEGABYTE,
			}

			appInstance2 := models.AppInstanceFields{
				State:   models.InstanceDown,
				Details: "failure",
				Since:   testtime.MustParse("Mon Jan 2 15:04:05 -0700 MST 2006", "Mon Apr 1 15:04:05 -0700 MST 2012"),
			}

			instances := []models.AppInstanceFields{appInstance, appInstance2}
			appInstancesRepo.GetInstancesReturns(instances, nil)

			appSummaryRepo.GetSummarySummary = app
			requirementsFactory.Application = app

			pluginAppModel = &plugin_models.GetAppModel{}
			deps.PluginModels.Application = pluginAppModel
			updateCommandDependency(true)
		})

		It("populates the plugin model upon execution", func() {
			runCommand("my-app")
			Ω(pluginAppModel.Name).To(Equal("my-app"))
			Ω(pluginAppModel.State).To(Equal("started"))
			Ω(pluginAppModel.Guid).To(Equal("app-guid"))
			Ω(pluginAppModel.BuildpackUrl).To(Equal("http://123.com"))
			Ω(pluginAppModel.Command).To(Equal("command1"))
			Ω(pluginAppModel.Diego).To(BeFalse())
			Ω(pluginAppModel.DetectedStartCommand).To(Equal("detected_command"))
			Ω(pluginAppModel.DiskQuota).To(Equal(int64(100)))
			Ω(pluginAppModel.EnvironmentVars).To(Equal(map[string]interface{}{"test": 123}))
			Ω(pluginAppModel.InstanceCount).To(Equal(2))
			Ω(pluginAppModel.Memory).To(Equal(int64(256)))
			Ω(pluginAppModel.RunningInstances).To(Equal(2))
			Ω(pluginAppModel.HealthCheckTimeout).To(Equal(100))
			Ω(pluginAppModel.SpaceGuid).To(Equal("guids_in_spaaace"))
			Ω(pluginAppModel.PackageUpdatedAt.String()).To(Equal(time.Date(2009, time.November, 10, 15, 0, 0, 0, time.UTC).String()))
			Ω(pluginAppModel.PackageState).To(Equal("STAGED"))
			Ω(pluginAppModel.StagingFailedReason).To(Equal("no reason"))
			Ω(pluginAppModel.Stack.Name).To(Equal("fake_stack"))
			Ω(pluginAppModel.Stack.Guid).To(Equal("123-123-123"))
			Ω(pluginAppModel.Routes[0].Host).To(Equal("foo"))
			Ω(pluginAppModel.Routes[0].Guid).To(Equal("foo-guid"))
			Ω(pluginAppModel.Routes[0].Domain.Name).To(Equal("example.com"))
			Ω(pluginAppModel.Routes[0].Domain.Guid).To(Equal("domain1-guid"))
			Ω(pluginAppModel.Routes[0].Domain.Shared).To(BeTrue())
			Ω(pluginAppModel.Routes[0].Domain.OwningOrganizationGuid).To(Equal("org-123"))
			Ω(pluginAppModel.Services[0].Guid).To(Equal("s1-guid"))
			Ω(pluginAppModel.Services[0].Name).To(Equal("s1-service"))
			Ω(pluginAppModel.Instances[0].State).To(Equal("running"))
			Ω(pluginAppModel.Instances[0].Details).To(Equal("normal"))
			Ω(pluginAppModel.Instances[0].CpuUsage).To(Equal(float64(1.0)))
			Ω(pluginAppModel.Instances[0].DiskQuota).To(Equal(int64(1 * formatters.GIGABYTE)))
			Ω(pluginAppModel.Instances[0].DiskUsage).To(Equal(int64(32 * formatters.MEGABYTE)))
			Ω(pluginAppModel.Instances[0].MemQuota).To(Equal(int64(64 * formatters.MEGABYTE)))
			Ω(pluginAppModel.Instances[0].MemUsage).To(Equal(int64(13 * formatters.MEGABYTE)))

			Ω(pluginAppModel.Routes[1].Host).To(Equal("my-app"))
		})
	})

	Describe("displaying a summary of an app", func() {
		BeforeEach(func() {
			app = makeAppWithRoute("my-app")
			appInstance := models.AppInstanceFields{
				State:     models.InstanceRunning,
				Since:     testtime.MustParse("Mon Jan 2 15:04:05 -0700 MST 2006", "Mon Jan 2 15:04:05 -0700 MST 2012"),
				CpuUsage:  1.0,
				DiskQuota: 1 * formatters.GIGABYTE,
				DiskUsage: 32 * formatters.MEGABYTE,
				MemQuota:  64 * formatters.MEGABYTE,
				MemUsage:  13 * formatters.BYTE,
			}

			appInstance2 := models.AppInstanceFields{
				State:   models.InstanceDown,
				Details: "failure",
				Since:   testtime.MustParse("Mon Jan 2 15:04:05 -0700 MST 2006", "Mon Apr 1 15:04:05 -0700 MST 2012"),
			}

			instances := []models.AppInstanceFields{appInstance, appInstance2}

			appSummaryRepo.GetSummarySummary = app
			appInstancesRepo.GetInstancesReturns(instances, nil)
			requirementsFactory.Application = app
			updateCommandDependency(false)
		})

		Context("When app is a diego app", func() {
			It("uses noaa log library to gather metrics", func() {
				app.Diego = true
				appSummaryRepo.GetSummarySummary = app
				requirementsFactory.Application = app

				updateCommandDependency(false)
				runCommand("my-app")
				Ω(appLogsNoaaRepo.GetContainerMetricsCallCount()).To(Equal(1))
			})
			It("gracefully handles when /instances is down but /noaa is not", func() {
				app.Diego = true
				appSummaryRepo.GetSummarySummary = app
				requirementsFactory.Application = app
				appInstancesRepo.GetInstancesReturns([]models.AppInstanceFields{}, errors.New("danger will robinson"))
				updateCommandDependency(false)

				runCommand("my-app")
				Ω(appLogsNoaaRepo.GetContainerMetricsCallCount()).To(Equal(0))

			})
		})

		Context("When app is not a diego app", func() {
			It("does not use noaa log library to gather metrics", func() {
				app.Diego = false
				appSummaryRepo.GetSummarySummary = app
				requirementsFactory.Application = app

				updateCommandDependency(false)
				runCommand("my-app")
				Ω(appLogsNoaaRepo.GetContainerMetricsCallCount()).To(Equal(0))
			})
		})

		Context("Displaying buildpack info", func() {
			It("Shows 'Buildpack' when buildpack is set", func() {
				app.Diego = false
				app.Buildpack = "go_buildpack"
				app.DetectedBuildpack = "should_not_display"
				appSummaryRepo.GetSummarySummary = app
				requirementsFactory.Application = app

				updateCommandDependency(false)
				runCommand("my-app")

				Expect(appSummaryRepo.GetSummaryAppGuid).To(Equal("app-guid"))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"buildpack", "go_buildpack"},
				))
			})

			It("Shows 'DetectedBuildpack' when detected buildpack is set and 'Buildpack' is not set", func() {
				app.Diego = false
				app.DetectedBuildpack = "go_buildpack"
				appSummaryRepo.GetSummarySummary = app
				requirementsFactory.Application = app

				updateCommandDependency(false)
				runCommand("my-app")

				Expect(appSummaryRepo.GetSummaryAppGuid).To(Equal("app-guid"))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"buildpack", "go_buildpack"},
				))
			})

			It("Shows 'Unknown' when there is no buildpack set", func() {
				app.Diego = false
				appSummaryRepo.GetSummarySummary = app
				requirementsFactory.Application = app

				updateCommandDependency(false)
				runCommand("my-app")

				Expect(appSummaryRepo.GetSummaryAppGuid).To(Equal("app-guid"))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"buildpack", "unknown"},
				))
			})

		})

		It("displays a summary of the app", func() {
			app.Diego = false
			appSummaryRepo.GetSummarySummary = app
			requirementsFactory.Application = app

			updateCommandDependency(false)
			runCommand("my-app")

			Expect(appSummaryRepo.GetSummaryAppGuid).To(Equal("app-guid"))

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Showing health and status", "my-app"},
				[]string{"state", "started"},
				[]string{"instances", "2/2"},
				[]string{"usage", "256M x 2 instances"},
				[]string{"urls", "my-app.example.com", "foo.example.com"},
				[]string{"last uploaded", "Tue Nov 10 15:00:00 UTC 2009"},
				[]string{"#0", "running", "2012-01-02 03:04:05 PM", "100.0%", "13 of 64M", "32M of 1G"},
				[]string{"#1", "down", "2012-04-01 03:04:05 PM", "0%", "0 of 0", "0 of 0", "failure"},
				[]string{"stack", "fake_stack"},
			))
		})

		Describe("when the package updated at is nil", func() {
			BeforeEach(func() {
				appSummaryRepo.GetSummarySummary.PackageUpdatedAt = nil
				updateCommandDependency(false)
			})

			It("should output whatever greg sez", func() {
				runCommand("my-app")
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"last uploaded", "unknown"},
				))
			})
		})
	})

	Describe("when the app is not running", func() {
		BeforeEach(func() {
			application := models.Application{}
			application.Name = "my-app"
			application.Guid = "my-app-guid"
			application.State = "stopped"
			application.InstanceCount = 2
			application.RunningInstances = 0
			application.Memory = 256
			now := time.Now()
			application.PackageUpdatedAt = &now

			appSummaryRepo.GetSummarySummary = application
			requirementsFactory.Application = application

			updateCommandDependency(false)
		})

		It("displays nice output when the app is stopped", func() {
			appSummaryRepo.GetSummaryErrorCode = errors.APP_STOPPED

			updateCommandDependency(false)
			runCommand("my-app")

			Expect(appSummaryRepo.GetSummaryAppGuid).To(Equal("my-app-guid"))
			Expect(appInstancesRepo.GetInstancesArgsForCall(0)).To(Equal("my-app-guid"))

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Showing health and status", "my-app", "my-org", "my-space", "my-user"},
				[]string{"state", "stopped"},
				[]string{"instances", "0/2"},
				[]string{"usage", "256M x 2 instances"},
				[]string{"no running instances"},
			))
		})

		It("displays nice output when the app has not yet finished staging", func() {
			appSummaryRepo.GetSummaryErrorCode = errors.APP_NOT_STAGED
			runCommand("my-app")

			Expect(appSummaryRepo.GetSummaryAppGuid).To(Equal("my-app-guid"))
			Expect(appInstancesRepo.GetInstancesArgsForCall(0)).To(Equal("my-app-guid"))

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Showing health and status", "my-app", "my-org", "my-space", "my-user"},
				[]string{"state", "stopped"},
				[]string{"instances", "0/2"},
				[]string{"usage", "256M x 2 instances"},
				[]string{"no running instances"},
			))
		})
	})

	Describe("when running instances is unknown", func() {
		BeforeEach(func() {
			app := makeAppWithRoute("my-app")
			app.RunningInstances = -1
			appInstance := models.AppInstanceFields{
				State:     models.InstanceRunning,
				Since:     testtime.MustParse("Mon Jan 2 15:04:05 -0700 MST 2006", "Mon Jan 2 15:04:05 -0700 MST 2012"),
				CpuUsage:  5.0,
				DiskQuota: 4 * formatters.GIGABYTE,
				DiskUsage: 3 * formatters.GIGABYTE,
				MemQuota:  2 * formatters.GIGABYTE,
				MemUsage:  1 * formatters.GIGABYTE,
			}

			appInstance2 := models.AppInstanceFields{
				State: models.InstanceRunning,
				Since: testtime.MustParse("Mon Jan 2 15:04:05 -0700 MST 2006", "Mon Apr 1 15:04:05 -0700 MST 2012"),
			}

			instances := []models.AppInstanceFields{appInstance, appInstance2}

			appSummaryRepo.GetSummarySummary = app
			appInstancesRepo.GetInstancesReturns(instances, nil)
			requirementsFactory.Application = app

			updateCommandDependency(false)
		})

		It("displays a '?' for running instances", func() {
			runCommand("my-app")

			Expect(appSummaryRepo.GetSummaryAppGuid).To(Equal("app-guid"))

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Showing health and status", "my-app"},
				[]string{"state", "started"},
				[]string{"instances", "?/2"},
				[]string{"usage", "256M x 2 instances"},
				[]string{"urls", "my-app.example.com", "foo.example.com"},
				[]string{"#0", "running", "2012-01-02 03:04:05 PM", "500.0%", "1G of 2G", "3G of 4G"},
				[]string{"#1", "running", "2012-04-01 03:04:05 PM", "0%", "0 of 0", "0 of 0"},
			))
		})
	})

	Describe("when the user passes the --guid flag", func() {
		var app models.Application
		BeforeEach(func() {
			app = makeAppWithRoute("my-app")

			requirementsFactory.Application = app
		})

		It("displays guid for the requested app", func() {
			runCommand("--guid", "my-app")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{app.Guid},
			))
			Expect(ui.Outputs).ToNot(ContainSubstrings(
				[]string{"Showing health and status", "my-app"},
			))
		})
	})
})

func makeAppWithRoute(appName string) models.Application {
	application := models.Application{}
	application.Name = appName
	application.Guid = "app-guid"
	application.BuildpackUrl = "http://123.com"
	application.Command = "command1"
	application.Diego = false
	application.DetectedStartCommand = "detected_command"
	application.DiskQuota = 100
	application.EnvironmentVars = map[string]interface{}{"test": 123}
	application.RunningInstances = 2
	application.HealthCheckTimeout = 100
	application.SpaceGuid = "guids_in_spaaace"
	application.PackageState = "STAGED"
	application.StagingFailedReason = "no reason"
	application.State = "started"
	application.InstanceCount = 2
	application.RunningInstances = 2
	application.Memory = 256

	t := time.Date(2009, time.November, 10, 15, 0, 0, 0, time.UTC)
	application.PackageUpdatedAt = &t

	services := models.ServicePlanSummary{
		Guid: "s1-guid",
		Name: "s1-service",
	}

	application.Services = []models.ServicePlanSummary{services}

	domain := models.DomainFields{Guid: "domain1-guid", Name: "example.com", OwningOrganizationGuid: "org-123", Shared: true}

	route := models.RouteSummary{Host: "foo", Guid: "foo-guid", Domain: domain}
	secondRoute := models.RouteSummary{Host: appName, Domain: domain}

	application.Stack = &models.Stack{
		Name: "fake_stack",
		Guid: "123-123-123",
	}
	application.Routes = []models.RouteSummary{route, secondRoute}

	return application
}

package application_test

import (
	"time"

	testAppInstanaces "github.com/cloudfoundry/cli/cf/api/app_instances/fakes"
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	. "github.com/cloudfoundry/cli/cf/commands/application"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/formatters"
	"github.com/cloudfoundry/cli/cf/models"
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
		configRepo          core_config.ReadWriter
		appSummaryRepo      *testapi.FakeAppSummaryRepo
		appInstancesRepo    *testAppInstanaces.FakeAppInstancesRepository
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		appSummaryRepo = &testapi.FakeAppSummaryRepo{}
		appInstancesRepo = &testAppInstanaces.FakeAppInstancesRepository{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{
			LoginSuccess:         true,
			TargetedSpaceSuccess: true,
		}
	})

	runCommand := func(args ...string) bool {
		cmd := NewShowApp(ui, configRepo, appSummaryRepo, appInstancesRepo)
		return testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails if not logged in", func() {
			requirementsFactory.LoginSuccess = false
			Expect(runCommand("cf-plays-dwarf-fortress")).To(BeFalse())
		})

		It("fails if a space is not targeted", func() {
			requirementsFactory.TargetedSpaceSuccess = false
			Expect(runCommand("cf-plays-dwarf-fortress")).To(BeFalse())
		})

		It("fails with usage when no arguments are given", func() {
			passed := runCommand()
			Expect(ui.FailedWithUsage).To(BeTrue())
			Expect(passed).To(BeFalse())
		})

	})

	Describe("displaying a summary of an app", func() {
		BeforeEach(func() {
			app := makeAppWithRoute("my-app")
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
				State: models.InstanceDown,
				Since: testtime.MustParse("Mon Jan 2 15:04:05 -0700 MST 2006", "Mon Apr 1 15:04:05 -0700 MST 2012"),
			}

			instances := []models.AppInstanceFields{appInstance, appInstance2}

			appSummaryRepo.GetSummarySummary = app
			appInstancesRepo.GetInstancesReturns(instances, nil)
			requirementsFactory.Application = app
		})

		It("displays a summary of the app", func() {
			runCommand("my-app")

			Expect(appSummaryRepo.GetSummaryAppGuid).To(Equal("app-guid"))

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Showing health and status", "my-app"},
				[]string{"state", "started"},
				[]string{"instances", "2/2"},
				[]string{"usage", "256M x 2 instances"},
				[]string{"urls", "my-app.example.com", "foo.example.com"},
				[]string{"last uploaded", "Wed Oct 24 19:54:00 UTC 2012"},
				[]string{"#0", "running", "2012-01-02 03:04:05 PM", "100.0%", "13 of 64M", "32M of 1G"},
				[]string{"#1", "down", "2012-04-01 03:04:05 PM", "0%", "0 of 0", "0 of 0"},
			))
		})

		Describe("when the package updated at is nil", func() {
			BeforeEach(func() {
				appSummaryRepo.GetSummarySummary.PackageUpdatedAt = nil
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
		})

		It("displays nice output when the app is stopped", func() {
			appSummaryRepo.GetSummaryErrorCode = errors.APP_STOPPED
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
		})

		It("displays a '?' for running instances", func() {
			runCommand("my-app")

			Expect(appSummaryRepo.GetSummaryAppGuid).To(Equal("app-guid"))
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

	domain := models.DomainFields{}
	domain.Name = "example.com"

	route := models.RouteSummary{Host: "foo", Domain: domain}
	secondRoute := models.RouteSummary{Host: appName, Domain: domain}
	packgeUpdatedAt, _ := time.Parse("2006-01-02T15:04:05Z07:00", "2012-10-24T19:54:00Z")

	application.State = "started"
	application.InstanceCount = 2
	application.RunningInstances = 2
	application.Memory = 256
	application.Routes = []models.RouteSummary{route, secondRoute}
	application.PackageUpdatedAt = &packgeUpdatedAt

	return application
}

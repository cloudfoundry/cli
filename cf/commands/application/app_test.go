package application_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	. "github.com/cloudfoundry/cli/cf/commands/application"
	"github.com/cloudfoundry/cli/cf/configuration"
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
		configRepo          configuration.ReadWriter
		appSummaryRepo      *testapi.FakeAppSummaryRepo
		appStatsRepo        *testapi.FakeAppStatsRepo
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		appSummaryRepo = &testapi.FakeAppSummaryRepo{}
		appStatsRepo = &testapi.FakeAppStatsRepo{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{
			LoginSuccess:         true,
			TargetedSpaceSuccess: true,
		}
	})

	runCommand := func(args ...string) {
		cmd := NewShowApp(ui, configRepo, appSummaryRepo, appStatsRepo)
		testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails if not logged in", func() {
			requirementsFactory.LoginSuccess = false
			runCommand("cf-plays-dwarf-fortress")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("fails if a space is not targeted", func() {
			requirementsFactory.TargetedSpaceSuccess = false
			runCommand("cf-plays-dwarf-fortress")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("fails with usage when no arguments are given", func() {
			runCommand()
			Expect(ui.FailedWithUsage).To(BeTrue())
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

	})

	Describe("displaying a summary of an app", func() {
		BeforeEach(func() {
			app := makeAppWithRoute("my-app")

			curTime := testtime.MustParse("Mon Jan 2 15:04:05 -0700 MST 2006", "Mon Jul 1 15:04:05 -0700 MST 2014")
			startTime := testtime.MustParse("Mon Jan 2 15:04:05 -0700 MST 2006", "Mon Jan 2 15:04:05 -0700 MST 2014")

			uptime := uint64((curTime.Sub(startTime)).Seconds())

			appStats := models.AppStatsFields{
				State: models.InstanceRunning,
				Stats: models.AppStatsStatsFields{
					Uptime:    uptime,
					DiskQuota: 1 * formatters.GIGABYTE,
					MemQuota:  64 * formatters.MEGABYTE,
					Usage: models.AppStatsUsageFields{
						Time: curTime.Format("2006-01-02 15:04:05 +0000"),
						Disk: 32 * formatters.MEGABYTE,
						Cpu:  1.0,
						Mem:  13 * formatters.BYTE,
					},
				},
			}

			startTime2 := testtime.MustParse("Mon Jan 2 15:04:05 -0700 MST 2006", "Mon Apr 1 15:04:05 -0700 MST 2012")
			uptime2 := uint64((curTime.Sub(startTime2)).Seconds())
			appStats2 := models.AppStatsFields{
				State: models.InstanceDown,
				Stats: models.AppStatsStatsFields{
					Uptime: uptime2,
					Usage: models.AppStatsUsageFields{
						Time: curTime.Format("2006-01-02 15:04:05 +0000"),
					},
				},
			}

			stats := []models.AppStatsFields{appStats, appStats2}

			appSummaryRepo.GetSummarySummary = app
			appStatsRepo.GetStatsResponses = [][]models.AppStatsFields{stats}
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
				[]string{"#0", "running", "2014-01-02 03:04:05 PM", "100.0%", "13 of 64M", "32M of 1G"},
				[]string{"#1", "down", "2012-04-01 03:04:05 PM", "0%", "0 of 0", "0 of 0"},
			))
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

			appSummaryRepo.GetSummarySummary = application
			requirementsFactory.Application = application
		})

		It("displays nice output when the app is stopped", func() {
			appSummaryRepo.GetSummaryErrorCode = errors.APP_STOPPED
			runCommand("my-app")

			Expect(appSummaryRepo.GetSummaryAppGuid).To(Equal("my-app-guid"))
			Expect(appStatsRepo.GetStatsAppGuid).To(Equal("my-app-guid"))

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
			Expect(appStatsRepo.GetStatsAppGuid).To(Equal("my-app-guid"))

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Showing health and status", "my-app", "my-org", "my-space", "my-user"},
				[]string{"state", "stopped"},
				[]string{"instances", "0/2"},
				[]string{"usage", "256M x 2 instances"},
				[]string{"no running instances"},
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

	application.State = "started"
	application.InstanceCount = 2
	application.RunningInstances = 2
	application.Memory = 256
	application.Routes = []models.RouteSummary{route, secondRoute}

	return application
}

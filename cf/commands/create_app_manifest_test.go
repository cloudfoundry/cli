package commands_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	testAppInstanaces "github.com/cloudfoundry/cli/cf/api/app_instances/fakes"
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	. "github.com/cloudfoundry/cli/cf/commands"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
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

var _ = Describe("create-app-manifest Command", func() {
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
		cmd := NewCreateAppManifest(ui, configRepo, appSummaryRepo, appInstancesRepo)
		return testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	FDescribe("requirements", func() {
		It("fails if not logged in", func() {
			requirementsFactory.LoginSuccess = false
			Expect(runCommand("cf-plays-dwarf-fortress")).To(BeFalse())
		})

		It("fails if a space is not targeted", func() {
			requirementsFactory.TargetedSpaceSuccess = false
			Expect(runCommand("cf-plays-dwarf-fortress")).To(BeFalse())
		})

		It("fails with usage when not provided exactly one arg", func() {
			passed := runCommand()
			Expect(ui.FailedWithUsage).To(BeTrue())
			Expect(passed).To(BeFalse())
		})

	})

	Describe("displaying a summary of an app", func() {
		var (
			manifestPath string
		)

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

			wd, err := os.Getwd()
			Expect(err).ToNot(HaveOccurred())

			manifestPath = filepath.Join(wd, "my-app-manifest.yml")
		})

		It("downloads a manifest for the app", func() {
			runCommand("my-app")

			Expect(appSummaryRepo.GetSummaryAppGuid).To(Equal("app-guid"))

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Creating an app manifest from current settings of app", "my-app"},
				[]string{"Manifest has been created in"},
			))

			_, err := os.Stat(manifestPath)
			Expect(err).ToNot(HaveOccurred())
		})

		It("the manifest contains information about the app", func() {
			//expect file contents to be right name memory etc
			_, err := os.Stat(manifestPath)
			file, err := ioutil.ReadFile(manifestPath)
			Expect(err).ToNot(HaveOccurred())
			fileContents := string(file)

			Expect(fileContents).To(ContainSubstrings(
				[]string{"name:", "my-app"},
			))
		})

		//it downloads to a different path when -p is provided
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

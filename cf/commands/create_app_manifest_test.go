package commands_test

import (
	"time"

	testAppInstanaces "github.com/cloudfoundry/cli/cf/api/app_instances/fakes"
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	. "github.com/cloudfoundry/cli/cf/commands"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/formatters"
	testManifest "github.com/cloudfoundry/cli/cf/manifest/fakes"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	testtime "github.com/cloudfoundry/cli/testhelpers/time"

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
		fakeManifest        *testManifest.FakeAppManifest
	)

	BeforeEach(func() {
		fakeManifest = &testManifest.FakeAppManifest{}
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
		cmd := NewCreateAppManifest(ui, configRepo, appSummaryRepo, fakeManifest)
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

		It("fails with usage when not provided exactly one arg", func() {
			passed := runCommand()
			Expect(ui.FailedWithUsage).To(BeTrue())
			Expect(passed).To(BeFalse())
		})

	})

	Describe("creating app manifest", func() {
		var (
			appInstance  models.AppInstanceFields
			appInstance2 models.AppInstanceFields
			instances    []models.AppInstanceFields
		)

		BeforeEach(func() {
			appInstance = models.AppInstanceFields{
				State:     models.InstanceRunning,
				Since:     testtime.MustParse("Mon Jan 2 15:04:05 -0700 MST 2006", "Mon Jan 2 15:04:05 -0700 MST 2012"),
				CpuUsage:  1.0,
				DiskQuota: 1 * formatters.GIGABYTE,
				DiskUsage: 32 * formatters.MEGABYTE,
				MemQuota:  64 * formatters.MEGABYTE,
				MemUsage:  13 * formatters.BYTE,
			}

			appInstance2 = models.AppInstanceFields{
				State: models.InstanceDown,
				Since: testtime.MustParse("Mon Jan 2 15:04:05 -0700 MST 2006", "Mon Apr 1 15:04:05 -0700 MST 2012"),
			}

			instances = []models.AppInstanceFields{appInstance, appInstance2}
			appInstancesRepo.GetInstancesReturns(instances, nil)
		})

		Context("app with Services, Routes, Environment Vars", func() {
			BeforeEach(func() {
				app := makeAppWithOptions("my-app")
				appSummaryRepo.GetSummarySummary = app
				requirementsFactory.Application = app
			})

			It("creates a manifest with services, routes and environment vars", func() {
				runCommand("my-app")
				Ω(fakeManifest.MemoryCallCount()).To(Equal(1))
				Ω(fakeManifest.EnvironmentVarsCallCount()).To(Equal(1))
				Ω(fakeManifest.HealthCheckTimeoutCallCount()).To(Equal(1))
				Ω(fakeManifest.InstancesCallCount()).To(Equal(1))
				Ω(fakeManifest.DomainCallCount()).To(Equal(1))
				Ω(fakeManifest.ServiceCallCount()).To(Equal(1))
				Ω(fakeManifest.StartupCommandCallCount()).To(Equal(1))
			})
		})

		Context("Env Vars will be written in aplhabetical order", func() {
			BeforeEach(func() {
				app := makeAppWithMultipleEnvVars("my-app")
				appSummaryRepo.GetSummarySummary = app
				requirementsFactory.Application = app
			})

			It("calls manifest EnvironmentVars() aphlhabetically", func() {
				runCommand("my-app")
				Ω(fakeManifest.EnvironmentVarsCallCount()).To(Equal(4))
				_, k, _ := fakeManifest.EnvironmentVarsArgsForCall(0)
				Ω(k).To(Equal("abc"))
				_, k, _ = fakeManifest.EnvironmentVarsArgsForCall(1)
				Ω(k).To(Equal("bar"))
				_, k, _ = fakeManifest.EnvironmentVarsArgsForCall(2)
				Ω(k).To(Equal("foo"))
				_, k, _ = fakeManifest.EnvironmentVarsArgsForCall(3)
				Ω(k).To(Equal("xyz"))
			})
		})

		Context("Env Vars can be in different types (string, float64, bool)", func() {
			BeforeEach(func() {
				app := makeAppWithMultipleEnvVars("my-app")
				appSummaryRepo.GetSummarySummary = app
				requirementsFactory.Application = app
			})

			It("calls manifest EnvironmentVars() aphlhabetically", func() {
				runCommand("my-app")
				Ω(fakeManifest.EnvironmentVarsCallCount()).To(Equal(4))
				_, _, v := fakeManifest.EnvironmentVarsArgsForCall(0)
				Ω(v).To(Equal("\"abc\""))
				_, _, v = fakeManifest.EnvironmentVarsArgsForCall(1)
				Ω(v).To(Equal("10"))
				_, _, v = fakeManifest.EnvironmentVarsArgsForCall(2)
				Ω(v).To(Equal("true"))
				_, _, v = fakeManifest.EnvironmentVarsArgsForCall(3)
				Ω(v).To(Equal("false"))
			})
		})

		Context("app without Services, Routes, Environment Vars", func() {
			BeforeEach(func() {
				app := makeAppWithoutOptions("my-app")
				appSummaryRepo.GetSummarySummary = app
				requirementsFactory.Application = app
			})

			It("creates a manifest with services, routes and environment vars", func() {
				runCommand("my-app")
				Ω(fakeManifest.MemoryCallCount()).To(Equal(1))
				Ω(fakeManifest.EnvironmentVarsCallCount()).To(Equal(0))
				Ω(fakeManifest.HealthCheckTimeoutCallCount()).To(Equal(0))
				Ω(fakeManifest.InstancesCallCount()).To(Equal(1))
				Ω(fakeManifest.DomainCallCount()).To(Equal(0))
				Ω(fakeManifest.ServiceCallCount()).To(Equal(0))
			})
		})

		Context("when the flag -p is supplied", func() {
			BeforeEach(func() {
				app := makeAppWithoutOptions("my-app")
				appSummaryRepo.GetSummarySummary = app
				requirementsFactory.Application = app
			})

			It("creates a manifest with services, routes and environment vars", func() {
				filePath := "another/location/manifest.yml"
				runCommand("-p", filePath, "my-app")
				Ω(fakeManifest.FileSavePathArgsForCall(0)).To(Equal(filePath))
			})
		})

		Context("when no -p flag is supplied", func() {
			BeforeEach(func() {
				app := makeAppWithoutOptions("my-app2")
				appSummaryRepo.GetSummarySummary = app
				requirementsFactory.Application = app
			})

			It("creates a manifest named <app-name>_manifest.yml", func() {
				runCommand("my-app2")
				Ω(fakeManifest.FileSavePathArgsForCall(0)).To(Equal("./my-app2_manifest.yml"))
			})
		})

	})
})

func makeAppWithOptions(appName string) models.Application {
	application := models.Application{}
	application.Name = appName
	application.Guid = "app-guid"
	application.Command = "run main.go"

	domain := models.DomainFields{}
	domain.Name = "example.com"

	route := models.RouteSummary{Host: "foo", Domain: domain}
	secondRoute := models.RouteSummary{Host: appName, Domain: domain}
	packgeUpdatedAt, _ := time.Parse("2006-01-02T15:04:05Z07:00", "2012-10-24T19:54:00Z")

	application.State = "started"
	application.InstanceCount = 2
	application.RunningInstances = 2
	application.Memory = 256
	application.HealthCheckTimeout = 100
	application.Routes = []models.RouteSummary{route, secondRoute}
	application.PackageUpdatedAt = &packgeUpdatedAt

	envMap := make(map[string]interface{})
	envMap["foo"] = "bar"
	application.EnvironmentVars = envMap

	application.Services = append(application.Services, models.ServicePlanSummary{
		Guid: "",
		Name: "server1",
	})

	return application
}

func makeAppWithoutOptions(appName string) models.Application {
	application := models.Application{}
	application.Name = appName
	application.Guid = "app-guid"
	packgeUpdatedAt, _ := time.Parse("2006-01-02T15:04:05Z07:00", "2012-10-24T19:54:00Z")

	application.State = "started"
	application.InstanceCount = 2
	application.RunningInstances = 2
	application.Memory = 256
	application.PackageUpdatedAt = &packgeUpdatedAt

	return application
}

func makeAppWithMultipleEnvVars(appName string) models.Application {
	application := models.Application{}
	application.Name = appName
	application.Guid = "app-guid"
	packgeUpdatedAt, _ := time.Parse("2006-01-02T15:04:05Z07:00", "2012-10-24T19:54:00Z")

	application.State = "started"
	application.InstanceCount = 2
	application.RunningInstances = 2
	application.Memory = 256
	application.PackageUpdatedAt = &packgeUpdatedAt

	envMap := make(map[string]interface{})
	envMap["foo"] = bool(true)
	envMap["abc"] = "abc"
	envMap["xyz"] = bool(false)
	envMap["bar"] = float64(10)
	application.EnvironmentVars = envMap

	return application
}

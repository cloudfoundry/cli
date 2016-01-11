package application_test

import (
	"encoding/json"
	"time"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/commands/application"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/formatters"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/plugin/models"

	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	fakeappinstances "github.com/cloudfoundry/cli/cf/api/app_instances/fakes"
	fakeapi "github.com/cloudfoundry/cli/cf/api/fakes"
	fakerequirements "github.com/cloudfoundry/cli/cf/requirements/fakes"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type passingRequirement struct {
	Name string
}

func (r passingRequirement) Execute() bool {
	return true
}

var _ = Describe("App", func() {
	var (
		ui               *testterm.FakeUI
		appSummaryRepo   *fakeapi.FakeAppSummaryRepository
		appInstancesRepo *fakeappinstances.FakeAppInstancesRepository
		getAppModel      *plugin_models.GetAppModel

		cmd         command_registry.Command
		deps        command_registry.Dependency
		factory     *fakerequirements.FakeFactory
		flagContext flags.FlagContext

		loginRequirement         requirements.Requirement
		targetedSpaceRequirement requirements.Requirement
		applicationRequirement   *fakerequirements.FakeApplicationRequirement
	)

	BeforeEach(func() {
		cmd = &application.ShowApp{}
		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)

		ui = &testterm.FakeUI{}

		getAppModel = &plugin_models.GetAppModel{}

		repoLocator := api.RepositoryLocator{}
		appSummaryRepo = &fakeapi.FakeAppSummaryRepository{}
		repoLocator = repoLocator.SetAppSummaryRepository(appSummaryRepo)
		appInstancesRepo = &fakeappinstances.FakeAppInstancesRepository{}
		repoLocator = repoLocator.SetAppInstancesRepository(appInstancesRepo)

		deps = command_registry.Dependency{
			Ui:     ui,
			Config: testconfig.NewRepositoryWithDefaults(),
			PluginModels: &command_registry.PluginModels{
				Application: getAppModel,
			},
			RepoLocator: repoLocator,
		}

		cmd.SetDependency(deps, false)

		factory = &fakerequirements.FakeFactory{}

		loginRequirement = &passingRequirement{}
		factory.NewLoginRequirementReturns(loginRequirement)

		targetedSpaceRequirement = &passingRequirement{}
		factory.NewTargetedSpaceRequirementReturns(targetedSpaceRequirement)

		applicationRequirement = &fakerequirements.FakeApplicationRequirement{}
		factory.NewApplicationRequirementReturns(applicationRequirement)
	})

	Describe("Requirements", func() {
		Context("when not provided exactly one arg", func() {
			BeforeEach(func() {
				flagContext.Parse("app-name", "extra-arg")
			})

			It("fails with usage", func() {
				Expect(func() { cmd.Requirements(factory, flagContext) }).To(Panic())
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Incorrect Usage. Requires an argument"},
					[]string{"NAME"},
					[]string{"USAGE"},
				))
			})
		})

		Context("when provided exactly one arg", func() {
			BeforeEach(func() {
				flagContext.Parse("app-name")
			})

			It("returns a LoginRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(factory.NewLoginRequirementCallCount()).To(Equal(1))

				Expect(actualRequirements).To(ContainElement(loginRequirement))
			})

			It("returns a TargetedSpaceRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(factory.NewTargetedSpaceRequirementCallCount()).To(Equal(1))

				Expect(actualRequirements).To(ContainElement(targetedSpaceRequirement))
			})

			It("returns an ApplicationRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(factory.NewApplicationRequirementCallCount()).To(Equal(1))
				Expect(factory.NewApplicationRequirementArgsForCall(0)).To(Equal("app-name"))

				Expect(actualRequirements).To(ContainElement(applicationRequirement))
			})
		})
	})

	Describe("Execute", func() {
		var (
			getApplicationModel models.Application
			getAppSummaryModel  models.Application
			appInstanceFields   []models.AppInstanceFields
			getAppSummaryErr    error
		)

		BeforeEach(func() {
			err := flagContext.Parse("app-name")
			Expect(err).NotTo(HaveOccurred())
			_, err = cmd.Requirements(factory, flagContext)
			Expect(err).NotTo(HaveOccurred())

			paginatedApplicationResources := resources.PaginatedApplicationResources{}
			err = json.Unmarshal([]byte(getApplicationJSON), &paginatedApplicationResources)
			Expect(err).NotTo(HaveOccurred())

			getApplicationModel = paginatedApplicationResources.Resources[0].ToModel()

			applicationFromSummary := api.ApplicationFromSummary{}
			err = json.Unmarshal([]byte(getSummaryJSON), &applicationFromSummary)
			Expect(err).NotTo(HaveOccurred())

			getAppSummaryModel = applicationFromSummary.ToModel()

			appInstanceFields = []models.AppInstanceFields{
				{
					State:     models.InstanceRunning,
					Details:   "fake-instance-details",
					Since:     time.Date(2015, time.November, 19, 1, 1, 17, 0, time.UTC),
					CpuUsage:  float64(0.25),
					DiskUsage: int64(1 * formatters.GIGABYTE),
					DiskQuota: int64(2 * formatters.GIGABYTE),
					MemUsage:  int64(24 * formatters.MEGABYTE),
					MemQuota:  int64(32 * formatters.MEGABYTE),
				},
			}

			applicationRequirement.GetApplicationReturns(getApplicationModel)
			appSummaryRepo.GetSummaryReturns(getAppSummaryModel, getAppSummaryErr)
			appInstancesRepo.GetInstancesReturns(appInstanceFields, nil)
		})

		It("gets the application summary", func() {
			cmd.Execute(flagContext)
			Expect(appSummaryRepo.GetSummaryCallCount()).To(Equal(1))
		})

		It("gets the app instances", func() {
			cmd.Execute(flagContext)
			Expect(appInstancesRepo.GetInstancesCallCount()).To(Equal(1))
		})

		It("gets the application from the application requirement", func() {
			cmd.Execute(flagContext)
			Expect(applicationRequirement.GetApplicationCallCount()).To(Equal(1))
		})

		It("prints a summary of the app", func() {
			cmd.Execute(flagContext)
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Showing health and status for app fake-app-name"},
				[]string{"requested state: started"},
				[]string{"instances: 1/1"},
				[]string{"usage: 1G x 1 instances"},
				[]string{"urls: fake-route-host.fake-route-domain-name"},
				[]string{"last uploaded: Thu Nov 19 01:00:15 UTC 2015"},
				[]string{"stack: fake-stack-name"},
				// buildpack tested separately
				[]string{"#0", "running", "2015-11-19 01:01:17 AM", "25.0%", "24M of 32M", "1G of 2G"},
			))
		})

		Context("when getting the application summary fails because the app is stopped", func() {
			BeforeEach(func() {
				getAppSummaryModel.RunningInstances = 0
				getAppSummaryModel.InstanceCount = 1
				getAppSummaryModel.State = "stopped"
				appSummaryRepo.GetSummaryReturns(getAppSummaryModel, errors.NewHttpError(400, errors.APP_STOPPED, "error"))
			})

			It("prints appropriate output", func() {
				cmd.Execute(flagContext)
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Showing health and status", "fake-app-name", "my-org", "my-space", "my-user"},
					[]string{"state", "stopped"},
					[]string{"instances", "0/1"},
					[]string{"usage", "1G x 1 instances"},
					[]string{"There are no running instances of this app."},
				))
			})
		})

		Context("when getting the application summary fails because the app has not yet finished staged", func() {
			BeforeEach(func() {
				getAppSummaryModel.RunningInstances = 0
				getAppSummaryModel.InstanceCount = 1
				getAppSummaryModel.State = "stopped"
				appSummaryRepo.GetSummaryReturns(getAppSummaryModel, errors.NewHttpError(400, errors.APP_NOT_STAGED, "error"))
			})

			It("prints appropriate output", func() {
				cmd.Execute(flagContext)
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Showing health and status", "fake-app-name", "my-org", "my-space", "my-user"},
					[]string{"state", "stopped"},
					[]string{"instances", "0/1"},
					[]string{"usage", "1G x 1 instances"},
					[]string{"There are no running instances of this app."},
				))
			})
		})

		Context("when getting the application summary fails for any other reason", func() {
			BeforeEach(func() {
				getAppSummaryModel.RunningInstances = 0
				getAppSummaryModel.InstanceCount = 1
				appSummaryRepo.GetSummaryReturns(getAppSummaryModel, errors.New("an-error"))
			})

			It("panics and prints a failure message", func() {
				Expect(func() { cmd.Execute(flagContext) }).To(Panic())
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"an-error"},
				))
			})

			Context("when the app is stopped", func() {
				BeforeEach(func() {
					getAppSummaryModel.State = "stopped"
					appSummaryRepo.GetSummaryReturns(getAppSummaryModel, errors.New("an-error"))
				})

				It("prints appropriate output", func() {
					cmd.Execute(flagContext)
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Showing health and status", "fake-app-name", "my-org", "my-space", "my-user"},
						[]string{"state", "stopped"},
						[]string{"instances", "0/1"},
						[]string{"usage", "1G x 1 instances"},
						[]string{"There are no running instances of this app."},
					))
				})
			})
		})

		Context("when getting the app instances fails", func() {
			BeforeEach(func() {
				appInstancesRepo.GetInstancesReturns([]models.AppInstanceFields{}, errors.New("an-error"))
			})

			It("panics and prints a failure message", func() {
				Expect(func() { cmd.Execute(flagContext) }).To(Panic())
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"an-error"},
				))
			})

			Context("when the app is stopped", func() {
				BeforeEach(func() {
					getAppSummaryModel.RunningInstances = 0
					getAppSummaryModel.State = "stopped"
					appSummaryRepo.GetSummaryReturns(getAppSummaryModel, nil)
				})

				It("prints appropriate output", func() {
					cmd.Execute(flagContext)
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Showing health and status", "fake-app-name", "my-org", "my-space", "my-user"},
						[]string{"state", "stopped"},
						[]string{"instances", "0/1"},
						[]string{"usage", "1G x 1 instances"},
						[]string{"There are no running instances of this app."},
					))
				})
			})
		})

		Context("when the package updated at is missing", func() {
			BeforeEach(func() {
				getAppSummaryModel.PackageUpdatedAt = nil
				appSummaryRepo.GetSummaryReturns(getAppSummaryModel, nil)
			})

			It("prints 'unknown' as last uploaded", func() {
				cmd.Execute(flagContext)

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"last uploaded: unknown"},
				))
			})
		})

		Context("when the GetApplication model includes a buildpack", func() {
			// this should be the GetAppSummary model
			BeforeEach(func() {
				getApplicationModel.Buildpack = "fake-buildpack"
				getApplicationModel.DetectedBuildpack = ""
				applicationRequirement.GetApplicationReturns(getApplicationModel)
			})

			It("prints the buildpack", func() {
				cmd.Execute(flagContext)

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"buildpack", "fake-buildpack"},
				))
			})
		})

		Context("when the GetApplication Model includes a detected buildpack", func() {
			// this should be the GetAppSummary model
			BeforeEach(func() {
				getApplicationModel.Buildpack = ""
				getApplicationModel.DetectedBuildpack = "fake-detected-buildpack"
				applicationRequirement.GetApplicationReturns(getApplicationModel)
			})

			It("prints the detected buildpack", func() {
				cmd.Execute(flagContext)

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"buildpack", "fake-detected-buildpack"},
				))
			})
		})

		Context("when the GetApplication Model does not include a buildpack or detected buildpack", func() {
			// this should be the GetAppSummary model
			BeforeEach(func() {
				getApplicationModel.Buildpack = ""
				getApplicationModel.DetectedBuildpack = ""
				applicationRequirement.GetApplicationReturns(getApplicationModel)
			})

			It("prints the 'unknown' as the buildpack", func() {
				cmd.Execute(flagContext)
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"buildpack", "unknown"},
				))
			})
		})

		Context("when running instances is -1", func() {
			BeforeEach(func() {
				getAppSummaryModel.RunningInstances = -1
				appSummaryRepo.GetSummaryReturns(getAppSummaryModel, nil)
			})

			It("displays a '?' for running instances", func() {
				cmd.Execute(flagContext)
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"instances", "?/1"},
				))
			})
		})

		Context("when the --guid flag is passed", func() {
			BeforeEach(func() {
				flagContext.Parse("app-name", "--guid")
			})

			It("only prints the guid for the app", func() {
				cmd.Execute(flagContext)
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"fake-app-guid"},
				))
				Expect(ui.Outputs).ToNot(ContainSubstrings(
					[]string{"Showing health and status", "my-app"},
				))
			})
		})

		Context("when called from a plugin", func() {
			BeforeEach(func() {
				cmd.SetDependency(deps, true)
			})

			It("populates the plugin model", func() {
				cmd.Execute(flagContext)

				// from AppRequirement model
				Expect(getAppModel.Stack.Name).To(Equal("fake-stack-name"))
				Expect(getAppModel.Stack.Guid).To(Equal("fake-stack-guid"))

				// from GetAppSummary model
				Expect(getAppModel.Name).To(Equal("fake-app-name"))
				Expect(getAppModel.State).To(Equal("started"))
				Expect(getAppModel.Guid).To(Equal("fake-app-guid"))
				Expect(getAppModel.Command).To(Equal("fake-command"))
				Expect(getAppModel.Diego).To(BeTrue())
				Expect(getAppModel.DetectedStartCommand).To(Equal("fake-detected-start-command"))
				Expect(getAppModel.DiskQuota).To(Equal(int64(1024)))
				Expect(getAppModel.EnvironmentVars).To(Equal(map[string]interface{}{"fake-env-var": "fake-env-var-value"}))
				Expect(getAppModel.InstanceCount).To(Equal(1))
				Expect(getAppModel.Memory).To(Equal(int64(1024)))
				Expect(getAppModel.RunningInstances).To(Equal(1))
				Expect(getAppModel.HealthCheckTimeout).To(Equal(0))
				Expect(getAppModel.SpaceGuid).To(Equal("fake-space-guid"))
				Expect(getAppModel.PackageUpdatedAt.String()).To(Equal(time.Date(2015, time.November, 19, 1, 0, 15, 0, time.UTC).String()))
				Expect(getAppModel.PackageState).To(Equal("STAGED"))
				Expect(getAppModel.StagingFailedReason).To(BeEmpty())
				Expect(getAppModel.BuildpackUrl).To(Equal("fake-buildpack"))
				Expect(getAppModel.Routes[0].Host).To(Equal("fake-route-host"))
				Expect(getAppModel.Routes[0].Guid).To(Equal("fake-route-guid"))
				Expect(getAppModel.Routes[0].Domain.Name).To(Equal("fake-route-domain-name"))
				Expect(getAppModel.Routes[0].Domain.Guid).To(Equal("fake-route-domain-guid"))
				Expect(getAppModel.Services[0].Guid).To(Equal("fake-service-guid"))
				Expect(getAppModel.Services[0].Name).To(Equal("fake-service-name"))

				// from GetInstances model
				Expect(getAppModel.Instances[0].State).To(Equal("running"))
				Expect(getAppModel.Instances[0].Details).To(Equal("fake-instance-details"))
				Expect(getAppModel.Instances[0].CpuUsage).To(Equal(float64(0.25)))
				Expect(getAppModel.Instances[0].DiskUsage).To(Equal(int64(1 * formatters.GIGABYTE)))
				Expect(getAppModel.Instances[0].DiskQuota).To(Equal(int64(2 * formatters.GIGABYTE)))
				Expect(getAppModel.Instances[0].MemUsage).To(Equal(int64(24 * formatters.MEGABYTE)))
				Expect(getAppModel.Instances[0].MemQuota).To(Equal(int64(32 * formatters.MEGABYTE)))
			})
		})
	})
})

var getApplicationJSON string = `{
  "total_results": 1,
  "total_pages": 1,
  "prev_url": null,
  "next_url": null,
  "resources": [
    {
      "metadata": {
        "guid": "fake-app-guid",
        "url": "fake-url",
        "created_at": "2015-11-19T01:00:12Z",
        "updated_at": "2015-11-19T01:01:04Z"
      },
      "entity": {
        "name": "fake-app-name",
        "production": false,
        "space_guid": "fake-space-guid",
        "stack_guid": "fake-stack-guid",
        "buildpack": null,
        "detected_buildpack": "fake-detected-buildpack",
				"environment_json": {
					"fake-env-var": "fake-env-var-value"
				},
				"memory": 1024,
        "instances": 1,
        "disk_quota": 1024,
        "state": "started",
        "version": "fake-version",
        "command": "fake-command",
        "console": false,
        "debug": null,
        "staging_task_id": "fake-staging-task-id",
        "package_state": "STAGED",
        "health_check_type": "port",
        "health_check_timeout": null,
        "staging_failed_reason": null,
        "staging_failed_description": null,
        "diego": true,
        "docker_image": null,
        "package_updated_at": "2015-11-19T01:00:15Z",
        "detected_start_command": "fake-detected-start-command",
        "enable_ssh": true,
        "docker_credentials_json": {
          "redacted_message": "[PRIVATE DATA HIDDEN]"
        },
        "ports": null,
        "space_url": "fake-space-url",
        "space": {
          "metadata": {
            "guid": "fake-space-guid",
            "url": "fake-space-url",
            "created_at": "2014-05-12T23:36:57Z",
            "updated_at": null
          },
          "entity": {
            "name": "fake-space-name",
            "organization_guid": "fake-space-organization-guid",
            "space_quota_definition_guid": null,
            "allow_ssh": true,
            "organization_url": "fake-space-organization-url",
            "developers_url": "fake-space-developers-url",
            "managers_url": "fake-space-managers-url",
            "auditors_url": "fake-space-auditors-url",
            "apps_url": "fake-space-apps-url",
            "routes_url": "fake-space-routes-url",
            "domains_url": "fake-space-domains-url",
            "service_instances_url": "fake-space-service-instances-url",
            "app_events_url": "fake-space-app-events-url",
            "events_url": "fake-space-events-url",
            "security_groups_url": "fake-space-security-groups-url"
          }
        },
        "stack_url": "fake-stack-url",
        "stack": {
          "metadata": {
            "guid": "fake-stack-guid",
            "url": "fake-stack-url",
            "created_at": "2015-03-04T18:58:42Z",
            "updated_at": null
          },
          "entity": {
            "name": "fake-stack-name",
            "description": "fake-stack-description"
          }
        },
        "events_url": "fake-events-url",
        "service_bindings_url": "fake-service-bindings-url",
        "service_bindings": [],
        "routes_url": "fake-routes-url",
        "routes": [
          {
            "metadata": {
              "guid": "fake-route-guid",
              "url": "fake-route-url",
              "created_at": "2014-05-13T21:38:42Z",
              "updated_at": null
            },
            "entity": {
              "host": "fake-route-host",
              "path": "",
              "domain_guid": "fake-route-domain-guid",
              "space_guid": "fake-route-space-guid",
              "service_instance_guid": null,
              "port": 0,
              "domain_url": "fake-route-domain-url",
              "space_url": "fake-route-space-url",
              "apps_url": "fake-route-apps-url"
            }
          }
        ]
      }
    }
  ]
}`

var getSummaryJSON string = `{
	"guid": "fake-app-guid",
	"name": "fake-app-name",
	"routes": [
	{
		"guid": "fake-route-guid",
		"host": "fake-route-host",
		"domain": {
			"guid": "fake-route-domain-guid",
			"name": "fake-route-domain-name"
		}
	}
	],
	"running_instances": 1,
	"services": [
	{
		"guid": "fake-service-guid",
		"name": "fake-service-name",
		"bound_app_count": 1,
		"last_operation": null,
		"dashboard_url": null,
		"service_plan": {
			"guid": "fake-service-plan-guid",
			"name": "fake-service-plan-name",
			"service": {
				"guid": "fake-service-plan-service-guid",
				"label": "fake-service-plan-service-label",
				"provider": null,
				"version": null
			}
		}
	}
	],
	"available_domains": [
	{
		"guid": "fake-available-domain-guid",
		"name": "fake-available-domain-name",
		"owning_organization_guid": "fake-owning-organization-guid"
	}
	],
	"production": false,
	"space_guid": "fake-space-guid",
	"stack_guid": "fake-stack-guid",
	"buildpack": "fake-buildpack",
	"detected_buildpack": "fake-detected-buildpack",
	"environment_json": {
		"fake-env-var": "fake-env-var-value"
	},
	"memory": 1024,
	"instances": 1,
	"disk_quota": 1024,
	"state": "STARTED",
	"version": "fake-version",
	"command": "fake-command",
	"console": false,
	"debug": null,
	"staging_task_id": "fake-staging-task-id",
	"package_state": "STAGED",
	"health_check_type": "port",
	"health_check_timeout": null,
	"staging_failed_reason": null,
	"staging_failed_description": null,
	"diego": true,
	"docker_image": null,
	"package_updated_at": "2015-11-19T01:00:15Z",
	"detected_start_command": "fake-detected-start-command",
	"enable_ssh": true,
	"docker_credentials_json": {
		"redacted_message": "[PRIVATE DATA HIDDEN]"
	},
	"ports": null
}`

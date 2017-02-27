package application_test

import (
	"encoding/json"
	"time"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/api/resources"
	"code.cloudfoundry.org/cli/cf/api/stacks/stacksfakes"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/commands/application"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/formatters"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	"code.cloudfoundry.org/cli/plugin/models"

	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/api/appinstances/appinstancesfakes"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("App", func() {
	var (
		ui               *testterm.FakeUI
		appSummaryRepo   *apifakes.FakeAppSummaryRepository
		appInstancesRepo *appinstancesfakes.FakeAppInstancesRepository
		stackRepo        *stacksfakes.FakeStackRepository
		getAppModel      *plugin_models.GetAppModel

		cmd         commandregistry.Command
		deps        commandregistry.Dependency
		factory     *requirementsfakes.FakeFactory
		flagContext flags.FlagContext

		loginRequirement         requirements.Requirement
		targetedSpaceRequirement requirements.Requirement
		applicationRequirement   *requirementsfakes.FakeApplicationRequirement
	)

	BeforeEach(func() {
		cmd = &application.ShowApp{}
		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)

		ui = &testterm.FakeUI{}

		getAppModel = &plugin_models.GetAppModel{}

		repoLocator := api.RepositoryLocator{}
		appSummaryRepo = new(apifakes.FakeAppSummaryRepository)
		repoLocator = repoLocator.SetAppSummaryRepository(appSummaryRepo)
		appInstancesRepo = new(appinstancesfakes.FakeAppInstancesRepository)
		repoLocator = repoLocator.SetAppInstancesRepository(appInstancesRepo)
		stackRepo = new(stacksfakes.FakeStackRepository)
		repoLocator = repoLocator.SetStackRepository(stackRepo)

		deps = commandregistry.Dependency{
			UI:     ui,
			Config: testconfig.NewRepositoryWithDefaults(),
			PluginModels: &commandregistry.PluginModels{
				Application: getAppModel,
			},
			RepoLocator: repoLocator,
		}

		cmd.SetDependency(deps, false)

		factory = new(requirementsfakes.FakeFactory)

		loginRequirement = &passingRequirement{}
		factory.NewLoginRequirementReturns(loginRequirement)

		targetedSpaceRequirement = &passingRequirement{}
		factory.NewTargetedSpaceRequirementReturns(targetedSpaceRequirement)

		applicationRequirement = new(requirementsfakes.FakeApplicationRequirement)
		factory.NewApplicationRequirementReturns(applicationRequirement)
	})

	Describe("Requirements", func() {
		Context("when not provided exactly one arg", func() {
			BeforeEach(func() {
				flagContext.Parse("app-name", "extra-arg")
			})

			It("fails with usage", func() {
				_, err := cmd.Requirements(factory, flagContext)
				Expect(err).To(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
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
			appStackModel       models.Stack
			appInstanceFields   []models.AppInstanceFields
			getAppSummaryErr    error
			err                 error
		)

		BeforeEach(func() {
			err := flagContext.Parse("app-name")
			Expect(err).NotTo(HaveOccurred())
			cmd.Requirements(factory, flagContext)

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
					CPUUsage:  float64(0.25),
					DiskUsage: int64(1 * formatters.GIGABYTE),
					DiskQuota: int64(2 * formatters.GIGABYTE),
					MemUsage:  int64(24 * formatters.MEGABYTE),
					MemQuota:  int64(32 * formatters.MEGABYTE),
				},
			}

			appStackModel = models.Stack{
				GUID: "fake-stack-guid",
				Name: "fake-stack-name",
			}

			applicationRequirement.GetApplicationReturns(getApplicationModel)
			appSummaryRepo.GetSummaryReturns(getAppSummaryModel, getAppSummaryErr)
			appInstancesRepo.GetInstancesReturns(appInstanceFields, nil)
			stackRepo.FindByGUIDReturns(appStackModel, nil)
		})

		JustBeforeEach(func() {
			err = cmd.Execute(flagContext)
		})

		It("gets the application summary", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(appSummaryRepo.GetSummaryCallCount()).To(Equal(1))
		})

		It("gets the app instances", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(appInstancesRepo.GetInstancesCallCount()).To(Equal(1))
		})

		It("gets the application from the application requirement", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(applicationRequirement.GetApplicationCallCount()).To(Equal(1))
		})

		It("gets the stack name from the stack repository", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(stackRepo.FindByGUIDCallCount()).To(Equal(1))
		})

		It("prints a summary of the app", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Showing health and status for app fake-app-name"},
				[]string{"requested state: started"},
				[]string{"instances: 1/1"},
				// Commented to hide app-ports for release #117189491
				// []string{"app ports: 8080, 9090"},
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
				appSummaryRepo.GetSummaryReturns(getAppSummaryModel, errors.NewHTTPError(400, errors.InstancesError, "error"))
			})

			It("prints appropriate output", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
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
				appSummaryRepo.GetSummaryReturns(getAppSummaryModel, errors.NewHTTPError(400, errors.NotStaged, "error"))
			})

			It("prints appropriate output", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
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

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("an-error"))
			})

			Context("when the app is stopped", func() {
				BeforeEach(func() {
					getAppSummaryModel.State = "stopped"
					appSummaryRepo.GetSummaryReturns(getAppSummaryModel, errors.New("an-error"))
				})

				It("prints appropriate output", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(ui.Outputs()).To(ContainSubstrings(
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

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("an-error"))
			})

			Context("when the app is stopped", func() {
				BeforeEach(func() {
					getAppSummaryModel.RunningInstances = 0
					getAppSummaryModel.State = "stopped"
					appSummaryRepo.GetSummaryReturns(getAppSummaryModel, nil)
				})

				It("prints appropriate output", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(ui.Outputs()).To(ContainSubstrings(
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
				Expect(err).NotTo(HaveOccurred())

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"last uploaded: unknown"},
				))
			})
		})

		Context("when the application has no app ports", func() {
			BeforeEach(func() {
				getAppSummaryModel.AppPorts = []int{}
				appSummaryRepo.GetSummaryReturns(getAppSummaryModel, nil)
			})

			It("does not print 'app ports'", func() {
				Expect(err).NotTo(HaveOccurred())

				Expect(ui.Outputs()).NotTo(ContainSubstrings(
					[]string{"app ports:"},
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
				Expect(err).NotTo(HaveOccurred())

				Expect(ui.Outputs()).To(ContainSubstrings(
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
				Expect(err).NotTo(HaveOccurred())

				Expect(ui.Outputs()).To(ContainSubstrings(
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
				Expect(err).NotTo(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
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
				Expect(err).NotTo(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"instances", "?/1"},
				))
			})
		})

		Context("when the --guid flag is passed", func() {
			BeforeEach(func() {
				flagContext.Parse("app-name", "--guid")
			})

			It("only prints the guid for the app", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"fake-app-guid"},
				))
				Expect(ui.Outputs()).ToNot(ContainSubstrings(
					[]string{"Showing health and status", "my-app"},
				))
			})
		})

		Context("when called from a plugin", func() {
			BeforeEach(func() {
				cmd.SetDependency(deps, true)
			})

			Context("when the app is running", func() {
				It("populates the plugin model", func() {
					Expect(err).NotTo(HaveOccurred())

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
					Expect(getAppModel.AppPorts).To(Equal([]int{8080, 9090}))
					Expect(getAppModel.Routes[0].Host).To(Equal("fake-route-host"))
					Expect(getAppModel.Routes[0].Guid).To(Equal("fake-route-guid"))
					Expect(getAppModel.Routes[0].Domain.Name).To(Equal("fake-route-domain-name"))
					Expect(getAppModel.Routes[0].Domain.Guid).To(Equal("fake-route-domain-guid"))
					Expect(getAppModel.Routes[0].Path).To(Equal("some-path"))
					Expect(getAppModel.Routes[0].Port).To(Equal(3333))
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

			Context("when the app is stopped but instance is returning back an error", func() {
				BeforeEach(func() {
					getAppSummaryModel.State = "stopped"
					appSummaryRepo.GetSummaryReturns(getAppSummaryModel, nil)

					var instances []models.AppInstanceFields //Very important since this is a nil body
					appInstancesRepo.GetInstancesReturns(instances, errors.New("Bonzi"))
				})

				It("populates the plugin model with empty sets", func() {
					Expect(err).NotTo(HaveOccurred())

					Expect(getAppModel.Instances).ToNot(BeNil())
					Expect(getAppModel.Instances).To(BeEmpty())
				})
			})

			Context("when the there are no routes", func() {
				BeforeEach(func() {
					app := models.Application{
						Stack: &models.Stack{
							GUID: "stack-guid",
							Name: "stack-name",
						},
					}
					appSummaryRepo.GetSummaryReturns(app, nil)
				})

				It("populates the plugin model with empty sets", func() {
					Expect(err).NotTo(HaveOccurred())

					Expect(getAppModel.Routes).ToNot(BeNil())
					Expect(getAppModel.Routes).To(BeEmpty())
				})
			})

			Context("when the there are no services", func() {
				BeforeEach(func() {
					app := models.Application{
						Stack: &models.Stack{
							GUID: "stack-guid",
							Name: "stack-name",
						},
					}
					appSummaryRepo.GetSummaryReturns(app, nil)
				})

				It("populates the plugin model with empty sets", func() {
					Expect(err).NotTo(HaveOccurred())

					Expect(getAppModel.Services).ToNot(BeNil())
					Expect(getAppModel.Services).To(BeEmpty())
				})
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
        "ports": [
          8080,
          9090
				],
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
		},
		"path": "some-path",
    "port": 3333
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
	"ports": [
		8080,
		9090
	]
}`

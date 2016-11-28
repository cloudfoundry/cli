package service_test

import (
	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/commands/service"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"

	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	"fmt"

	"code.cloudfoundry.org/cli/cf/api/applications/applicationsfakes"
	"code.cloudfoundry.org/cli/plugin/models"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("service command", func() {
	var (
		ui                         *testterm.FakeUI
		deps                       commandregistry.Dependency
		flagContext                flags.FlagContext
		reqFactory                 *requirementsfakes.FakeFactory
		loginRequirement           requirements.Requirement
		targetedSpaceRequirement   requirements.Requirement
		serviceInstanceRequirement *requirementsfakes.FakeServiceInstanceRequirement
		pluginCall                 bool

		cmd *service.ShowService
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		pluginCall = false

		appRepo := new(applicationsfakes.FakeRepository)
		appRepo.GetAppStub = func(appGUID string) (models.Application, error) {
			if appGUID == "app1-guid" {
				return models.Application{
					ApplicationFields: models.ApplicationFields{
						Name: "app1",
					},
				}, nil
			}
			return models.Application{}, fmt.Errorf("Called stubbed applications repo GetApp with incorrect app GUID\nExpected \"app1-guid\"\nGot \"%s\"\n", appGUID)
		}

		deps = commandregistry.Dependency{
			UI:           ui,
			PluginModels: &commandregistry.PluginModels{},
			RepoLocator:  api.RepositoryLocator{}.SetApplicationRepository(appRepo),
		}

		cmd = &service.ShowService{}

		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
		reqFactory = new(requirementsfakes.FakeFactory)

		loginRequirement = &passingRequirement{Name: "login-requirement"}
		reqFactory.NewLoginRequirementReturns(loginRequirement)
		targetedSpaceRequirement = &passingRequirement{Name: "targeted-space-requirement"}
		reqFactory.NewTargetedSpaceRequirementReturns(targetedSpaceRequirement)
		serviceInstanceRequirement = &requirementsfakes.FakeServiceInstanceRequirement{}
		reqFactory.NewServiceInstanceRequirementReturns(serviceInstanceRequirement)
	})

	Describe("Requirements", func() {
		BeforeEach(func() {
			cmd.SetDependency(deps, pluginCall)
		})

		Context("when not provided exactly 1 argument", func() {
			It("fails", func() {
				err := flagContext.Parse("too", "many")
				Expect(err).NotTo(HaveOccurred())
				_, err = cmd.Requirements(reqFactory, flagContext)
				Expect(err).To(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Incorrect Usage", "Requires an argument"},
				))
			})
		})

		Context("when provided exactly one arg", func() {
			var actualRequirements []requirements.Requirement

			BeforeEach(func() {
				err := flagContext.Parse("service-name")
				Expect(err).NotTo(HaveOccurred())
				actualRequirements, err = cmd.Requirements(reqFactory, flagContext)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns a LoginRequirement", func() {
				Expect(reqFactory.NewLoginRequirementCallCount()).To(Equal(1))
				Expect(actualRequirements).To(ContainElement(loginRequirement))
			})

			It("returns a TargetedSpaceRequirement", func() {
				Expect(reqFactory.NewTargetedSpaceRequirementCallCount()).To(Equal(1))
				Expect(actualRequirements).To(ContainElement(targetedSpaceRequirement))
			})

			It("returns a ServiceInstanceRequirement", func() {
				Expect(reqFactory.NewServiceInstanceRequirementCallCount()).To(Equal(1))
				Expect(actualRequirements).To(ContainElement(serviceInstanceRequirement))
			})
		})
	})

	Describe("Execute", func() {
		var serviceInstance models.ServiceInstance

		BeforeEach(func() {
			serviceInstance = models.ServiceInstance{
				ServiceInstanceFields: models.ServiceInstanceFields{
					GUID: "service1-guid",
					Name: "service1",
					LastOperation: models.LastOperationFields{
						Type:        "create",
						State:       "in progress",
						Description: "creating resource - step 1",
						CreatedAt:   "created-date",
						UpdatedAt:   "updated-date",
					},
					DashboardURL: "some-url",
				},
				ServiceBindings: []models.ServiceBindingFields{
					models.ServiceBindingFields{
						AppGUID: "app1-guid",
					},
				},
				ServicePlan: models.ServicePlanFields{
					GUID: "plan-guid",
					Name: "plan-name",
				},
				ServiceOffering: models.ServiceOfferingFields{
					Label:            "mysql",
					DocumentationURL: "http://documentation.url",
					Description:      "the-description",
				},
			}
		})

		JustBeforeEach(func() {
			serviceInstanceRequirement.GetServiceInstanceReturns(serviceInstance)
			cmd.SetDependency(deps, pluginCall)
			cmd.Requirements(reqFactory, flagContext)
			cmd.Execute(flagContext)
		})

		Context("when invoked by a plugin", func() {
			var (
				pluginModel *plugin_models.GetService_Model
			)

			BeforeEach(func() {
				pluginModel = &plugin_models.GetService_Model{}
				deps.PluginModels.Service = pluginModel
				pluginCall = true
				err := flagContext.Parse("service1")
				Expect(err).NotTo(HaveOccurred())
			})

			It("populates the plugin model upon execution", func() {
				Expect(pluginModel.Name).To(Equal("service1"))
				Expect(pluginModel.Guid).To(Equal("service1-guid"))
				Expect(pluginModel.LastOperation.Type).To(Equal("create"))
				Expect(pluginModel.LastOperation.State).To(Equal("in progress"))
				Expect(pluginModel.LastOperation.Description).To(Equal("creating resource - step 1"))
				Expect(pluginModel.LastOperation.CreatedAt).To(Equal("created-date"))
				Expect(pluginModel.LastOperation.UpdatedAt).To(Equal("updated-date"))
				Expect(pluginModel.LastOperation.Type).To(Equal("create"))
				Expect(pluginModel.ServicePlan.Name).To(Equal("plan-name"))
				Expect(pluginModel.ServicePlan.Guid).To(Equal("plan-guid"))
				Expect(pluginModel.ServiceOffering.DocumentationUrl).To(Equal("http://documentation.url"))
				Expect(pluginModel.ServiceOffering.Name).To(Equal("mysql"))
			})
		})

		Context("when the service is externally provided", func() {
			Context("when only the service name is specified", func() {
				BeforeEach(func() {
					err := flagContext.Parse("service1")
					Expect(err).NotTo(HaveOccurred())
				})

				It("shows the service", func() {
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Service instance:", "service1"},
						[]string{"Service: ", "mysql"},
						[]string{"Bound apps: ", "app1"},
						[]string{"Plan: ", "plan-name"},
						[]string{"Description: ", "the-description"},
						[]string{"Documentation url: ", "http://documentation.url"},
						[]string{"Dashboard: ", "some-url"},
						[]string{"Last Operation"},
						[]string{"Status: ", "create in progress"},
						[]string{"Message: ", "creating resource - step 1"},
						[]string{"Started: ", "created-date"},
						[]string{"Updated: ", "updated-date"},
					))
				})

				Context("when the service instance CreatedAt is empty", func() {
					BeforeEach(func() {
						serviceInstance.LastOperation.CreatedAt = ""
					})

					It("does not output the Started line", func() {
						Expect(ui.Outputs()).To(ContainSubstrings(
							[]string{"Service instance:", "service1"},
							[]string{"Service: ", "mysql"},
							[]string{"Bound apps: ", "app1"},
							[]string{"Plan: ", "plan-name"},
							[]string{"Description: ", "the-description"},
							[]string{"Documentation url: ", "http://documentation.url"},
							[]string{"Dashboard: ", "some-url"},
							[]string{"Last Operation"},
							[]string{"Status: ", "create in progress"},
							[]string{"Message: ", "creating resource - step 1"},
							[]string{"Updated: ", "updated-date"},
						))
						Expect(ui.Outputs()).ToNot(ContainSubstrings(
							[]string{"Started: "},
						))
					})
				})

				Context("when the state is 'in progress'", func() {
					BeforeEach(func() {
						serviceInstance.LastOperation.State = "in progress"
					})

					It("shows status: `create in progress`", func() {
						Expect(ui.Outputs()).To(ContainSubstrings(
							[]string{"Status: ", "create in progress"},
						))
					})
				})

				Context("when the state is 'succeeded'", func() {
					BeforeEach(func() {
						serviceInstance.LastOperation.State = "succeeded"
					})

					It("shows status: `create succeeded`", func() {
						Expect(ui.Outputs()).To(ContainSubstrings(
							[]string{"Status: ", "create succeeded"},
						))
					})
				})

				Context("when the state is 'failed'", func() {
					BeforeEach(func() {
						serviceInstance.LastOperation.State = "failed"
					})

					It("shows status: `create failed`", func() {
						Expect(ui.Outputs()).To(ContainSubstrings(
							[]string{"Status: ", "create failed"},
						))
					})
				})

				Context("when the state is empty", func() {
					BeforeEach(func() {
						serviceInstance.LastOperation.State = ""
					})

					It("shows status: ``", func() {
						Expect(ui.Outputs()).To(ContainSubstrings(
							[]string{"Status: ", ""},
						))
					})
				})
			})

			Context("when the guid flag is provided", func() {
				BeforeEach(func() {
					err := flagContext.Parse("--guid", "service1")
					Expect(err).NotTo(HaveOccurred())
				})

				It("shows only the service guid", func() {
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"service1-guid"},
					))

					Expect(ui.Outputs()).ToNot(ContainSubstrings(
						[]string{"Service instance:", "service1"},
					))
				})
			})
		})

		Context("when the service is user provided", func() {
			BeforeEach(func() {
				serviceInstance = models.ServiceInstance{
					ServiceInstanceFields: models.ServiceInstanceFields{
						Name: "service1",
						GUID: "service1-guid",
					},
					ServiceBindings: []models.ServiceBindingFields{
						models.ServiceBindingFields{
							AppGUID: "app1-guid",
						},
					},
				}

				err := flagContext.Parse("service1")
				Expect(err).NotTo(HaveOccurred())
			})

			It("shows user provided services", func() {
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Service instance: ", "service1"},
					[]string{"Service: ", "user-provided"},
					[]string{"Bound apps: ", "app1"},
				))
			})
		})

		Context("when the service has tags", func() {
			BeforeEach(func() {
				serviceInstance = models.ServiceInstance{
					ServiceInstanceFields: models.ServiceInstanceFields{
						Tags: []string{"tag1", "tag2"},
					},
					ServicePlan: models.ServicePlanFields{GUID: "plan-guid", Name: "plan-name"},
				}

				err := flagContext.Parse("service1")
				Expect(err).NotTo(HaveOccurred())
			})

			It("includes the tags in the output", func() {
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Tags: ", "tag1, tag2"},
				))
			})
		})
	})
})

var _ = Describe("ServiceInstanceStateToStatus", func() {
	var operationType string
	Context("when the service is not user provided", func() {
		isUserProvided := false

		Context("when operationType is `create`", func() {
			BeforeEach(func() { operationType = "create" })

			It("returns status: `create in progress` when state: `in progress`", func() {
				status := service.InstanceStateToStatus(operationType, "in progress", isUserProvided)
				Expect(status).To(Equal("create in progress"))
			})

			It("returns status: `create succeeded` when state: `succeeded`", func() {
				status := service.InstanceStateToStatus(operationType, "succeeded", isUserProvided)
				Expect(status).To(Equal("create succeeded"))
			})

			It("returns status: `create failed` when state: `failed`", func() {
				status := service.InstanceStateToStatus(operationType, "failed", isUserProvided)
				Expect(status).To(Equal("create failed"))
			})

			It("returns status: `` when state: ``", func() {
				status := service.InstanceStateToStatus(operationType, "", isUserProvided)
				Expect(status).To(Equal(""))
			})
		})
	})

	Context("when the service is user provided", func() {
		isUserProvided := true

		It("returns status: `` when state: ``", func() {
			status := service.InstanceStateToStatus(operationType, "", isUserProvided)
			Expect(status).To(Equal(""))
		})
	})
})

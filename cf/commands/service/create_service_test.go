package service_test

import (
	"fmt"
	"io/ioutil"
	"os"

	"code.cloudfoundry.org/cli/cf/actors/servicebuilder/servicebuilderfakes"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"

	"code.cloudfoundry.org/cli/cf/api/apifakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	"code.cloudfoundry.org/cli/cf/commandregistry"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
)

var _ = Describe("create-service command", func() {
	var (
		ui                  *testterm.FakeUI
		config              coreconfig.Repository
		requirementsFactory *requirementsfakes.FakeFactory
		serviceRepo         *apifakes.FakeServiceRepository
		serviceBuilder      *servicebuilderfakes.FakeServiceBuilder

		offering1 models.ServiceOffering
		offering2 models.ServiceOffering
		deps      commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.Config = config
		deps.RepoLocator = deps.RepoLocator.SetServiceRepository(serviceRepo)
		deps.ServiceBuilder = serviceBuilder
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("create-service").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = new(requirementsfakes.FakeFactory)
		requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Passing{})
		serviceRepo = new(apifakes.FakeServiceRepository)
		serviceBuilder = new(servicebuilderfakes.FakeServiceBuilder)

		offering1 = models.ServiceOffering{}
		offering1.Label = "cleardb"
		offering1.Plans = []models.ServicePlanFields{{
			Name: "spark",
			GUID: "cleardb-spark-guid",
			Free: true,
		}, {
			Name: "expensive",
			GUID: "luxury-guid",
			Free: false,
		}}

		offering2 = models.ServiceOffering{}
		offering2.Label = "postgres"

		serviceBuilder.GetServicesByNameForSpaceWithPlansReturns(models.ServiceOfferings([]models.ServiceOffering{offering1, offering2}), nil)
	})

	var callCreateService = func(args []string) bool {
		return testcmd.RunCLICommand("create-service", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("requirements", func() {
		It("passes when logged in and a space is targeted", func() {
			Expect(callCreateService([]string{"cleardb", "spark", "my-cleardb-service"})).To(BeTrue())
		})

		It("fails when not logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
			Expect(callCreateService([]string{"cleardb", "spark", "my-cleardb-service"})).To(BeFalse())
		})

		It("fails when a space is not targeted", func() {
			requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Failing{Message: "not targeted"})
			Expect(callCreateService([]string{"cleardb", "spark", "my-cleardb-service"})).To(BeFalse())
		})
	})

	It("successfully creates a service", func() {
		callCreateService([]string{"cleardb", "spark", "my-cleardb-service"})

		spaceGUID, serviceName := serviceBuilder.GetServicesByNameForSpaceWithPlansArgsForCall(0)
		Expect(spaceGUID).To(Equal(config.SpaceFields().GUID))
		Expect(serviceName).To(Equal("cleardb"))
		Expect(ui.Outputs()).To(ContainSubstrings(
			[]string{"Creating service instance", "my-cleardb-service", "my-org", "my-space", "my-user"},
			[]string{"OK"},
		))
		name, planGUID, _, _ := serviceRepo.CreateServiceInstanceArgsForCall(0)
		Expect(name).To(Equal("my-cleardb-service"))
		Expect(planGUID).To(Equal("cleardb-spark-guid"))
	})

	Context("when passing in tags", func() {
		It("sucessfully creates a service and passes the tags as json", func() {
			callCreateService([]string{"cleardb", "spark", "my-cleardb-service", "-t", "tag1, tag2,tag3,  tag4"})

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Creating service instance", "my-cleardb-service", "my-org", "my-space", "my-user"},
				[]string{"OK"},
			))
			_, _, _, tags := serviceRepo.CreateServiceInstanceArgsForCall(0)
			Expect(tags).To(ConsistOf("tag1", "tag2", "tag3", "tag4"))
		})
	})

	Context("when passing arbitrary params", func() {
		Context("as a json string", func() {
			It("successfully creates a service and passes the params as a json string", func() {
				callCreateService([]string{"cleardb", "spark", "my-cleardb-service", "-c", `{"foo": "bar"}`})

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Creating service instance", "my-cleardb-service", "my-org", "my-space", "my-user"},
					[]string{"OK"},
				))
				_, _, params, _ := serviceRepo.CreateServiceInstanceArgsForCall(0)
				Expect(params).To(Equal(map[string]interface{}{"foo": "bar"}))
			})

			Context("that are not valid json", func() {
				It("returns an error to the UI", func() {
					callCreateService([]string{"cleardb", "spark", "my-cleardb-service", "-c", `bad-json`})

					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object."},
					))
				})
			})
		})

		Context("as a file that contains json", func() {
			var jsonFile *os.File
			var params string

			BeforeEach(func() {
				params = "{\"foo\": \"bar\"}"
			})

			AfterEach(func() {
				if jsonFile != nil {
					jsonFile.Close()
					os.Remove(jsonFile.Name())
				}
			})

			JustBeforeEach(func() {
				var err error
				jsonFile, err = ioutil.TempFile("", "")
				Expect(err).ToNot(HaveOccurred())

				err = ioutil.WriteFile(jsonFile.Name(), []byte(params), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			})

			It("successfully creates a service and passes the params as a json", func() {
				callCreateService([]string{"cleardb", "spark", "my-cleardb-service", "-c", jsonFile.Name()})

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Creating service instance", "my-cleardb-service", "my-org", "my-space", "my-user"},
					[]string{"OK"},
				))
				_, _, params, _ := serviceRepo.CreateServiceInstanceArgsForCall(0)
				Expect(params).To(Equal(map[string]interface{}{"foo": "bar"}))
			})

			Context("that are not valid json", func() {
				BeforeEach(func() {
					params = "bad-json"
				})

				It("returns an error to the UI", func() {
					callCreateService([]string{"cleardb", "spark", "my-cleardb-service", "-c", jsonFile.Name()})

					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object."},
					))
				})
			})
		})
	})

	Context("when service creation is asynchronous", func() {
		var serviceInstance models.ServiceInstance

		BeforeEach(func() {
			serviceInstance = models.ServiceInstance{
				ServiceInstanceFields: models.ServiceInstanceFields{
					Name: "my-cleardb-service",
					LastOperation: models.LastOperationFields{
						Type:        "create",
						State:       "in progress",
						Description: "fake service instance description",
					},
				},
			}
			serviceRepo.FindInstanceByNameReturns(serviceInstance, nil)
		})

		It("successfully starts async service creation", func() {
			callCreateService([]string{"cleardb", "spark", "my-cleardb-service"})

			spaceGUID, serviceName := serviceBuilder.GetServicesByNameForSpaceWithPlansArgsForCall(0)
			Expect(spaceGUID).To(Equal(config.SpaceFields().GUID))
			Expect(serviceName).To(Equal("cleardb"))

			creatingServiceMessage := fmt.Sprintf("Create in progress. Use 'cf services' or 'cf service %s' to check operation status.", serviceInstance.ServiceInstanceFields.Name)

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Creating service instance", "my-cleardb-service", "my-org", "my-space", "my-user"},
				[]string{"OK"},
				[]string{creatingServiceMessage},
			))
			name, planGUID, _, _ := serviceRepo.CreateServiceInstanceArgsForCall(0)
			Expect(name).To(Equal("my-cleardb-service"))
			Expect(planGUID).To(Equal("cleardb-spark-guid"))
		})

		It("fails when service instance could is created but cannot be found", func() {
			serviceRepo.FindInstanceByNameReturns(models.ServiceInstance{}, errors.New("Error finding instance"))
			callCreateService([]string{"cleardb", "spark", "fake-service-instance-name"})

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Creating service instance fake-service-instance-name in org my-org / space my-space as my-user..."},
				[]string{"FAILED"},
				[]string{"Error finding instance"}))
		})
	})

	Describe("warning the user about paid services", func() {
		It("does not warn the user when the service is free", func() {
			callCreateService([]string{"cleardb", "spark", "my-free-cleardb-service"})
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Creating service instance", "my-free-cleardb-service", "my-org", "my-space", "my-user"},
				[]string{"OK"},
			))
			Expect(ui.Outputs()).NotTo(ContainSubstrings([]string{"will incurr a cost"}))
		})

		It("warns the user when the service is not free", func() {
			callCreateService([]string{"cleardb", "expensive", "my-expensive-cleardb-service"})
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Creating service instance", "my-expensive-cleardb-service", "my-org", "my-space", "my-user"},
				[]string{"OK"},
				[]string{"Attention: The plan `expensive` of service `cleardb` is not free.  The instance `my-expensive-cleardb-service` will incur a cost.  Contact your administrator if you think this is in error."},
			))
		})
	})

	It("warns the user when the service already exists with the same service plan", func() {
		serviceRepo.CreateServiceInstanceReturns(errors.NewModelAlreadyExistsError("Service", "my-cleardb-service"))

		callCreateService([]string{"cleardb", "spark", "my-cleardb-service"})

		Expect(ui.Outputs()).To(ContainSubstrings(
			[]string{"Creating service instance", "my-cleardb-service"},
			[]string{"OK"},
			[]string{"my-cleardb-service", "already exists"},
		))

		name, planGUID, _, _ := serviceRepo.CreateServiceInstanceArgsForCall(0)
		Expect(name).To(Equal("my-cleardb-service"))
		Expect(planGUID).To(Equal("cleardb-spark-guid"))
	})

	Context("When there are multiple services with the same label", func() {
		It("finds the plan even if it has to search multiple services", func() {
			offering2.Label = "cleardb"

			serviceRepo.CreateServiceInstanceReturns(errors.NewModelAlreadyExistsError("Service", "my-cleardb-service"))
			callCreateService([]string{"cleardb", "spark", "my-cleardb-service"})

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Creating service instance", "my-cleardb-service", "my-org", "my-space", "my-user"},
				[]string{"OK"},
			))

			name, planGUID, _, _ := serviceRepo.CreateServiceInstanceArgsForCall(0)
			Expect(name).To(Equal("my-cleardb-service"))
			Expect(planGUID).To(Equal("cleardb-spark-guid"))
		})
	})
})

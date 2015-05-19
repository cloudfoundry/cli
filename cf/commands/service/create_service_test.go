package service_test

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry/cli/cf/actors/service_builder/fakes"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/generic"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/service"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("create-service command", func() {
	var (
		ui                  *testterm.FakeUI
		config              core_config.Repository
		requirementsFactory *testreq.FakeReqFactory
		cmd                 CreateService
		serviceRepo         *testapi.FakeServiceRepo
		serviceBuilder      *fakes.FakeServiceBuilder

		offering1 models.ServiceOffering
		offering2 models.ServiceOffering
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
		serviceRepo = &testapi.FakeServiceRepo{}
		serviceBuilder = &fakes.FakeServiceBuilder{}
		cmd = NewCreateService(ui, config, serviceRepo, serviceBuilder)

		offering1 = models.ServiceOffering{}
		offering1.Label = "cleardb"
		offering1.Plans = []models.ServicePlanFields{{
			Name: "spark",
			Guid: "cleardb-spark-guid",
			Free: true,
		}, {
			Name: "expensive",
			Guid: "luxury-guid",
			Free: false,
		}}

		offering2 = models.ServiceOffering{}
		offering2.Label = "postgres"

		serviceBuilder.GetServicesByNameForSpaceWithPlansReturns(models.ServiceOfferings([]models.ServiceOffering{offering1, offering2}), nil)
	})

	var callCreateService = func(args []string) bool {
		return testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("passes when logged in and a space is targeted", func() {
			Expect(callCreateService([]string{"cleardb", "spark", "my-cleardb-service"})).To(BeTrue())
		})

		It("fails when not logged in", func() {
			requirementsFactory.LoginSuccess = false
			Expect(callCreateService([]string{"cleardb", "spark", "my-cleardb-service"})).To(BeFalse())
		})

		It("fails when a space is not targeted", func() {
			requirementsFactory.TargetedSpaceSuccess = false
			Expect(callCreateService([]string{"cleardb", "spark", "my-cleardb-service"})).To(BeFalse())
		})
	})

	It("successfully creates a service", func() {
		callCreateService([]string{"cleardb", "spark", "my-cleardb-service"})

		spaceGuid, serviceName := serviceBuilder.GetServicesByNameForSpaceWithPlansArgsForCall(0)
		Expect(spaceGuid).To(Equal(config.SpaceFields().Guid))
		Expect(serviceName).To(Equal("cleardb"))
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Creating service instance", "my-cleardb-service", "my-org", "my-space", "my-user"},
			[]string{"OK"},
		))
		Expect(serviceRepo.CreateServiceInstanceArgs.Name).To(Equal("my-cleardb-service"))
		Expect(serviceRepo.CreateServiceInstanceArgs.PlanGuid).To(Equal("cleardb-spark-guid"))
	})

	Context("when passing arbitrary params", func() {
		Context("as a json string", func() {
			It("successfully creates a service and passes the params as a json string", func() {
				callCreateService([]string{"cleardb", "spark", "my-cleardb-service", "-c", `{"foo": "bar"}`})

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Creating service instance", "my-cleardb-service", "my-org", "my-space", "my-user"},
					[]string{"OK"},
				))
				Expect(serviceRepo.CreateServiceInstanceArgs.Params).To(Equal(map[string]interface{}{"foo": "bar"}))
			})

			Context("that are not valid json", func() {
				It("returns an error to the UI", func() {
					callCreateService([]string{"cleardb", "spark", "my-cleardb-service", "-c", `bad-json`})

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"Invalid JSON provided in -c argument"},
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

			It("successfully creates a service and passes the params as a json string", func() {
				callCreateService([]string{"cleardb", "spark", "my-cleardb-service", "-c", jsonFile.Name()})

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Creating service instance", "my-cleardb-service", "my-org", "my-space", "my-user"},
					[]string{"OK"},
				))
				Expect(serviceRepo.CreateServiceInstanceArgs.Params).To(Equal(map[string]interface{}{"foo": "bar"}))
			})

			Context("that are not valid json", func() {
				BeforeEach(func() {
					params = "bad-json"
				})

				It("returns an error to the UI", func() {
					callCreateService([]string{"cleardb", "spark", "my-cleardb-service", "-c", jsonFile.Name()})

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"Invalid JSON provided in -c argument"},
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
			serviceRepo.FindInstanceByNameMap = generic.NewMap()
			serviceRepo.FindInstanceByNameMap.Set("my-cleardb-service", serviceInstance)
		})

		It("successfully starts async service creation", func() {
			callCreateService([]string{"cleardb", "spark", "my-cleardb-service"})

			spaceGuid, serviceName := serviceBuilder.GetServicesByNameForSpaceWithPlansArgsForCall(0)
			Expect(spaceGuid).To(Equal(config.SpaceFields().Guid))
			Expect(serviceName).To(Equal("cleardb"))

			creatingServiceMessage := fmt.Sprintf("Create in progress. Use 'cf services' or 'cf service %s' to check operation status.", serviceInstance.ServiceInstanceFields.Name)

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Creating service instance", "my-cleardb-service", "my-org", "my-space", "my-user"},
				[]string{"OK"},
				[]string{creatingServiceMessage},
			))
			Expect(serviceRepo.CreateServiceInstanceArgs.Name).To(Equal("my-cleardb-service"))
			Expect(serviceRepo.CreateServiceInstanceArgs.PlanGuid).To(Equal("cleardb-spark-guid"))
		})

		It("fails when service instance could is created but cannot be found", func() {
			serviceRepo.FindInstanceByNameErr = true
			callCreateService([]string{"cleardb", "spark", "fake-service-instance-name"})

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Creating service instance fake-service-instance-name in org my-org / space my-space as my-user..."},
				[]string{"FAILED"},
				[]string{"Error finding instance"}))
		})
	})

	Describe("warning the user about paid services", func() {
		It("does not warn the user when the service is free", func() {
			callCreateService([]string{"cleardb", "spark", "my-free-cleardb-service"})
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Creating service instance", "my-free-cleardb-service", "my-org", "my-space", "my-user"},
				[]string{"OK"},
			))
			Expect(ui.Outputs).NotTo(ContainSubstrings([]string{"will incurr a cost"}))
		})

		It("warns the user when the service is not free", func() {
			callCreateService([]string{"cleardb", "expensive", "my-expensive-cleardb-service"})
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Creating service instance", "my-expensive-cleardb-service", "my-org", "my-space", "my-user"},
				[]string{"OK"},
				[]string{"Attention: The plan `expensive` of service `cleardb` is not free.  The instance `my-expensive-cleardb-service` will incur a cost.  Contact your administrator if you think this is in error."},
			))
		})
	})

	It("warns the user when the service already exists with the same service plan", func() {
		serviceRepo.CreateServiceInstanceReturns.Error = errors.NewModelAlreadyExistsError("Service", "my-cleardb-service")

		callCreateService([]string{"cleardb", "spark", "my-cleardb-service"})

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Creating service instance", "my-cleardb-service"},
			[]string{"OK"},
			[]string{"my-cleardb-service", "already exists"},
		))
		Expect(serviceRepo.CreateServiceInstanceArgs.Name).To(Equal("my-cleardb-service"))
		Expect(serviceRepo.CreateServiceInstanceArgs.PlanGuid).To(Equal("cleardb-spark-guid"))
	})

	Context("When there are multiple services with the same label", func() {
		It("finds the plan even if it has to search multiple services", func() {
			offering2.Label = "cleardb"

			serviceRepo.CreateServiceInstanceReturns.Error = errors.NewModelAlreadyExistsError("Service", "my-cleardb-service")
			callCreateService([]string{"cleardb", "spark", "my-cleardb-service"})

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Creating service instance", "my-cleardb-service", "my-org", "my-space", "my-user"},
				[]string{"OK"},
			))
			Expect(serviceRepo.CreateServiceInstanceArgs.Name).To(Equal("my-cleardb-service"))
			Expect(serviceRepo.CreateServiceInstanceArgs.PlanGuid).To(Equal("cleardb-spark-guid"))
		})
	})
})

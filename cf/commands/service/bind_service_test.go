package service_test

import (
	"io/ioutil"
	"net/http"
	"os"

	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cli/cf/errors"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
)

var _ = Describe("bind-service command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *requirementsfakes.FakeFactory
		config              coreconfig.Repository
		serviceBindingRepo  *apifakes.FakeServiceBindingRepository
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.Config = config
		deps.RepoLocator = deps.RepoLocator.SetServiceBindingRepository(serviceBindingRepo)
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("bind-service").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = new(requirementsfakes.FakeFactory)
		serviceBindingRepo = new(apifakes.FakeServiceBindingRepository)
	})

	var callBindService = func(args []string) bool {
		return testcmd.RunCLICommand("bind-service", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	It("fails requirements when not logged in", func() {
		requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
		Expect(callBindService([]string{"service", "app"})).To(BeFalse())
	})

	Context("when logged in", func() {
		var (
			app             models.Application
			serviceInstance models.ServiceInstance
		)

		BeforeEach(func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})

			app = models.Application{
				ApplicationFields: models.ApplicationFields{
					Name: "my-app",
					GUID: "my-app-guid",
				},
			}
			serviceInstance = models.ServiceInstance{
				ServiceInstanceFields: models.ServiceInstanceFields{
					Name: "my-service",
					GUID: "my-service-guid",
				},
			}
			applicationReq := new(requirementsfakes.FakeApplicationRequirement)
			applicationReq.GetApplicationReturns(app)
			requirementsFactory.NewApplicationRequirementReturns(applicationReq)

			serviceInstanceReq := new(requirementsfakes.FakeServiceInstanceRequirement)
			serviceInstanceReq.GetServiceInstanceReturns(serviceInstance)
			requirementsFactory.NewServiceInstanceRequirementReturns(serviceInstanceReq)
		})

		It("binds a service instance to an app", func() {
			callBindService([]string{"my-app", "my-service"})

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Binding service", "my-service", "my-app", "my-org", "my-space", "my-user"},
				[]string{"OK"},
				[]string{"TIP", "my-app"},
			))

			Expect(serviceBindingRepo.CreateCallCount()).To(Equal(1))
			serviceInstanceGUID, applicationGUID, _ := serviceBindingRepo.CreateArgsForCall(0)
			Expect(serviceInstanceGUID).To(Equal("my-service-guid"))
			Expect(applicationGUID).To(Equal("my-app-guid"))
		})

		It("warns the user when the service instance is already bound to the given app", func() {
			serviceBindingRepo.CreateReturns(errors.NewHTTPError(http.StatusBadRequest, errors.ServiceBindingAppServiceTaken, ""))
			callBindService([]string{"my-app", "my-service"})

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Binding service"},
				[]string{"OK"},
				[]string{"my-app", "is already bound", "my-service"},
			))
		})

		It("warns the user when the error is non HTTPError ", func() {
			serviceBindingRepo.CreateReturns(errors.New("1001"))
			callBindService([]string{"my-app1", "my-service1"})
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Binding service", "my-service", "my-app", "my-org", "my-space", "my-user"},
				[]string{"FAILED"},
				[]string{"1001"},
			))
		})

		It("fails with usage when called without a service instance and app", func() {
			callBindService([]string{"my-service"})
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))

			ui = &testterm.FakeUI{}
			callBindService([]string{"my-app"})
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))

			ui = &testterm.FakeUI{}
			callBindService([]string{"my-app", "my-service"})
			Expect(ui.Outputs()).ToNot(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))
		})

		Context("when passing arbitrary params", func() {
			Context("as a json string", func() {
				It("successfully creates a service and passes the params as a json string", func() {
					callBindService([]string{"my-app", "my-service", "-c", `{"foo": "bar"}`})

					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Binding service", "my-service", "my-app", "my-org", "my-space", "my-user"},
						[]string{"OK"},
						[]string{"TIP"},
					))

					Expect(serviceBindingRepo.CreateCallCount()).To(Equal(1))
					serviceInstanceGUID, applicationGUID, createParams := serviceBindingRepo.CreateArgsForCall(0)
					Expect(serviceInstanceGUID).To(Equal("my-service-guid"))
					Expect(applicationGUID).To(Equal("my-app-guid"))
					Expect(createParams).To(Equal(map[string]interface{}{"foo": "bar"}))
				})

				Context("that are not valid json", func() {
					It("returns an error to the UI", func() {
						callBindService([]string{"my-app", "my-service", "-c", `bad-json`})

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
					callBindService([]string{"my-app", "my-service", "-c", jsonFile.Name()})

					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Binding service", "my-service", "my-app", "my-org", "my-space", "my-user"},
						[]string{"OK"},
						[]string{"TIP"},
					))

					Expect(serviceBindingRepo.CreateCallCount()).To(Equal(1))
					serviceInstanceGUID, applicationGUID, createParams := serviceBindingRepo.CreateArgsForCall(0)
					Expect(serviceInstanceGUID).To(Equal("my-service-guid"))
					Expect(applicationGUID).To(Equal("my-app-guid"))
					Expect(createParams).To(Equal(map[string]interface{}{"foo": "bar"}))
				})

				Context("that are not valid json", func() {
					BeforeEach(func() {
						params = "bad-json"
					})

					It("returns an error to the UI", func() {
						callBindService([]string{"my-app", "my-service", "-c", jsonFile.Name()})

						Expect(ui.Outputs()).To(ContainSubstrings(
							[]string{"FAILED"},
							[]string{"Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object."},
						))
					})
				})
			})
		})
	})
})

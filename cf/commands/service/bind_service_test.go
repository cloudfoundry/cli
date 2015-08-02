package service_test

import (
	"io/ioutil"
	"os"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("bind-service command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		config              core_config.Repository
		serviceBindingRepo  *testapi.FakeServiceBindingRepo
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.Config = config
		deps.RepoLocator = deps.RepoLocator.SetServiceBindingRepository(serviceBindingRepo)
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("bind-service").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{}
		serviceBindingRepo = &testapi.FakeServiceBindingRepo{}
	})

	var callBindService = func(args []string) bool {
		return testcmd.RunCliCommand("bind-service", args, requirementsFactory, updateCommandDependency, false)
	}

	It("fails requirements when not logged in", func() {
		Expect(callBindService([]string{"service", "app"})).To(BeFalse())
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		It("binds a service instance to an app", func() {
			app := models.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			serviceInstance := models.ServiceInstance{}
			serviceInstance.Name = "my-service"
			serviceInstance.Guid = "my-service-guid"
			requirementsFactory.Application = app
			requirementsFactory.ServiceInstance = serviceInstance
			callBindService([]string{"my-app", "my-service"})

			Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
			Expect(requirementsFactory.ServiceInstanceName).To(Equal("my-service"))

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Binding service", "my-service", "my-app", "my-org", "my-space", "my-user"},
				[]string{"OK"},
				[]string{"TIP", "my-app"},
			))
			Expect(serviceBindingRepo.CreateServiceInstanceGuid).To(Equal("my-service-guid"))
			Expect(serviceBindingRepo.CreateApplicationGuid).To(Equal("my-app-guid"))
		})

		It("warns the user when the service instance is already bound to the given app", func() {
			app := models.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			serviceInstance := models.ServiceInstance{}
			serviceInstance.Name = "my-service"
			serviceInstance.Guid = "my-service-guid"
			requirementsFactory.Application = app
			requirementsFactory.ServiceInstance = serviceInstance
			serviceBindingRepo = &testapi.FakeServiceBindingRepo{CreateErrorCode: "90003"}
			callBindService([]string{"my-app", "my-service"})

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Binding service"},
				[]string{"OK"},
				[]string{"my-app", "is already bound", "my-service"},
			))
		})

		It("warns the user when the error is non HttpError ", func() {
			app := models.Application{}
			app.Name = "my-app1"
			app.Guid = "my-app1-guid1"
			serviceInstance := models.ServiceInstance{}
			serviceInstance.Name = "my-service1"
			serviceInstance.Guid = "my-service1-guid1"
			requirementsFactory.Application = app
			requirementsFactory.ServiceInstance = serviceInstance
			serviceBindingRepo = &testapi.FakeServiceBindingRepo{CreateNonHttpErrCode: "1001"}
			callBindService([]string{"my-app1", "my-service1"})
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Binding service", "my-service", "my-app", "my-org", "my-space", "my-user"},
				[]string{"FAILED"},
				[]string{"1001"},
			))
		})

		It("fails with usage when called without a service instance and app", func() {
			callBindService([]string{"my-service"})
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))

			ui = &testterm.FakeUI{}
			callBindService([]string{"my-app"})
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))

			ui = &testterm.FakeUI{}
			callBindService([]string{"my-app", "my-service"})
			Expect(ui.Outputs).ToNot(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))
		})

		Context("when passing arbitrary params", func() {
			var (
				app             models.Application
				serviceInstance models.ServiceInstance
			)

			BeforeEach(func() {
				app = models.Application{}
				app.Name = "my-app"
				app.Guid = "my-app-guid"

				serviceInstance = models.ServiceInstance{}
				serviceInstance.Name = "my-service"
				serviceInstance.Guid = "my-service-guid"

				requirementsFactory.Application = app
				requirementsFactory.ServiceInstance = serviceInstance
			})

			Context("as a json string", func() {
				It("successfully creates a service and passes the params as a json string", func() {
					callBindService([]string{"my-app", "my-service", "-c", `{"foo": "bar"}`})

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Binding service", "my-service", "my-app", "my-org", "my-space", "my-user"},
						[]string{"OK"},
						[]string{"TIP"},
					))
					Expect(serviceBindingRepo.CreateServiceInstanceGuid).To(Equal("my-service-guid"))
					Expect(serviceBindingRepo.CreateApplicationGuid).To(Equal("my-app-guid"))
					Expect(serviceBindingRepo.CreateParams).To(Equal(map[string]interface{}{"foo": "bar"}))
				})

				Context("that are not valid json", func() {
					It("returns an error to the UI", func() {
						callBindService([]string{"my-app", "my-service", "-c", `bad-json`})

						Expect(ui.Outputs).To(ContainSubstrings(
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

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Binding service", "my-service", "my-app", "my-org", "my-space", "my-user"},
						[]string{"OK"},
						[]string{"TIP"},
					))
					Expect(serviceBindingRepo.CreateServiceInstanceGuid).To(Equal("my-service-guid"))
					Expect(serviceBindingRepo.CreateApplicationGuid).To(Equal("my-app-guid"))
					Expect(serviceBindingRepo.CreateParams).To(Equal(map[string]interface{}{"foo": "bar"}))
				})

				Context("that are not valid json", func() {
					BeforeEach(func() {
						params = "bad-json"
					})

					It("returns an error to the UI", func() {
						callBindService([]string{"my-app", "my-service", "-c", jsonFile.Name()})

						Expect(ui.Outputs).To(ContainSubstrings(
							[]string{"FAILED"},
							[]string{"Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object."},
						))
					})
				})
			})
		})
	})
})

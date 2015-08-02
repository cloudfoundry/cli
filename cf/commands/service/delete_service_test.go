package service_test

import (
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

var _ = Describe("delete-service command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		serviceRepo         *testapi.FakeServiceRepo
		serviceInstance     models.ServiceInstance
		configRepo          core_config.Repository
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetServiceRepository(serviceRepo)
		deps.Config = configRepo
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("delete-service").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{
			Inputs: []string{"yes"},
		}

		configRepo = testconfig.NewRepositoryWithDefaults()
		serviceRepo = &testapi.FakeServiceRepo{}
		requirementsFactory = &testreq.FakeReqFactory{
			LoginSuccess: true,
		}
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("delete-service", args, requirementsFactory, updateCommandDependency, false)
	}

	Context("when not logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = false
		})

		It("does not pass requirements", func() {
			Expect(runCommand("vestigial-service")).To(BeFalse())
		})
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		It("fails with usage when not provided exactly one arg", func() {
			runCommand()
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires an argument"},
			))
		})

		Context("when the service exists", func() {
			Context("and the service deletion is asynchronous", func() {
				BeforeEach(func() {
					serviceInstance = models.ServiceInstance{}
					serviceInstance.Name = "my-service"
					serviceInstance.Guid = "my-service-guid"
					serviceInstance.LastOperation.Type = "delete"
					serviceInstance.LastOperation.State = "in progress"
					serviceInstance.LastOperation.Description = "delete"
					serviceRepo.FindInstanceByNameServiceInstance = serviceInstance
				})

				Context("when the command is confirmed", func() {
					It("deletes the service", func() {
						runCommand("my-service")

						Expect(ui.Prompts).To(ContainSubstrings([]string{"Really delete the service my-service"}))

						Expect(ui.Outputs).To(ContainSubstrings(
							[]string{"Deleting service", "my-service", "my-org", "my-space", "my-user"},
							[]string{"OK"},
							[]string{"Delete in progress. Use 'cf services' or 'cf service my-service' to check operation status."},
						))

						Expect(serviceRepo.DeleteServiceServiceInstance).To(Equal(serviceInstance))
					})
				})

				It("skips confirmation when the -f flag is given", func() {
					runCommand("-f", "foo.com")

					Expect(ui.Prompts).To(BeEmpty())
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Deleting service", "foo.com"},
						[]string{"OK"},
						[]string{"Delete in progress. Use 'cf services' or 'cf service foo.com' to check operation status."},
					))
				})
			})

			Context("and the service deletion is synchronous", func() {
				BeforeEach(func() {
					serviceInstance = models.ServiceInstance{}
					serviceInstance.Name = "my-service"
					serviceInstance.Guid = "my-service-guid"
					serviceRepo.FindInstanceByNameServiceInstance = serviceInstance
				})

				Context("when the command is confirmed", func() {
					It("deletes the service", func() {
						runCommand("my-service")

						Expect(ui.Prompts).To(ContainSubstrings([]string{"Really delete the service my-service"}))

						Expect(ui.Outputs).To(ContainSubstrings(
							[]string{"Deleting service", "my-service", "my-org", "my-space", "my-user"},
							[]string{"OK"},
						))

						Expect(serviceRepo.DeleteServiceServiceInstance).To(Equal(serviceInstance))
					})
				})

				It("skips confirmation when the -f flag is given", func() {
					runCommand("-f", "foo.com")

					Expect(ui.Prompts).To(BeEmpty())
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Deleting service", "foo.com"},
						[]string{"OK"},
					))
				})
			})
		})

		Context("when the service does not exist", func() {
			BeforeEach(func() {
				serviceRepo.FindInstanceByNameNotFound = true
			})

			It("warns the user the service does not exist", func() {
				runCommand("-f", "my-service")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Deleting service", "my-service"},
					[]string{"OK"},
				))

				Expect(ui.WarnOutputs).To(ContainSubstrings([]string{"my-service", "does not exist"}))
			})
		})
	})
})

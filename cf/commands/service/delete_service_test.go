package service_test

import (
	. "github.com/cloudfoundry/cli/cf/commands/service"
	"github.com/cloudfoundry/cli/cf/models"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
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
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{
			Inputs: []string{"yes"},
		}

		serviceRepo = &testapi.FakeServiceRepo{}
		requirementsFactory = &testreq.FakeReqFactory{
			LoginSuccess: true,
		}
	})

	runCommand := func(args ...string) {
		configRepo := testconfig.NewRepositoryWithDefaults()
		cmd := NewDeleteService(ui, configRepo, serviceRepo)
		testcmd.RunCommand(cmd, testcmd.NewContext("delete-service", args), requirementsFactory)
	}

	Context("when not logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = false
		})

		It("does not pass requirements", func() {
			runCommand("vestigial-service")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		It("fails with usage when no service name is given", func() {
			runCommand()
			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		Context("when the service exists", func() {
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

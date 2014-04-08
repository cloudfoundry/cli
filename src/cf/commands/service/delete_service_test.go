package service_test

import (
	. "cf/commands/service"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
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

					testassert.SliceContains(ui.Prompts, testassert.Lines{
						{"Really delete the service my-service"},
					})

					testassert.SliceContains(ui.Outputs, testassert.Lines{
						{"Deleting service", "my-service", "my-org", "my-space", "my-user"},
						{"OK"},
					})

					Expect(serviceRepo.DeleteServiceServiceInstance).To(Equal(serviceInstance))
				})
			})

			It("skips confirmation when the -f flag is given", func() {
				runCommand("-f", "foo.com")

				Expect(ui.Prompts).To(BeEmpty())
				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"Deleting service", "foo.com"},
					{"OK"},
				})
			})
		})

		Context("when the service does not exist", func() {
			BeforeEach(func() {
				serviceRepo.FindInstanceByNameNotFound = true
			})

			It("warns the user the service does not exist", func() {
				runCommand("-f", "my-service")

				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"Deleting service", "my-service"},
					{"OK"},
				})

				testassert.SliceContains(ui.WarnOutputs, testassert.Lines{
					{"my-service", "does not exist"},
				})
			})
		})
	})
})

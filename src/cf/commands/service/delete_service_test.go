package service_test

import (
	"cf/api"
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

func callDeleteService(confirmation string, args []string, reqFactory *testreq.FakeReqFactory, serviceRepo api.ServiceRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = &testterm.FakeUI{
		Inputs: []string{confirmation},
	}
	ctxt := testcmd.NewContext("delete-service", args)

	config := testconfig.NewRepositoryWithDefaults()

	cmd := NewDeleteService(fakeUI, config, serviceRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

var _ = Describe("delete-service command", func() {
	var (
		requirementsFactory *testreq.FakeReqFactory
		serviceRepo         *testapi.FakeServiceRepo
		serviceInstance     models.ServiceInstance
	)

	BeforeEach(func() {
		requirementsFactory = &testreq.FakeReqFactory{}
		serviceRepo = &testapi.FakeServiceRepo{}
	})

	Context("when not logged in", func() {
		It("does not pass requirements", func() {
			cmd := NewDeleteService(&testterm.FakeUI{}, testconfig.NewRepository(), serviceRepo)
			testcmd.RunCommand(cmd, testcmd.NewContext("delete-service", []string{"vestigal-service"}), requirementsFactory)
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		It("fails with usage when no arguments are given", func() {
			ui := callDeleteService("", []string{"-f"}, requirementsFactory, serviceRepo)
			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		Context("when the service exists", func() {
			BeforeEach(func() {
				serviceInstance = models.ServiceInstance{}
				serviceInstance.Name = "my-service"
				serviceInstance.Guid = "my-service-guid"
				serviceRepo = &testapi.FakeServiceRepo{FindInstanceByNameServiceInstance: serviceInstance}
			})

			Context("when the command is confirmed", func() {
				It("deletes the service", func() {
					ui := callDeleteService("Y", []string{"my-service"}, requirementsFactory, serviceRepo)

					testassert.SliceContains(ui.Prompts, testassert.Lines{
						{"Are you sure"},
					})

					testassert.SliceContains(ui.Outputs, testassert.Lines{
						{"Deleting service", "my-service", "my-org", "my-space", "my-user"},
						{"OK"},
					})

					Expect(serviceRepo.DeleteServiceServiceInstance).To(Equal(serviceInstance))
				})
			})

			It("skips confirmation when the -f flag is given", func() {
				ui := callDeleteService("", []string{"-f", "foo.com"}, requirementsFactory, serviceRepo)

				Expect(ui.Prompts).To(BeEmpty())
				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"Deleting service", "foo.com"},
					{"OK"},
				})
			})
		})

		Context("when the service does not exist", func() {
			BeforeEach(func() {
				serviceRepo = &testapi.FakeServiceRepo{FindInstanceByNameNotFound: true}
			})

			It("warns the user the service does not exist", func() {
				ui := callDeleteService("", []string{"-f", "my-service"}, requirementsFactory, serviceRepo)

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

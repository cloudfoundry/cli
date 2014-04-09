package servicebroker_test

import (
	. "cf/commands/servicebroker"
	"cf/configuration"
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

var _ = Describe("delete-service-broker command", func() {
	var (
		ui                  *testterm.FakeUI
		configRepo          configuration.ReadWriter
		brokerRepo          *testapi.FakeServiceBrokerRepo
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{Inputs: []string{"y"}}
		brokerRepo = &testapi.FakeServiceBrokerRepo{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
	})

	runCommand := func(args ...string) {
		cmd := NewDeleteServiceBroker(ui, configRepo, brokerRepo)
		testcmd.RunCommand(cmd, testcmd.NewContext("delete-service-broker", args), requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails with usage when called without a broker's name", func() {
			runCommand()
			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		It("fails requirements when not logged in", func() {
			requirementsFactory.LoginSuccess = false
			runCommand("-f", "my-broker")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	Context("when the service broker exists", func() {
		BeforeEach(func() {
			brokerRepo.FindByNameServiceBroker = models.ServiceBroker{
				Name: "service-broker-to-delete",
				Guid: "service-broker-to-delete-guid",
			}
		})

		It("deletes the service broker with the given name", func() {
			runCommand("service-broker-to-delete")

			Expect(brokerRepo.FindByNameName).To(Equal("service-broker-to-delete"))
			Expect(brokerRepo.DeletedServiceBrokerGuid).To(Equal("service-broker-to-delete-guid"))
			Expect(len(ui.Outputs)).To(Equal(2))
			testassert.SliceContains(ui.Prompts, testassert.Lines{
				{"Really delete the service-broker service-broker-to-delete"},
			})

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Deleting service broker", "service-broker-to-delete", "my-user"},
				{"OK"},
			})
		})

		It("does not prompt when the -f flag is provided", func() {
			runCommand("-f", "service-broker-to-delete")

			Expect(brokerRepo.FindByNameName).To(Equal("service-broker-to-delete"))
			Expect(brokerRepo.DeletedServiceBrokerGuid).To(Equal("service-broker-to-delete-guid"))

			Expect(ui.Prompts).To(BeEmpty())
			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Deleting service broker", "service-broker-to-delete", "my-user"},
				{"OK"},
			})
		})
	})

	Context("when the service broker does not exist", func() {
		BeforeEach(func() {
			brokerRepo.FindByNameNotFound = true
		})

		It("warns the user", func() {
			ui.Inputs = []string{}
			runCommand("-f", "service-broker-to-delete")

			Expect(brokerRepo.FindByNameName).To(Equal("service-broker-to-delete"))
			Expect(brokerRepo.DeletedServiceBrokerGuid).To(Equal(""))
			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Deleting service broker", "service-broker-to-delete"},
				{"OK"},
			})

			testassert.SliceContains(ui.WarnOutputs, testassert.Lines{
				{"service-broker-to-delete", "does not exist"},
			})
		})
	})
})

package servicebroker_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	. "github.com/cloudfoundry/cli/cf/commands/servicebroker"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("rename-service-broker command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		configRepo          configuration.ReadWriter
		serviceBrokerRepo   *testapi.FakeServiceBrokerRepo
	)

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()

		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		serviceBrokerRepo = &testapi.FakeServiceBrokerRepo{}
	})

	runCommand := func(args ...string) {
		testcmd.RunCommand(NewRenameServiceBroker(ui, configRepo, serviceBrokerRepo), args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails with usage when not invoked with exactly two args", func() {
			requirementsFactory.LoginSuccess = true
			runCommand("welp")
			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		It("fails when not logged in", func() {
			runCommand("okay", "DO---IIIIT")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			broker := models.ServiceBroker{}
			broker.Name = "my-found-broker"
			broker.Guid = "my-found-broker-guid"
			serviceBrokerRepo.FindByNameServiceBroker = broker
		})

		It("renames the given service broker", func() {
			runCommand("my-broker", "my-new-broker")
			Expect(serviceBrokerRepo.FindByNameName).To(Equal("my-broker"))

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Renaming service broker", "my-found-broker", "my-new-broker", "my-user"},
				[]string{"OK"},
			))

			Expect(serviceBrokerRepo.RenamedServiceBrokerGuid).To(Equal("my-found-broker-guid"))
			Expect(serviceBrokerRepo.RenamedServiceBrokerName).To(Equal("my-new-broker"))
		})
	})
})

package servicebroker_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/servicebroker"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("create-service-broker command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		configRepo          core_config.ReadWriter
		serviceBrokerRepo   *testapi.FakeServiceBrokerRepo
	)

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()

		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		serviceBrokerRepo = &testapi.FakeServiceBrokerRepo{}
	})

	runCommand := func(args ...string) {
		testcmd.RunCommand(NewCreateServiceBroker(ui, configRepo, serviceBrokerRepo), args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails with usage when called without exactly four args", func() {
			requirementsFactory.LoginSuccess = true
			runCommand("whoops", "not-enough", "args")
			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		It("fails when not logged in", func() {
			runCommand("Just", "Enough", "Args", "Provided")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		It("creates a service broker, obviously", func() {
			runCommand("my-broker", "my-username", "my-password", "http://example.com")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Creating service broker", "my-broker", "my-user"},
				[]string{"OK"},
			))

			Expect(serviceBrokerRepo.CreateName).To(Equal("my-broker"))
			Expect(serviceBrokerRepo.CreateUrl).To(Equal("http://example.com"))
			Expect(serviceBrokerRepo.CreateUsername).To(Equal("my-username"))
			Expect(serviceBrokerRepo.CreatePassword).To(Equal("my-password"))
		})
	})
})

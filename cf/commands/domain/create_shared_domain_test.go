package domain_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/configuration"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/domain"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing with ginkgo", func() {
	var (
		requirementsFactory *testreq.FakeReqFactory
		ui                  *testterm.FakeUI
		domainRepo          *testapi.FakeDomainRepository
		configRepo          configuration.ReadWriter
	)
	BeforeEach(func() {
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
		domainRepo = &testapi.FakeDomainRepository{}
		configRepo = testconfig.NewRepositoryWithAccessToken(configuration.TokenInfo{Username: "my-user"})
	})

	runCommand := func(args ...string) {
		ui = new(testterm.FakeUI)
		cmd := NewCreateSharedDomain(ui, configRepo, domainRepo)
		testcmd.RunCommand(cmd, args, requirementsFactory)
		return
	}

	It("TestShareDomainRequirements", func() {
		runCommand("example.com")
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: false}
		runCommand("example.com")
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})

	It("TestShareDomainFailsWithUsage", func() {
		runCommand()
		Expect(ui.FailedWithUsage).To(BeTrue())

		runCommand("example.com")
		Expect(ui.FailedWithUsage).To(BeFalse())
	})

	It("TestShareDomain", func() {
		runCommand("example.com")

		Expect(domainRepo.CreateSharedDomainName).To(Equal("example.com"))
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Creating shared domain", "example.com", "my-user"},
			[]string{"OK"},
		))
	})
})

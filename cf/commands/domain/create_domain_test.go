package domain_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("create domain command", func() {

	var (
		requirementsFactory *testreq.FakeReqFactory
		ui                  *testterm.FakeUI
		domainRepo          *testapi.FakeDomainRepository
		configRepo          core_config.Repository
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetDomainRepository(domainRepo)
		deps.Config = configRepo
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("create-domain").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
		domainRepo = &testapi.FakeDomainRepository{}
		configRepo = testconfig.NewRepositoryWithAccessToken(core_config.TokenInfo{Username: "my-user"})
	})

	runCommand := func(args ...string) bool {
		ui = new(testterm.FakeUI)
		return testcmd.RunCliCommand("create-domain", args, requirementsFactory, updateCommandDependency, false)
	}

	It("fails with usage", func() {
		runCommand("")
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Incorrect Usage", "Requires", "arguments"},
		))

		runCommand("org1")
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Incorrect Usage", "Requires", "arguments"},
		))

		runCommand("org1", "example.com")
		Expect(ui.Outputs).ToNot(ContainSubstrings(
			[]string{"Incorrect Usage", "Requires", "arguments"},
		))

	})

	Context("checks login", func() {
		It("passes when logged in", func() {
			Expect(runCommand("my-org", "example.com")).To(BeTrue())
			Expect(requirementsFactory.OrganizationName).To(Equal("my-org"))
		})

		It("fails when not logged in", func() {
			requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: false}

			Expect(runCommand("my-org", "example.com")).To(BeFalse())
		})
	})

	It("creates a domain", func() {
		org := models.Organization{}
		org.Name = "myOrg"
		org.Guid = "myOrg-guid"
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, Organization: org}
		runCommand("myOrg", "example.com")

		domainName, domainOwningOrgGUID := domainRepo.CreateArgsForCall(0)
		Expect(domainName).To(Equal("example.com"))
		Expect(domainOwningOrgGUID).To(Equal("myOrg-guid"))
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Creating domain", "example.com", "myOrg", "my-user"},
			[]string{"OK"},
		))
	})
})

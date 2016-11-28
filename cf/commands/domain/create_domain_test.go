package domain_test

import (
	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("create domain command", func() {

	var (
		requirementsFactory *requirementsfakes.FakeFactory
		ui                  *testterm.FakeUI
		domainRepo          *apifakes.FakeDomainRepository
		configRepo          coreconfig.Repository
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.RepoLocator = deps.RepoLocator.SetDomainRepository(domainRepo)
		deps.Config = configRepo
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("create-domain").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		requirementsFactory = new(requirementsfakes.FakeFactory)
		requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		domainRepo = new(apifakes.FakeDomainRepository)
		configRepo = testconfig.NewRepositoryWithAccessToken(coreconfig.TokenInfo{Username: "my-user"})
	})

	runCommand := func(args ...string) bool {
		ui = new(testterm.FakeUI)
		return testcmd.RunCLICommand("create-domain", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	It("fails with usage", func() {
		runCommand("")
		Expect(ui.Outputs()).To(ContainSubstrings(
			[]string{"Incorrect Usage", "Requires", "arguments"},
		))

		runCommand("org1")
		Expect(ui.Outputs()).To(ContainSubstrings(
			[]string{"Incorrect Usage", "Requires", "arguments"},
		))
	})

	Context("checks login", func() {
		It("passes when logged in", func() {
			fakeOrgRequirement := new(requirementsfakes.FakeOrganizationRequirement)
			fakeOrgRequirement.GetOrganizationReturns(models.Organization{
				OrganizationFields: models.OrganizationFields{
					Name: "my-org",
				},
			})
			requirementsFactory.NewOrganizationRequirementReturns(fakeOrgRequirement)
			Expect(runCommand("my-org", "example.com")).To(BeTrue())
			Expect(ui.Outputs()).To(ContainSubstrings([]string{"my-org"}))
		})

		It("fails when not logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})

			Expect(runCommand("my-org", "example.com")).To(BeFalse())
		})
	})

	It("creates a domain", func() {
		org := models.Organization{}
		org.Name = "myOrg"
		org.GUID = "myOrg-guid"
		fakeOrgRequirement := new(requirementsfakes.FakeOrganizationRequirement)
		fakeOrgRequirement.GetOrganizationReturns(org)
		requirementsFactory.NewOrganizationRequirementReturns(fakeOrgRequirement)
		requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		runCommand("myOrg", "example.com")

		domainName, domainOwningOrgGUID := domainRepo.CreateArgsForCall(0)
		Expect(domainName).To(Equal("example.com"))
		Expect(domainOwningOrgGUID).To(Equal("myOrg-guid"))
		Expect(ui.Outputs()).To(ContainSubstrings(
			[]string{"Creating domain", "example.com", "myOrg", "my-user"},
			[]string{"OK"},
		))
	})
})

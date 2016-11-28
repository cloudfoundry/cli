package domain_test

import (
	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
)

var _ = Describe("delete-domain command", func() {
	var (
		ui                  *testterm.FakeUI
		configRepo          coreconfig.Repository
		domainRepo          *apifakes.FakeDomainRepository
		requirementsFactory *requirementsfakes.FakeFactory
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.RepoLocator = deps.RepoLocator.SetDomainRepository(domainRepo)
		deps.Config = configRepo
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("delete-domain").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{
			Inputs: []string{"yes"},
		}

		domainRepo = new(apifakes.FakeDomainRepository)

		requirementsFactory = new(requirementsfakes.FakeFactory)
		requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		requirementsFactory.NewTargetedOrgRequirementReturns(new(requirementsfakes.FakeTargetedOrgRequirement))

		fakeOrgRequirement := new(requirementsfakes.FakeOrganizationRequirement)
		fakeOrgRequirement.GetOrganizationReturns(models.Organization{
			OrganizationFields: models.OrganizationFields{
				Name: "my-org",
			},
		})
		requirementsFactory.NewOrganizationRequirementReturns(fakeOrgRequirement)

		configRepo = testconfig.NewRepositoryWithDefaults()
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("delete-domain", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("requirements", func() {
		It("fails when the user is not logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})

			Expect(runCommand("foo.com")).To(BeFalse())
		})

		It("fails when the an org is not targetted", func() {
			targetedOrganizationReq := new(requirementsfakes.FakeTargetedOrgRequirement)
			targetedOrganizationReq.ExecuteReturns(errors.New("not targeted"))
			requirementsFactory.NewTargetedOrgRequirementReturns(targetedOrganizationReq)

			Expect(runCommand("foo.com")).To(BeFalse())
		})
	})

	Context("when the domain is shared", func() {
		BeforeEach(func() {
			domainRepo.FindByNameInOrgReturns(
				models.DomainFields{
					Name:   "foo1.com",
					GUID:   "foo1-guid",
					Shared: true,
				}, nil)
		})
		It("informs the user that the domain is shared", func() {
			runCommand("foo1.com")

			Expect(domainRepo.DeleteCallCount()).To(BeZero())
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"FAILED"},
				[]string{"domain"},
				[]string{"foo1.com"},
				[]string{"is a shared domain, not an owned domain."},
				[]string{"TIP"},
				[]string{"Use `cf delete-shared-domain` to delete shared domains."},
			))

		})
	})
	Context("when the domain exists", func() {
		BeforeEach(func() {
			domainRepo.FindByNameInOrgReturns(
				models.DomainFields{
					Name: "foo.com",
					GUID: "foo-guid",
				}, nil)
		})

		It("deletes domains", func() {
			runCommand("foo.com")

			Expect(domainRepo.DeleteArgsForCall(0)).To(Equal("foo-guid"))

			Expect(ui.Prompts).To(ContainSubstrings([]string{"Really delete the domain foo.com"}))
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Deleting domain", "foo.com", "my-user"},
				[]string{"OK"},
			))
		})

		Context("when there is an error deleting the domain", func() {
			BeforeEach(func() {
				domainRepo.DeleteReturns(errors.New("failed badly"))
			})

			It("show the error the user", func() {
				runCommand("foo.com")

				Expect(domainRepo.DeleteArgsForCall(0)).To(Equal("foo-guid"))

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Deleting domain", "foo.com"},
					[]string{"FAILED"},
					[]string{"foo.com"},
					[]string{"failed badly"},
				))
			})
		})

		Context("when the user does not confirm", func() {
			BeforeEach(func() {
				ui.Inputs = []string{"no"}
			})

			It("does nothing", func() {
				runCommand("foo.com")

				Expect(domainRepo.DeleteCallCount()).To(BeZero())

				Expect(ui.Prompts).To(ContainSubstrings([]string{"delete", "foo.com"}))

				Expect(ui.Outputs()).To(BeEmpty())
			})
		})

		Context("when the user provides the -f flag", func() {
			BeforeEach(func() {
				ui.Inputs = []string{}
			})

			It("skips confirmation", func() {
				runCommand("-f", "foo.com")

				Expect(domainRepo.DeleteArgsForCall(0)).To(Equal("foo-guid"))
				Expect(ui.Prompts).To(BeEmpty())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Deleting domain", "foo.com"},
					[]string{"OK"},
				))
			})
		})
	})

	Context("when a domain with the given name doesn't exist", func() {
		BeforeEach(func() {
			domainRepo.FindByNameInOrgReturns(models.DomainFields{}, errors.NewModelNotFoundError("Domain", "foo.com"))
		})

		It("fails", func() {
			runCommand("foo.com")

			Expect(domainRepo.DeleteCallCount()).To(BeZero())

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"OK"},
				[]string{"foo.com", "not found"},
			))
		})
	})

	Context("when there is an error finding the domain", func() {
		BeforeEach(func() {
			domainRepo.FindByNameInOrgReturns(models.DomainFields{}, errors.New("failed badly"))
		})

		It("shows the error to the user", func() {
			runCommand("foo.com")

			Expect(domainRepo.DeleteCallCount()).To(BeZero())

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"FAILED"},
				[]string{"foo.com"},
				[]string{"failed badly"},
			))
		})
	})
})

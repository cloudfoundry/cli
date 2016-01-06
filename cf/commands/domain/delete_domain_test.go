package domain_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("delete-domain command", func() {
	var (
		ui                  *testterm.FakeUI
		configRepo          core_config.Repository
		domainRepo          *testapi.FakeDomainRepository
		requirementsFactory *testreq.FakeReqFactory
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetDomainRepository(domainRepo)
		deps.Config = configRepo
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("delete-domain").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{
			Inputs: []string{"yes"},
		}

		domainRepo = &testapi.FakeDomainRepository{}
		requirementsFactory = &testreq.FakeReqFactory{
			LoginSuccess:       true,
			TargetedOrgSuccess: true,
		}
		configRepo = testconfig.NewRepositoryWithDefaults()
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("delete-domain", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("requirements", func() {
		It("fails when the user is not logged in", func() {
			requirementsFactory.LoginSuccess = false

			Expect(runCommand("foo.com")).To(BeFalse())
		})

		It("fails when the an org is not targetted", func() {
			requirementsFactory.TargetedOrgSuccess = false

			Expect(runCommand("foo.com")).To(BeFalse())
		})
	})

	Context("when the domain is shared", func() {
		BeforeEach(func() {
			domainRepo.FindByNameInOrgReturns(
				models.DomainFields{
					Name:   "foo1.com",
					Guid:   "foo1-guid",
					Shared: true,
				}, nil)
		})
		It("informs the user that the domain is shared", func() {
			runCommand("foo1.com")

			Expect(domainRepo.DeleteCallCount()).To(BeZero())
			Expect(ui.Outputs).To(ContainSubstrings(
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
					Guid: "foo-guid",
				}, nil)
		})

		It("deletes domains", func() {
			runCommand("foo.com")

			Expect(domainRepo.DeleteArgsForCall(0)).To(Equal("foo-guid"))

			Expect(ui.Prompts).To(ContainSubstrings([]string{"Really delete the domain foo.com"}))
			Expect(ui.Outputs).To(ContainSubstrings(
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

				Expect(ui.Outputs).To(ContainSubstrings(
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

				Expect(ui.Outputs).To(BeEmpty())
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
				Expect(ui.Outputs).To(ContainSubstrings(
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

			Expect(ui.Outputs).To(ContainSubstrings(
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

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"FAILED"},
				[]string{"foo.com"},
				[]string{"failed badly"},
			))
		})
	})
})

package domain_test

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/flags"

	fakeapi "github.com/cloudfoundry/cli/cf/api/fakes"
	fakerequirements "github.com/cloudfoundry/cli/cf/requirements/fakes"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"

	"github.com/cloudfoundry/cli/cf/commands/domain"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ListDomains", func() {
	var (
		ui             *testterm.FakeUI
		routingApiRepo *fakeapi.FakeRoutingApiRepository
		domainRepo     *fakeapi.FakeDomainRepository
		configRepo     core_config.Repository

		cmd         domain.ListDomains
		deps        command_registry.Dependency
		factory     *fakerequirements.FakeFactory
		flagContext flags.FlagContext

		loginRequirement       requirements.Requirement
		targetedOrgRequirement *fakerequirements.FakeTargetedOrgRequirement

		domainFields []models.DomainFields
		routerGroups models.RouterGroups
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		routingApiRepo = &fakeapi.FakeRoutingApiRepository{}
		repoLocator := deps.RepoLocator.SetRoutingApiRepository(routingApiRepo)

		domainRepo = &fakeapi.FakeDomainRepository{}
		repoLocator = repoLocator.SetDomainRepository(domainRepo)

		deps = command_registry.Dependency{
			Ui:          ui,
			Config:      configRepo,
			RepoLocator: repoLocator,
		}

		cmd = domain.ListDomains{}
		cmd.SetDependency(deps, false)

		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)

		factory = &fakerequirements.FakeFactory{}
		loginRequirement = &passingRequirement{Name: "LoginRequirement"}
		factory.NewLoginRequirementReturns(loginRequirement)

		targetedOrgRequirement = &fakerequirements.FakeTargetedOrgRequirement{}
		factory.NewTargetedOrgRequirementReturns(targetedOrgRequirement)

		domainRepo.ListDomainsForOrgStub = func(orgGuid string, cb func(models.DomainFields) bool) error {
			for _, field := range domainFields {
				if !cb(field) {
					break
				}
			}
			return nil
		}

		routerGroups = models.RouterGroups{
			models.RouterGroup{
				Guid: "router-group-guid",
				Name: "my-router-name1",
				Type: "tcp",
			},
		}
		routingApiRepo.ListRouterGroupsStub = func(cb func(models.RouterGroup) bool) error {
			for _, routerGroup := range routerGroups {
				if !cb(routerGroup) {
					break
				}
			}
			return nil
		}
	})

	Describe("Requirements", func() {
		Context("when arguments are provided", func() {
			var cmd command_registry.Command
			var flagContext flags.FlagContext

			BeforeEach(func() {
				cmd = &domain.ListDomains{}
				cmd.SetDependency(deps, false)
				flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
			})

			It("should fail with usage", func() {
				flagContext.Parse("blahblah")

				reqs := cmd.Requirements(factory, flagContext)

				err := testcmd.RunRequirements(reqs)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Incorrect Usage"))
				Expect(err.Error()).To(ContainSubstring("No argument required"))
			})
		})

		Context("when provided no arguments", func() {
			BeforeEach(func() {
				flagContext.Parse()
			})

			It("does not fail with usage", func() {
				cmd.Requirements(factory, flagContext)
				Expect(ui.Outputs).NotTo(ContainSubstrings(
					[]string{"Incorrect Usage. No argument required"},
					[]string{"NAME"},
					[]string{"USAGE"},
				))
			})

			It("returns a LoginRequirement", func() {
				actualRequirements := cmd.Requirements(factory, flagContext)
				Expect(factory.NewLoginRequirementCallCount()).To(Equal(1))
				Expect(actualRequirements).To(ContainElement(loginRequirement))
			})

			It("returns a TargetedOrgRequirement", func() {
				actualRequirements := cmd.Requirements(factory, flagContext)
				Expect(factory.NewTargetedOrgRequirementCallCount()).To(Equal(1))
				Expect(actualRequirements).To(ContainElement(targetedOrgRequirement))
			})
		})
	})

	Describe("Execute", func() {
		It("prints getting domains message", func() {
			cmd.Execute(flagContext)
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting domains in org my-org"},
			))
		})

		It("tries to get the list of domains for org", func() {
			cmd.Execute(flagContext)
			Expect(domainRepo.ListDomainsForOrgCallCount()).To(Equal(1))
			orgGuid, _ := domainRepo.ListDomainsForOrgArgsForCall(0)
			Expect(orgGuid).To(Equal("my-org-guid"))
		})

		It("prints no domains found message", func() {
			cmd.Execute(flagContext)
			Expect(ui.Outputs).To(BeInDisplayOrder(
				[]string{"name", "status"},
				[]string{"No domains found"},
			))
		})

		Context("when list domains for org returns error", func() {
			BeforeEach(func() {
				domainRepo.ListDomainsForOrgReturns(errors.New("org-domain-err"))
			})

			It("fails with message", func() {
				Expect(func() { cmd.Execute(flagContext) }).To(Panic())
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Failed fetching domains."},
					[]string{"org-domain-err"},
				))
			})
		})

		Context("when domains are found", func() {
			BeforeEach(func() {
				domainFields = []models.DomainFields{
					models.DomainFields{Shared: false, Name: "Private-domain1"},
					models.DomainFields{Shared: false, Name: "Private-domain2", RouterGroupTypes: []string{"tcp", "bazquux"}},
					models.DomainFields{Shared: true, Name: "Shared-domain1"},
					models.DomainFields{Shared: true, Name: "Shared-domain2", RouterGroupTypes: []string{"tcp", "foobar"}},
				}
			})

			AfterEach(func() {
				domainFields = []models.DomainFields{}
			})

			It("does not print no domains found message", func() {
				cmd.Execute(flagContext)
				Expect(ui.Outputs).NotTo(ContainSubstrings(
					[]string{"No domains found"},
				))
			})

			It("prints the domain information", func() {
				cmd.Execute(flagContext)
				Expect(ui.Outputs).To(BeInDisplayOrder(
					[]string{"name", "status", "type"},
					[]string{"Shared-domain1", "shared"},
					[]string{"Shared-domain2", "shared", "tcp, foobar"},
					[]string{"Private-domain1", "owned"},
					[]string{"Private-domain2", "owned", "tcp, bazquux"},
				))
			})
		})
	})
})

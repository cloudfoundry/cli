package domain_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"

	"code.cloudfoundry.org/cli/cf/api/apifakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"

	"code.cloudfoundry.org/cli/cf/commands/domain"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ListDomains", func() {
	var (
		ui             *testterm.FakeUI
		routingAPIRepo *apifakes.FakeRoutingAPIRepository
		domainRepo     *apifakes.FakeDomainRepository
		configRepo     coreconfig.Repository

		cmd         domain.ListDomains
		deps        commandregistry.Dependency
		factory     *requirementsfakes.FakeFactory
		flagContext flags.FlagContext

		loginRequirement       requirements.Requirement
		targetedOrgRequirement *requirementsfakes.FakeTargetedOrgRequirement

		domainFields []models.DomainFields
		routerGroups models.RouterGroups
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		routingAPIRepo = new(apifakes.FakeRoutingAPIRepository)
		repoLocator := deps.RepoLocator.SetRoutingAPIRepository(routingAPIRepo)

		domainRepo = new(apifakes.FakeDomainRepository)
		repoLocator = repoLocator.SetDomainRepository(domainRepo)

		deps = commandregistry.Dependency{
			UI:          ui,
			Config:      configRepo,
			RepoLocator: repoLocator,
		}

		cmd = domain.ListDomains{}
		cmd.SetDependency(deps, false)

		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)

		factory = new(requirementsfakes.FakeFactory)
		loginRequirement = &passingRequirement{Name: "LoginRequirement"}
		factory.NewLoginRequirementReturns(loginRequirement)

		targetedOrgRequirement = new(requirementsfakes.FakeTargetedOrgRequirement)
		factory.NewTargetedOrgRequirementReturns(targetedOrgRequirement)

		domainRepo.ListDomainsForOrgStub = func(orgGUID string, cb func(models.DomainFields) bool) error {
			for _, field := range domainFields {
				if !cb(field) {
					break
				}
			}
			return nil
		}

		routerGroups = models.RouterGroups{
			models.RouterGroup{
				GUID: "router-group-guid",
				Name: "my-router-name1",
				Type: "tcp",
			},
		}
		routingAPIRepo.ListRouterGroupsStub = func(cb func(models.RouterGroup) bool) error {
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
			var cmd commandregistry.Command
			var flagContext flags.FlagContext

			BeforeEach(func() {
				cmd = &domain.ListDomains{}
				cmd.SetDependency(deps, false)
				flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
			})

			It("should fail with usage", func() {
				flagContext.Parse("blahblah")

				reqs, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())

				err = testcmd.RunRequirements(reqs)
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
				_, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(ui.Outputs()).NotTo(ContainSubstrings(
					[]string{"Incorrect Usage. No argument required"},
					[]string{"NAME"},
					[]string{"USAGE"},
				))
			})

			It("returns a LoginRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(factory.NewLoginRequirementCallCount()).To(Equal(1))
				Expect(actualRequirements).To(ContainElement(loginRequirement))
			})

			It("returns a TargetedOrgRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(factory.NewTargetedOrgRequirementCallCount()).To(Equal(1))
				Expect(actualRequirements).To(ContainElement(targetedOrgRequirement))
			})
		})
	})

	Describe("Execute", func() {
		var err error

		JustBeforeEach(func() {
			err = cmd.Execute(flagContext)
		})

		It("prints getting domains message", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Getting domains in org my-org"},
			))
		})

		It("tries to get the list of domains for org", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(domainRepo.ListDomainsForOrgCallCount()).To(Equal(1))
			orgGUID, _ := domainRepo.ListDomainsForOrgArgsForCall(0)
			Expect(orgGUID).To(Equal("my-org-guid"))
		})

		It("prints no domains found message", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(ui.Outputs()).To(BeInDisplayOrder(
				[]string{"name", "status"},
				[]string{"No domains found"},
			))
		})

		Context("when list domains for org returns error", func() {
			BeforeEach(func() {
				domainRepo.ListDomainsForOrgReturns(errors.New("org-domain-err"))
			})

			It("fails with message", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Failed fetching domains."))
				Expect(err.Error()).To(ContainSubstring("org-domain-err"))
			})
		})

		Context("when domains are found", func() {
			BeforeEach(func() {
				domainFields = []models.DomainFields{
					{Shared: false, Name: "Private-domain1"},
					{Shared: false, Name: "Private-domain2", RouterGroupType: "tcp"},
					{Shared: true, Name: "Shared-domain1"},
					{Shared: true, Name: "Shared-domain2", RouterGroupType: "foobar"},
				}
			})

			AfterEach(func() {
				domainFields = []models.DomainFields{}
			})

			It("does not print no domains found message", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(ui.Outputs()).NotTo(ContainSubstrings(
					[]string{"No domains found"},
				))
			})

			It("prints the domain information", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(ui.Outputs()).To(BeInDisplayOrder(
					[]string{"name", "status", "type"},
					[]string{"Shared-domain1", "shared"},
					[]string{"Shared-domain2", "shared", "foobar"},
					[]string{"Private-domain1", "owned"},
					[]string{"Private-domain2", "owned", "tcp"},
				))
			})
		})
	})
})

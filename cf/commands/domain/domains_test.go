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
		Context("when provided one arg", func() {
			BeforeEach(func() {
				flagContext.Parse("arg-1")
			})

			It("fails with usage", func() {
				Expect(func() {
					cmd.Requirements(factory, flagContext)
				}).To(Panic())
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Incorrect Usage. No argument required"},
					[]string{"NAME"},
					[]string{"USAGE"},
				))
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
		It("prints getting domains", func() {
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

		Context("when list domans for org returns error", func() {
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

		Context("when there are no domains for org", func() {
			BeforeEach(func() {
				domainFields = []models.DomainFields{}
				cmd.Execute(flagContext)
			})

			It("prints no domains found message", func() {
				Expect(ui.Outputs).To(BeInDisplayOrder(
					[]string{"name", "status", "type"},
					[]string{"No domains found"},
				))
			})
		})

		Context("when domains are found", func() {
			BeforeEach(func() {
				domainFields = []models.DomainFields{
					models.DomainFields{Shared: false, Name: "Private-domain1"},
					models.DomainFields{Shared: true, Name: "Shared-domain1"},
				}
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
					[]string{"Private-domain1", "owned"},
				))
			})

			Context("when routing api endpoint is not set", func() {
				BeforeEach(func() {
					configRepo.SetRoutingApiEndpoint("")
				})

				It("does not panic", func() {
					Expect(func() { cmd.Execute(flagContext) }).NotTo(Panic())
				})

				Context("when returned domain has a router-group-guid", func() {
					BeforeEach(func() {
						domainFields = []models.DomainFields{
							models.DomainFields{
								Shared:          true,
								Name:            "Shared-domain1",
								RouterGroupGuid: "router-group-guid"},
						}
					})

					It("panics with error about missing routing api endpoint", func() {
						Expect(func() { cmd.Execute(flagContext) }).To(Panic())
						Expect(ui.Outputs).To(ContainSubstrings(
							[]string{"Routing API URI missing."},
						))
					})
				})
			})

			Context("when routing api endpoint is set", func() {
				BeforeEach(func() {
					configRepo.SetRoutingApiEndpoint("routing-api-endpoint")

					domainFields = []models.DomainFields{
						models.DomainFields{Shared: true,
							Name:            "Shared-domain1",
							RouterGroupGuid: "router-group-guid"},
					}
				})

				It("prints domain information with router group type", func() {
					cmd.Execute(flagContext)
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"name", "status", "type"},
						[]string{"Shared-domain1", "shared", "tcp"},
					))
				})

				Context("when list router groups returns error", func() {
					BeforeEach(func() {
						routingApiRepo.ListRouterGroupsReturns(errors.New("router-group-err"))
					})

					It("fails with error message", func() {
						Expect(func() { cmd.Execute(flagContext) }).To(Panic())
						Expect(ui.Outputs).To(ContainSubstrings(
							[]string{"FAILED"},
							[]string{"Failed fetching router groups."},
							[]string{"router-group-err"},
						))
					})
				})

				Context("when router group does not exists", func() {
					BeforeEach(func() {
						routerGroups = models.RouterGroups{}
					})

					It("prints the invalid router group message", func() {
						Expect(func() { cmd.Execute(flagContext) }).To(Panic())
						Expect(ui.Outputs).To(ContainSubstrings(
							[]string{"FAILED"},
							[]string{"Invalid router group guid: router-group-guid"},
						))
					})

					It("does not print table header", func() {
						Expect(func() { cmd.Execute(flagContext) }).To(Panic())
						Expect(ui.Outputs).ToNot(ContainSubstrings(
							[]string{"name", "status", "type"},
						))
					})
				})
			})
		})
	})
})

package domain_test

import (
	"errors"

	"github.com/blang/semver"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/flags"

	fakeapi "github.com/cloudfoundry/cli/cf/api/fakes"
	fakerequirements "github.com/cloudfoundry/cli/cf/requirements/fakes"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"

	"github.com/cloudfoundry/cli/cf/commands/domain"
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type passingRequirement struct {
	Name string
}

func (r passingRequirement) Execute() error {
	return nil
}

var _ = Describe("CreateSharedDomain", func() {
	var (
		ui             *testterm.FakeUI
		routingApiRepo *fakeapi.FakeRoutingApiRepository
		domainRepo     *fakeapi.FakeDomainRepository
		configRepo     core_config.Repository

		cmd         domain.CreateSharedDomain
		deps        command_registry.Dependency
		factory     *fakerequirements.FakeFactory
		flagContext flags.FlagContext

		loginRequirement         requirements.Requirement
		routingApiRequirement    requirements.Requirement
		minAPIVersionRequirement requirements.Requirement

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

		cmd = domain.CreateSharedDomain{}
		cmd.SetDependency(deps, false)

		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)

		factory = &fakerequirements.FakeFactory{}

		loginRequirement = &passingRequirement{Name: "Login"}
		factory.NewLoginRequirementReturns(loginRequirement)

		routingApiRequirement = &passingRequirement{Name: "RoutingApi"}
		factory.NewRoutingAPIRequirementReturns(routingApiRequirement)

		minAPIVersionRequirement = &passingRequirement{"MinApiVersionRequirement"}
		factory.NewMinAPIVersionRequirementReturns(minAPIVersionRequirement)

		routingApiRepo.ListRouterGroupsStub = func(cb func(models.RouterGroup) bool) error {
			for _, r := range routerGroups {
				if !cb(r) {
					break
				}
			}
			return nil
		}
	})

	Describe("Requirements", func() {
		Context("when not provided exactly one arg", func() {
			BeforeEach(func() {
				flagContext.Parse("arg-1", "extra-arg")
			})

			It("fails with usage", func() {
				Expect(func() { cmd.Requirements(factory, flagContext) }).To(Panic())
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Incorrect Usage. Requires DOMAIN as an argument"},
					[]string{"NAME"},
					[]string{"USAGE"},
				))
			})
		})

		Context("when provided exactly one arg", func() {
			BeforeEach(func() {
				flagContext.Parse("domain-name")
			})

			It("does not fail with usage", func() {
				cmd.Requirements(factory, flagContext)
				Expect(ui.Outputs).NotTo(ContainSubstrings(
					[]string{"Incorrect Usage. Requires DOMAIN as an argument"},
					[]string{"NAME"},
					[]string{"USAGE"},
				))
			})

			It("returns a LoginRequirement", func() {
				actualRequirements := cmd.Requirements(factory, flagContext)
				Expect(factory.NewLoginRequirementCallCount()).To(Equal(1))
				Expect(actualRequirements).To(ContainElement(loginRequirement))
			})

			It("does not return a RoutingApiRequirement", func() {
				actualRequirements := cmd.Requirements(factory, flagContext)
				Expect(factory.NewRoutingAPIRequirementCallCount()).To(Equal(0))
				Expect(actualRequirements).ToNot(ContainElement(routingApiRequirement))
			})

			It("does not return a MinAPIVersionRequirement", func() {
				actualRequirements := cmd.Requirements(factory, flagContext)
				Expect(actualRequirements).NotTo(ContainElement(minAPIVersionRequirement))
			})

			Context("when router-group flag is set", func() {
				BeforeEach(func() {
					flagContext.Parse("domain-name", "--router-group", "route-group-name")
				})

				It("returns a LoginRequirement", func() {
					actualRequirements := cmd.Requirements(factory, flagContext)
					Expect(factory.NewLoginRequirementCallCount()).To(Equal(1))
					Expect(actualRequirements).To(ContainElement(loginRequirement))
				})

				It("returns a RoutingApiRequirement", func() {
					actualRequirements := cmd.Requirements(factory, flagContext)

					Expect(factory.NewRoutingAPIRequirementCallCount()).To(Equal(1))
					Expect(actualRequirements).To(ContainElement(routingApiRequirement))
				})

				It("returns a MinAPIVersionRequirement", func() {
					expectedVersion, err := semver.Make("2.36.0")
					Expect(err).NotTo(HaveOccurred())

					actualRequirements := cmd.Requirements(factory, flagContext)

					Expect(factory.NewMinAPIVersionRequirementCallCount()).To(Equal(1))
					feature, requiredVersion := factory.NewMinAPIVersionRequirementArgsForCall(0)
					Expect(feature).To(Equal("Option '--router-group'"))
					Expect(requiredVersion).To(Equal(expectedVersion))
					Expect(actualRequirements).To(ContainElement(minAPIVersionRequirement))
				})
			})
		})
	})

	Describe("Execute", func() {
		Context("when router-group flag is set", func() {
			BeforeEach(func() {
				routerGroups = models.RouterGroups{
					models.RouterGroup{
						Name: "router-group-name",
						Guid: "router-group-guid",
						Type: "router-group-type",
					},
				}
				flagContext.Parse("domain-name", "--router-group", "router-group-name")
			})

			It("tries to retrieve the router group", func() {
				cmd.Execute(flagContext)
				Expect(routingApiRepo.ListRouterGroupsCallCount()).To(Equal(1))
			})

			It("prints a message", func() {
				cmd.Execute(flagContext)
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Creating shared domain domain-name"},
				))
			})

			It("tries to create a shared domain with router group", func() {
				cmd.Execute(flagContext)
				Expect(domainRepo.CreateSharedDomainCallCount()).To(Equal(1))
				domainName, routerGroupGuid := domainRepo.CreateSharedDomainArgsForCall(0)
				Expect(domainName).To(Equal("domain-name"))
				Expect(routerGroupGuid).To(Equal("router-group-guid"))
			})

			It("prints success message", func() {
				cmd.Execute(flagContext)
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"OK"},
				))
			})

			Context("when listing router groups returns an error", func() {
				BeforeEach(func() {
					routingApiRepo.ListRouterGroupsReturns(errors.New("router-group-error"))
				})

				It("fails with error message", func() {
					Expect(func() { cmd.Execute(flagContext) }).To(Panic())
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"router-group-error"},
					))
				})
			})

			Context("when router group is not found", func() {
				BeforeEach(func() {
					routerGroups = models.RouterGroups{}
				})

				It("fails with a message", func() {
					Expect(func() { cmd.Execute(flagContext) }).To(Panic())
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"Router group router-group-name not found"},
					))
				})
			})
		})

		Context("when router-group flag is not set", func() {
			BeforeEach(func() {
				flagContext.Parse("domain-name")
			})

			It("does not try to retrieve the router group", func() {
				cmd.Execute(flagContext)
				Expect(routingApiRepo.ListRouterGroupsCallCount()).To(Equal(0))
			})

			It("prints a message", func() {
				cmd.Execute(flagContext)
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Creating shared domain domain-name"},
				))
			})

			It("tries to create a shared domain without router group", func() {
				cmd.Execute(flagContext)
				Expect(domainRepo.CreateSharedDomainCallCount()).To(Equal(1))
				domainName, routerGroupGuid := domainRepo.CreateSharedDomainArgsForCall(0)
				Expect(domainName).To(Equal("domain-name"))
				Expect(routerGroupGuid).To(Equal(""))
			})

			It("prints success message", func() {
				cmd.Execute(flagContext)
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"OK"},
				))
			})
		})

		Context("when creating shared domain returns error", func() {
			BeforeEach(func() {
				flagContext.Parse("domain-name")
				domainRepo.CreateSharedDomainReturns(errors.New("create-domain-error"))
			})

			It("fails with error", func() {
				Expect(func() { cmd.Execute(flagContext) }).To(Panic())
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"create-domain-error"},
				))
			})
		})
	})
})

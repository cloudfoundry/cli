package domain_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	"github.com/blang/semver"

	"code.cloudfoundry.org/cli/cf/api/apifakes"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"

	"code.cloudfoundry.org/cli/cf/commands/domain"
	"code.cloudfoundry.org/cli/cf/models"
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
		routingAPIRepo *apifakes.FakeRoutingAPIRepository
		domainRepo     *apifakes.FakeDomainRepository
		configRepo     coreconfig.Repository

		cmd         domain.CreateSharedDomain
		deps        commandregistry.Dependency
		factory     *requirementsfakes.FakeFactory
		flagContext flags.FlagContext

		loginRequirement         requirements.Requirement
		routingAPIRequirement    requirements.Requirement
		minAPIVersionRequirement requirements.Requirement

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

		cmd = domain.CreateSharedDomain{}
		cmd.SetDependency(deps, false)

		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)

		factory = new(requirementsfakes.FakeFactory)

		loginRequirement = &passingRequirement{Name: "Login"}
		factory.NewLoginRequirementReturns(loginRequirement)

		routingAPIRequirement = &passingRequirement{Name: "RoutingApi"}
		factory.NewRoutingAPIRequirementReturns(routingAPIRequirement)

		minAPIVersionRequirement = &passingRequirement{"MinAPIVersionRequirement"}
		factory.NewMinAPIVersionRequirementReturns(minAPIVersionRequirement)

		routingAPIRepo.ListRouterGroupsStub = func(cb func(models.RouterGroup) bool) error {
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
				_, err := cmd.Requirements(factory, flagContext)
				Expect(err).To(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
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
				_, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())

				Expect(ui.Outputs()).NotTo(ContainSubstrings(
					[]string{"Incorrect Usage. Requires DOMAIN as an argument"},
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

			It("does not return a RoutingAPIRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(factory.NewRoutingAPIRequirementCallCount()).To(Equal(0))
				Expect(actualRequirements).ToNot(ContainElement(routingAPIRequirement))
			})

			It("does not return a MinAPIVersionRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(actualRequirements).NotTo(ContainElement(minAPIVersionRequirement))
			})

			Context("when router-group flag is set", func() {
				BeforeEach(func() {
					flagContext.Parse("domain-name", "--router-group", "route-group-name")
				})

				It("returns a LoginRequirement", func() {
					actualRequirements, err := cmd.Requirements(factory, flagContext)
					Expect(err).NotTo(HaveOccurred())
					Expect(factory.NewLoginRequirementCallCount()).To(Equal(1))
					Expect(actualRequirements).To(ContainElement(loginRequirement))
				})

				It("returns a RoutingAPIRequirement", func() {
					actualRequirements, err := cmd.Requirements(factory, flagContext)
					Expect(err).NotTo(HaveOccurred())

					Expect(factory.NewRoutingAPIRequirementCallCount()).To(Equal(1))
					Expect(actualRequirements).To(ContainElement(routingAPIRequirement))
				})

				It("returns a MinAPIVersionRequirement", func() {
					expectedVersion, err := semver.Make("2.36.0")
					Expect(err).NotTo(HaveOccurred())

					actualRequirements, err := cmd.Requirements(factory, flagContext)
					Expect(err).NotTo(HaveOccurred())

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
		var err error

		JustBeforeEach(func() {
			err = cmd.Execute(flagContext)
		})

		Context("when router-group flag is set", func() {
			BeforeEach(func() {
				routerGroups = models.RouterGroups{
					models.RouterGroup{
						Name: "router-group-name",
						GUID: "router-group-guid",
						Type: "router-group-type",
					},
				}
				flagContext.Parse("domain-name", "--router-group", "router-group-name")
			})

			It("tries to retrieve the router group", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(routingAPIRepo.ListRouterGroupsCallCount()).To(Equal(1))
			})

			It("prints a message", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Creating shared domain domain-name"},
				))
			})

			It("tries to create a shared domain with router group", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(domainRepo.CreateSharedDomainCallCount()).To(Equal(1))
				domainName, routerGroupGUID := domainRepo.CreateSharedDomainArgsForCall(0)
				Expect(domainName).To(Equal("domain-name"))
				Expect(routerGroupGUID).To(Equal("router-group-guid"))
			})

			It("prints success message", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"OK"},
				))
			})

			Context("when listing router groups returns an error", func() {
				BeforeEach(func() {
					routingAPIRepo.ListRouterGroupsReturns(errors.New("router-group-error"))
				})

				It("fails with error message", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("router-group-error"))
				})
			})

			Context("when router group is not found", func() {
				BeforeEach(func() {
					routerGroups = models.RouterGroups{}
				})

				It("fails with a message", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("Router group router-group-name not found"))
				})
			})
		})

		Context("when router-group flag is not set", func() {
			BeforeEach(func() {
				flagContext.Parse("domain-name")
			})

			It("does not try to retrieve the router group", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(routingAPIRepo.ListRouterGroupsCallCount()).To(Equal(0))
			})

			It("prints a message", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Creating shared domain domain-name"},
				))
			})

			It("tries to create a shared domain without router group", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(domainRepo.CreateSharedDomainCallCount()).To(Equal(1))
				domainName, routerGroupGUID := domainRepo.CreateSharedDomainArgsForCall(0)
				Expect(domainName).To(Equal("domain-name"))
				Expect(routerGroupGUID).To(Equal(""))
			})

			It("prints success message", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
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
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("create-domain-error"))
			})
		})
	})
})

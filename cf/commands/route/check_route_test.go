package route_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/commands/route"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	"github.com/blang/semver"

	"code.cloudfoundry.org/cli/cf/api/apifakes"

	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CheckRoute", func() {
	var (
		ui         *testterm.FakeUI
		configRepo coreconfig.Repository
		routeRepo  *apifakes.FakeRouteRepository
		domainRepo *apifakes.FakeDomainRepository

		cmd         commandregistry.Command
		deps        commandregistry.Dependency
		factory     *requirementsfakes.FakeFactory
		flagContext flags.FlagContext

		loginRequirement         requirements.Requirement
		targetedOrgRequirement   *requirementsfakes.FakeTargetedOrgRequirement
		minAPIVersionRequirement requirements.Requirement
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}

		configRepo = testconfig.NewRepositoryWithDefaults()
		routeRepo = new(apifakes.FakeRouteRepository)
		repoLocator := deps.RepoLocator.SetRouteRepository(routeRepo)

		domainRepo = new(apifakes.FakeDomainRepository)
		repoLocator = repoLocator.SetDomainRepository(domainRepo)

		deps = commandregistry.Dependency{
			UI:          ui,
			Config:      configRepo,
			RepoLocator: repoLocator,
		}

		cmd = &route.CheckRoute{}
		cmd.SetDependency(deps, false)

		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)

		factory = new(requirementsfakes.FakeFactory)

		loginRequirement = &passingRequirement{Name: "login-requirement"}
		factory.NewLoginRequirementReturns(loginRequirement)

		targetedOrgRequirement = new(requirementsfakes.FakeTargetedOrgRequirement)
		factory.NewTargetedOrgRequirementReturns(targetedOrgRequirement)

		minAPIVersionRequirement = &passingRequirement{Name: "min-api-version-requirement"}
		factory.NewMinAPIVersionRequirementReturns(minAPIVersionRequirement)
	})

	Describe("Requirements", func() {
		Context("when not provided exactly two args", func() {
			BeforeEach(func() {
				flagContext.Parse("app-name")
			})

			It("fails with usage", func() {
				_, err := cmd.Requirements(factory, flagContext)
				Expect(err).To(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Incorrect Usage. Requires host and domain as arguments"},
				))
			})
		})

		Context("when provided exactly two args", func() {
			BeforeEach(func() {
				flagContext.Parse("domain-name", "host-name")
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

			Context("when a path is passed", func() {
				BeforeEach(func() {
					flagContext.Parse("domain-name", "hostname", "--path", "the-path")
				})

				It("returns a MinAPIVersionRequirement as the first requirement", func() {
					actualRequirements, err := cmd.Requirements(factory, flagContext)
					Expect(err).NotTo(HaveOccurred())

					expectedVersion, err := semver.Make("2.36.0")
					Expect(err).NotTo(HaveOccurred())

					Expect(factory.NewMinAPIVersionRequirementCallCount()).To(Equal(1))
					feature, requiredVersion := factory.NewMinAPIVersionRequirementArgsForCall(0)
					Expect(feature).To(Equal("Option '--path'"))
					Expect(requiredVersion).To(Equal(expectedVersion))
					Expect(actualRequirements[0]).To(Equal(minAPIVersionRequirement))
				})
			})

			Context("when a path is not passed", func() {
				BeforeEach(func() {
					flagContext.Parse("domain-name")
				})

				It("does not return a MinAPIVersionRequirement", func() {
					actualRequirements, err := cmd.Requirements(factory, flagContext)
					Expect(err).NotTo(HaveOccurred())
					Expect(factory.NewMinAPIVersionRequirementCallCount()).To(Equal(0))
					Expect(actualRequirements).NotTo(ContainElement(minAPIVersionRequirement))
				})
			})
		})
	})

	Describe("Execute", func() {
		var err error

		BeforeEach(func() {
			err := flagContext.Parse("host-name", "domain-name")
			Expect(err).NotTo(HaveOccurred())
			cmd.Requirements(factory, flagContext)
			configRepo.SetOrganizationFields(models.OrganizationFields{
				GUID: "fake-org-guid",
				Name: "fake-org-name",
			})
		})

		JustBeforeEach(func() {
			err = cmd.Execute(flagContext)
		})

		It("tells the user that it is checking for the route", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Checking for route"},
			))
		})

		It("tries to find the domain", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(domainRepo.FindByNameInOrgCallCount()).To(Equal(1))

			domainName, orgGUID := domainRepo.FindByNameInOrgArgsForCall(0)
			Expect(domainName).To(Equal("domain-name"))
			Expect(orgGUID).To(Equal("fake-org-guid"))
		})

		Context("when it finds the domain successfully", func() {
			var actualDomain models.DomainFields

			BeforeEach(func() {
				actualDomain = models.DomainFields{
					GUID: "domain-guid",
					Name: "domain-name",
				}
				domainRepo.FindByNameInOrgReturns(actualDomain, nil)
			})

			It("checks if the route exists", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(routeRepo.CheckIfExistsCallCount()).To(Equal(1))
				hostName, domain, path := routeRepo.CheckIfExistsArgsForCall(0)
				Expect(hostName).To(Equal("host-name"))
				Expect(actualDomain).To(Equal(domain))
				Expect(path).To(Equal(""))
			})

			Context("when a path is passed", func() {
				BeforeEach(func() {
					flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
					err := flagContext.Parse("hostname", "domain-name", "--path", "the-path")
					Expect(err).NotTo(HaveOccurred())
					cmd.Requirements(factory, flagContext)
				})

				It("checks if the route exists", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(routeRepo.CheckIfExistsCallCount()).To(Equal(1))
					_, _, path := routeRepo.CheckIfExistsArgsForCall(0)
					Expect(path).To(Equal("the-path"))
				})

				Context("when finding the route succeeds and the route exists", func() {
					BeforeEach(func() {
						routeRepo.CheckIfExistsReturns(true, nil)
					})

					It("tells the user the route exists", func() {
						Expect(err).NotTo(HaveOccurred())
						Expect(ui.Outputs()).To(ContainSubstrings([]string{"Route hostname.domain-name/the-path does exist"}))
					})
				})

				Context("when finding the route succeeds and the route does not exist", func() {
					BeforeEach(func() {
						routeRepo.CheckIfExistsReturns(false, nil)
					})

					It("tells the user the route exists", func() {
						Expect(err).NotTo(HaveOccurred())
						Expect(ui.Outputs()).To(ContainSubstrings([]string{"Route hostname.domain-name/the-path does not exist"}))
					})
				})
			})

			Context("when finding the route succeeds and the route exists", func() {
				BeforeEach(func() {
					routeRepo.CheckIfExistsReturns(true, nil)
				})

				It("tells the user OK", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(ui.Outputs()).To(ContainSubstrings([]string{"OK"}))
				})

				It("tells the user the route exists", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(ui.Outputs()).To(ContainSubstrings([]string{"Route", "does exist"}))
				})
			})

			Context("when finding the route succeeds and the route does not exist", func() {
				BeforeEach(func() {
					routeRepo.CheckIfExistsReturns(false, nil)
				})

				It("tells the user OK", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(ui.Outputs()).To(ContainSubstrings([]string{"OK"}))
				})

				It("tells the user the route does not exist", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(ui.Outputs()).To(ContainSubstrings([]string{"Route", "does not exist"}))
				})
			})

			Context("when finding the route results in an error", func() {
				BeforeEach(func() {
					routeRepo.CheckIfExistsReturns(false, errors.New("check-if-exists-err"))
				})

				It("fails with error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("check-if-exists-err"))
				})
			})
		})

		Context("when finding the domain results in an error", func() {
			BeforeEach(func() {
				domainRepo.FindByNameInOrgReturns(models.DomainFields{}, errors.New("find-by-name-in-org-err"))
			})

			It("fails with error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("find-by-name-in-org-err"))
			})
		})
	})
})

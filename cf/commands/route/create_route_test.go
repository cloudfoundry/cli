package route_test

import (
	"errors"

	"github.com/blang/semver"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/commands/route"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/flags"

	fakeapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/requirements"
	fakerequirements "github.com/cloudfoundry/cli/cf/requirements/fakes"

	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type passingRequirement struct {
	Name string
}

func (r passingRequirement) Execute() bool {
	return true
}

var _ = Describe("CreateRoute", func() {
	var (
		ui         *testterm.FakeUI
		routeRepo  *fakeapi.FakeRouteRepository
		configRepo core_config.Repository

		cmd         route.CreateRoute
		deps        command_registry.Dependency
		factory     *fakerequirements.FakeFactory
		flagContext flags.FlagContext

		spaceRequirement         *fakerequirements.FakeSpaceRequirement
		domainRequirement        *fakerequirements.FakeDomainRequirement
		minAPIVersionRequirement requirements.Requirement
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		routeRepo = &fakeapi.FakeRouteRepository{}
		repoLocator := deps.RepoLocator.SetRouteRepository(routeRepo)

		deps = command_registry.Dependency{
			Ui:          ui,
			Config:      configRepo,
			RepoLocator: repoLocator,
		}

		cmd = route.CreateRoute{}
		cmd.SetDependency(deps, false)

		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)

		factory = &fakerequirements.FakeFactory{}

		spaceRequirement = &fakerequirements.FakeSpaceRequirement{}
		space := models.Space{}
		space.Guid = "space-guid"
		space.Name = "space-name"
		spaceRequirement.GetSpaceReturns(space)
		factory.NewSpaceRequirementReturns(spaceRequirement)

		domainRequirement = &fakerequirements.FakeDomainRequirement{}
		domainRequirement.GetDomainReturns(models.DomainFields{
			Guid: "domain-guid",
			Name: "domain-name",
		})
		factory.NewDomainRequirementReturns(domainRequirement)

		minAPIVersionRequirement = &passingRequirement{}
		factory.NewMinAPIVersionRequirementReturns(minAPIVersionRequirement)
	})

	Describe("Requirements", func() {
		Context("when not provided exactly two args", func() {
			BeforeEach(func() {
				flagContext.Parse("space-name")
			})

			It("fails with usage", func() {
				Expect(func() { cmd.Requirements(factory, flagContext) }).To(Panic())
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Incorrect Usage. Requires SPACE and DOMAIN as arguments"},
					[]string{"NAME"},
					[]string{"USAGE"},
				))
			})
		})

		Context("when provided exactly two args", func() {
			BeforeEach(func() {
				flagContext.Parse("space-name", "domain-name")
			})

			It("returns a SpaceRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(factory.NewSpaceRequirementCallCount()).To(Equal(1))
				Expect(factory.NewSpaceRequirementArgsForCall(0)).To(Equal("space-name"))

				Expect(actualRequirements).To(ContainElement(spaceRequirement))
			})

			It("returns a DomainRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(factory.NewDomainRequirementCallCount()).To(Equal(1))
				Expect(factory.NewDomainRequirementArgsForCall(0)).To(Equal("domain-name"))

				Expect(actualRequirements).To(ContainElement(domainRequirement))
			})
		})

		Context("when the --path option is given", func() {
			BeforeEach(func() {
				flagContext.Parse("space-name", "domain-name", "--path", "path")
			})

			It("returns a MinAPIVersionRequirement", func() {
				expectedVersion, err := semver.Make("2.36.0")
				Expect(err).NotTo(HaveOccurred())

				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())

				Expect(factory.NewMinAPIVersionRequirementCallCount()).To(Equal(1))
				feature, requiredVersion := factory.NewMinAPIVersionRequirementArgsForCall(0)
				Expect(feature).To(Equal("Option '--path'"))
				Expect(requiredVersion).To(Equal(expectedVersion))
				Expect(actualRequirements).To(ContainElement(minAPIVersionRequirement))
			})
		})

		Context("when the --path option is not given", func() {
			BeforeEach(func() {
				flagContext.Parse("space-name", "domain-name")
			})

			It("does not return a MinAPIVersionRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(actualRequirements).NotTo(ContainElement(minAPIVersionRequirement))
			})
		})
	})

	Describe("Execute", func() {
		BeforeEach(func() {
			err := flagContext.Parse("space-name", "domain-name", "-n", "hostname", "--path", "path")
			Expect(err).NotTo(HaveOccurred())
			_, err = cmd.Requirements(factory, flagContext)
			Expect(err).NotTo(HaveOccurred())
		})

		It("attempts to create a route in the space", func() {
			cmd.Execute(flagContext)

			Expect(routeRepo.CreateInSpaceCallCount()).To(Equal(1))
			hostname, path, domain, space := routeRepo.CreateInSpaceArgsForCall(0)
			Expect(hostname).To(Equal("hostname"))
			Expect(path).To(Equal("path"))
			Expect(domain).To(Equal("domain-guid"))
			Expect(space).To(Equal("space-guid"))
		})

		Context("when creating the route fails", func() {
			BeforeEach(func() {
				routeRepo.CreateInSpaceReturns(models.Route{}, errors.New("create-error"))
			})

			It("attempts to find the route", func() {
				Expect(func() { cmd.Execute(flagContext) }).To(Panic())
				Expect(routeRepo.FindCallCount()).To(Equal(1))
			})

			Context("when finding the route fails", func() {
				BeforeEach(func() {
					routeRepo.FindReturns(models.Route{}, errors.New("find-error"))
				})

				It("fails with the original error", func() {
					Expect(func() { cmd.Execute(flagContext) }).To(Panic())
					Expect(ui.Outputs).To(ContainSubstrings([]string{"FAILED"}, []string{"create-error"}))
				})
			})

			Context("when a route with the same space guid, but different domain guid is found", func() {
				It("fails with the original error", func() {
					Expect(func() { cmd.Execute(flagContext) }).To(Panic())
					Expect(ui.Outputs).To(ContainSubstrings([]string{"FAILED"}, []string{"create-error"}))
				})
			})

			Context("when a route with the same domain guid, but different space guid is found", func() {
				It("fails with the original error", func() {
					Expect(func() { cmd.Execute(flagContext) }).To(Panic())
					Expect(ui.Outputs).To(ContainSubstrings([]string{"FAILED"}, []string{"create-error"}))
				})
			})

			Context("when a route with the same domain and space guid is found", func() {
				BeforeEach(func() {
					routeRepo.FindReturns(models.Route{
						Host: "hostname",
						Path: "path",
						Domain: models.DomainFields{
							Guid: "domain-guid",
							Name: "domain-name",
						},
						Space: models.SpaceFields{
							Guid: "space-guid",
						},
					}, nil)
				})

				It("prints a message", func() {
					cmd.Execute(flagContext)
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"OK"},
						[]string{"Route hostname.domain-name/path already exists"},
					))
				})
			})
		})
	})

	Describe("CreateRoute", func() {
		var domainFields models.DomainFields
		var spaceFields models.SpaceFields

		BeforeEach(func() {
			domainFields = models.DomainFields{
				Guid: "domain-guid",
				Name: "domain-name",
			}
			spaceFields = models.SpaceFields{
				Guid: "space-guid",
				Name: "space-name",
			}
		})

		It("attempts to create a route in the space", func() {
			cmd.CreateRoute("hostname", "path", domainFields, spaceFields)

			Expect(routeRepo.CreateInSpaceCallCount()).To(Equal(1))
			hostname, path, domain, space := routeRepo.CreateInSpaceArgsForCall(0)
			Expect(hostname).To(Equal("hostname"))
			Expect(path).To(Equal("path"))
			Expect(domain).To(Equal(domainFields.Guid))
			Expect(space).To(Equal(spaceFields.Guid))
		})

		Context("when creating the route fails", func() {
			BeforeEach(func() {
				routeRepo.CreateInSpaceReturns(models.Route{}, errors.New("create-error"))
			})

			It("attempts to find the route", func() {
				cmd.CreateRoute("hostname", "path", domainFields, spaceFields)
				Expect(routeRepo.FindCallCount()).To(Equal(1))
			})

			Context("when finding the route fails", func() {
				BeforeEach(func() {
					routeRepo.FindReturns(models.Route{}, errors.New("find-error"))
				})

				It("returns the original error", func() {
					_, err := cmd.CreateRoute("hostname", "path", domainFields, spaceFields)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("create-error"))
				})
			})

			Context("when a route with the same space guid, but different domain guid is found", func() {
				It("returns the original error", func() {
					_, err := cmd.CreateRoute("hostname", "path", domainFields, spaceFields)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("create-error"))
				})
			})

			Context("when a route with the same domain guid, but different space guid is found", func() {
				It("returns the original error", func() {
					_, err := cmd.CreateRoute("hostname", "path", domainFields, spaceFields)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("create-error"))
				})
			})

			Context("when a route with the same domain and space guid is found", func() {
				BeforeEach(func() {
					routeRepo.FindReturns(models.Route{
						Host: "hostname",
						Path: "path",
						Domain: models.DomainFields{
							Guid: "domain-guid",
							Name: "domain-name",
						},
						Space: models.SpaceFields{
							Guid: "space-guid",
						},
					}, nil)
				})

				It("prints a message that it already exists", func() {
					cmd.CreateRoute("hostname", "path", domainFields, spaceFields)
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"OK"},
						[]string{"Route hostname.domain-name/path already exists"}))
				})
			})
		})

		Context("when creating the route succeeds", func() {
			BeforeEach(func() {
				routeRepo.CreateInSpaceReturns(models.Route{}, nil)
			})

			It("prints a success message", func() {
				cmd.CreateRoute("hostname", "path", domainFields, spaceFields)
				Expect(ui.Outputs).To(ContainSubstrings([]string{"OK"}))
			})
		})
	})
})

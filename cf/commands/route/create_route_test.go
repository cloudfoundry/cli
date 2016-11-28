package route_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/commands/route"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	"github.com/blang/semver"

	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/requirements"

	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CreateRoute", func() {
	var (
		ui         *testterm.FakeUI
		routeRepo  *apifakes.FakeRouteRepository
		configRepo coreconfig.Repository

		cmd         commandregistry.Command
		deps        commandregistry.Dependency
		factory     *requirementsfakes.FakeFactory
		flagContext flags.FlagContext

		spaceRequirement         *requirementsfakes.FakeSpaceRequirement
		domainRequirement        *requirementsfakes.FakeDomainRequirement
		minAPIVersionRequirement requirements.Requirement
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		routeRepo = new(apifakes.FakeRouteRepository)
		repoLocator := deps.RepoLocator.SetRouteRepository(routeRepo)

		deps = commandregistry.Dependency{
			UI:          ui,
			Config:      configRepo,
			RepoLocator: repoLocator,
		}

		cmd = &route.CreateRoute{}
		cmd.SetDependency(deps, false)

		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)

		factory = new(requirementsfakes.FakeFactory)

		spaceRequirement = new(requirementsfakes.FakeSpaceRequirement)
		space := models.Space{}
		space.GUID = "space-guid"
		space.Name = "space-name"
		spaceRequirement.GetSpaceReturns(space)
		factory.NewSpaceRequirementReturns(spaceRequirement)

		domainRequirement = new(requirementsfakes.FakeDomainRequirement)
		domainRequirement.GetDomainReturns(models.DomainFields{
			GUID: "domain-guid",
			Name: "domain-name",
		})
		factory.NewDomainRequirementReturns(domainRequirement)

		minAPIVersionRequirement = &passingRequirement{}
		factory.NewMinAPIVersionRequirementReturns(minAPIVersionRequirement)
	})

	Describe("Requirements", func() {
		Context("when not provided exactly two args", func() {
			BeforeEach(func() {
				err := flagContext.Parse("space-name")
				Expect(err).NotTo(HaveOccurred())
			})

			It("fails with usage", func() {
				_, err := cmd.Requirements(factory, flagContext)
				Expect(err).To(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Incorrect Usage. Requires SPACE and DOMAIN as arguments"},
					[]string{"NAME"},
					[]string{"USAGE"},
				))
			})
		})

		Context("when provided exactly two args", func() {
			BeforeEach(func() {
				err := flagContext.Parse("space-name", "domain-name")
				Expect(err).NotTo(HaveOccurred())
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
				err := flagContext.Parse("space-name", "domain-name", "--path", "path")
				Expect(err).NotTo(HaveOccurred())
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

		Context("when the --port option is given", func() {
			BeforeEach(func() {
				err := flagContext.Parse("space-name", "domain-name", "--port", "9090")
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns a MinAPIVersionRequirement", func() {
				expectedVersion, err := semver.Make("2.53.0")
				Expect(err).NotTo(HaveOccurred())

				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())

				Expect(factory.NewMinAPIVersionRequirementCallCount()).To(Equal(1))
				feature, requiredVersion := factory.NewMinAPIVersionRequirementArgsForCall(0)
				Expect(feature).To(Equal("Option '--port'"))
				Expect(requiredVersion).To(Equal(expectedVersion))
				Expect(actualRequirements).To(ContainElement(minAPIVersionRequirement))
			})
		})

		Context("when the --random-port option is given", func() {
			BeforeEach(func() {
				err := flagContext.Parse("space-name", "domain-name", "--random-port")
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns a MinAPIVersionRequirement", func() {
				expectedVersion, err := semver.Make("2.53.0")
				Expect(err).NotTo(HaveOccurred())

				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())

				Expect(factory.NewMinAPIVersionRequirementCallCount()).To(Equal(1))
				feature, requiredVersion := factory.NewMinAPIVersionRequirementArgsForCall(0)
				Expect(feature).To(Equal("Option '--random-port'"))
				Expect(requiredVersion).To(Equal(expectedVersion))
				Expect(actualRequirements).To(ContainElement(minAPIVersionRequirement))
			})
		})

		Context("when the --path option is not given", func() {
			BeforeEach(func() {
				err := flagContext.Parse("space-name", "domain-name")
				Expect(err).NotTo(HaveOccurred())
			})

			It("does not return a MinAPIVersionRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(actualRequirements).NotTo(ContainElement(minAPIVersionRequirement))
			})
		})

		Context("when both --port and --hostname are given", func() {
			BeforeEach(func() {
				err := flagContext.Parse("space-name", "domain-name", "--port", "9090", "--hostname", "host")
				Expect(err).NotTo(HaveOccurred())
			})

			It("fails with error", func() {
				_, err := cmd.Requirements(factory, flagContext)
				Expect(err).To(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Cannot specify port together with hostname and/or path."},
				))
			})
		})

		Context("when both --port and --path are given", func() {
			BeforeEach(func() {
				err := flagContext.Parse("space-name", "domain-name", "--port", "9090", "--path", "path")
				Expect(err).NotTo(HaveOccurred())
			})

			It("fails with error", func() {
				_, err := cmd.Requirements(factory, flagContext)
				Expect(err).To(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Cannot specify port together with hostname and/or path."},
				))
			})
		})

		Context("when both --port and --random-port are given", func() {
			BeforeEach(func() {
				err := flagContext.Parse("space-name", "domain-name", "--port", "9090", "--random-port")
				Expect(err).NotTo(HaveOccurred())
			})

			It("fails with error", func() {
				_, err := cmd.Requirements(factory, flagContext)
				Expect(err).To(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Cannot specify random-port together with port, hostname and/or path."},
				))
			})
		})

		Context("when both --random-port and --hostname are given", func() {
			BeforeEach(func() {
				err := flagContext.Parse("space-name", "domain-name", "--hostname", "host", "--random-port")
				Expect(err).NotTo(HaveOccurred())
			})

			It("fails with error", func() {
				_, err := cmd.Requirements(factory, flagContext)
				Expect(err).To(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Cannot specify random-port together with port, hostname and/or path."},
				))
			})
		})

		Context("when --random-port and --path are given", func() {
			BeforeEach(func() {
				err := flagContext.Parse("space-name", "domain-name", "--path", "path", "--random-port")
				Expect(err).NotTo(HaveOccurred())
			})

			It("fails with error", func() {
				_, err := cmd.Requirements(factory, flagContext)
				Expect(err).To(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Cannot specify random-port together with port, hostname and/or path."},
				))
			})
		})
	})

	Describe("Execute", func() {
		var err error

		BeforeEach(func() {
			err := flagContext.Parse("space-name", "domain-name")
			Expect(err).NotTo(HaveOccurred())
			cmd.Requirements(factory, flagContext)
		})

		JustBeforeEach(func() {
			err = cmd.Execute(flagContext)
		})

		It("attempts to create a route in the space", func() {
			Expect(err).NotTo(HaveOccurred())

			Expect(routeRepo.CreateInSpaceCallCount()).To(Equal(1))
			hostname, path, domain, space, port, randomPort := routeRepo.CreateInSpaceArgsForCall(0)
			Expect(hostname).To(Equal(""))
			Expect(path).To(Equal(""))
			Expect(domain).To(Equal("domain-guid"))
			Expect(space).To(Equal("space-guid"))
			Expect(port).To(Equal(0))
			Expect(randomPort).To(BeFalse())
		})

		Context("when the --path option is given", func() {
			BeforeEach(func() {
				err := flagContext.Parse("space-name", "domain-name", "--path", "some-path")
				Expect(err).NotTo(HaveOccurred())
			})

			It("tries to create a route with the path", func() {
				Expect(err).NotTo(HaveOccurred())

				Expect(routeRepo.CreateInSpaceCallCount()).To(Equal(1))
				_, path, _, _, _, _ := routeRepo.CreateInSpaceArgsForCall(0)
				Expect(path).To(Equal("some-path"))
			})
		})

		Context("when the --random-port option is given", func() {
			BeforeEach(func() {
				err := flagContext.Parse("space-name", "domain-name", "--random-port")
				Expect(err).NotTo(HaveOccurred())
			})

			It("tries to create a route with a random port", func() {
				Expect(err).NotTo(HaveOccurred())

				Expect(routeRepo.CreateInSpaceCallCount()).To(Equal(1))
				_, _, _, _, _, randomPort := routeRepo.CreateInSpaceArgsForCall(0)
				Expect(randomPort).To(BeTrue())
			})
		})

		Context("when the --port option is given", func() {
			BeforeEach(func() {
				err := flagContext.Parse("space-name", "domain-name", "--port", "9090")
				Expect(err).NotTo(HaveOccurred())
			})

			It("tries to create a route with the port", func() {
				Expect(err).NotTo(HaveOccurred())

				Expect(routeRepo.CreateInSpaceCallCount()).To(Equal(1))
				_, _, _, _, port, _ := routeRepo.CreateInSpaceArgsForCall(0)
				Expect(port).To(Equal(9090))
			})
		})

		Context("when the --hostname option is given", func() {
			BeforeEach(func() {
				err := flagContext.Parse("space-name", "domain-name", "--hostname", "host")
				Expect(err).NotTo(HaveOccurred())
			})

			It("tries to create a route with the hostname", func() {
				Expect(err).NotTo(HaveOccurred())

				Expect(routeRepo.CreateInSpaceCallCount()).To(Equal(1))
				host, _, _, _, _, _ := routeRepo.CreateInSpaceArgsForCall(0)
				Expect(host).To(Equal("host"))
			})
		})

		Context("when creating the route fails", func() {
			BeforeEach(func() {
				routeRepo.CreateInSpaceReturns(models.Route{}, errors.New("create-error"))

				err := flagContext.Parse("space-name", "domain-name", "--port", "9090", "--hostname", "hostname", "--path", "/path")
				Expect(err).NotTo(HaveOccurred())
			})

			It("attempts to find the route", func() {
				Expect(err).To(HaveOccurred())
				Expect(routeRepo.FindCallCount()).To(Equal(1))

				host, domain, path, port := routeRepo.FindArgsForCall(0)
				Expect(host).To(Equal("hostname"))
				Expect(domain.Name).To(Equal("domain-name"))
				Expect(path).To(Equal("/path"))
				Expect(port).To(Equal(9090))
			})

			Context("when finding the route fails", func() {
				BeforeEach(func() {
					routeRepo.FindReturns(models.Route{}, errors.New("find-error"))
				})

				It("fails with the original error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("create-error"))
				})
			})

			Context("when a route with the same space guid, but different domain guid is found", func() {
				It("fails with the original error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("create-error"))
				})
			})

			Context("when a route with the same domain guid, but different space guid is found", func() {
				It("fails with the original error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("create-error"))
				})
			})

			Context("when a route with the same domain and space guid is found", func() {
				BeforeEach(func() {
					routeRepo.FindReturns(models.Route{
						Domain: models.DomainFields{
							GUID: "domain-guid",
							Name: "domain-name",
						},
						Space: models.SpaceFields{
							GUID: "space-guid",
						},
					}, nil)
				})

				It("prints a message", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"OK"},
						[]string{"Route domain-name already exists"},
					))
				})
			})
		})
	})

	Describe("CreateRoute", func() {
		var domainFields models.DomainFields
		var spaceFields models.SpaceFields
		var rc route.Creator

		BeforeEach(func() {
			domainFields = models.DomainFields{
				GUID: "domain-guid",
				Name: "domain-name",
			}
			spaceFields = models.SpaceFields{
				GUID: "space-guid",
				Name: "space-name",
			}

			var ok bool
			rc, ok = cmd.(route.Creator)
			Expect(ok).To(BeTrue())
		})

		It("attempts to create a route in the space", func() {
			rc.CreateRoute("hostname", "path", 9090, true, domainFields, spaceFields)

			Expect(routeRepo.CreateInSpaceCallCount()).To(Equal(1))
			hostname, path, domain, space, port, randomPort := routeRepo.CreateInSpaceArgsForCall(0)
			Expect(hostname).To(Equal("hostname"))
			Expect(path).To(Equal("path"))
			Expect(domain).To(Equal(domainFields.GUID))
			Expect(space).To(Equal(spaceFields.GUID))
			Expect(port).To(Equal(9090))
			Expect(randomPort).To(BeTrue())
		})

		Context("when creating the route fails", func() {
			BeforeEach(func() {
				routeRepo.CreateInSpaceReturns(models.Route{}, errors.New("create-error"))
			})

			It("attempts to find the route", func() {
				rc.CreateRoute("hostname", "path", 0, false, domainFields, spaceFields)
				Expect(routeRepo.FindCallCount()).To(Equal(1))
			})

			Context("when finding the route fails", func() {
				BeforeEach(func() {
					routeRepo.FindReturns(models.Route{}, errors.New("find-error"))
				})

				It("returns the original error", func() {
					_, err := rc.CreateRoute("hostname", "path", 0, false, domainFields, spaceFields)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("create-error"))
				})
			})

			Context("when a route with the same space guid, but different domain guid is found", func() {
				It("returns the original error", func() {
					_, err := rc.CreateRoute("hostname", "path", 0, false, domainFields, spaceFields)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("create-error"))
				})
			})

			Context("when a route with the same domain guid, but different space guid is found", func() {
				It("returns the original error", func() {
					_, err := rc.CreateRoute("hostname", "path", 0, false, domainFields, spaceFields)
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
							GUID: "domain-guid",
							Name: "domain-name",
						},
						Space: models.SpaceFields{
							GUID: "space-guid",
						},
					}, nil)
				})

				It("prints a message that it already exists", func() {
					rc.CreateRoute("hostname", "path", 0, false, domainFields, spaceFields)
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"OK"},
						[]string{"Route hostname.domain-name/path already exists"}))
				})
			})
		})

		Context("when creating the route succeeds", func() {
			var route models.Route

			JustBeforeEach(func() {
				routeRepo.CreateInSpaceReturns(route, nil)
			})

			It("prints a success message", func() {
				rc.CreateRoute("hostname", "path", 0, false, domainFields, spaceFields)
				Expect(ui.Outputs()).To(ContainSubstrings([]string{"OK"}))
			})

			Context("when --random-port is specified", func() {
				BeforeEach(func() {
					route = models.Route{
						Host:   "some-host",
						Domain: domainFields,
						Port:   9090,
					}
				})

				It("print a success message with created route", func() {
					rc.CreateRoute("hostname", "path", 0, true, domainFields, spaceFields)
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"OK"},
						[]string{"Route domain-name:9090 has been created"},
					))
				})
			})
		})
	})
})

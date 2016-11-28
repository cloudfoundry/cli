package route_test

import (
	"strings"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/commands/route"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
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

var _ = Describe("DeleteRoute", func() {
	var (
		ui         *testterm.FakeUI
		configRepo coreconfig.Repository
		routeRepo  *apifakes.FakeRouteRepository

		cmd         commandregistry.Command
		deps        commandregistry.Dependency
		factory     *requirementsfakes.FakeFactory
		flagContext flags.FlagContext

		loginRequirement         requirements.Requirement
		domainRequirement        *requirementsfakes.FakeDomainRequirement
		minAPIVersionRequirement requirements.Requirement

		fakeDomain models.DomainFields
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

		cmd = &route.DeleteRoute{}
		cmd.SetDependency(deps, false)

		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)

		factory = new(requirementsfakes.FakeFactory)

		loginRequirement = &passingRequirement{Name: "login-requirement"}
		factory.NewLoginRequirementReturns(loginRequirement)

		domainRequirement = new(requirementsfakes.FakeDomainRequirement)
		factory.NewDomainRequirementReturns(domainRequirement)

		fakeDomain = models.DomainFields{
			GUID: "fake-domain-guid",
			Name: "fake-domain-name",
		}
		domainRequirement.GetDomainReturns(fakeDomain)

		minAPIVersionRequirement = &passingRequirement{Name: "min-api-version-requirement"}
		factory.NewMinAPIVersionRequirementReturns(minAPIVersionRequirement)
	})

	Describe("Help text", func() {
		var usage []string

		BeforeEach(func() {
			dr := &route.DeleteRoute{}
			up := commandregistry.CLICommandUsagePresenter(dr)
			usage = strings.Split(up.Usage(), "\n")
		})

		It("has a HTTP route usage", func() {
			Expect(usage).To(ContainElement("   Delete an HTTP route:"))
			Expect(usage).To(ContainElement("      cf delete-route DOMAIN [--hostname HOSTNAME] [--path PATH] [-f]"))
		})

		It("has a TCP route usage", func() {
			Expect(usage).To(ContainElement("   Delete a TCP route:"))
			Expect(usage).To(ContainElement("      cf delete-route DOMAIN --port PORT [-f]"))
		})

		It("has a TCP route example", func() {
			Expect(usage).To(ContainElement("   cf delete-route example.com --port 50000                 # example.com:50000"))
		})

		It("has a TCP option", func() {
			Expect(usage).To(ContainElement("   --port              Port used to identify the TCP route"))
		})
	})

	Describe("Requirements", func() {
		Context("when not provided exactly one arg", func() {
			BeforeEach(func() {
				flagContext.Parse("app-name", "extra-arg")
			})

			It("fails with usage", func() {
				_, err := cmd.Requirements(factory, flagContext)
				Expect(err).To(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Incorrect Usage. Requires an argument"},
				))
			})
		})

		Context("when provided exactly one arg", func() {
			BeforeEach(func() {
				flagContext.Parse("domain-name")
			})

			It("returns a LoginRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(factory.NewLoginRequirementCallCount()).To(Equal(1))
				Expect(actualRequirements).To(ContainElement(loginRequirement))
			})

			It("returns a DomainRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(factory.NewDomainRequirementCallCount()).To(Equal(1))

				Expect(factory.NewDomainRequirementArgsForCall(0)).To(Equal("domain-name"))
				Expect(actualRequirements).To(ContainElement(domainRequirement))
			})

			Context("when a path is passed", func() {
				BeforeEach(func() {
					flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
					flagContext.Parse("domain-name", "--path", "the-path")
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

			Describe("deleting a tcp route", func() {
				Context("when passing port with a hostname", func() {
					BeforeEach(func() {
						flagContext.Parse("example.com", "--port", "8080", "--hostname", "something-else")
					})

					It("fails", func() {
						_, err := cmd.Requirements(factory, flagContext)
						Expect(err).To(HaveOccurred())
						Expect(ui.Outputs()).To(ContainSubstrings(
							[]string{"FAILED"},
							[]string{"Cannot specify port together with hostname and/or path."},
						))
					})
				})

				Context("when passing port with a path", func() {
					BeforeEach(func() {
						flagContext.Parse("example.com", "--port", "8080", "--path", "something-else")
					})

					It("fails", func() {
						_, err := cmd.Requirements(factory, flagContext)
						Expect(err).To(HaveOccurred())
						Expect(ui.Outputs()).To(ContainSubstrings(
							[]string{"FAILED"},
							[]string{"Cannot specify port together with hostname and/or path."},
						))
					})
				})

				Context("when a port is passed", func() {
					BeforeEach(func() {
						flagContext.Parse("example.com", "--port", "8080")
					})

					It("returns a MinAPIVersionRequirement as the first requirement", func() {
						actualRequirements, err := cmd.Requirements(factory, flagContext)
						Expect(err).NotTo(HaveOccurred())

						expectedVersion, err := semver.Make("2.53.0")
						Expect(err).NotTo(HaveOccurred())

						Expect(factory.NewMinAPIVersionRequirementCallCount()).To(Equal(1))
						feature, requiredVersion := factory.NewMinAPIVersionRequirementArgsForCall(0)
						Expect(feature).To(Equal("Option '--port'"))
						Expect(requiredVersion).To(Equal(expectedVersion))
						Expect(actualRequirements[0]).To(Equal(minAPIVersionRequirement))
					})
				})
			})
		})
	})

	Describe("Execute", func() {
		var err error

		BeforeEach(func() {
			err := flagContext.Parse("domain-name")
			Expect(err).NotTo(HaveOccurred())
			cmd.Requirements(factory, flagContext)
		})

		JustBeforeEach(func() {
			err = cmd.Execute(flagContext)
		})

		Context("when passed -n flag", func() {
			BeforeEach(func() {
				ui.Inputs = []string{"n"}
			})

			It("asks the user if they would like to proceed", func() {
				Expect(err).NotTo(HaveOccurred())
				Eventually(func() []string { return ui.Prompts }).Should(ContainSubstrings(
					[]string{"Really delete the route"},
				))
			})
		})

		Context("when the response is to proceed", func() {
			BeforeEach(func() {
				ui.Inputs = []string{"y"}
			})

			It("tries to find the route", func() {
				Expect(err).NotTo(HaveOccurred())
				Eventually(routeRepo.FindCallCount).Should(Equal(1))
				host, domain, path, port := routeRepo.FindArgsForCall(0)
				Expect(host).To(Equal(""))
				Expect(path).To(Equal(""))
				Expect(port).To(Equal(0))
				Expect(domain).To(Equal(fakeDomain))
			})

			Context("when a path is passed", func() {
				BeforeEach(func() {
					err := flagContext.Parse("domain-name", "-f", "--path", "the-path")
					Expect(err).NotTo(HaveOccurred())
					cmd.Requirements(factory, flagContext)
				})

				It("tries to find the route with the path", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(routeRepo.FindCallCount()).To(Equal(1))
					_, _, path, _ := routeRepo.FindArgsForCall(0)
					Expect(path).To(Equal("the-path"))
				})
			})

			Context("when a port is passed", func() {
				BeforeEach(func() {
					err := flagContext.Parse("domain-name", "-f", "--port", "60000")
					Expect(err).NotTo(HaveOccurred())
					cmd.Requirements(factory, flagContext)
				})

				It("tries to find the route with the port", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(routeRepo.FindCallCount()).To(Equal(1))
					_, _, _, port := routeRepo.FindArgsForCall(0)
					Expect(port).To(Equal(60000))
				})
			})

			Context("when the route can be found", func() {
				BeforeEach(func() {
					routeRepo.FindReturns(models.Route{
						GUID: "route-guid",
					}, nil)
				})

				It("tries to delete the route", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(routeRepo.DeleteCallCount()).To(Equal(1))
					Expect(routeRepo.DeleteArgsForCall(0)).To(Equal("route-guid"))
				})

				Context("when deleting the route succeeds", func() {
					BeforeEach(func() {
						routeRepo.DeleteReturns(nil)
					})

					It("tells the user that it succeeded", func() {
						Expect(err).NotTo(HaveOccurred())
						Expect(ui.Outputs()).To(ContainSubstrings(
							[]string{"OK"},
						))
					})
				})

				Context("when deleting the route fails", func() {
					BeforeEach(func() {
						routeRepo.DeleteReturns(errors.New("delete-err"))
					})

					It("fails with error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(Equal("delete-err"))
					})
				})
			})

			Context("when there is an error finding the route", func() {
				BeforeEach(func() {
					routeRepo.FindReturns(models.Route{}, errors.New("find-err"))
				})

				It("fails with error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("find-err"))
				})

				It("does not try to delete the route", func() {
					Expect(err).To(HaveOccurred())
					Expect(routeRepo.DeleteCallCount()).To(BeZero())
				})
			})

			Context("when there is a ModelNotFoundError when finding the route", func() {
				BeforeEach(func() {
					routeRepo.FindReturns(models.Route{}, errors.NewModelNotFoundError("model-type", "model-name"))
				})

				It("tells the user that it could not delete the route", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Unable to delete, route", "does not exist"},
					))
				})

				It("does not try to delete the route", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(routeRepo.DeleteCallCount()).To(BeZero())
				})
			})

		})

		Context("when the response is not to proceed", func() {
			BeforeEach(func() {
				ui.Inputs = []string{"n"}
			})

			It("does not try to delete the route", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(routeRepo.DeleteCallCount()).To(Equal(0))
			})
		})

		Context("when force is set", func() {
			BeforeEach(func() {
				err := flagContext.Parse("domain-name", "-f")
				Expect(err).NotTo(HaveOccurred())
			})

			It("does not ask the user if they would like to proceed", func() {
				Expect(err).NotTo(HaveOccurred())
				Consistently(func() []string { return ui.Prompts }).ShouldNot(ContainSubstrings(
					[]string{"Really delete the route"},
				))
			})
		})
	})
})

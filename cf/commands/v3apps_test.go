package commands_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/commands"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	"code.cloudfoundry.org/cli/cf/terminal"

	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/v3/models"
	"code.cloudfoundry.org/cli/cf/v3/repository/repositoryfakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("V3Apps", func() {
	var (
		ui         *testterm.FakeUI
		routeRepo  *apifakes.FakeRouteRepository
		configRepo coreconfig.Repository
		repository *repositoryfakes.FakeRepository

		cmd         commandregistry.Command
		deps        commandregistry.Dependency
		factory     *requirementsfakes.FakeFactory
		flagContext flags.FlagContext

		loginRequirement         requirements.Requirement
		targetedSpaceRequirement requirements.Requirement
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}

		routeRepo = new(apifakes.FakeRouteRepository)
		repoLocator := deps.RepoLocator.SetRouteRepository(routeRepo)
		repository = new(repositoryfakes.FakeRepository)
		repoLocator = repoLocator.SetV3Repository(repository)

		configRepo = testconfig.NewRepositoryWithDefaults()
		deps = commandregistry.Dependency{
			UI:          ui,
			Config:      configRepo,
			RepoLocator: repoLocator,
		}

		cmd = &commands.V3Apps{}
		cmd.SetDependency(deps, false)

		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)

		factory = new(requirementsfakes.FakeFactory)

		loginRequirement = &passingRequirement{}
		factory.NewLoginRequirementReturns(loginRequirement)

		targetedSpaceRequirement = &passingRequirement{}
		factory.NewTargetedSpaceRequirementReturns(targetedSpaceRequirement)
	})

	Describe("Requirements", func() {
		It("returns a LoginRequirement", func() {
			actualRequirements, err := cmd.Requirements(factory, flagContext)
			Expect(err).NotTo(HaveOccurred())
			Expect(factory.NewLoginRequirementCallCount()).To(Equal(1))

			Expect(actualRequirements).To(ContainElement(loginRequirement))
		})

		It("returns a TargetedSpaceRequirement", func() {
			actualRequirements, err := cmd.Requirements(factory, flagContext)
			Expect(err).NotTo(HaveOccurred())
			Expect(factory.NewTargetedSpaceRequirementCallCount()).To(Equal(1))

			Expect(actualRequirements).To(ContainElement(targetedSpaceRequirement))
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

	Describe("Execute", func() {
		var runCLIErr error

		BeforeEach(func() {
			cmd.Requirements(factory, flagContext)
			repository.GetProcessesReturns([]models.V3Process{{Type: "web"}, {Type: "web"}}, nil)
			repository.GetRoutesReturns([]models.V3Route{{}, {}}, nil)
		})

		JustBeforeEach(func() {
			runCLIErr = cmd.Execute(flagContext)
		})

		It("attemps to get applications for the targeted space", func() {
			Expect(runCLIErr).NotTo(HaveOccurred())
			Expect(repository.GetApplicationsCallCount()).To(Equal(1))
		})

		Context("when getting the applications succeeds", func() {
			BeforeEach(func() {
				repository.GetApplicationsReturns([]models.V3Application{
					{
						Name:                  "app-1-name",
						DesiredState:          "STOPPED",
						TotalDesiredInstances: 1,
						Links: models.Links{
							Processes: models.Link{
								Href: "/v3/apps/app-1-guid/processes",
							},
							Routes: models.Link{
								Href: "/v3/apps/app-1-guid/routes",
							},
						},
					},
					{
						Name:                  "app-2-name",
						DesiredState:          "RUNNING",
						TotalDesiredInstances: 2,
						Links: models.Links{
							Processes: models.Link{
								Href: "/v3/apps/app-2-guid/processes",
							},
							Routes: models.Link{
								Href: "/v3/apps/app-2-guid/routes",
							},
						},
					},
				}, nil)
			})

			It("tries to get the processes for each application", func() {
				Expect(runCLIErr).NotTo(HaveOccurred())
				Expect(repository.GetProcessesCallCount()).To(Equal(2))
				calls := make([]string, repository.GetProcessesCallCount())
				for i := range calls {
					calls[i] = repository.GetProcessesArgsForCall(i)
				}
				Expect(calls).To(ContainElement("/v3/apps/app-1-guid/processes"))
				Expect(calls).To(ContainElement("/v3/apps/app-2-guid/processes"))
			})

			Context("when getting all processes succeeds", func() {
				BeforeEach(func() {
					repository.GetProcessesStub = func(path string) ([]models.V3Process, error) {
						if repository.GetProcessesCallCount() == 1 {
							return []models.V3Process{
								{
									Type:       "web",
									Instances:  1,
									MemoryInMB: 1024,
									DiskInMB:   2048,
								},
							}, nil
						}

						return []models.V3Process{
							{
								Type:       "web",
								Instances:  2,
								MemoryInMB: 512,
								DiskInMB:   1024,
							},
						}, nil
					}
				})

				It("tries to get the routes for each application", func() {
					Expect(runCLIErr).NotTo(HaveOccurred())
					Expect(repository.GetRoutesCallCount()).To(Equal(2))
					calls := make([]string, repository.GetRoutesCallCount())
					for i := range calls {
						calls[i] = repository.GetRoutesArgsForCall(i)
					}
					Expect(calls).To(ContainElement("/v3/apps/app-1-guid/routes"))
					Expect(calls).To(ContainElement("/v3/apps/app-2-guid/routes"))
				})

				Context("when getting the routes succeeds", func() {
					BeforeEach(func() {
						repository.GetRoutesStub = func(path string) ([]models.V3Route, error) {
							if repository.GetRoutesCallCount() == 1 {
								return []models.V3Route{
									{
										Host: "route-1-host",
										Path: "/route-1-path",
									},
									{
										Host: "route-1-host-2",
										Path: "",
									},
								}, nil
							}

							return []models.V3Route{
								{
									Host: "route-2-host",
									Path: "",
								},
							}, nil
						}
					})

					It("prints a table of the results", func() {
						Expect(runCLIErr).NotTo(HaveOccurred())
						outputs := make([]string, len(ui.Outputs()))
						for i := range ui.Outputs() {
							outputs[i] = terminal.Decolorize(ui.Outputs()[i])
						}
						Expect(outputs).To(ConsistOf(
							MatchRegexp(`name.*requested state.*instances.*memory.*disk.*urls`),
							MatchRegexp("app-1-name.*stopped.*1.*1G.*2G.*route-1-host/route-1-path, route-1-host-2"),
							MatchRegexp("app-2-name.*running.*2.*512M.*1G.*route-2-host"),
						))
					})
				})

				Context("when getting the routes fails", func() {
					BeforeEach(func() {
						repository.GetRoutesReturns([]models.V3Route{}, errors.New("get-routes-err"))
					})

					It("fails with error", func() {
						Expect(runCLIErr).To(HaveOccurred())
						Expect(runCLIErr.Error()).To(Equal("get-routes-err"))
					})
				})
			})

			Context("when getting any process fails", func() {
				BeforeEach(func() {
					repository.GetProcessesStub = func(path string) ([]models.V3Process, error) {
						if repository.GetProcessesCallCount() == 0 {
							return []models.V3Process{
								{
									Type:       "web",
									Instances:  1,
									MemoryInMB: 1024,
									DiskInMB:   1024,
								},
							}, nil
						}

						return []models.V3Process{}, errors.New("get-processes-err")
					}
				})

				It("fails with error", func() {
					Expect(runCLIErr).To(HaveOccurred())
					Expect(runCLIErr.Error()).To(Equal("get-processes-err"))
				})
			})
		})

		Context("when getting the applications fails", func() {
			BeforeEach(func() {
				repository.GetApplicationsReturns([]models.V3Application{}, errors.New("get-applications-err"))
			})

			It("fails with error", func() {
				Expect(runCLIErr).To(HaveOccurred())
				Expect(runCLIErr.Error()).To(Equal("get-applications-err"))
			})
		})
	})
})

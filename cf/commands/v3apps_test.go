package commands_test

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/commands"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"

	fakeapi "github.com/cloudfoundry/cli/cf/api/fakes"
	fakerequirements "github.com/cloudfoundry/cli/cf/requirements/fakes"
	"github.com/cloudfoundry/cli/cf/v3/models"
	fakerepository "github.com/cloudfoundry/cli/cf/v3/repository/fakes"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("V3Apps", func() {
	var (
		ui         *testterm.FakeUI
		routeRepo  *fakeapi.FakeRouteRepository
		configRepo core_config.Repository
		repository *fakerepository.FakeRepository

		cmd         command_registry.Command
		deps        command_registry.Dependency
		factory     *fakerequirements.FakeFactory
		flagContext flags.FlagContext

		loginRequirement         requirements.Requirement
		targetedSpaceRequirement requirements.Requirement
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}

		routeRepo = &fakeapi.FakeRouteRepository{}
		repoLocator := deps.RepoLocator.SetRouteRepository(routeRepo)
		repository = &fakerepository.FakeRepository{}
		repoLocator = repoLocator.SetV3Repository(repository)

		configRepo = testconfig.NewRepositoryWithDefaults()
		deps = command_registry.Dependency{
			Ui:          ui,
			Config:      configRepo,
			RepoLocator: repoLocator,
		}

		cmd = &commands.V3Apps{}
		cmd.SetDependency(deps, false)

		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)

		factory = &fakerequirements.FakeFactory{}

		loginRequirement = &passingRequirement{}
		factory.NewLoginRequirementReturns(loginRequirement)

		targetedSpaceRequirement = &passingRequirement{}
		factory.NewTargetedSpaceRequirementReturns(targetedSpaceRequirement)
	})

	Describe("Requirements", func() {
		It("returns a LoginRequirement", func() {
			actualRequirements := cmd.Requirements(factory, flagContext)
			Expect(factory.NewLoginRequirementCallCount()).To(Equal(1))

			Expect(actualRequirements).To(ContainElement(loginRequirement))
		})

		It("returns a TargetedSpaceRequirement", func() {
			actualRequirements := cmd.Requirements(factory, flagContext)
			Expect(factory.NewTargetedSpaceRequirementCallCount()).To(Equal(1))

			Expect(actualRequirements).To(ContainElement(targetedSpaceRequirement))
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

	Describe("Execute", func() {
		BeforeEach(func() {
			cmd.Requirements(factory, flagContext)
			repository.GetProcessesReturns([]models.V3Process{{Type: "web"}, {Type: "web"}}, nil)
			repository.GetRoutesReturns([]models.V3Route{{}, {}}, nil)
		})

		It("attemps to get applications for the targeted space", func() {
			cmd.Execute(flagContext)
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
				cmd.Execute(flagContext)
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
					cmd.Execute(flagContext)
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
						cmd.Execute(flagContext)
						outputs := make([]string, len(ui.Outputs))
						for i := range ui.Outputs {
							outputs[i] = terminal.Decolorize(ui.Outputs[i])
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
						Expect(func() { cmd.Execute(flagContext) }).To(Panic())
						Expect(ui.Outputs).To(ContainSubstrings(
							[]string{"FAILED"},
							[]string{"get-routes-err"},
						))
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
					Expect(func() { cmd.Execute(flagContext) }).To(Panic())
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"get-processes-err"},
					))
				})
			})
		})

		Context("when getting the applications fails", func() {
			BeforeEach(func() {
				repository.GetApplicationsReturns([]models.V3Application{}, errors.New("get-applications-err"))
			})

			It("fails with error", func() {
				Expect(func() { cmd.Execute(flagContext) }).To(Panic())
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"get-applications-err"},
				))
			})
		})
	})
})

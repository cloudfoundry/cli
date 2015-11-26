package application_test

import (
	testApplication "github.com/cloudfoundry/cli/cf/api/applications/fakes"
	testCopyApplication "github.com/cloudfoundry/cli/cf/api/copy_application_source/fakes"
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	testorg "github.com/cloudfoundry/cli/cf/api/organizations/fakes"
	appCmdFakes "github.com/cloudfoundry/cli/cf/commands/application/fakes"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CopySource", func() {

	var (
		ui                  *testterm.FakeUI
		config              core_config.Repository
		requirementsFactory *testreq.FakeReqFactory
		authRepo            *testapi.FakeAuthenticationRepository
		appRepo             *testApplication.FakeApplicationRepository
		copyAppSourceRepo   *testCopyApplication.FakeCopyApplicationSourceRepository
		spaceRepo           *testapi.FakeSpaceRepository
		orgRepo             *testorg.FakeOrganizationRepository
		appRestarter        *appCmdFakes.FakeApplicationRestarter
		OriginalCommand     command_registry.Command
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetAuthenticationRepository(authRepo)
		deps.RepoLocator = deps.RepoLocator.SetApplicationRepository(appRepo)
		deps.RepoLocator = deps.RepoLocator.SetCopyApplicationSourceRepository(copyAppSourceRepo)
		deps.RepoLocator = deps.RepoLocator.SetSpaceRepository(spaceRepo)
		deps.RepoLocator = deps.RepoLocator.SetOrganizationRepository(orgRepo)
		deps.Config = config

		//inject fake 'command dependency' into registry
		command_registry.Register(appRestarter)

		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("copy-source").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
		authRepo = &testapi.FakeAuthenticationRepository{}
		appRepo = &testApplication.FakeApplicationRepository{}
		copyAppSourceRepo = &testCopyApplication.FakeCopyApplicationSourceRepository{}
		spaceRepo = &testapi.FakeSpaceRepository{}
		orgRepo = &testorg.FakeOrganizationRepository{}
		config = testconfig.NewRepositoryWithDefaults()

		//save original command and restore later
		OriginalCommand = command_registry.Commands.FindCommand("restart")

		appRestarter = &appCmdFakes.FakeApplicationRestarter{}
		//setup fakes to correctly interact with command_registry
		appRestarter.SetDependencyStub = func(_ command_registry.Dependency, _ bool) command_registry.Command {
			return appRestarter
		}
		appRestarter.MetaDataReturns(command_registry.CommandMetadata{Name: "restart"})
	})

	AfterEach(func() {
		command_registry.Register(OriginalCommand)
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("copy-source", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("requirement failures", func() {
		It("when not logged in", func() {
			requirementsFactory.LoginSuccess = false
			Expect(runCommand("source-app", "target-app")).ToNot(HavePassedRequirements())
		})

		It("when a space is not targeted", func() {
			requirementsFactory.TargetedSpaceSuccess = false
			Expect(runCommand("source-app", "target-app")).ToNot(HavePassedRequirements())
		})

		It("when provided too many args", func() {
			Expect(runCommand("source-app", "target-app", "too-much", "app-name")).ToNot(HavePassedRequirements())
		})
	})

	Describe("Passing requirements", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = true
		})

		Context("refreshing the auth token", func() {
			It("makes a call for the app token", func() {
				runCommand("source-app", "target-app")
				Expect(authRepo.RefreshTokenCalled).To(BeTrue())
			})

			Context("when refreshing the auth token fails", func() {
				BeforeEach(func() {
					authRepo.RefreshTokenError = errors.New("I accidentally the UAA")
				})

				It("it displays an error", func() {
					runCommand("source-app", "target-app")
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"accidentally the UAA"},
					))
				})
			})

			Describe("when retrieving the app token succeeds", func() {
				var (
					sourceApp, targetApp models.Application
				)

				BeforeEach(func() {
					sourceApp = models.Application{
						ApplicationFields: models.ApplicationFields{
							Name: "source-app",
							Guid: "source-app-guid",
						},
					}
					appRepo.ReadReturns(sourceApp, nil)

					targetApp = models.Application{
						ApplicationFields: models.ApplicationFields{
							Name: "target-app",
							Guid: "target-app-guid",
						},
					}
					appRepo.ReadFromSpaceReturns(targetApp, nil)
				})

				Describe("when no parameters are passed", func() {
					It("obtains both the source and target application from the same space", func() {
						runCommand("source-app", "target-app")

						targetAppName, spaceGuid := appRepo.ReadFromSpaceArgsForCall(0)
						Expect(targetAppName).To(Equal("target-app"))
						Expect(spaceGuid).To(Equal("my-space-guid"))

						Expect(appRepo.ReadArgsForCall(0)).To(Equal("source-app"))

						sourceAppGuid, targetAppGuid := copyAppSourceRepo.CopyApplicationArgsForCall(0)
						Expect(sourceAppGuid).To(Equal("source-app-guid"))
						Expect(targetAppGuid).To(Equal("target-app-guid"))

						appArg, orgName, spaceName := appRestarter.ApplicationRestartArgsForCall(0)
						Expect(appArg).To(Equal(targetApp))
						Expect(orgName).To(Equal(config.OrganizationFields().Name))
						Expect(spaceName).To(Equal(config.SpaceFields().Name))

						Expect(ui.Outputs).To(ContainSubstrings(
							[]string{"Copying source from app", "source-app", "to target app", "target-app", "in org my-org / space my-space as my-user..."},
							[]string{"Note: this may take some time"},
							[]string{"OK"},
						))
					})

					Context("Failures", func() {
						It("if we cannot obtain the source application", func() {
							appRepo.ReadReturns(models.Application{}, errors.New("could not find source app"))
							runCommand("source-app", "target-app")

							Expect(ui.Outputs).To(ContainSubstrings(
								[]string{"FAILED"},
								[]string{"could not find source app"},
							))
						})

						It("fails if we cannot obtain the target application", func() {
							appRepo.ReadFromSpaceReturns(models.Application{}, errors.New("could not find target app"))
							runCommand("source-app", "target-app")

							Expect(ui.Outputs).To(ContainSubstrings(
								[]string{"FAILED"},
								[]string{"could not find target app"},
							))
						})
					})
				})

				Describe("when a space is provided, but not an org", func() {
					It("send the correct target appplication for the current org and target space", func() {
						space := models.Space{}
						space.Name = "space-name"
						space.Guid = "model-space-guid"
						spaceRepo.FindByNameReturns(space, nil)

						runCommand("-s", "space-name", "source-app", "target-app")

						targetAppName, spaceGuid := appRepo.ReadFromSpaceArgsForCall(0)
						Expect(targetAppName).To(Equal("target-app"))
						Expect(spaceGuid).To(Equal("model-space-guid"))

						Expect(ui.Outputs).To(ContainSubstrings(
							[]string{"Copying source from app", "source-app", "to target app", "target-app", "in org my-org / space space-name as my-user..."},
							[]string{"Note: this may take some time"},
							[]string{"OK"},
						))
					})

					Context("Failures", func() {
						It("when we cannot find the provided space", func() {
							spaceRepo.FindByNameReturns(models.Space{}, errors.New("Error finding space by name."))

							runCommand("-s", "space-name", "source-app", "target-app")
							Expect(ui.Outputs).To(ContainSubstrings(
								[]string{"FAILED"},
								[]string{"Error finding space by name."},
							))
						})
					})
				})

				Describe("when an org and space name are passed as parameters", func() {
					It("send the correct target application for the space and org", func() {
						orgRepo.FindByNameReturns(models.Organization{
							Spaces: []models.SpaceFields{
								{
									Name: "space-name",
									Guid: "space-guid",
								},
							},
						}, nil)

						runCommand("-o", "org-name", "-s", "space-name", "source-app", "target-app")

						targetAppName, spaceGuid := appRepo.ReadFromSpaceArgsForCall(0)
						Expect(targetAppName).To(Equal("target-app"))
						Expect(spaceGuid).To(Equal("space-guid"))

						sourceAppGuid, targetAppGuid := copyAppSourceRepo.CopyApplicationArgsForCall(0)
						Expect(sourceAppGuid).To(Equal("source-app-guid"))
						Expect(targetAppGuid).To(Equal("target-app-guid"))

						appArg, orgName, spaceName := appRestarter.ApplicationRestartArgsForCall(0)
						Expect(appArg).To(Equal(targetApp))
						Expect(orgName).To(Equal("org-name"))
						Expect(spaceName).To(Equal("space-name"))

						Expect(ui.Outputs).To(ContainSubstrings(
							[]string{"Copying source from app source-app to target app target-app in org org-name / space space-name as my-user..."},
							[]string{"Note: this may take some time"},
							[]string{"OK"},
						))
					})

					Context("failures", func() {
						It("cannot just accept an organization and no space", func() {
							runCommand("-o", "org-name", "source-app", "target-app")

							Expect(ui.Outputs).To(ContainSubstrings(
								[]string{"FAILED"},
								[]string{"Please provide the space within the organization containing the target application"},
							))
						})

						It("when we cannot find the provided org", func() {
							orgRepo.FindByNameReturns(models.Organization{}, errors.New("Could not find org"))
							runCommand("-o", "org-name", "-s", "space-name", "source-app", "target-app")

							Expect(ui.Outputs).To(ContainSubstrings(
								[]string{"FAILED"},
								[]string{"Could not find org"},
							))
						})

						It("when the org does not contain the space name provide", func() {
							orgRepo.FindByNameReturns(models.Organization{}, nil)
							runCommand("-o", "org-name", "-s", "space-name", "source-app", "target-app")

							Expect(ui.Outputs).To(ContainSubstrings(
								[]string{"FAILED"},
								[]string{"Could not find space space-name in organization org-name"},
							))
						})

						It("when the targeted app does not exist in the targeted org and space", func() {
							orgRepo.FindByNameReturns(models.Organization{
								Spaces: []models.SpaceFields{
									{
										Name: "space-name",
										Guid: "space-guid",
									},
								},
							}, nil)

							appRepo.ReadFromSpaceReturns(models.Application{}, errors.New("could not find app"))
							runCommand("-o", "org-name", "-s", "space-name", "source-app", "target-app")

							Expect(ui.Outputs).To(ContainSubstrings(
								[]string{"FAILED"},
								[]string{"could not find app"},
							))
						})
					})
				})

				Describe("when the --no-restart flag is passed", func() {
					It("does not restart the target application", func() {
						runCommand("--no-restart", "source-app", "target-app")
						Expect(appRestarter.ApplicationRestartCallCount()).To(Equal(0))
					})
				})
			})
		})
	})
})

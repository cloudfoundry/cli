package application_test

import (
	testApplication "github.com/cloudfoundry/cli/cf/api/applications/fakes"
	testCopyApplication "github.com/cloudfoundry/cli/cf/api/copy_application_source/fakes"
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	testorg "github.com/cloudfoundry/cli/cf/api/organizations/fakes"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/application"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CopySource", func() {

	var (
		ui                  *testterm.FakeUI
		config              core_config.ReadWriter
		requirementsFactory *testreq.FakeReqFactory
		authRepo            *testapi.FakeAuthenticationRepository
		appRepo             *testApplication.FakeApplicationRepository
		copyAppSourceRepo   *testCopyApplication.FakeCopyApplicationSourceRepository
		spaceRepo           *testapi.FakeSpaceRepository
		orgRepo             *testorg.FakeOrganizationRepository
		appRestarter        *testcmd.FakeAppRestarter
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
		authRepo = &testapi.FakeAuthenticationRepository{}
		appRepo = &testApplication.FakeApplicationRepository{}
		copyAppSourceRepo = &testCopyApplication.FakeCopyApplicationSourceRepository{}
		spaceRepo = &testapi.FakeSpaceRepository{}
		orgRepo = &testorg.FakeOrganizationRepository{}
		appRestarter = &testcmd.FakeAppRestarter{}
		config = testconfig.NewRepositoryWithDefaults()
	})

	runCommand := func(args ...string) bool {
		cmd := NewCopySource(ui, config, authRepo, appRepo, orgRepo, spaceRepo, copyAppSourceRepo, appRestarter)
		return testcmd.RunCommand(cmd, args, requirementsFactory)
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
					appRepo.ReadReturns.App = sourceApp

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

						Expect(appRepo.ReadArgs.Name).To(Equal("source-app"))

						sourceAppGuid, targetAppGuid := copyAppSourceRepo.CopyApplicationArgsForCall(0)
						Expect(sourceAppGuid).To(Equal("source-app-guid"))
						Expect(targetAppGuid).To(Equal("target-app-guid"))

						Expect(appRestarter.AppToRestart).To(Equal(targetApp))

						Expect(ui.Outputs).To(ContainSubstrings(
							[]string{"Copying source from app", "source-app", "to target app", "target-app", "in org my-org / space my-space as my-user..."},
							[]string{"Note: this may take some time"},
							[]string{"OK"},
						))
					})

					Context("Failures", func() {
						It("if we cannot obtain the source application", func() {
							appRepo.ReadReturns.Error = errors.New("could not find source app")
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
						spaceRepo.Spaces = []models.Space{
							{
								SpaceFields: models.SpaceFields{
									Name: "space-name",
									Guid: "model-space-guid",
								},
							},
						}

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
							spaceRepo.FindByNameErr = true

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

						Expect(appRestarter.AppToRestart).To(Equal(targetApp))

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
						Expect(appRestarter.AppToRestart).To(Equal(models.Application{}))
					})
				})
			})
		})
	})
})

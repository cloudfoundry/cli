package application_test

import (
	testApplication "github.com/cloudfoundry/cli/cf/api/applications/fakes"
	testCopyApplication "github.com/cloudfoundry/cli/cf/api/copy_application_source/fakes"
	testorg "github.com/cloudfoundry/cli/cf/api/organizations/fakes"
	"github.com/cloudfoundry/cli/cf/models"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
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
		cmd                 *CopySource
		ui                  *testterm.FakeUI
		config              core_config.ReadWriter
		requirementsFactory *testreq.FakeReqFactory
		authRepo            *testapi.FakeAuthenticationRepository
		appRepo             *testApplication.FakeApplicationRepository
		copyAppSourceRepo   *testCopyApplication.FakeCopyApplicationSourceRepository
		spaceRepo           *testapi.FakeSpaceRepository
		orgRepo             *testorg.FakeOrganizationRepository
	)

	BeforeEach(func() {
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
		authRepo = &testapi.FakeAuthenticationRepository{}
		appRepo = &testApplication.FakeApplicationRepository{}
		copyAppSourceRepo = &testCopyApplication.FakeCopyApplicationSourceRepository{}
		spaceRepo = &testapi.FakeSpaceRepository{}
		orgRepo = &testorg.FakeOrganizationRepository{}
		config = testconfig.NewRepositoryWithDefaults()

		ui = new(testterm.FakeUI)
		cmd = NewCopySource(ui, config, authRepo, appRepo, orgRepo, spaceRepo, copyAppSourceRepo)
	})

	copySource := func(args ...string) {
		testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {

		It("fails when not logged in", func() {
			requirementsFactory.LoginSuccess = false
			copySource("source-app", "target-app")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("fails when a space is not targeted", func() {
			requirementsFactory.TargetedSpaceSuccess = false
			copySource("source-app", "target-app")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("fails when provided too many args", func() {
			copySource("source-app", "target-app", "path", "too-much", "app-name")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	Describe("Passing requirements", func() {
		Context("refreshing the auth token", func() {
			It("makes a call for the app token", func() { // so clean
				copySource("source-app", "target-app")

				Expect(authRepo.RefreshTokenCalled).To(BeTrue())
			})

			Context("when refreshing the auth token fails", func() {
				BeforeEach(func() {
					authRepo.RefreshTokenError = errors.New("I accidentally the UAA")
				})

				It("it displays an error", func() {
					copySource("source-app", "target-app")

					Expect(ui.Outputs).ToNot(ContainSubstrings(
						[]string{"Error refreshing auth token"},
					))
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"accidentally the UAA"},
					))
				})
			})

			Describe("when retrieving the app token succeeds", func() {
				var sourceApp, targetApp models.Application
				Context("It should obtain the source app guid", func() {
					BeforeEach(func() {
						sourceApp = models.Application{}
						sourceApp.Name = "source-app"
						sourceApp.Guid = "source-app-guid"
						appRepo.ReadReturns.App = sourceApp

						targetApp = models.Application{}
						targetApp.Name = "target-app"
						targetApp.Guid = "target-app-guid"
						appRepo.ReadFromSpaceReturns(targetApp, nil)

						copySource("source-app", "target-app")
					})

					It("copies the application source", func() {
						Expect(appRepo.ReadArgs.Name).To(Equal("source-app"))
						targetAppName, spaceGuid := appRepo.ReadFromSpaceArgsForCall(0)
						Expect(targetAppName).To(Equal("target-app"))
						Expect(spaceGuid).To(Equal("my-space-guid"))

						sourceAppGuid, targetAppGuid := copyAppSourceRepo.CopyApplicationArgsForCall(0)
						Expect(sourceAppGuid).To(Equal("source-app-guid"))
						Expect(targetAppGuid).To(Equal("target-app-guid"))

					})
				})
			})
		})
	})
})

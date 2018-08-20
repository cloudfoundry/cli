package v3_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2v3action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v3"
	"code.cloudfoundry.org/cli/command/v3/v3fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("app Command", func() {
	var (
		cmd                 v3.AppCommand
		testUI              *ui.UI
		fakeConfig          *commandfakes.FakeConfig
		fakeSharedActor     *commandfakes.FakeSharedActor
		fakeActor           *v3fakes.FakeAppActor
		fakeAppSummaryActor *v3fakes.FakeAppSummaryActor
		binaryName          string
		executeErr          error
		app                 string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeAppActor)
		fakeAppSummaryActor = new(v3fakes.FakeAppSummaryActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		app = "some-app"

		cmd = v3.AppCommand{
			RequiredArgs: flag.AppName{AppName: app},

			UI:              testUI,
			Config:          fakeConfig,
			SharedActor:     fakeSharedActor,
			Actor:           fakeActor,
			AppSummaryActor: fakeAppSummaryActor,
		}

		fakeConfig.TargetedOrganizationReturns(configv3.Organization{
			Name: "some-org",
			GUID: "some-org-guid",
		})
		fakeConfig.TargetedSpaceReturns(configv3.Space{
			Name: "some-space",
			GUID: "some-space-guid",
		})

		fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("the API version is below the minimum", func() {
		BeforeEach(func() {
			fakeActor.CloudControllerAPIVersionReturns(ccversion.MinV3ClientVersion)
		})

		It("returns a MinimumAPIVersionNotMetError", func() {
			Expect(executeErr).To(MatchError(translatableerror.MinimumCFAPIVersionNotMetError{
				CurrentVersion: ccversion.MinV3ClientVersion,
				MinimumVersion: ccversion.MinVersionApplicationFlowV3,
			}))
		})
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeActor.CloudControllerAPIVersionReturns(ccversion.MinVersionApplicationFlowV3)
			fakeSharedActor.CheckTargetReturns(actionerror.NoOrganizationTargetedError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NoOrganizationTargetedError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	When("the user is not logged in", func() {
		var expectedErr error

		BeforeEach(func() {
			fakeActor.CloudControllerAPIVersionReturns(ccversion.MinVersionApplicationFlowV3)
			expectedErr = errors.New("some current user error")
			fakeConfig.CurrentUserReturns(configv3.User{}, expectedErr)
		})

		It("return an error", func() {
			Expect(executeErr).To(Equal(expectedErr))
		})
	})

	When("the --guid flag is provided", func() {
		BeforeEach(func() {
			fakeActor.CloudControllerAPIVersionReturns(ccversion.MinVersionApplicationFlowV3)
			cmd.GUID = true
		})

		When("no errors occur", func() {
			BeforeEach(func() {
				fakeActor.GetApplicationByNameAndSpaceReturns(
					v3action.Application{GUID: "some-guid"},
					v3action.Warnings{"warning-1", "warning-2"},
					nil)
			})

			It("displays the application guid and all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("some-guid"))
				Expect(testUI.Err).To(Say("warning-1"))
				Expect(testUI.Err).To(Say("warning-2"))

				Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
				appName, spaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
				Expect(appName).To(Equal("some-app"))
				Expect(spaceGUID).To(Equal("some-space-guid"))
			})
		})

		When("an error is encountered getting the app", func() {
			When("the error is translatable", func() {
				BeforeEach(func() {
					fakeActor.GetApplicationByNameAndSpaceReturns(
						v3action.Application{},
						v3action.Warnings{"warning-1", "warning-2"},
						actionerror.ApplicationNotFoundError{Name: "some-app"})
				})

				It("returns a translatable error and all warnings", func() {
					Expect(executeErr).To(MatchError(actionerror.ApplicationNotFoundError{Name: "some-app"}))

					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))
				})
			})

			When("the error is not translatable", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("get app summary error")
					fakeActor.GetApplicationByNameAndSpaceReturns(
						v3action.Application{},
						v3action.Warnings{"warning-1", "warning-2"},
						expectedErr)
				})

				It("returns the error and all warnings", func() {
					Expect(executeErr).To(MatchError(expectedErr))

					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))
				})
			})
		})
	})

	When("the --guid is not passed", func() {
		When("getting the application summary returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				fakeActor.CloudControllerAPIVersionReturns(ccversion.MinVersionApplicationFlowV3)
				expectedErr = actionerror.ApplicationNotFoundError{Name: app}
				fakeAppSummaryActor.GetApplicationSummaryByNameAndSpaceReturns(v2v3action.ApplicationSummary{}, v2v3action.Warnings{"warning-1", "warning-2"}, expectedErr)
			})

			It("returns the error and prints warnings", func() {
				Expect(executeErr).To(Equal(actionerror.ApplicationNotFoundError{Name: app}))

				Expect(testUI.Out).To(Say("Showing health and status for app some-app in org some-org / space some-space as steve\\.\\.\\."))

				Expect(testUI.Err).To(Say("warning-1"))
				Expect(testUI.Err).To(Say("warning-2"))
			})
		})

		When("getting the application summary is successful", func() {
			BeforeEach(func() {
				fakeActor.CloudControllerAPIVersionReturns(ccversion.MinVersionApplicationFlowV3)
				summary := v2v3action.ApplicationSummary{
					ApplicationSummary: v3action.ApplicationSummary{
						Application: v3action.Application{
							Name:  "some-app",
							State: constant.ApplicationStarted,
						},
						CurrentDroplet: v3action.Droplet{
							Stack: "cflinuxfs2",
							Buildpacks: []v3action.Buildpack{
								{
									Name:         "ruby_buildpack",
									DetectOutput: "some-detect-output",
								},
								{
									Name:         "some-buildpack",
									DetectOutput: "",
								},
							},
						},
						ProcessSummaries: v3action.ProcessSummaries{
							{
								Process: v3action.Process{
									Type:    constant.ProcessTypeWeb,
									Command: "some-command-1",
								},
							},
							{
								Process: v3action.Process{
									Type:    "console",
									Command: "some-command-2",
								},
							},
						},
					},
				}
				fakeAppSummaryActor.GetApplicationSummaryByNameAndSpaceReturns(summary, v2v3action.Warnings{"warning-1", "warning-2"}, nil)
			})

			It("prints the application summary and outputs warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("(?m)Showing health and status for app some-app in org some-org / space some-space as steve\\.\\.\\.\n\n"))
				Expect(testUI.Out).To(Say("name:\\s+some-app"))
				Expect(testUI.Out).To(Say("requested state:\\s+started"))
				Expect(testUI.Out).ToNot(Say("start command:"))

				Expect(testUI.Err).To(Say("warning-1"))
				Expect(testUI.Err).To(Say("warning-2"))

				Expect(fakeAppSummaryActor.GetApplicationSummaryByNameAndSpaceCallCount()).To(Equal(1))
				appName, spaceGUID, withObfuscatedValues := fakeAppSummaryActor.GetApplicationSummaryByNameAndSpaceArgsForCall(0)
				Expect(appName).To(Equal("some-app"))
				Expect(spaceGUID).To(Equal("some-space-guid"))
				Expect(withObfuscatedValues).To(BeFalse())
			})
		})
	})
})

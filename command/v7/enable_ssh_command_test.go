package v7_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/ui"
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("enable-ssh Command", func() {
	var (
		cmd                EnableSSHCommand
		testUI             *ui.UI
		fakeConfig         *commandfakes.FakeConfig
		fakeSharedActor    *commandfakes.FakeSharedActor
		fakeEnableSSHActor *v7fakes.FakeEnableSSHActor

		binaryName      string
		currentUserName string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeEnableSSHActor = new(v7fakes.FakeEnableSSHActor)

		cmd = EnableSSHCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeEnableSSHActor,
		}

		cmd.RequiredArgs.AppName = "some-app"

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		currentUserName = "some-user"
		fakeConfig.CurrentUserNameReturns(currentUserName, nil)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: "faceman"}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	When("the user is logged in", func() {
		When("no errors occur", func() {
			BeforeEach(func() {
				fakeEnableSSHActor.GetApplicationByNameAndSpaceReturns(
					v7action.Application{Name: "some-app", GUID: "some-app-guid"},
					v7action.Warnings{"some-get-app-warnings"},
					nil,
				)
				fakeEnableSSHActor.GetAppFeatureReturns(
					ccv3.ApplicationFeature{Enabled: false, Name: "ssh"},
					v7action.Warnings{"some-feature-warnings"},
					nil,
				)
				fakeEnableSSHActor.UpdateAppFeatureReturns(
					v7action.Warnings{"some-update-ssh-warnings"},
					nil,
				)
			})

			It("enables ssh on the app", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeEnableSSHActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))

				appName, spaceGUID := fakeEnableSSHActor.GetApplicationByNameAndSpaceArgsForCall(0)
				Expect(appName).To(Equal(cmd.RequiredArgs.AppName))
				Expect(spaceGUID).To(Equal(cmd.Config.TargetedSpace().GUID))

				Expect(fakeEnableSSHActor.GetAppFeatureCallCount()).To(Equal(1))

				appGUID, featureName := fakeEnableSSHActor.GetAppFeatureArgsForCall(0)
				Expect(appGUID).To(Equal("some-app-guid"))
				Expect(featureName).To(Equal("ssh"))

				Expect(fakeEnableSSHActor.UpdateAppFeatureCallCount()).To(Equal(1))
				app, enabled, featureName := fakeEnableSSHActor.UpdateAppFeatureArgsForCall(0)
				Expect(app.Name).To(Equal("some-app"))
				Expect(enabled).To(Equal(true))
				Expect(featureName).To(Equal("ssh"))

				Expect(testUI.Err).To(Say("some-get-app-warnings"))
				Expect(testUI.Err).To(Say("some-feature-warnings"))
				Expect(testUI.Err).To(Say("some-update-ssh-warnings"))
				Expect(testUI.Out).To(Say(`Enabling ssh support for app %s as %s\.\.\.`, appName, currentUserName))
				Expect(testUI.Out).To(Say("OK"))
			})
		})

		When("app ssh is already enabled", func() {
			BeforeEach(func() {
				fakeEnableSSHActor.UpdateAppFeatureReturns(
					v7action.Warnings{"ssh support for app 'some-app' is already enabled.", "some-other-warnings"},
					nil,
				)
				fakeEnableSSHActor.GetAppFeatureReturns(
					ccv3.ApplicationFeature{Enabled: true, Name: "ssh"},
					v7action.Warnings{},
					nil,
				)
			})

			It("shows the app ssh is already enabled", func() {
				Expect(testUI.Out).To(Say("ssh support for app 'some-app' is already enabled."))
				Expect(testUI.Out).To(Say("OK"))
			})
		})

		When("an error occurs", func() {
			When("GetApp action errors", func() {
				When("no user is found", func() {
					var returnedErr error

					BeforeEach(func() {
						returnedErr = actionerror.ApplicationNotFoundError{Name: "some-app"}
						fakeEnableSSHActor.GetApplicationByNameAndSpaceReturns(
							v7action.Application{},
							nil,
							returnedErr)
					})

					It("returns the same error", func() {
						Expect(executeErr).To(HaveOccurred())
						Expect(executeErr).To(MatchError(returnedErr))
					})
				})
			})

			When("GetAppFeature action errors", func() {
				returnedErr := errors.New("some-error")
				BeforeEach(func() {

					fakeEnableSSHActor.GetAppFeatureReturns(
						ccv3.ApplicationFeature{},
						nil,
						returnedErr,
					)
				})

				It("returns the same error", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(executeErr).To(MatchError(returnedErr))
				})
			})

			When("Enable ssh action errors", func() {
				var returnedErr error

				BeforeEach(func() {
					returnedErr = errors.New("some-error")
					fakeEnableSSHActor.GetApplicationByNameAndSpaceReturns(
						v7action.Application{Name: "some-app"},
						v7action.Warnings{"some-warning"},
						nil,
					)
					fakeEnableSSHActor.UpdateAppFeatureReturns(nil, returnedErr)
				})

				It("returns the same error", func() {
					Expect(executeErr).To(MatchError(returnedErr))
				})
			})
		})
	})
})

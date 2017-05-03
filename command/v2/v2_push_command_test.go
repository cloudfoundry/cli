package v2_test

import (
	"errors"
	"os"

	"code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/pushaction/manifest"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v2"
	"code.cloudfoundry.org/cli/command/v2/v2fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("v2-push Command", func() {
	var (
		cmd             V2PushCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v2fakes.FakeV2PushActor
		input           *Buffer
		binaryName      string

		appName    string
		executeErr error
		pwd        string
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v2fakes.FakeV2PushActor)

		cmd = V2PushCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		appName = "some-app"
		cmd.OptionalArgs.AppName = appName
		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

		var err error
		pwd, err = os.Getwd()
		Expect(err).ToNot(HaveOccurred())
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(sharedaction.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(command.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			_, checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	Context("when the user is logged in, and org and space are targeted", func() {
		BeforeEach(func() {
			fakeConfig.HasTargetedOrganizationReturns(true)
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{GUID: "some-org-guid", Name: "some-org"})
			fakeConfig.HasTargetedSpaceReturns(true)
			fakeConfig.TargetedSpaceReturns(configv3.Space{GUID: "some-space-guid", Name: "some-space"})
			fakeConfig.CurrentUserReturns(configv3.User{Name: "some-user"}, nil)
		})

		Context("when the push settings are valid", func() {
			var appManifests []manifest.Application

			BeforeEach(func() {
				appManifests = []manifest.Application{
					{
						Name: appName,
						Path: pwd,
					},
				}
				fakeActor.MergeAndValidateSettingsAndManifestsReturns(appManifests, nil)
			})

			Context("when the settings can be converted to a valid config", func() {
				var appConfigs []pushaction.ApplicationConfig

				BeforeEach(func() {
					appConfigs = []pushaction.ApplicationConfig{
						{
							DesiredApplication: v2action.Application{Name: appName},
							TargetedSpaceGUID:  "some-space-guid",
							Path:               pwd,
						},
					}
					fakeActor.ConvertToApplicationConfigsReturns(appConfigs, pushaction.Warnings{"some-config-warnings"}, nil)
				})

				Context("when the push is successful", func() {
					var (
						eventStream    chan pushaction.Event
						warningsStream chan pushaction.Warnings
						errorStream    chan error
					)

					BeforeEach(func() {
						eventStream = make(chan pushaction.Event)
						warningsStream = make(chan pushaction.Warnings)
						errorStream = make(chan error)

						fakeActor.ApplyReturns(eventStream, warningsStream, errorStream)

						go func() {
							defer GinkgoRecover()

							Eventually(eventStream).Should(BeSent(pushaction.ApplicationCreated))
							Eventually(eventStream).Should(BeSent(pushaction.ApplicationUpdated))
							Eventually(eventStream).Should(BeSent(pushaction.RouteCreated))
							Eventually(eventStream).Should(BeSent(pushaction.RouteBound))
							Eventually(eventStream).Should(BeSent(pushaction.UploadingApplication))
							Eventually(eventStream).Should(BeSent(pushaction.UploadComplete))
							Eventually(eventStream).Should(BeSent(pushaction.Complete))
							Eventually(warningsStream).Should(BeSent(pushaction.Warnings{"apply-1", "apply-2"}))
							close(eventStream)
							close(warningsStream)
							close(errorStream)
						}()
					})

					AfterEach(func() {
						Eventually(eventStream).Should(BeClosed())
						Eventually(warningsStream).Should(BeClosed())
						Eventually(errorStream).Should(BeClosed())
					})

					It("merges app manifest and flags", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(fakeActor.MergeAndValidateSettingsAndManifestsCallCount()).To(Equal(1))
						cmdSettings, _ := fakeActor.MergeAndValidateSettingsAndManifestsArgsForCall(0)
						Expect(cmdSettings).To(Equal(pushaction.CommandLineSettings{
							Name: appName,
							Path: pwd,
						}))
					})

					It("converts the manifests to app configs and outputs config warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Err).To(Say("some-config-warnings"))

						Expect(fakeActor.ConvertToApplicationConfigsCallCount()).To(Equal(1))
						orgGUID, spaceGUID, manifests := fakeActor.ConvertToApplicationConfigsArgsForCall(0)
						Expect(orgGUID).To(Equal("some-org-guid"))
						Expect(spaceGUID).To(Equal("some-space-guid"))
						Expect(manifests).To(Equal(appManifests))
					})

					It("outputs flavor text prior to generating app configuration", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Out).To(Say("Getting app info..."))
					})

					It("applies each of the application configurations", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(fakeActor.ApplyCallCount()).To(Equal(1))
						Expect(fakeActor.ApplyArgsForCall(0)).To(Equal(appConfigs[0]))
					})

					It("displays app events and warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("Creating app %s in org %s / space %s as %s...", appName, "some-org", "some-space", "some-user"))
						Expect(testUI.Out).To(Say("Updating app %s in org %s / space %s as %s...", appName, "some-org", "some-space", "some-user"))
						Expect(testUI.Out).To(Say("Creating routes..."))
						Expect(testUI.Out).To(Say("Binding routes..."))
						Expect(testUI.Out).To(Say("Uploading application..."))
						Expect(testUI.Out).To(Say("Upload complete"))

						Expect(testUI.Err).To(Say("some-config-warnings"))
						Expect(testUI.Err).To(Say("apply-1"))
						Expect(testUI.Err).To(Say("apply-2"))
					})

					It("displays app staging logs", func() {
						Skip("will fill in later")

					})

					It("displays the app info", func() {
						Skip("will fill in later")
					})
				})

				Context("when the push errors", func() {
					var (
						expectedErr    error
						eventStream    chan pushaction.Event
						warningsStream chan pushaction.Warnings
						errorStream    chan error
					)

					BeforeEach(func() {
						expectedErr = errors.New("no wayz dude")
						eventStream = make(chan pushaction.Event)
						warningsStream = make(chan pushaction.Warnings)
						errorStream = make(chan error)

						fakeActor.ApplyReturns(eventStream, warningsStream, errorStream)

						go func() {
							defer GinkgoRecover()

							Eventually(warningsStream).Should(BeSent(pushaction.Warnings{"apply-1", "apply-2"}))
							Eventually(errorStream).Should(BeSent(expectedErr))
							close(eventStream)
							close(warningsStream)
							close(errorStream)
						}()
					})

					AfterEach(func() {
						Eventually(eventStream).Should(BeClosed())
						Eventually(warningsStream).Should(BeClosed())
						Eventually(errorStream).Should(BeClosed())
					})

					It("outputs the warnings and returns the error", func() {
						Expect(executeErr).To(MatchError(expectedErr))

						Expect(testUI.Err).To(Say("some-config-warnings"))
						Expect(testUI.Err).To(Say("apply-1"))
						Expect(testUI.Err).To(Say("apply-2"))
					})
				})
			})

			Context("when there is an error converting the app setting into a config", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("no wayz dude")
					fakeActor.ConvertToApplicationConfigsReturns(nil, pushaction.Warnings{"some-config-warnings"}, expectedErr)
				})

				It("outputs the warnings and returns the error", func() {
					Expect(executeErr).To(MatchError(expectedErr))

					Expect(testUI.Err).To(Say("some-config-warnings"))
				})
			})
		})

		Context("when the push settings are invalid", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("no wayz dude")
				fakeActor.MergeAndValidateSettingsAndManifestsReturns(nil, expectedErr)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(expectedErr))
			})
		})
	})
})

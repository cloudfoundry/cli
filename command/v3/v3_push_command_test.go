package v3_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/v3action"
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
	. "github.com/onsi/gomega/gstruct"
)

type Step struct {
	Error    error
	Event    pushaction.Event
	Warnings pushaction.Warnings
}

func FillInValues(tuples []Step, state pushaction.PushState) func(state pushaction.PushState, progressBar pushaction.ProgressBar) (<-chan pushaction.PushState, <-chan pushaction.Event, <-chan pushaction.Warnings, <-chan error) {
	return func(state pushaction.PushState, progressBar pushaction.ProgressBar) (<-chan pushaction.PushState, <-chan pushaction.Event, <-chan pushaction.Warnings, <-chan error) {
		stateStream := make(chan pushaction.PushState)
		eventStream := make(chan pushaction.Event)
		warningsStream := make(chan pushaction.Warnings)
		errorStream := make(chan error)

		go func() {
			defer close(stateStream)
			defer close(eventStream)
			defer close(warningsStream)
			defer close(errorStream)

			for _, tuple := range tuples {
				warningsStream <- tuple.Warnings
				if tuple.Error != nil {
					errorStream <- tuple.Error
					return
				} else {
					eventStream <- tuple.Event
				}
			}

			eventStream <- pushaction.Complete
			stateStream <- state
		}()

		return stateStream, eventStream, warningsStream, errorStream
	}
}

var _ = Describe("v3-push Command", func() {
	var (
		cmd              v3.V3PushCommand
		testUI           *ui.UI
		fakeConfig       *commandfakes.FakeConfig
		fakeSharedActor  *commandfakes.FakeSharedActor
		fakeActor        *v3fakes.FakeV3PushActor
		fakeVersionActor *v3fakes.FakeV3PushVersionActor
		binaryName       string
		executeErr       error

		userName  string
		spaceName string
		orgName   string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeV3PushActor)
		fakeVersionActor = new(v3fakes.FakeV3PushVersionActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		fakeConfig.ExperimentalReturns(true) // TODO: Delete once we remove the experimental flag

		cmd = v3.V3PushCommand{
			RequiredArgs: flag.AppName{AppName: "some-app"},
			UI:           testUI,
			Config:       fakeConfig,
			Actor:        fakeActor,
			VersionActor: fakeVersionActor,
			SharedActor:  fakeSharedActor,
		}

		userName = "banana"
		spaceName = "some-space"
		orgName = "some-org"
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when the API version is below the minimum", func() {
		BeforeEach(func() {
			fakeVersionActor.CloudControllerAPIVersionReturns("0.0.0")
		})

		It("returns a MinimumAPIVersionNotMetError", func() {
			Expect(executeErr).To(MatchError(translatableerror.MinimumAPIVersionNotMetError{
				CurrentVersion: "0.0.0",
				MinimumVersion: ccversion.MinVersionV3,
			}))
		})
	})

	Context("when the API version is met", func() {
		BeforeEach(func() {
			fakeVersionActor.CloudControllerAPIVersionReturns(ccversion.MinVersionV3)
		})

		Context("when checking target fails", func() {
			BeforeEach(func() {
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

		Context("when checking target fails because the user is not logged in", func() {
			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

				Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
				checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
				Expect(checkTargetedOrg).To(BeTrue())
				Expect(checkTargetedSpace).To(BeTrue())
			})
		})

		Context("when the user is logged in", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{Name: userName}, nil)

				fakeConfig.TargetedOrganizationReturns(configv3.Organization{
					Name: orgName,
					GUID: "some-org-guid",
				})
				fakeConfig.TargetedSpaceReturns(configv3.Space{
					Name: spaceName,
					GUID: "some-space-guid",
				})
			})

			It("displays the experimental warning", func() {
				Expect(testUI.Err).To(Say("This command is in EXPERIMENTAL stage and may change without notice"))
			})

			Context("when getting app settings is successful", func() {
				BeforeEach(func() {
					fakeActor.GeneratePushStateReturns(
						[]pushaction.PushState{
							{
								Application: v3action.Application{Name: "some-app"},
							},
						},
						pushaction.Warnings{"some-warning-1"}, nil)
				})

				Context("when the app is successfully actualized", func() {
					BeforeEach(func() {
						fakeActor.ActualizeStub = FillInValues([]Step{
							{
								Event:    pushaction.SkipingApplicationCreation,
								Warnings: pushaction.Warnings{"skipping app creation warnings"},
							},
							{
								Event:    pushaction.CreatedApplication,
								Warnings: pushaction.Warnings{"app creation warnings"},
							},
						}, pushaction.PushState{})
					})

					It("generates a push state with the specified app path", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(testUI.Out).To(Say("Getting app info..."))
						Expect(testUI.Err).To(Say("some-warning-1"))

						Expect(fakeActor.GeneratePushStateCallCount()).To(Equal(1))
						settings, spaceGUID := fakeActor.GeneratePushStateArgsForCall(0)
						Expect(settings).To(MatchFields(IgnoreExtras, Fields{
							"Name": Equal("some-app"),
						}))
						Expect(spaceGUID).To(Equal("some-space-guid"))
					})

					It("actualizes the application and displays events/warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("Updating app some-app..."))
						Expect(testUI.Err).To(Say("skipping app creation warnings"))

						Expect(testUI.Out).To(Say("Creating app some-app..."))
						Expect(testUI.Err).To(Say("app creation warnings"))
					})
				})
			})

			Context("when getting app settings returns an error", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("fasdflkjdkjinenwnnondsokekm foo")
					fakeActor.GeneratePushStateReturns(nil, pushaction.Warnings{"some-warning-1"}, expectedErr)
				})

				It("generates a push state with the specified app path", func() {
					Expect(executeErr).To(MatchError(expectedErr))
					Expect(testUI.Err).To(Say("some-warning-1"))
				})
			})

			Context("when app path is specified", func() {
				BeforeEach(func() {
					cmd.AppPath = "some/app/path"
				})

				It("generates a push state with the specified app path", func() {
					Expect(fakeActor.GeneratePushStateCallCount()).To(Equal(1))
					settings, spaceGUID := fakeActor.GeneratePushStateArgsForCall(0)
					Expect(settings).To(MatchFields(IgnoreExtras, Fields{
						"Name":            Equal("some-app"),
						"ProvidedAppPath": Equal("some/app/path"),
					}))
					Expect(spaceGUID).To(Equal("some-space-guid"))
				})
			})

			Context("when buildpack is specified", func() {
				BeforeEach(func() {
					cmd.Buildpacks = []string{"some-buildpack-1", "some-buildpack-2"}
				})

				It("generates a push state with the specified buildpacks", func() {
					Expect(fakeActor.GeneratePushStateCallCount()).To(Equal(1))
					settings, spaceGUID := fakeActor.GeneratePushStateArgsForCall(0)
					Expect(settings).To(MatchFields(IgnoreExtras, Fields{
						"Name":       Equal("some-app"),
						"Buildpacks": Equal([]string{"some-buildpack-1", "some-buildpack-2"}),
					}))
					Expect(spaceGUID).To(Equal("some-space-guid"))
				})
			})
		})
	})
})

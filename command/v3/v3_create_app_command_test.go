package v3_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
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

var _ = Describe("v3-create-app Command", func() {
	var (
		cmd             v3.V3CreateAppCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v3fakes.FakeV3CreateAppActor
		binaryName      string
		executeErr      error
		app             string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeV3CreateAppActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		app = "some-app"

		cmd = v3.V3CreateAppCommand{
			UI:           testUI,
			Config:       fakeConfig,
			SharedActor:  fakeSharedActor,
			Actor:        fakeActor,
			RequiredArgs: flag.AppName{AppName: app},
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(sharedaction.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(translatableerror.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			_, checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{Name: "banana"}, nil)
			fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "some-space", GUID: "some-space-guid"})
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-org"})
		})

		Context("when the create is successful", func() {
			BeforeEach(func() {
				fakeActor.CreateApplicationByNameAndSpaceReturns(v3action.Application{}, v3action.Warnings{"I am a warning", "I am also a warning"}, nil)
			})

			It("displays the header and ok", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("Creating V3 app some-app in org some-org / space some-space as banana..."))
				Expect(testUI.Out).To(Say("OK"))

				Expect(testUI.Err).To(Say("I am a warning"))
				Expect(testUI.Err).To(Say("I am also a warning"))

				Expect(fakeActor.CreateApplicationByNameAndSpaceCallCount()).To(Equal(1))

				createAppInput := fakeActor.CreateApplicationByNameAndSpaceArgsForCall(0)
				Expect(createAppInput).To(Equal(v3action.CreateApplicationInput{
					AppName:   app,
					SpaceGUID: "some-space-guid",
				}))
			})
		})

		Context("when the create is unsuccessful", func() {
			Context("due to an unexpected error", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("I am an error")
					fakeActor.CreateApplicationByNameAndSpaceReturns(v3action.Application{}, v3action.Warnings{"I am a warning", "I am also a warning"}, expectedErr)
				})

				It("displays the header and error", func() {
					Expect(executeErr).To(MatchError(expectedErr))

					Expect(testUI.Out).To(Say("Creating V3 app some-app in org some-org / space some-space as banana..."))

					Expect(testUI.Err).To(Say("I am a warning"))
					Expect(testUI.Err).To(Say("I am also a warning"))
				})
			})

			Context("due to an ApplicationAlreadyExistsError", func() {
				BeforeEach(func() {
					fakeActor.CreateApplicationByNameAndSpaceReturns(v3action.Application{}, v3action.Warnings{"I am a warning", "I am also a warning"}, v3action.ApplicationAlreadyExistsError{})
				})

				It("displays the header and ok", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Creating V3 app some-app in org some-org / space some-space as banana..."))
					Expect(testUI.Out).To(Say("OK"))

					Expect(testUI.Err).To(Say("I am a warning"))
					Expect(testUI.Err).To(Say("I am also a warning"))
					Expect(testUI.Err).To(Say("App %s already exists.", app))
				})
			})
		})
	})
})

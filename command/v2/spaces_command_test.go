package v2_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v2"
	"code.cloudfoundry.org/cli/command/v2/v2fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("spaces Command", func() {
	var (
		cmd             SpacesCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v2fakes.FakeSpacesActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v2fakes.FakeSpacesActor)

		cmd = SpacesCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when an error is encountered checking if the environment is setup correctly", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrgArg, checkTargetedSpaceArg := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrgArg).To(BeTrue())
			Expect(checkTargetedSpaceArg).To(BeFalse())
		})
	})

	Context("when the user is logged in and an org is targeted", func() {
		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				GUID: "some-org-guid",
				Name: "some-org",
			})
		})

		Context("when getting the current user fails", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{}, errors.New("get-user-error"))
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("get-user-error"))
			})
		})

		Context("when getting the current user succeeds", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(
					configv3.User{Name: "some-user"},
					nil)
			})

			Context("when there are no spaces", func() {
				BeforeEach(func() {
					fakeActor.GetOrganizationSpacesReturns(
						[]v2action.Space{},
						v2action.Warnings{"get-spaces-warning"},
						nil)
				})

				It("displays that there are no spaces", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Getting spaces in org some-org as some-user\\.\\.\\."))
					Expect(testUI.Out).To(Say(""))
					Expect(testUI.Out).To(Say("No spaces found\\."))

					Expect(testUI.Err).To(Say("get-spaces-warning"))

					Expect(fakeActor.GetOrganizationSpacesCallCount()).To(Equal(1))
					Expect(fakeActor.GetOrganizationSpacesArgsForCall(0)).To(Equal("some-org-guid"))
				})
			})

			Context("when there are multiple spaces", func() {
				BeforeEach(func() {
					fakeActor.GetOrganizationSpacesReturns(
						[]v2action.Space{
							{Name: "space-1"},
							{Name: "space-2"},
						},
						v2action.Warnings{"get-spaces-warning"},
						nil)
				})

				It("displays all the spaces in the org", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Getting spaces in org some-org as some-user\\.\\.\\."))
					Expect(testUI.Out).To(Say(""))
					Expect(testUI.Out).To(Say("name"))
					Expect(testUI.Out).To(Say("space-1"))
					Expect(testUI.Out).To(Say("space-2"))

					Expect(testUI.Err).To(Say("get-spaces-warning"))

					Expect(fakeActor.GetOrganizationSpacesCallCount()).To(Equal(1))
					Expect(fakeActor.GetOrganizationSpacesArgsForCall(0)).To(Equal("some-org-guid"))
				})
			})

			Context("when a translatable error is encountered getting spaces", func() {
				BeforeEach(func() {
					fakeActor.GetOrganizationSpacesReturns(
						nil,
						v2action.Warnings{"get-spaces-warning"},
						actionerror.OrganizationNotFoundError{Name: "not-found-org"})
				})

				It("returns a translatable error", func() {
					Expect(executeErr).To(MatchError(actionerror.OrganizationNotFoundError{Name: "not-found-org"}))

					Expect(testUI.Out).To(Say("Getting spaces in org some-org as some-user\\.\\.\\."))
					Expect(testUI.Out).To(Say(""))

					Expect(testUI.Err).To(Say("get-spaces-warning"))

					Expect(fakeActor.GetOrganizationSpacesCallCount()).To(Equal(1))
					Expect(fakeActor.GetOrganizationSpacesArgsForCall(0)).To(Equal("some-org-guid"))
				})
			})
		})
	})
})

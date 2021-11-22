package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("enable-org-isolation Command", func() {
	var (
		cmd              EnableOrgIsolationCommand
		testUI           *ui.UI
		fakeConfig       *commandfakes.FakeConfig
		fakeSharedActor  *commandfakes.FakeSharedActor
		fakeActor        *v7fakes.FakeActor
		binaryName       string
		executeErr       error
		isolationSegment string
		org              string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = EnableOrgIsolationCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		org = "some-org"
		isolationSegment = "segment1"
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeFalse())
			Expect(checkTargetedSpace).To(BeFalse())
		})
	})

	When("the user is logged in", func() {
		BeforeEach(func() {
			fakeActor.GetCurrentUserReturns(configv3.User{Name: "banana"}, nil)

			cmd.RequiredArgs.OrganizationName = org
			cmd.RequiredArgs.IsolationSegmentName = isolationSegment
		})

		When("the enable is successful", func() {
			BeforeEach(func() {
				fakeActor.EntitleIsolationSegmentToOrganizationByNameReturns(v7action.Warnings{"I am a warning", "I am also a warning"}, nil)
			})

			It("displays the header and ok", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("Enabling isolation segment segment1 for org some-org as banana..."))
				Expect(testUI.Out).To(Say("OK"))

				Expect(testUI.Err).To(Say("I am a warning"))
				Expect(testUI.Err).To(Say("I am also a warning"))

				Expect(fakeActor.EntitleIsolationSegmentToOrganizationByNameCallCount()).To(Equal(1))

				isolationSegmentName, orgName := fakeActor.EntitleIsolationSegmentToOrganizationByNameArgsForCall(0)
				Expect(orgName).To(Equal(org))
				Expect(isolationSegmentName).To(Equal(isolationSegment))
			})
		})

		When("the enable is unsuccessful", func() {
			Context("due to an unexpected error", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("I am an error")
					fakeActor.EntitleIsolationSegmentToOrganizationByNameReturns(v7action.Warnings{"I am a warning", "I am also a warning"}, expectedErr)
				})

				It("displays the header and error", func() {
					Expect(executeErr).To(MatchError(expectedErr))

					Expect(testUI.Out).To(Say("Enabling isolation segment segment1 for org some-org as banana..."))

					Expect(testUI.Err).To(Say("I am a warning"))
					Expect(testUI.Err).To(Say("I am also a warning"))
				})
			})

			When("the isolation segment does not exist", func() {
				BeforeEach(func() {
					fakeActor.EntitleIsolationSegmentToOrganizationByNameReturns(v7action.Warnings{"I am a warning", "I am also a warning"}, actionerror.IsolationSegmentNotFoundError{Name: "segment1"})
				})

				It("displays all warnings and the isolation segment not found error", func() {
					Expect(testUI.Err).To(Say("I am a warning"))
					Expect(testUI.Err).To(Say("I am also a warning"))
					Expect(executeErr).To(MatchError(actionerror.IsolationSegmentNotFoundError{Name: "segment1"}))
				})
			})

			When("the organization does not exist", func() {
				BeforeEach(func() {
					fakeActor.EntitleIsolationSegmentToOrganizationByNameReturns(
						v7action.Warnings{"I am a warning", "I am also a warning"},
						actionerror.OrganizationNotFoundError{Name: "some-org"})
				})

				It("displays all warnings and the org not found error", func() {
					Expect(executeErr).To(MatchError(actionerror.OrganizationNotFoundError{Name: "some-org"}))
				})
			})

		})
	})
})

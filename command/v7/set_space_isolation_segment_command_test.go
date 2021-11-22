package v7_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("set-space-isolation-segment Command", func() {
	var (
		cmd              SetSpaceIsolationSegmentCommand
		testUI           *ui.UI
		fakeConfig       *commandfakes.FakeConfig
		fakeSharedActor  *commandfakes.FakeSharedActor
		fakeActor        *v7fakes.FakeActor
		binaryName       string
		executeErr       error
		isolationSegment string
		space            string
		org              string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = SetSpaceIsolationSegmentCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		space = "some-space"
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
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeFalse())
		})
	})

	When("the user is logged in", func() {
		BeforeEach(func() {
			fakeActor.GetCurrentUserReturns(configv3.User{Name: "banana"}, nil)
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: org,
				GUID: "some-org-guid",
			})

			cmd.RequiredArgs.SpaceName = space
			cmd.RequiredArgs.IsolationSegmentName = isolationSegment
		})

		When("the space lookup is unsuccessful", func() {
			BeforeEach(func() {
				fakeActor.GetSpaceByNameAndOrganizationReturns(resources.Space{}, v7action.Warnings{"I am a warning", "I am also a warning"}, actionerror.SpaceNotFoundError{Name: space})
			})

			It("returns the warnings and error", func() {
				Expect(executeErr).To(MatchError(actionerror.SpaceNotFoundError{Name: space}))
				Expect(testUI.Err).To(Say("I am a warning"))
				Expect(testUI.Err).To(Say("I am also a warning"))
			})
		})

		When("the space lookup is successful", func() {
			BeforeEach(func() {
				fakeActor.GetSpaceByNameAndOrganizationReturns(resources.Space{
					Name: space,
					GUID: "some-space-guid",
				}, v7action.Warnings{"I am a warning", "I am also a warning"}, nil)
			})

			When("the entitlement is successful", func() {
				BeforeEach(func() {
					fakeActor.AssignIsolationSegmentToSpaceByNameAndSpaceReturns(v7action.Warnings{"entitlement-warning", "banana"}, nil)
				})

				It("Displays the header and okay", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Updating isolation segment of space %s in org %s as banana...", space, org))
					Expect(testUI.Out).To(Say("OK"))

					Expect(testUI.Err).To(Say("I am a warning"))
					Expect(testUI.Err).To(Say("I am also a warning"))
					Expect(testUI.Err).To(Say("entitlement-warning"))
					Expect(testUI.Err).To(Say("banana"))

					Expect(testUI.Out).To(Say("TIP: Restart applications in this space to relocate them to this isolation segment."))

					Expect(fakeActor.AssignIsolationSegmentToSpaceByNameAndSpaceCallCount()).To(Equal(1))
					isolationSegmentName, spaceGUID := fakeActor.AssignIsolationSegmentToSpaceByNameAndSpaceArgsForCall(0)
					Expect(isolationSegmentName).To(Equal(isolationSegment))
					Expect(spaceGUID).To(Equal("some-space-guid"))
				})
			})

			When("the entitlement errors", func() {
				BeforeEach(func() {
					fakeActor.AssignIsolationSegmentToSpaceByNameAndSpaceReturns(v7action.Warnings{"entitlement-warning", "banana"}, actionerror.IsolationSegmentNotFoundError{Name: "segment1"})
				})

				It("returns the warnings and error", func() {
					Expect(testUI.Out).To(Say("Updating isolation segment of space %s in org %s as banana...", space, org))
					Expect(testUI.Err).To(Say("I am a warning"))
					Expect(testUI.Err).To(Say("I am also a warning"))
					Expect(testUI.Err).To(Say("entitlement-warning"))
					Expect(testUI.Err).To(Say("banana"))
					Expect(executeErr).To(MatchError(actionerror.IsolationSegmentNotFoundError{Name: "segment1"}))
				})
			})
		})
	})
})

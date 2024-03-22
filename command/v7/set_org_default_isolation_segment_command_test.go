package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v7action"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/resources"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("set-org-default-isolation-segment Command", func() {
	var (
		cmd              SetOrgDefaultIsolationSegmentCommand
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

		cmd = SetOrgDefaultIsolationSegmentCommand{
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

		fakeActor.GetCurrentUserReturns(configv3.User{Name: "banana"}, nil)

		cmd.RequiredArgs.OrganizationName = org
		cmd.RequiredArgs.IsolationSegmentName = isolationSegment

		fakeActor.GetOrganizationByNameReturns(resources.Organization{
			Name: org,
			GUID: "some-org-guid",
		}, v7action.Warnings{"org-warning-1", "org-warning-2"}, nil)
		fakeActor.GetIsolationSegmentByNameReturns(resources.IsolationSegment{GUID: "some-iso-guid"}, v7action.Warnings{"iso-seg-warning-1", "iso-seg-warning-2"}, nil)
		fakeActor.SetOrganizationDefaultIsolationSegmentReturns(v7action.Warnings{"isolation-set-warning-1", "isolation-set-warning-2"}, nil)
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

	When("fetching the user fails", func() {
		BeforeEach(func() {
			fakeActor.GetCurrentUserReturns(configv3.User{}, errors.New("some-error"))
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError("some-error"))
		})
	})

	It("prints out its intentions", func() {
		Expect(testUI.Out).To(Say(`Setting isolation segment %s to default on org %s as banana\.\.\.`, isolationSegment, org))
	})

	When("the org lookup is unsuccessful", func() {
		BeforeEach(func() {
			fakeActor.GetOrganizationByNameReturns(resources.Organization{}, v7action.Warnings{"org-warning-1", "org-warning-2"}, actionerror.OrganizationNotFoundError{Name: org})
		})

		It("returns the warnings and error", func() {
			Expect(executeErr).To(MatchError(actionerror.OrganizationNotFoundError{Name: org}))
			Expect(testUI.Err).To(Say("org-warning-1"))
			Expect(testUI.Err).To(Say("org-warning-2"))
		})
	})

	It("prints the org fetch warnings", func() {
		Expect(testUI.Err).To(Say("org-warning-1"))
		Expect(testUI.Err).To(Say("org-warning-2"))
	})

	When("the isolation segment lookup is unsuccessful", func() {
		BeforeEach(func() {
			fakeActor.GetIsolationSegmentByNameReturns(resources.IsolationSegment{}, v7action.Warnings{"iso-seg-warning-1", "iso-seg-warning-2"}, actionerror.IsolationSegmentNotFoundError{Name: isolationSegment})
		})

		It("returns the warnings and error", func() {
			Expect(executeErr).To(MatchError(actionerror.IsolationSegmentNotFoundError{Name: isolationSegment}))
			Expect(testUI.Err).To(Say("iso-seg-warning-1"))
			Expect(testUI.Err).To(Say("iso-seg-warning-2"))
		})
	})

	It("prints the iso segment fetch warnings", func() {
		Expect(testUI.Err).To(Say("iso-seg-warning-1"))
		Expect(testUI.Err).To(Say("iso-seg-warning-2"))
	})

	When("setting the default isolation to an org errors", func() {
		BeforeEach(func() {
			fakeActor.SetOrganizationDefaultIsolationSegmentReturns(v7action.Warnings{"isolation-set-warning-1", "isolation-set-warning-2"}, actionerror.IsolationSegmentNotFoundError{Name: isolationSegment})
		})

		It("returns the warnings and error", func() {
			Expect(testUI.Err).To(Say("isolation-set-warning-1"))
			Expect(testUI.Err).To(Say("isolation-set-warning-2"))
			Expect(executeErr).To(MatchError(actionerror.IsolationSegmentNotFoundError{Name: isolationSegment}))
		})
	})

	It("prints the iso segment set warnings", func() {
		Expect(testUI.Err).To(Say("isolation-set-warning-1"))
		Expect(testUI.Err).To(Say("isolation-set-warning-2"))
	})

	It("Displays the header and okay", func() {
		Expect(executeErr).ToNot(HaveOccurred())

		Expect(testUI.Out).To(Say("OK"))
		Expect(testUI.Out).To(Say("TIP: Restart applications in this organization to relocate them to this isolation segment."))

		Expect(fakeActor.SetOrganizationDefaultIsolationSegmentCallCount()).To(Equal(1))
		orgGUID, isoSegGUID := fakeActor.SetOrganizationDefaultIsolationSegmentArgsForCall(0)
		Expect(orgGUID).To(Equal("some-org-guid"))
		Expect(isoSegGUID).To(Equal("some-iso-guid"))
	})
})

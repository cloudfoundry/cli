package v3_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v3"
	"code.cloudfoundry.org/cli/command/v3/v3fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("set-org-default-isolation-segment Command", func() {
	var (
		cmd              v3.SetOrgDefaultIsolationSegmentCommand
		testUI           *ui.UI
		fakeConfig       *commandfakes.FakeConfig
		fakeSharedActor  *commandfakes.FakeSharedActor
		fakeActor        *v3fakes.FakeSetOrgDefaultIsolationSegmentActor
		fakeActorV2      *v3fakes.FakeSetOrgDefaultIsolationSegmentActorV2
		binaryName       string
		executeErr       error
		isolationSegment string
		org              string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeSetOrgDefaultIsolationSegmentActor)
		fakeActorV2 = new(v3fakes.FakeSetOrgDefaultIsolationSegmentActorV2)

		cmd = v3.SetOrgDefaultIsolationSegmentCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
			ActorV2:     fakeActorV2,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		org = "some-org"
		isolationSegment = "segment1"

		fakeActor.CloudControllerAPIVersionReturns(ccversion.MinVersionIsolationSegmentV3)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when the API version is below the minimum", func() {
		BeforeEach(func() {
			fakeActor.CloudControllerAPIVersionReturns("0.0.0")
		})

		It("returns a MinimumAPIVersionNotMetError", func() {
			Expect(executeErr).To(MatchError(translatableerror.MinimumAPIVersionNotMetError{
				CurrentVersion: "0.0.0",
				MinimumVersion: ccversion.MinVersionIsolationSegmentV3,
			}))
		})
	})

	Context("when checking target fails", func() {
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

	Context("when fetching the user fails", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{}, errors.New("some-error"))
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError("some-error"))
		})
	})

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{Name: "banana"}, nil)

			cmd.RequiredArgs.OrganizationName = org
			cmd.RequiredArgs.IsolationSegmentName = isolationSegment
		})

		Context("when the org lookup is unsuccessful", func() {
			BeforeEach(func() {
				fakeActorV2.GetOrganizationByNameReturns(v2action.Organization{}, v2action.Warnings{"I am a warning", "I am also a warning"}, actionerror.OrganizationNotFoundError{Name: org})
			})

			It("returns the warnings and error", func() {
				Expect(executeErr).To(MatchError(actionerror.OrganizationNotFoundError{Name: org}))
				Expect(testUI.Err).To(Say("I am a warning"))
				Expect(testUI.Err).To(Say("I am also a warning"))
			})
		})

		Context("when the org lookup is successful", func() {
			BeforeEach(func() {
				fakeActorV2.GetOrganizationByNameReturns(v2action.Organization{
					Name: org,
					GUID: "some-org-guid",
				}, v2action.Warnings{"org-warning-1", "org-warning-2"}, nil)
			})

			Context("when the isolation segment lookup is unsuccessful", func() {
				BeforeEach(func() {
					fakeActor.GetIsolationSegmentByNameReturns(v3action.IsolationSegment{}, v3action.Warnings{"iso-seg-warning-1", "iso-seg-warning-2"}, actionerror.IsolationSegmentNotFoundError{Name: isolationSegment})
				})

				It("returns the warnings and error", func() {
					Expect(executeErr).To(MatchError(actionerror.IsolationSegmentNotFoundError{Name: isolationSegment}))
					Expect(testUI.Err).To(Say("org-warning-1"))
					Expect(testUI.Err).To(Say("org-warning-2"))
					Expect(testUI.Err).To(Say("iso-seg-warning-1"))
					Expect(testUI.Err).To(Say("iso-seg-warning-2"))
				})
			})

			Context("when the entitlement is successful", func() {
				BeforeEach(func() {
					fakeActor.GetIsolationSegmentByNameReturns(v3action.IsolationSegment{GUID: "some-iso-guid"}, v3action.Warnings{"iso-seg-warning-1", "iso-seg-warning-2"}, nil)
					fakeActor.SetOrganizationDefaultIsolationSegmentReturns(v3action.Warnings{"entitlement-warning", "banana"}, nil)
				})

				It("Displays the header and okay", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Setting isolation segment %s to default on org %s as banana\\.\\.\\.", isolationSegment, org))
					Expect(testUI.Out).To(Say("OK"))

					Expect(testUI.Err).To(Say("org-warning-1"))
					Expect(testUI.Err).To(Say("org-warning-2"))
					Expect(testUI.Err).To(Say("iso-seg-warning-1"))
					Expect(testUI.Err).To(Say("iso-seg-warning-2"))
					Expect(testUI.Err).To(Say("entitlement-warning"))
					Expect(testUI.Err).To(Say("banana"))

					Expect(testUI.Out).To(Say("In order to move running applications to this isolation segment, they must be restarted\\."))

					Expect(fakeActor.SetOrganizationDefaultIsolationSegmentCallCount()).To(Equal(1))
					orgGUID, isoSegGUID := fakeActor.SetOrganizationDefaultIsolationSegmentArgsForCall(0)
					Expect(orgGUID).To(Equal("some-org-guid"))
					Expect(isoSegGUID).To(Equal("some-iso-guid"))
				})

				Context("when the entitlement errors", func() {
					BeforeEach(func() {
						fakeActor.SetOrganizationDefaultIsolationSegmentReturns(v3action.Warnings{"entitlement-warning", "banana"}, actionerror.IsolationSegmentNotFoundError{Name: isolationSegment})
					})

					It("returns the warnings and error", func() {
						Expect(testUI.Out).To(Say("Setting isolation segment %s to default on org %s as banana\\.\\.\\.", isolationSegment, org))
						Expect(testUI.Err).To(Say("org-warning-1"))
						Expect(testUI.Err).To(Say("org-warning-2"))
						Expect(testUI.Err).To(Say("iso-seg-warning-1"))
						Expect(testUI.Err).To(Say("iso-seg-warning-2"))
						Expect(testUI.Err).To(Say("entitlement-warning"))
						Expect(testUI.Err).To(Say("banana"))
						Expect(executeErr).To(MatchError(actionerror.IsolationSegmentNotFoundError{Name: isolationSegment}))
					})
				})
			})
		})
	})
})

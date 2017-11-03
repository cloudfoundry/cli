package v3_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
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

var _ = Describe("disable-org-isolation Command", func() {
	var (
		cmd              v3.DisableOrgIsolationCommand
		testUI           *ui.UI
		fakeConfig       *commandfakes.FakeConfig
		fakeSharedActor  *commandfakes.FakeSharedActor
		fakeActor        *v3fakes.FakeDisableOrgIsolationActor
		binaryName       string
		executeErr       error
		isolationSegment string
		org              string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeDisableOrgIsolationActor)

		cmd = v3.DisableOrgIsolationCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		org = "org1"
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

	Context("when user is logged in", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{Name: "admin"}, nil)
			cmd.RequiredArgs.OrganizationName = org
			cmd.RequiredArgs.IsolationSegmentName = isolationSegment
		})

		It("Outputs a mesaage", func() {
			Expect(testUI.Out).To(Say("Removing entitlement to isolation segment segment1 from org org1 as admin..."))
		})

		Context("when revoking is successful", func() {
			BeforeEach(func() {
				fakeActor.RevokeIsolationSegmentFromOrganizationByNameReturns(v3action.Warnings{"warning 1", "warning 2"}, nil)
			})

			It("Isolation segnment is revoked from org", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("OK"))
				Expect(testUI.Err).To(Say("warning 1"))
				Expect(testUI.Err).To(Say("warning 2"))

				Expect(fakeActor.RevokeIsolationSegmentFromOrganizationByNameCallCount()).To(Equal(1))
				actualIsolationSegmentName, actualOrgName := fakeActor.RevokeIsolationSegmentFromOrganizationByNameArgsForCall(0)
				Expect(actualIsolationSegmentName).To(Equal(isolationSegment))
				Expect(actualOrgName).To(Equal(org))
			})
		})

		Context("generic error while revoking segment isolation", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("ZOMG")
				fakeActor.RevokeIsolationSegmentFromOrganizationByNameReturns(v3action.Warnings{"warning 1", "warning 2"}, expectedErr)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(expectedErr))
				Expect(testUI.Err).To(Say("warning 1"))
				Expect(testUI.Err).To(Say("warning 2"))
			})
		})
	})
})

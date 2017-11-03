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

var _ = Describe("enable-org-isolation Command", func() {
	var (
		cmd              v3.EnableOrgIsolationCommand
		testUI           *ui.UI
		fakeConfig       *commandfakes.FakeConfig
		fakeSharedActor  *commandfakes.FakeSharedActor
		fakeActor        *v3fakes.FakeEnableOrgIsolationActor
		binaryName       string
		executeErr       error
		isolationSegment string
		org              string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeEnableOrgIsolationActor)

		cmd = v3.EnableOrgIsolationCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
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

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{Name: "banana"}, nil)

			cmd.RequiredArgs.OrganizationName = org
			cmd.RequiredArgs.IsolationSegmentName = isolationSegment
		})

		Context("when the enable is successful", func() {
			BeforeEach(func() {
				fakeActor.EntitleIsolationSegmentToOrganizationByNameReturns(v3action.Warnings{"I am a warning", "I am also a warning"}, nil)
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

		Context("when the enable is unsuccessful", func() {
			Context("due to an unexpected error", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("I am an error")
					fakeActor.EntitleIsolationSegmentToOrganizationByNameReturns(v3action.Warnings{"I am a warning", "I am also a warning"}, expectedErr)
				})

				It("displays the header and error", func() {
					Expect(executeErr).To(MatchError(expectedErr))

					Expect(testUI.Out).To(Say("Enabling isolation segment segment1 for org some-org as banana..."))

					Expect(testUI.Err).To(Say("I am a warning"))
					Expect(testUI.Err).To(Say("I am also a warning"))
				})
			})

			Context("when the isolation segment does not exist", func() {
				BeforeEach(func() {
					fakeActor.EntitleIsolationSegmentToOrganizationByNameReturns(v3action.Warnings{"I am a warning", "I am also a warning"}, actionerror.IsolationSegmentNotFoundError{Name: "segment1"})
				})

				It("displays all warnings and the isolation segment not found error", func() {
					Expect(testUI.Err).To(Say("I am a warning"))
					Expect(testUI.Err).To(Say("I am also a warning"))
					Expect(executeErr).To(MatchError(actionerror.IsolationSegmentNotFoundError{Name: "segment1"}))
				})
			})

			Context("when the organization does not exist", func() {
				BeforeEach(func() {
					fakeActor.EntitleIsolationSegmentToOrganizationByNameReturns(
						v3action.Warnings{"I am a warning", "I am also a warning"},
						actionerror.OrganizationNotFoundError{Name: "some-org"})
				})

				It("displays all warnings and the org not found error", func() {
					Expect(executeErr).To(MatchError(actionerror.OrganizationNotFoundError{Name: "some-org"}))
				})
			})

		})
	})
})

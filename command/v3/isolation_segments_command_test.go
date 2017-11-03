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

var _ = Describe("isolation-segments Command", func() {
	var (
		cmd             v3.IsolationSegmentsCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v3fakes.FakeIsolationSegmentsActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeIsolationSegmentsActor)

		cmd = v3.IsolationSegmentsCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
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

	Context("when checking target does not fail", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{Name: "banana"}, nil)
		})

		Context("when an error is not encountered getting the isolation segment summaries", func() {
			Context("when there are isolation segments", func() {
				BeforeEach(func() {
					fakeActor.GetIsolationSegmentSummariesReturns(
						[]v3action.IsolationSegmentSummary{
							{
								Name:         "some-iso-1",
								EntitledOrgs: []string{},
							},
							{
								Name:         "some-iso-2",
								EntitledOrgs: []string{"some-org-1"},
							},
							{
								Name:         "some-iso-3",
								EntitledOrgs: []string{"some-org-1", "some-org-2"},
							},
						},
						v3action.Warnings{"warning-1", "warning-2"},
						nil,
					)
				})

				It("displays the isolation segment summaries and all warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Getting isolation segments as banana..."))
					Expect(testUI.Out).To(Say("OK\n\n"))
					Expect(testUI.Out).To(Say("name\\s+orgs"))
					Expect(testUI.Out).To(Say("some-iso-1"))
					Expect(testUI.Out).To(Say("some-iso-2\\s+some-org-1"))
					Expect(testUI.Out).To(Say("some-iso-3\\s+some-org-1, some-org-2"))

					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))

					Expect(fakeActor.GetIsolationSegmentSummariesCallCount()).To(Equal(1))
				})
			})

			Context("when there are no isolation segments", func() {
				BeforeEach(func() {
					fakeActor.GetIsolationSegmentSummariesReturns(
						[]v3action.IsolationSegmentSummary{},
						nil,
						nil,
					)
				})
				It("displays the empty table", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(testUI.Out).To(Say("Getting isolation segments as banana..."))
					Expect(testUI.Out).To(Say("OK\n\n"))
					Expect(testUI.Out).To(Say("name\\s+orgs"))
					Expect(testUI.Out).NotTo(Say("[a-zA-Z]+"))

					Expect(fakeActor.GetIsolationSegmentSummariesCallCount()).To(Equal(1))
				})
			})
		})

		Context("when an error is encountered getting the isolation segment summaries", func() {
			var expectedError error
			BeforeEach(func() {
				expectedError = errors.New("some-error")
				fakeActor.GetIsolationSegmentSummariesReturns(
					[]v3action.IsolationSegmentSummary{},
					v3action.Warnings{"warning-1", "warning-2"},
					expectedError,
				)
			})

			It("displays warnings and returns the error", func() {
				Expect(executeErr).To(MatchError(expectedError))

				Expect(testUI.Out).To(Say("Getting isolation segments as banana..."))
				Expect(testUI.Out).NotTo(Say("OK"))

				Expect(testUI.Err).To(Say("warning-1"))
				Expect(testUI.Err).To(Say("warning-2"))
			})
		})
	})
})

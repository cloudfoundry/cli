package v2_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v2"
	"code.cloudfoundry.org/cli/command/v2/v2fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("org Command", func() {
	var (
		cmd             OrgCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v2fakes.FakeOrgActor
		fakeActorV3     *v2fakes.FakeOrgActorV3
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v2fakes.FakeOrgActor)
		fakeActorV3 = new(v2fakes.FakeOrgActorV3)

		cmd = OrgCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
			ActorV3:     fakeActorV3,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		cmd.RequiredArgs.Organization = "some-org"
		fakeActorV3.CloudControllerAPIVersionReturns("3.12.0")
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when checking the target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(
				actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			targetedOrganizationRequired, targetedSpaceRequired := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(targetedOrganizationRequired).To(Equal(false))
			Expect(targetedSpaceRequired).To(Equal(false))
		})
	})

	Context("when the --guid flag is provided", func() {
		BeforeEach(func() {
			cmd.GUID = true
		})

		Context("when no errors occur", func() {
			BeforeEach(func() {
				fakeActor.GetOrganizationByNameReturns(
					v2action.Organization{GUID: "some-org-guid"},
					v2action.Warnings{"warning-1", "warning-2"},
					nil)
			})

			It("displays the org guid and outputs all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("some-org-guid"))
				Expect(testUI.Err).To(Say("warning-1"))
				Expect(testUI.Err).To(Say("warning-2"))

				Expect(fakeActor.GetOrganizationByNameCallCount()).To(Equal(1))
				orgName := fakeActor.GetOrganizationByNameArgsForCall(0)
				Expect(orgName).To(Equal("some-org"))
			})
		})

		Context("when getting the org returns an error", func() {
			Context("when the error is translatable", func() {
				BeforeEach(func() {
					fakeActor.GetOrganizationByNameReturns(
						v2action.Organization{},
						v2action.Warnings{"warning-1", "warning-2"},
						actionerror.OrganizationNotFoundError{Name: "some-org"})
				})

				It("returns a translatable error and outputs all warnings", func() {
					Expect(executeErr).To(MatchError(actionerror.OrganizationNotFoundError{Name: "some-org"}))

					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))
				})
			})

			Context("when the error is not translatable", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("get org error")
					fakeActor.GetOrganizationByNameReturns(
						v2action.Organization{},
						v2action.Warnings{"warning-1", "warning-2"},
						expectedErr)
				})

				It("returns the error and all warnings", func() {
					Expect(executeErr).To(MatchError(expectedErr))

					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))
				})
			})
		})
	})

	Context("when the --guid flag is not provided", func() {
		Context("when no errors occur", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(
					configv3.User{
						Name: "some-user",
					},
					nil)

				fakeActor.GetOrganizationSummaryByNameReturns(
					v2action.OrganizationSummary{
						Organization: v2action.Organization{
							Name: "some-org",
							GUID: "some-org-guid",
							DefaultIsolationSegmentGUID: "default-isolation-segment-guid",
						},
						DomainNames: []string{
							"a-shared.com",
							"b-private.com",
							"c-shared.com",
							"d-private.com",
						},
						QuotaName: "some-quota",
						SpaceNames: []string{
							"space1",
							"space2",
						},
					},
					v2action.Warnings{"warning-1", "warning-2"},
					nil)
			})

			Context("when the v3 actor is nil", func() {
				BeforeEach(func() {
					cmd.ActorV3 = nil
				})
				It("displays the org summary with no isolation segment row", func() {
					Expect(executeErr).To(BeNil())
					Expect(testUI.Out).ToNot(Say("isolation segments:"))
				})
			})

			Context("when api version is above 3.11.0", func() {
				BeforeEach(func() {
					fakeActorV3.GetIsolationSegmentsByOrganizationReturns(
						[]v3action.IsolationSegment{
							{
								Name: "isolation-segment-1",
								GUID: "default-isolation-segment-guid",
							}, {
								Name: "isolation-segment-2",
								GUID: "some-other-isolation-segment-guid",
							},
						},
						v3action.Warnings{"warning-3", "warning-4"},
						nil)
					fakeActorV3.CloudControllerAPIVersionReturns("3.12.0")
				})

				It("displays warnings and a table with org domains, org quota, spaces and isolation segments", func() {
					Expect(executeErr).To(BeNil())

					Expect(testUI.Out).To(Say("Getting info for org %s as some-user\\.\\.\\.", cmd.RequiredArgs.Organization))
					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))
					Expect(testUI.Err).To(Say("warning-3"))
					Expect(testUI.Err).To(Say("warning-4"))

					Expect(testUI.Out).To(Say("name:\\s+%s", cmd.RequiredArgs.Organization))
					Expect(testUI.Out).To(Say("domains:\\s+a-shared.com, b-private.com, c-shared.com, d-private.com"))
					Expect(testUI.Out).To(Say("quota:\\s+some-quota"))
					Expect(testUI.Out).To(Say("spaces:\\s+space1, space2"))
					Expect(testUI.Out).To(Say("isolation segments:\\s+isolation-segment-1 \\(default\\), isolation-segment-2"))

					Expect(fakeConfig.CurrentUserCallCount()).To(Equal(1))

					Expect(fakeActor.GetOrganizationSummaryByNameCallCount()).To(Equal(1))
					orgName := fakeActor.GetOrganizationSummaryByNameArgsForCall(0)
					Expect(orgName).To(Equal("some-org"))

					Expect(fakeActorV3.GetIsolationSegmentsByOrganizationCallCount()).To(Equal(1))
					orgGuid := fakeActorV3.GetIsolationSegmentsByOrganizationArgsForCall(0)
					Expect(orgGuid).To(Equal("some-org-guid"))
				})
			})

			Context("when api version is below 3.11.0", func() {
				BeforeEach(func() {
					fakeActorV3.CloudControllerAPIVersionReturns("3.10.0")
				})

				It("displays warnings and a table with org domains, org quota, spaces and isolation segments", func() {
					Expect(executeErr).To(BeNil())

					Expect(testUI.Out).To(Say("Getting info for org %s as some-user\\.\\.\\.", cmd.RequiredArgs.Organization))
					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))

					Expect(testUI.Out).To(Say("name:\\s+%s", cmd.RequiredArgs.Organization))
					Expect(testUI.Out).To(Say("domains:\\s+a-shared.com, b-private.com, c-shared.com, d-private.com"))
					Expect(testUI.Out).To(Say("quota:\\s+some-quota"))
					Expect(testUI.Out).To(Say("spaces:\\s+space1, space2"))
					Expect(testUI.Out).ToNot(Say("isolation segments:"))

					Expect(fakeConfig.CurrentUserCallCount()).To(Equal(1))

					Expect(fakeActor.GetOrganizationSummaryByNameCallCount()).To(Equal(1))
					orgName := fakeActor.GetOrganizationSummaryByNameArgsForCall(0)
					Expect(orgName).To(Equal("some-org"))
				})
			})
		})

		Context("when getting the current user returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("getting current user error")
				fakeConfig.CurrentUserReturns(
					configv3.User{},
					expectedErr)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(expectedErr))
			})
		})

		Context("when getting the org summary returns an error", func() {
			Context("when the error is translatable", func() {
				BeforeEach(func() {
					fakeActor.GetOrganizationSummaryByNameReturns(
						v2action.OrganizationSummary{},
						v2action.Warnings{"warning-1", "warning-2"},
						actionerror.OrganizationNotFoundError{Name: "some-org"})
				})

				It("returns a translatable error and outputs all warnings", func() {
					Expect(executeErr).To(MatchError(actionerror.OrganizationNotFoundError{Name: "some-org"}))

					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))
				})
			})

			Context("when the error is not translatable", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("get org error")
					fakeActor.GetOrganizationSummaryByNameReturns(
						v2action.OrganizationSummary{},
						v2action.Warnings{"warning-1", "warning-2"},
						expectedErr)
				})

				It("returns the error and all warnings", func() {
					Expect(executeErr).To(MatchError(expectedErr))

					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))
				})
			})
		})

		Context("when getting the org isolation segments returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("get org iso segs error")
				fakeActorV3.GetIsolationSegmentsByOrganizationReturns(
					nil,
					v3action.Warnings{"get iso seg warning"},
					expectedErr)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(expectedErr))
				Expect(testUI.Err).To(Say("get iso seg warning"))
			})
		})
	})
})

package v7_test

import (
	"errors"

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

var _ = Describe("org Command", func() {
	var (
		cmd             OrgCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = OrgCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		cmd.RequiredArgs.Organization = "some-org"
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("checking the target fails", func() {
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

	When("the --guid flag is provided", func() {
		BeforeEach(func() {
			cmd.GUID = true
		})

		When("no errors occur", func() {
			BeforeEach(func() {
				fakeActor.GetOrganizationByNameReturns(
					resources.Organization{GUID: "some-org-guid"},
					v7action.Warnings{"warning-1", "warning-2"},
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

		When("getting the org returns an error", func() {
			When("the error is translatable", func() {
				BeforeEach(func() {
					fakeActor.GetOrganizationByNameReturns(
						resources.Organization{},
						v7action.Warnings{"warning-1", "warning-2"},
						actionerror.OrganizationNotFoundError{Name: "some-org"})
				})

				It("returns a translatable error and outputs all warnings", func() {
					Expect(executeErr).To(MatchError(actionerror.OrganizationNotFoundError{Name: "some-org"}))

					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))
				})
			})

			When("the error is not translatable", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("get org error")
					fakeActor.GetOrganizationByNameReturns(
						resources.Organization{},
						v7action.Warnings{"warning-1", "warning-2"},
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

	When("the --guid flag is not provided", func() {
		When("no errors occur", func() {
			BeforeEach(func() {
				fakeActor.GetCurrentUserReturns(
					configv3.User{
						Name: "some-user",
					},
					nil)

				fakeActor.GetOrganizationSummaryByNameReturns(
					v7action.OrganizationSummary{
						Organization: resources.Organization{
							Name: "some-org",
							GUID: "some-org-guid",
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
						DefaultIsolationSegmentGUID: "default-isolation-segment-guid",
					},
					v7action.Warnings{"warning-1", "warning-2"},
					nil)
			})

			When("API version is above isolation segments minimum version", func() {
				When("something", func() {
					BeforeEach(func() {
						fakeActor.GetIsolationSegmentsByOrganizationReturns(
							[]resources.IsolationSegment{
								{
									Name: "isolation-segment-1",
									GUID: "default-isolation-segment-guid",
								}, {
									Name: "isolation-segment-2",
									GUID: "some-other-isolation-segment-guid",
								},
							},
							v7action.Warnings{"warning-3", "warning-4"},
							nil)
					})

					It("displays warnings and a table with org domains, quota, spaces and isolation segments", func() {
						Expect(executeErr).To(BeNil())

						Expect(testUI.Out).To(Say(`Getting info for org %s as some-user\.\.\.`, cmd.RequiredArgs.Organization))
						Expect(testUI.Err).To(Say("warning-1"))
						Expect(testUI.Err).To(Say("warning-2"))
						Expect(testUI.Err).To(Say("warning-3"))
						Expect(testUI.Err).To(Say("warning-4"))

						Expect(testUI.Out).To(Say(`name:\s+%s`, cmd.RequiredArgs.Organization))
						Expect(testUI.Out).To(Say(`domains:\s+a-shared.com, b-private.com, c-shared.com, d-private.com`))
						Expect(testUI.Out).To(Say(`quota:\s+some-quota`))
						Expect(testUI.Out).To(Say(`spaces:\s+space1, space2`))
						Expect(testUI.Out).To(Say(`isolation segments:\s+isolation-segment-1 \(default\), isolation-segment-2`))

						Expect(fakeActor.GetCurrentUserCallCount()).To(Equal(1))

						Expect(fakeActor.GetOrganizationSummaryByNameCallCount()).To(Equal(1))
						orgName := fakeActor.GetOrganizationSummaryByNameArgsForCall(0)
						Expect(orgName).To(Equal("some-org"))

						Expect(fakeActor.GetIsolationSegmentsByOrganizationCallCount()).To(Equal(1))
						orgGuid := fakeActor.GetIsolationSegmentsByOrganizationArgsForCall(0)
						Expect(orgGuid).To(Equal("some-org-guid"))
					})
				})

				When("getting the org isolation segments returns an error", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("get org iso segs error")
						fakeActor.GetIsolationSegmentsByOrganizationReturns(
							nil,
							v7action.Warnings{"get iso seg warning"},
							expectedErr)
					})

					It("returns the error and all warnings", func() {
						Expect(executeErr).To(MatchError(expectedErr))
						Expect(testUI.Err).To(Say("get iso seg warning"))
					})
				})
			})
		})

		When("getting the current user returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("getting current user error")
				fakeActor.GetCurrentUserReturns(
					configv3.User{},
					expectedErr)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(expectedErr))
			})
		})

		When("getting the org summary returns an error", func() {
			When("the error is translatable", func() {
				BeforeEach(func() {
					fakeActor.GetOrganizationSummaryByNameReturns(
						v7action.OrganizationSummary{},
						v7action.Warnings{"warning-1", "warning-2"},
						actionerror.OrganizationNotFoundError{Name: "some-org"})
				})

				It("returns a translatable error and outputs all warnings", func() {
					Expect(executeErr).To(MatchError(actionerror.OrganizationNotFoundError{Name: "some-org"}))

					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))
				})
			})

			When("the error is not translatable", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("get org error")
					fakeActor.GetOrganizationSummaryByNameReturns(
						v7action.OrganizationSummary{},
						v7action.Warnings{"warning-1", "warning-2"},
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
})

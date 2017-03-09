package v2_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/v2"
	"code.cloudfoundry.org/cli/command/v2/shared"
	"code.cloudfoundry.org/cli/command/v2/v2fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("org Command", func() {
	var (
		cmd             v2.OrgCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v2fakes.FakeOrgActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v2fakes.FakeOrgActor)

		cmd = v2.OrgCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		cmd.RequiredArgs.Organization = "some-org"
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when checking the target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(
				sharedaction.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(
				command.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			config, targetedOrganizationRequired, targetedSpaceRequired := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(config).To(Equal(fakeConfig))
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
						v2action.OrganizationNotFoundError{Name: "some-org"})
				})

				It("returns a translatable error and outputs all warnings", func() {
					Expect(executeErr).To(MatchError(shared.OrganizationNotFoundError{Name: "some-org"}))

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
						Name: "some-org",
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

			It("displays warnings and a table with org domains, org quota and spaces", func() {
				Expect(executeErr).To(BeNil())

				Eventually(testUI.Out).Should(Say("Getting info for org %s as some-user\\.\\.\\.", cmd.RequiredArgs.Organization))
				Expect(testUI.Err).To(Say("warning-1"))
				Expect(testUI.Err).To(Say("warning-2"))

				Eventually(testUI.Out).Should(Say("name:\\s+%s", cmd.RequiredArgs.Organization))

				Eventually(testUI.Out).Should(Say("domains:\\s+a-shared.com, b-private.com, c-shared.com, d-private.com"))

				Eventually(testUI.Out).Should(Say("quota:\\s+some-quota"))

				Eventually(testUI.Out).Should(Say("spaces:\\s+space1, space2"))

				Expect(fakeConfig.CurrentUserCallCount()).To(Equal(1))

				Expect(fakeActor.GetOrganizationSummaryByNameCallCount()).To(Equal(1))
				orgName := fakeActor.GetOrganizationSummaryByNameArgsForCall(0)
				Expect(orgName).To(Equal("some-org"))
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
						v2action.OrganizationNotFoundError{Name: "some-org"})
				})

				It("returns a translatable error and outputs all warnings", func() {
					Expect(executeErr).To(MatchError(shared.OrganizationNotFoundError{Name: "some-org"}))

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
	})
})

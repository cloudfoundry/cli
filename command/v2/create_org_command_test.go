package v2_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	. "code.cloudfoundry.org/cli/command/v2"
	"code.cloudfoundry.org/cli/command/v2/v2fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("CreateOrgCommand", func() {
	var (
		fakeConfig      *commandfakes.FakeConfig
		fakeActor       *v2fakes.FakeCreateOrgActor
		fakeSharedActor *commandfakes.FakeSharedActor
		testUI          *ui.UI
		orgName         string
		cmd             CreateOrgCommand

		executeErr error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeActor = new(v2fakes.FakeCreateOrgActor)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		orgName = "some-org"

		cmd = CreateOrgCommand{
			UI:           testUI,
			Config:       fakeConfig,
			Actor:        fakeActor,
			SharedActor:  fakeSharedActor,
			RequiredArgs: flag.Organization{Organization: orgName},
		}

		fakeConfig.ExperimentalReturns(true)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when checking the target fails", func() {
		var binaryName string

		BeforeEach(func() {
			binaryName = "faceman"
			fakeConfig.BinaryNameReturns(binaryName)
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: "faceman"}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeFalse())
			Expect(checkTargetedSpace).To(BeFalse())
		})
	})

	Context("when checking the target succeeds", func() {
		Context("when fetching the current user succeeds", func() {
			var username string

			BeforeEach(func() {
				username = "some-guy"

				fakeConfig.CurrentUserReturns(configv3.User{
					Name: username,
				}, nil)
			})

			Context("when creating the org succeeds", func() {
				BeforeEach(func() {
					fakeActor.CreateOrganizationReturns(
						v2action.Organization{GUID: "fake-org-id"},
						v2action.Warnings{"warn-1", "warn-2"},
						nil,
					)
				})

				Context("when making the user an org manager succeeds", func() {
					BeforeEach(func() {
						fakeActor.GrantOrgManagerByUsernameReturns(
							v2action.Warnings{"warn-role"},
							nil,
						)
					})

					It("creates the org and displays warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(fakeConfig.CurrentUserCallCount()).To(Equal(1))
						Expect(testUI.Out).To(Say("Creating org %s as %s\\.\\.\\.", orgName, username))
						Expect(testUI.Err).To(Say("warn-1\nwarn-2\n"))
						Expect(testUI.Out).To(Say("OK\n\n"))
						Expect(testUI.Out).To(Say("Assigning role OrgManager to user %s in org %s\\.\\.\\.", username, orgName))
						Expect(testUI.Err).To(Say("warn-role\n"))
						Expect(testUI.Out).To(Say("OK\n\n"))
						Expect(testUI.Out).To(Say(`TIP: Use 'cf target -o "%s"' to target new org`, orgName))

						Expect(fakeActor.CreateOrganizationCallCount()).To(Equal(1))
						Expect(fakeActor.CreateOrganizationArgsForCall(0)).To(Equal("some-org"))
						Expect(fakeActor.GrantOrgManagerByUsernameCallCount()).To(Equal(1))
						orgID, inputUsername := fakeActor.GrantOrgManagerByUsernameArgsForCall(0)
						Expect(orgID).To(Equal("fake-org-id"))
						Expect(inputUsername).To(Equal(username))
					})
				})

				Context("when making the user an org manager fails", func() {
					It("returns an error and prints warnings", func() {

					})
				})
			})

			Context("when creating the org fails", func() {
				BeforeEach(func() {
					fakeActor.CreateOrganizationReturns(
						v2action.Organization{},
						v2action.Warnings{"warn-1", "warn-2"},
						errors.New("failed to create"),
					)
				})

				It("returns an error and prints warnings", func() {
					Expect(executeErr).To(MatchError("failed to create"))
					Expect(testUI.Err).To(Say("warn-1\nwarn-2"))
				})
			})
		})

		Context("when fetching the user fails", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{}, errors.New("gotta log in"))
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError("gotta log in"))
			})
		})
	})
})

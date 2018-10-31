package v6_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	. "code.cloudfoundry.org/cli/command/v6"
	"code.cloudfoundry.org/cli/command/v6/v6fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("CreateOrgCommand", func() {
	var (
		fakeConfig      *commandfakes.FakeConfig
		fakeActor       *v6fakes.FakeCreateOrgActor
		fakeSharedActor *commandfakes.FakeSharedActor
		testUI          *ui.UI
		orgName         string
		cmd             CreateOrgCommand

		executeErr error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeActor = new(v6fakes.FakeCreateOrgActor)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		orgName = "some-org"

		cmd = CreateOrgCommand{
			UI:           testUI,
			Config:       fakeConfig,
			Actor:        fakeActor,
			SharedActor:  fakeSharedActor,
			RequiredArgs: flag.Organization{Organization: orgName},
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("checking the target fails", func() {
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

	When("checking the target succeeds", func() {
		When("fetching the current user succeeds", func() {
			var username string

			BeforeEach(func() {
				username = "some-guy"

				fakeConfig.CurrentUserReturns(configv3.User{
					Name: username,
				}, nil)
			})

			When("the -q is passed", func() {
				var quotaName string

				BeforeEach(func() {
					quotaName = "some-quota-name"
					cmd.Quota = quotaName
				})

				It("provides the quota when creating the org", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					_, inputQuota := fakeActor.CreateOrganizationArgsForCall(0)
					Expect(inputQuota).To(Equal("some-quota-name"))
				})
			})

			When("creating the org succeeds", func() {
				BeforeEach(func() {
					fakeActor.CreateOrganizationReturns(
						v2action.Organization{GUID: "fake-org-id"},
						v2action.Warnings{"warn-1", "warn-2"},
						nil,
					)
				})

				It("creates the org and displays warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(fakeConfig.CurrentUserCallCount()).To(Equal(1))
					Expect(testUI.Out).To(Say(`Creating org %s as %s\.\.\.`, orgName, username))
					Expect(testUI.Err).To(Say("warn-1\nwarn-2\n"))
					Expect(testUI.Out).To(Say("OK\n\n"))

					Expect(fakeActor.CreateOrganizationCallCount()).To(Equal(1))
					inputOrg, quota := fakeActor.CreateOrganizationArgsForCall(0)
					Expect(inputOrg).To(Equal("some-org"))
					Expect(quota).To(BeEmpty())
				})

				When("making the user an org manager succeeds", func() {
					BeforeEach(func() {
						fakeActor.GrantOrgManagerByUsernameReturns(
							v2action.Warnings{"warn-role"},
							nil,
						)
					})

					It("displays warnings and a tip", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(fakeConfig.CurrentUserCallCount()).To(Equal(1))
						Expect(testUI.Out).To(Say(`Assigning role OrgManager to user %s in org %s\.\.\.`, username, orgName))
						Expect(testUI.Err).To(Say("warn-role\n"))
						Expect(testUI.Out).To(Say("OK\n\n"))
						Expect(testUI.Out).To(Say(`TIP: Use 'cf target -o "%s"' to target new org`, orgName))

						Expect(fakeActor.GrantOrgManagerByUsernameCallCount()).To(Equal(1))
						orgID, inputUsername := fakeActor.GrantOrgManagerByUsernameArgsForCall(0)
						Expect(orgID).To(Equal("fake-org-id"))
						Expect(inputUsername).To(Equal(username))
					})
				})

				When("making the user an org manager fails", func() {
					BeforeEach(func() {
						fakeActor.GrantOrgManagerByUsernameReturns(
							v2action.Warnings{"warn-role"},
							errors.New("some-error"),
						)
					})

					It("returns an error and prints warnings", func() {
						Expect(executeErr).To(MatchError("some-error"))
						Expect(testUI.Err).To(Say("warn-role\n"))
					})
				})
			})

			When("creating the org fails", func() {
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

			When("creating the org failed because the name was taken", func() {
				BeforeEach(func() {
					fakeActor.CreateOrganizationReturns(
						v2action.Organization{},
						v2action.Warnings{"warn-1", "warn-2"},
						actionerror.OrganizationNameTakenError{Name: orgName},
					)
				})

				It("should print warnings and return nil because this error is not fatal", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(testUI.Err).To(Say("warn-1\nwarn-2"))
					Expect(testUI.Err).To(Say("Org %s already exists", orgName))
					Expect(fakeActor.GrantOrgManagerByUsernameCallCount()).To(Equal(0))
				})
			})
		})

		When("fetching the user fails", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{}, errors.New("gotta log in"))
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError("gotta log in"))
			})
		})
	})
})

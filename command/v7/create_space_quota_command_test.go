package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("create-space-quota Command", func() {
	var (
		cmd             v7.CreateSpaceCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeCreateSpaceActor
		binaryName      string
		executeErr      error

		spaceName string
		orgName   string
		userName  string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeCreateSpaceActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		spaceName = "some-space"
		orgName = ""
		userName = "some-user-name"
	})

	JustBeforeEach(func() {
		cmd = v7.CreateSpaceCommand{
			UI:           testUI,
			Config:       fakeConfig,
			SharedActor:  fakeSharedActor,
			Actor:        fakeActor,
			RequiredArgs: flag.Space{Space: spaceName},
			Organization: orgName,
		}

		executeErr = cmd.Execute(nil)
	})

	When("the environment is not set up correctly", func() {
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

	When("the environment is setup correctly", func() {
		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: "some-org-name",
				GUID: "some-org-guid",
			})
			fakeConfig.CurrentUserReturns(configv3.User{
				Name:   userName,
				Origin: "some-user-origin",
			}, nil)
		})

		It("prints text indicating it is creating a space", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Out).To(Say(`Creating space %s in org %s as %s\.\.\.`, spaceName, "some-org-name", userName))
		})

		When("creating the space errors", func() {
			BeforeEach(func() {
				fakeActor.CreateSpaceReturns(
					v7action.Space{},
					v7action.Warnings{"warnings-1", "warnings-2"},
					errors.New("err-create-space"),
				)
			})

			It("returns an error and displays warnings", func() {
				Expect(executeErr).To(MatchError("err-create-space"))
				Expect(testUI.Err).To(Say("warnings-1"))
				Expect(testUI.Err).To(Say("warnings-2"))
			})
		})

		When("creating the space and setting roles is successful", func() {
			BeforeEach(func() {
				fakeActor.CreateSpaceReturns(
					v7action.Space{GUID: "some-space-guid"},
					v7action.Warnings{"warnings-1", "warnings-2"},
					nil,
				)

				fakeActor.CreateSpaceRoleReturnsOnCall(0,
					v7action.Warnings{"create-space-manager-role-warning"},
					nil,
				)

				fakeActor.CreateSpaceRoleReturnsOnCall(1,
					v7action.Warnings{"create-space-developer-role-warning"},
					nil,
				)
			})

			It("creates the space in the targeted organization", func() {
				Expect(fakeActor.GetOrganizationByNameCallCount()).To(Equal(0))
				Expect(fakeActor.CreateSpaceCallCount()).To(Equal(1))
				expectedSpaceName, expectedOrgGUID := fakeActor.CreateSpaceArgsForCall(0)
				Expect(expectedSpaceName).To(Equal(spaceName))
				Expect(expectedOrgGUID).To(Equal("some-org-guid"))
			})

			It("sets the user as a space manager", func() {
				Expect(fakeActor.CreateSpaceRoleCallCount()).To(Equal(2))
				givenRoleType, givenOrgGuid, givenSpaceGUID, givenUserName, givenOrigin, givenIsClient := fakeActor.CreateSpaceRoleArgsForCall(0)
				Expect(givenRoleType).To(Equal(constant.SpaceManagerRole))
				Expect(givenOrgGuid).To(Equal("some-org-guid"))
				Expect(givenSpaceGUID).To(Equal("some-space-guid"))
				Expect(givenUserName).To(Equal("some-user-name"))
				Expect(givenOrigin).To(Equal("some-user-origin"))
				Expect(givenIsClient).To(BeFalse())
			})

			It("sets the user as a space developer", func() {
				Expect(fakeActor.CreateSpaceRoleCallCount()).To(Equal(2))
				givenRoleType, givenOrgGuid, givenSpaceGUID, givenUserName, givenOrigin, givenIsClient := fakeActor.CreateSpaceRoleArgsForCall(1)
				Expect(givenRoleType).To(Equal(constant.SpaceDeveloperRole))
				Expect(givenOrgGuid).To(Equal("some-org-guid"))
				Expect(givenSpaceGUID).To(Equal("some-space-guid"))
				Expect(givenUserName).To(Equal("some-user-name"))
				Expect(givenOrigin).To(Equal("some-user-origin"))
				Expect(givenIsClient).To(BeFalse())
			})

			It("prints all warnings, text indicating creation completion, ok and then a tip", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(testUI.Err).To(Say("warnings-1"))
				Expect(testUI.Err).To(Say("warnings-2"))
				Expect(testUI.Err).To(Say("create-space-manager-role-warning"))
				Expect(testUI.Err).To(Say("create-space-developer-role-warning"))

				Expect(testUI.Out).To(Say("Creating space some-space in org some-org-name as some-user-name..."))
				Expect(testUI.Out).To(Say("OK"))

				Expect(testUI.Out).To(Say("Assigning role SpaceManager to user some-user-name in org some-org-name / space some-space as some-user-name..."))
				Expect(testUI.Out).To(Say("OK"))

				Expect(testUI.Out).To(Say("Assigning role SpaceDeveloper to user some-user-name in org some-org-name / space some-space as some-user-name..."))
				Expect(testUI.Out).To(Say("OK"))

				Expect(testUI.Out).To(Say(`TIP: Use 'cf target -o "%s" -s "%s"' to target new space`, "some-org-name", spaceName))
			})
		})

		When("passing in a organization", func() {
			When("the organization exists", func() {
				BeforeEach(func() {
					fakeActor.CreateSpaceReturns(v7action.Space{}, v7action.Warnings{"warnings-1", "warnings-2"}, nil)
					orgName = "some-other-org"
					fakeActor.GetOrganizationByNameReturns(
						v7action.Organization{Name: "some-other-org", GUID: "some-other-org-guid"},
						v7action.Warnings{"get-org-warnings"},
						nil,
					)
				})

				It("prints all warnings, ok and then a tip", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(testUI.Err).To(Say("get-org-warnings"))
					Expect(testUI.Err).To(Say("warnings-1"))
					Expect(testUI.Err).To(Say("warnings-2"))
					Expect(testUI.Out).To(Say("OK"))
					Expect(testUI.Out).To(Say(`TIP: Use 'cf target -o "%s" -s "%s"' to target new space`, orgName, spaceName))
				})

				It("creates the space", func() {
					Expect(fakeActor.GetOrganizationByNameCallCount()).To(Equal(1))
					expectedOrgName := fakeActor.GetOrganizationByNameArgsForCall(0)
					Expect(expectedOrgName).To(Equal(orgName))

					Expect(fakeActor.CreateSpaceCallCount()).To(Equal(1))
					expectedSpaceName, expectedOrgGUID := fakeActor.CreateSpaceArgsForCall(0)
					Expect(expectedSpaceName).To(Equal(spaceName))
					Expect(expectedOrgGUID).To(Equal("some-other-org-guid"))
				})
			})
			When("the organization doesn't exist", func() {
				BeforeEach(func() {
					orgName = "some-other-org"
					fakeActor.GetOrganizationByNameReturns(
						v7action.Organization{},
						v7action.Warnings{"get-org-warnings"},
						errors.New("get-organization-error"),
					)
				})

				It("prints all warnings, ok and then a tip", func() {
					Expect(executeErr).To(MatchError(errors.New("get-organization-error")))
				})

				It(" does not create the space", func() {
					Expect(fakeActor.GetOrganizationByNameCallCount()).To(Equal(1))
					expectedOrgName := fakeActor.GetOrganizationByNameArgsForCall(0)
					Expect(expectedOrgName).To(Equal(orgName))

					Expect(fakeActor.CreateSpaceCallCount()).To(Equal(0))
				})
			})
		})

		When("the space already exists", func() {
			BeforeEach(func() {
				fakeActor.CreateSpaceReturns(v7action.Space{}, v7action.Warnings{"some-warning"}, actionerror.SpaceAlreadyExistsError{Space: spaceName})
			})

			It("displays all warnings, that the space already exists, and does not error", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Err).To(Say("some-warning"))
				Expect(testUI.Out).To(Say(`Creating space %s in org %s as %s\.\.\.`, spaceName, "some-org-name", userName))
				Expect(testUI.Out).To(Say(`Space '%s' already exists\.`, spaceName))
				Expect(testUI.Out).To(Say("OK"))
			})
		})

		When("setting the space manager role fails", func() {
			BeforeEach(func() {
				fakeActor.CreateSpaceRoleReturnsOnCall(0,
					v7action.Warnings{"create-space-manager-role-warning"},
					errors.New("create-space-manager-role-error"),
				)
			})

			It("displays warnings and returns the error", func() {
				Expect(testUI.Err).To(Say("create-space-manager-role-warning"))
				Expect(executeErr).To(MatchError("create-space-manager-role-error"))
			})
		})

		When("setting the space developer role fails", func() {
			BeforeEach(func() {
				fakeActor.CreateSpaceRoleReturnsOnCall(1,
					v7action.Warnings{"create-space-developer-role-warning"},
					errors.New("create-space-developer-role-error"),
				)
			})

			It("displays warnings and returns the error", func() {
				Expect(testUI.Err).To(Say("create-space-developer-role-warning"))
				Expect(executeErr).To(MatchError("create-space-developer-role-error"))
			})
		})
	})
})

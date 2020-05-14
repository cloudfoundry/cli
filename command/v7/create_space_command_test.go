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
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("create-space Command", func() {
	var (
		cmd             v7.CreateSpaceCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		binaryName      string
		executeErr      error

		spaceName string
		spaceGUID string
		orgGUID   string
		orgName   string
		userName  string
		quotaName string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		spaceName = "some-space"
		spaceGUID = "some-space-guid"
		orgGUID = "some-org-guid"
		orgName = ""
		quotaName = ""
		userName = "some-user-name"

		fakeConfig.TargetedOrganizationReturns(configv3.Organization{
			Name: "some-org-name",
			GUID: "some-org-guid",
		})
		fakeConfig.CurrentUserReturns(configv3.User{
			Name:   userName,
			Origin: "some-user-origin",
		}, nil)
		fakeActor.CreateSpaceReturns(v7action.Space{
			Name: spaceName,
			GUID: spaceGUID,
		}, v7action.Warnings{}, nil)

	})

	JustBeforeEach(func() {
		cmd = v7.CreateSpaceCommand{
			BaseCommand: v7.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
			RequiredArgs: flag.Space{Space: spaceName},
			Organization: orgName,
			Quota:        quotaName,
		}

		executeErr = cmd.Execute(nil)
	})

	When("the environment is not set up correctly (CheckTarget fails)", func() {
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

	When("passing in a organization", func() {
		When("getting the passed-in organization returns warnings", func() {
			BeforeEach(func() {
				orgName = "some-other-org"
				fakeActor.GetOrganizationByNameReturns(
					resources.Organization{},
					v7action.Warnings{"get-org-warnings"},
					nil,
				)
			})

			It("prints all warnings", func() {
				Expect(testUI.Err).To(Say("get-org-warnings"))
			})
		})

		When("the organization exists", func() {
			BeforeEach(func() {
				orgName = "some-other-org"
				fakeActor.GetOrganizationByNameReturns(
					resources.Organization{Name: "some-other-org", GUID: "some-other-org-guid"},
					v7action.Warnings{},
					nil,
				)
			})

			It("does not error", func() {
				Expect(executeErr).ToNot(HaveOccurred())
			})

			It("gets the org and passes it into the CreateSpaceActor", func() {
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
					resources.Organization{},
					v7action.Warnings{},
					errors.New("get-organization-error"),
				)
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(errors.New("get-organization-error")))
			})

			It("does not create the space", func() {
				Expect(fakeActor.GetOrganizationByNameCallCount()).To(Equal(1))
				expectedOrgName := fakeActor.GetOrganizationByNameArgsForCall(0)
				Expect(expectedOrgName).To(Equal(orgName))

				Expect(fakeActor.CreateSpaceCallCount()).To(Equal(0))
			})
		})
	})

	It("prints text indicating it is creating a space", func() {
		Expect(testUI.Out).To(Say(`Creating space %s in org %s as %s\.\.\.`, spaceName, "some-org-name", userName))
	})

	When("creating the space returns warnings", func() {
		BeforeEach(func() {
			fakeActor.CreateSpaceReturns(
				v7action.Space{},
				v7action.Warnings{"warnings-1", "warnings-2"},
				nil,
			)
		})
		It("prints the warnings", func() {
			Expect(testUI.Err).To(Say("warnings-1"))
			Expect(testUI.Err).To(Say("warnings-2"))
		})
	})

	When("the space already exists", func() {
		BeforeEach(func() {
			fakeActor.CreateSpaceReturns(v7action.Space{}, v7action.Warnings{}, actionerror.SpaceAlreadyExistsError{Space: spaceName})
		})

		It("displays that the space already exists, and does not error", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say(`Space '%s' already exists\.`, spaceName))
			Expect(testUI.Out).To(Say("OK"))
		})
	})

	When("creating the space errors", func() {
		BeforeEach(func() {
			fakeActor.CreateSpaceReturns(
				v7action.Space{},
				v7action.Warnings{},
				errors.New("err-create-space"),
			)
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError("err-create-space"))
		})
	})

	When("passing in a quota", func() {
		BeforeEach(func() {
			quotaName = "some-quota"
		})

		It("prints text indicating it is applying a quota to the space", func() {
			Expect(testUI.Out).To(Say(`Setting space quota %s to space %s as %s\.\.\.`, quotaName, spaceName, userName))
		})

		When("setting the quota onto the space returns warnings", func() {
			BeforeEach(func() {
				fakeActor.ApplySpaceQuotaByNameReturns(
					v7action.Warnings{"quota-warnings-1", "quota-warnings-2"},
					nil,
				)
			})

			It("prints the warnings", func() {
				Expect(testUI.Err).To(Say("quota-warnings-1"))
				Expect(testUI.Err).To(Say("quota-warnings-2"))
			})
		})

		When("the quota does not exist", func() {
			BeforeEach(func() {
				fakeActor.ApplySpaceQuotaByNameReturns(v7action.Warnings{}, actionerror.SpaceQuotaNotFoundForNameError{Name: quotaName})
			})

			It("returns an error and displays warnings", func() {
				Expect(executeErr).To(MatchError(actionerror.SpaceQuotaNotFoundForNameError{Name: "some-quota"}))
				Expect(fakeActor.ApplySpaceQuotaByNameCallCount()).To(Equal(1))
				Expect(fakeActor.CreateSpaceRoleCallCount()).To(Equal(0))
			})
		})

		When("the quota exists", func() {
			BeforeEach(func() {
				fakeActor.ApplySpaceQuotaByNameReturns(v7action.Warnings{}, nil)
			})

			It("calls ApplySpaceQuotaByName() with correct parameters", func() {
				quota, passedSpaceGUID, passedOrgGUID := fakeActor.ApplySpaceQuotaByNameArgsForCall(0)
				Expect(quota).To(Equal(quotaName))
				Expect(passedSpaceGUID).To(Equal(spaceGUID))
				Expect(passedOrgGUID).To(Equal(orgGUID))
			})

			It("does not return an error and displays warnings", func() {
				Expect(testUI.Out).To(Say("OK")) // create space
				Expect(testUI.Out).To(Say("OK")) // apply quota
				Expect(testUI.Out).To(Say("OK")) // assign spaceManager
				Expect(testUI.Out).To(Say("OK")) // assign spaceDeveloper
				Expect(executeErr).To(Not(HaveOccurred()))
			})
		})
	})

	It("prints that it is assigning roles to the current user", func() {
		Expect(testUI.Out).To(Say(`Assigning role SpaceManager to user %s in org %s / space %s as %s\.\.\.`, userName, "some-org-name", spaceName, userName))
		Expect(testUI.Out).To(Say(`Assigning role SpaceDeveloper to user %s in org %s / space %s as %s\.\.\.`, userName, "some-org-name", spaceName, userName))
	})

	When("setting roles returns warnings", func() {
		BeforeEach(func() {
			fakeActor.CreateSpaceRoleReturnsOnCall(0,
				v7action.Warnings{"create-space-manager-role-warning"},
				nil,
			)

			fakeActor.CreateSpaceRoleReturnsOnCall(1,
				v7action.Warnings{"create-space-developer-role-warning"},
				nil,
			)
		})

		It("displays the warnings", func() {
			Expect(testUI.Err).To(Say("create-space-manager-role-warning"))
			Expect(testUI.Err).To(Say("create-space-developer-role-warning"))
		})
	})

	When("setting the space manager role fails", func() {
		BeforeEach(func() {
			fakeActor.CreateSpaceRoleReturnsOnCall(0,
				v7action.Warnings{},
				errors.New("create-space-manager-role-error"),
			)
		})

		It("displays warnings and returns the error", func() {
			Expect(executeErr).To(MatchError("create-space-manager-role-error"))
		})
	})

	When("setting the space developer role fails", func() {
		BeforeEach(func() {
			fakeActor.CreateSpaceRoleReturnsOnCall(1,
				v7action.Warnings{},
				errors.New("create-space-developer-role-error"),
			)
		})

		It("displays warnings and returns the error", func() {
			Expect(executeErr).To(MatchError("create-space-developer-role-error"))
		})
	})

	When("setting roles is successful", func() {
		BeforeEach(func() {
			fakeActor.CreateSpaceRoleReturnsOnCall(0,
				v7action.Warnings{},
				nil,
			)

			fakeActor.CreateSpaceRoleReturnsOnCall(1,
				v7action.Warnings{},
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

		It("prints ok and then a tip", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(testUI.Out).To(Say("OK")) // create space
			Expect(testUI.Out).To(Say("OK")) // assign spaceManager
			Expect(testUI.Out).To(Say("OK")) // assign spaceDeveloper

			Expect(testUI.Out).To(Say(`TIP: Use 'cf target -o "%s" -s "%s"' to target new space`, "some-org-name", spaceName))
		})
	})
})

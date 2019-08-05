package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/util/configv3"

	//"code.cloudfoundry.org/cli/actor/v7action"
	//"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/commandfakes"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
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
		fakeActor       *v7fakes.FakeCreateSpaceActor
		binaryName      string
		executeErr      error

		spaceName string
		orgName   string
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
			fakeConfig.CurrentUserReturns(configv3.User{Name: "the-user"}, nil)
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: "some-org",
				GUID: "some-org-guid",
			})
		})

		It("prints text indicating it is creating a space", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Out).To(Say(`Creating space %s in org %s as the-user\.\.\.`, spaceName, "some-org"))
		})

		When("creating the space errors", func() {
			BeforeEach(func() {
				fakeActor.CreateSpaceReturns(v7action.Space{}, v7action.Warnings{"warnings-1", "warnings-2"}, errors.New("err-create-space"))
			})

			It("returns an error and displays warnings", func() {
				Expect(executeErr).To(MatchError("err-create-space"))
				Expect(testUI.Err).To(Say("warnings-1"))
				Expect(testUI.Err).To(Say("warnings-2"))
			})
		})

		When("creating the space is successful", func() {
			BeforeEach(func() {
				fakeActor.CreateSpaceReturns(v7action.Space{}, v7action.Warnings{"warnings-1", "warnings-2"}, nil)
			})

			//TODO: modify or remove these tests upon set-space-role implementation. they are included in the tests commented out below
			It("creates the space in the targeted organization", func() {
				Expect(fakeActor.GetOrganizationByNameCallCount()).To(Equal(0))
				Expect(fakeActor.CreateSpaceCallCount()).To(Equal(1))
				expectedSpaceName, expectedOrgGUID := fakeActor.CreateSpaceArgsForCall(0)
				Expect(expectedSpaceName).To(Equal(spaceName))
				Expect(expectedOrgGUID).To(Equal("some-org-guid"))
			})

			It("prints all warnings, text indicating creation completion, ok and then a tip", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(testUI.Err).To(Say("warnings-1"))
				Expect(testUI.Err).To(Say("warnings-2"))
				Expect(testUI.Out).To(Say("OK"))
				Expect(testUI.Out).To(Say(`TIP: Use 'cf target -o "%s" -s "%s"' to target new space`, "some-org", spaceName))
			})

			//TODO: add these tests back once V7/V3 set-space-role is implemented and included in create-space
			//When("setting the user as a space manager is successful", func() {
			//	BeforeEach(func() {
			//		fakeActor.GrantSpaceManagerByUsernameReturns(v2action.Warnings{"set-space-manager-warning"}, nil)
			//	})
			//
			//	It("creates the space in the targeted organization", func() {
			//		Expect(fakeActor.GetOrganizationByNameCallCount()).To(Equal(0))
			//		Expect(fakeActor.CreateSpaceCallCount()).To(Equal(1))
			//		expectedSpaceName, expectedOrgGUID := fakeActor.CreateSpaceArgsForCall(0)
			//		Expect(expectedSpaceName).To(Equal(spaceName))
			//		Expect(expectedOrgGUID).To(Equal("some-org-guid"))
			//	})
			//
			//	It("returns warnings for creating the space", func() {
			//		Expect(testUI.Err).To(Say("warnings-1"))
			//		Expect(testUI.Err).To(Say("warnings-2"))
			//	})

			//It("sets the user as a space manager", func() {
			//	Expect(fakeActor.GrantSpaceManagerByUsernameCallCount()).To(Equal(1))
			//	expectedOrgGUID, expectedSpaceGUID, userName := fakeActor.GrantSpaceManagerByUsernameArgsForCall(0)
			//	Expect(expectedSpaceGUID).To(Equal("some-space-guid"))
			//	Expect(expectedOrgGUID).To(Equal("some-org-guid"))
			//	Expect(userName).To(Equal("the-user"))
			//})
			//
			//	When("setting the user as a space developer is successful", func() {
			//		BeforeEach(func() {
			//			fakeActor.GrantSpaceDeveloperByUsernameReturns(v2action.Warnings{"set-space-developer-warning"}, nil)
			//		})
			//
			//		It("sets the user as a space developer", func() {
			//			Expect(fakeActor.GrantSpaceDeveloperByUsernameCallCount()).To(Equal(1))
			//			expectedSpaceGUID, userName := fakeActor.GrantSpaceDeveloperByUsernameArgsForCall(0)
			//			Expect(expectedSpaceGUID).To(Equal("some-space-guid"))
			//			Expect(userName).To(Equal("the-user"))
			//		})
			//
			//		It("prints all warnings, text indicating creation completion, ok and then a tip", func() {
			//			Expect(executeErr).ToNot(HaveOccurred())
			//			Expect(testUI.Err).To(Say("set-space-manager-warning"))
			//			Expect(testUI.Err).To(Say("set-space-developer-warning"))
			//			Expect(testUI.Out).To(Say("OK"))
			//			Expect(testUI.Out).To(Say(`TIP: Use 'cf target -o "%s" -s "%s"' to target new space`, "some-org", spaceName))
			//		})
			//	})
			//	When("setting the user as a space developer fails", func() {
			//		BeforeEach(func() {
			//			fakeActor.GrantSpaceDeveloperByUsernameReturns(v2action.Warnings{"set-space-developer-warning"}, errors.New("set-space-developer-error"))
			//		})
			//
			//		It("doesn't set the user as a space developer", func() {
			//			Expect(fakeActor.GrantSpaceDeveloperByUsernameCallCount()).To(Equal(1))
			//			expectedSpaceGUID, userName := fakeActor.GrantSpaceDeveloperByUsernameArgsForCall(0)
			//			Expect(expectedSpaceGUID).To(Equal("some-space-guid"))
			//			Expect(userName).To(Equal("the-user"))
			//			Expect(testUI.Err).To(Say("set-space-manager-warning"))
			//			Expect(testUI.Err).To(Say("set-space-developer-warning"))
			//			Expect(executeErr).To(MatchError(errors.New("set-space-developer-error")))
			//		})
			//	})
			//})
			//When("setting the user as a space manager fails", func() {
			//	BeforeEach(func() {
			//		fakeActor.GrantSpaceManagerByUsernameReturns(v2action.Warnings{"set-space-manager-warning"}, errors.New("set-space-manager-error"))
			//	})
			//
			//	It("fails to set the user as a space manager", func() {
			//		Expect(fakeActor.GrantSpaceManagerByUsernameCallCount()).To(Equal(1))
			//		Expect(fakeActor.GrantSpaceDeveloperByUsernameCallCount()).To(Equal(0))
			//		expectedOrgGUID, expectedSpaceGUID, userName := fakeActor.GrantSpaceManagerByUsernameArgsForCall(0)
			//		Expect(expectedSpaceGUID).To(Equal("some-space-guid"))
			//		Expect(expectedOrgGUID).To(Equal("some-org-guid"))
			//		Expect(userName).To(Equal("the-user"))
			//		Expect(testUI.Err).To(Say("set-space-manager-warning"))
			//		Expect(executeErr).To(MatchError(errors.New("set-space-manager-error")))
			//	})
			//})
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
				Expect(testUI.Out).To(Say(`Creating space %s in org %s as the-user\.\.\.`, spaceName, "some-org"))
				Expect(testUI.Out).To(Say(`Space '%s' already exists\.`, spaceName))
				Expect(testUI.Out).To(Say("OK"))
			})
		})
	})

})

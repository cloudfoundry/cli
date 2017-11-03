package v2_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v2"
	"code.cloudfoundry.org/cli/command/v2/v2fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("security-groups Command", func() {
	var (
		cmd             SecurityGroupsCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v2fakes.FakeSecurityGroupsActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v2fakes.FakeSecurityGroupsActor)

		cmd = SecurityGroupsCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		fakeConfig.ExperimentalReturns(true)

		fakeConfig.CurrentUserReturns(configv3.User{Name: "some-user"}, nil)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when getting the current user fails", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = errors.New("getting user failed")
			fakeConfig.CurrentUserReturns(configv3.User{}, expectedErr)
		})

		It("returns the error", func() {
			Expect(executeErr).To(MatchError(expectedErr))
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

	Context("when the API version is low enough not to support fetching staging", func() {
		BeforeEach(func() {
			fakeActor.CloudControllerAPIVersionReturns("2.36.0")
		})

		It("makes the fetch indicating that staging should not be included", func() {
			Expect(executeErr).NotTo(HaveOccurred())

			Expect(fakeActor.CloudControllerAPIVersionCallCount()).To(Equal(1))
			Expect(fakeActor.GetSecurityGroupsWithOrganizationSpaceAndLifecycleCallCount()).To(Equal(1))
			Expect(fakeActor.GetSecurityGroupsWithOrganizationSpaceAndLifecycleArgsForCall(0)).To(BeFalse())

			Expect(fakeActor.GetSecurityGroupsWithOrganizationSpaceAndLifecycleCallCount()).To(Equal(1))
		})
	})

	Context("when the API version is high enough to support fetching staging", func() {
		BeforeEach(func() {
			fakeActor.CloudControllerAPIVersionReturns(ccversion.MinVersionLifecyleStagingV2)
		})

		Context("when the list of security groups is returned", func() {
			var secGroups []v2action.SecurityGroupWithOrganizationSpaceAndLifecycle

			BeforeEach(func() {
				secGroups = []v2action.SecurityGroupWithOrganizationSpaceAndLifecycle{
					{
						SecurityGroup: &v2action.SecurityGroup{Name: "seg-group-1"},
						Organization:  &v2action.Organization{Name: "org-11"},
						Space:         &v2action.Space{Name: "space-111"},
						Lifecycle:     ccv2.SecurityGroupLifecycleRunning,
					},
					{
						SecurityGroup: &v2action.SecurityGroup{Name: "seg-group-1"},
						Organization:  &v2action.Organization{Name: "org-12"},
						Space:         &v2action.Space{Name: "space-121"},
						Lifecycle:     ccv2.SecurityGroupLifecycleRunning,
					},
					{
						SecurityGroup: &v2action.SecurityGroup{Name: "seg-group-1"},
						Organization:  &v2action.Organization{Name: "org-12"},
						Space:         &v2action.Space{Name: "space-122"},
						Lifecycle:     ccv2.SecurityGroupLifecycleStaging,
					},
					{
						SecurityGroup: &v2action.SecurityGroup{Name: "seg-group-2"},
						Organization:  &v2action.Organization{},
						Space:         &v2action.Space{},
					},
					{
						SecurityGroup: &v2action.SecurityGroup{Name: "seg-group-3"},
						Organization:  &v2action.Organization{Name: "org-31"},
						Space:         &v2action.Space{Name: "space-311"},
						Lifecycle:     ccv2.SecurityGroupLifecycleRunning,
					},
					{
						SecurityGroup: &v2action.SecurityGroup{
							Name:           "seg-group-4",
							RunningDefault: true,
						},
						Organization: &v2action.Organization{Name: ""},
						Space:        &v2action.Space{Name: ""},
						Lifecycle:    ccv2.SecurityGroupLifecycleRunning,
					},
					{
						SecurityGroup: &v2action.SecurityGroup{
							Name:           "seg-group-4",
							StagingDefault: true,
						},
						Organization: &v2action.Organization{Name: ""},
						Space:        &v2action.Space{Name: ""},
						Lifecycle:    ccv2.SecurityGroupLifecycleStaging,
					},
				}
				fakeActor.GetSecurityGroupsWithOrganizationSpaceAndLifecycleReturns(secGroups, v2action.Warnings{"warning-1", "warning-2"}, nil)
			})

			It("displays a table containing the security groups, the spaces to which they are bound, the spaces' orgs, and the lifecycle of the app they were assigned to", func() {
				Expect(executeErr).To(BeNil())

				Expect(fakeActor.CloudControllerAPIVersionCallCount()).To(Equal(1))
				Expect(fakeActor.GetSecurityGroupsWithOrganizationSpaceAndLifecycleCallCount()).To(Equal(1))
				Expect(fakeActor.GetSecurityGroupsWithOrganizationSpaceAndLifecycleArgsForCall(0)).To(BeTrue())

				Expect(fakeActor.GetSecurityGroupsWithOrganizationSpaceAndLifecycleCallCount()).To(Equal(1))

				Expect(testUI.Out).To(Say("Getting security groups as some-user\\.\\.\\."))
				Expect(testUI.Out).To(Say("OK\\n\\n"))
				Expect(testUI.Out).To(Say("\\s+name\\s+organization\\s+space\\s+lifecycle"))
				Expect(testUI.Out).To(Say("#0\\s+seg-group-1\\s+org-11\\s+space-111\\s+running"))
				Expect(testUI.Out).To(Say("(?m)\\s+seg-group-1\\s+org-12\\s+space-121\\s+running"))
				Expect(testUI.Out).To(Say("(?m)\\s+seg-group-1\\s+org-12\\s+space-122\\s+staging"))
				Expect(testUI.Out).To(Say("#1\\s+seg-group-2\\s+"))
				Expect(testUI.Out).To(Say("#2\\s+seg-group-3\\s+org-31\\s+space-311\\s+running"))
				Expect(testUI.Out).To(Say("#3\\s+seg-group-4\\s+<all>\\s+<all>\\s+running"))
				Expect(testUI.Out).To(Say("(?m)\\s+seg-group-4\\s+<all>\\s+<all>\\s+staging"))
				Expect(testUI.Err).To(Say("warning-1"))
				Expect(testUI.Err).To(Say("warning-2"))
			})
		})

		Context("when an error is encountered fetching the security groups", func() {
			BeforeEach(func() {
				fakeActor.GetSecurityGroupsWithOrganizationSpaceAndLifecycleReturns(nil, v2action.Warnings{"warning-1", "warning-2"}, errors.New("generic"))
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("generic"))

				Expect(testUI.Err).To(Say("warning-1"))
				Expect(testUI.Err).To(Say("warning-2"))
			})
		})
	})
})

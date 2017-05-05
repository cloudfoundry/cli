package v2action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Security Group Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil)
	})

	Describe("GetSecurityGroupsWithOrganizationAndSpace", func() {
		var (
			secGroupOrgSpaces []SecurityGroupWithOrganizationAndSpace
			warnings          Warnings
			err               error
		)

		JustBeforeEach(func() {
			secGroupOrgSpaces, warnings, err = actor.GetSecurityGroupsWithOrganizationAndSpace()
		})

		Context("when an error occurs getting security groups", func() {
			var returnedError error

			BeforeEach(func() {
				returnedError = errors.New("get-security-groups-error")
				fakeCloudControllerClient.GetSecurityGroupsReturns(
					nil,
					ccv2.Warnings{"warning-1", "warning-2"},
					returnedError,
				)
			})

			It("returns the error and all warnings", func() {
				Expect(err).To(MatchError(returnedError))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})

		Context("when an error occurs getting spaces", func() {
			var returnedError error

			BeforeEach(func() {
				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]ccv2.SecurityGroup{
						{
							GUID: "security-group-guid-1",
							Name: "security-group-1",
						},
					},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil,
				)
				returnedError = errors.New("get-spaces-error")
				fakeCloudControllerClient.GetSpacesBySecurityGroupReturns(
					nil,
					ccv2.Warnings{"warning-3", "warning-4"},
					returnedError,
				)
			})

			It("returns the error and all warnings", func() {
				Expect(err).To(MatchError(returnedError))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3", "warning-4"))
				Expect(fakeCloudControllerClient.GetSecurityGroupsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetSecurityGroupsArgsForCall(0)).To(BeNil())
				Expect(fakeCloudControllerClient.GetSpacesBySecurityGroupCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetSpacesBySecurityGroupArgsForCall(0)).To(Equal("security-group-guid-1"))
			})
		})

		Context("when an error occurs getting an organization", func() {
			var returnedError error

			BeforeEach(func() {
				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]ccv2.SecurityGroup{
						{
							GUID: "security-group-guid-1",
							Name: "security-group-1",
						},
					},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil,
				)
				fakeCloudControllerClient.GetSpacesBySecurityGroupReturns(
					[]ccv2.Space{
						{
							GUID:             "space-guid-11",
							Name:             "space-11",
							OrganizationGUID: "org-guid-11",
						},
					},
					ccv2.Warnings{"warning-3", "warning-4"},
					nil,
				)
				returnedError = errors.New("get-org-error")
				fakeCloudControllerClient.GetOrganizationReturns(
					ccv2.Organization{},
					ccv2.Warnings{"warning-5", "warning-6"},
					returnedError,
				)
			})

			It("returns the error and all warnings", func() {
				Expect(err).To(MatchError(returnedError))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3", "warning-4", "warning-5", "warning-6"))
				Expect(fakeCloudControllerClient.GetSecurityGroupsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetSecurityGroupsArgsForCall(0)).To(BeNil())
				Expect(fakeCloudControllerClient.GetSpacesBySecurityGroupCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetSpacesBySecurityGroupArgsForCall(0)).To(Equal("security-group-guid-1"))
				Expect(fakeCloudControllerClient.GetOrganizationCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetOrganizationArgsForCall(0)).To(Equal("org-guid-11"))
			})
		})

		Context("when no errors are encountered", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]ccv2.SecurityGroup{
						{
							GUID: "security-group-guid-1",
							Name: "security-group-1",
						},
						{
							GUID: "security-group-guid-2",
							Name: "security-group-2",
						},
						{
							GUID: "security-group-guid-3",
							Name: "security-group-3",
						},
						{
							GUID: "security-group-guid-4",
							Name: "security-group-4",
						},
					},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil,
				)
				fakeCloudControllerClient.GetSpacesBySecurityGroupReturnsOnCall(0,
					[]ccv2.Space{
						{
							GUID:             "space-guid-13",
							Name:             "space-13",
							OrganizationGUID: "org-guid-13",
						},
						{
							GUID:             "space-guid-12",
							Name:             "space-12",
							OrganizationGUID: "org-guid-12",
						},
						{
							GUID:             "space-guid-11",
							Name:             "space-11",
							OrganizationGUID: "org-guid-11",
						},
					},
					ccv2.Warnings{"warning-3", "warning-4"},
					nil,
				)
				fakeCloudControllerClient.GetSpacesBySecurityGroupReturnsOnCall(1,
					[]ccv2.Space{
						{
							GUID:             "space-guid-21",
							Name:             "space-21",
							OrganizationGUID: "org-guid-21",
						},
						{
							GUID:             "space-guid-23",
							Name:             "space-23",
							OrganizationGUID: "org-guid-23",
						},
						{
							GUID:             "space-guid-22",
							Name:             "space-22",
							OrganizationGUID: "org-guid-11",
						},
					},
					ccv2.Warnings{"warning-5", "warning-6"},
					nil,
				)
				fakeCloudControllerClient.GetSpacesBySecurityGroupReturnsOnCall(2,
					[]ccv2.Space{},
					ccv2.Warnings{"warning-7", "warning-8"},
					nil,
				)
				fakeCloudControllerClient.GetSpacesBySecurityGroupReturnsOnCall(3,
					[]ccv2.Space{
						{
							GUID:             "space-guid-31",
							Name:             "space-31",
							OrganizationGUID: "org-guid-23",
						},
						{
							GUID:             "space-guid-32",
							Name:             "space-32",
							OrganizationGUID: "org-guid-11",
						},
						{
							GUID:             "space-guid-33",
							Name:             "space-33",
							OrganizationGUID: "org-guid-33",
						},
					},
					ccv2.Warnings{"warning-9", "warning-10"},
					nil,
				)
				fakeCloudControllerClient.GetOrganizationReturnsOnCall(0,
					ccv2.Organization{
						GUID: "org-guid-13",
						Name: "org-13",
					},
					ccv2.Warnings{"warning-11", "warning-12"},
					nil,
				)
				fakeCloudControllerClient.GetOrganizationReturnsOnCall(1,
					ccv2.Organization{
						GUID: "org-guid-12",
						Name: "org-12",
					},
					ccv2.Warnings{"warning-13", "warning-14"},
					nil,
				)
				fakeCloudControllerClient.GetOrganizationReturnsOnCall(2,
					ccv2.Organization{
						GUID: "org-guid-11",
						Name: "org-11",
					},
					ccv2.Warnings{"warning-15", "warning-16"},
					nil,
				)
				fakeCloudControllerClient.GetOrganizationReturnsOnCall(3,
					ccv2.Organization{
						GUID: "org-guid-21",
						Name: "org-21",
					},
					ccv2.Warnings{"warning-17", "warning-18"},
					nil,
				)
				fakeCloudControllerClient.GetOrganizationReturnsOnCall(4,
					ccv2.Organization{
						GUID: "org-guid-23",
						Name: "org-23",
					},
					ccv2.Warnings{"warning-19", "warning-20"},
					nil,
				)
				fakeCloudControllerClient.GetOrganizationReturnsOnCall(5,
					ccv2.Organization{
						GUID: "org-guid-33",
						Name: "org-33",
					},
					ccv2.Warnings{"warning-25", "warning-26"},
					nil,
				)
			})

			It("returns a slice of SecurityGroupWithOrganizationAndSpace and all warnings", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf(
					"warning-1", "warning-2",
					"warning-3", "warning-4",
					"warning-5", "warning-6",
					"warning-7", "warning-8",
					"warning-9", "warning-10",
					"warning-11", "warning-12",
					"warning-13", "warning-14",
					"warning-15", "warning-16",
					"warning-17", "warning-18",
					"warning-19", "warning-20",
					"warning-25", "warning-26",
				))
				expected := []SecurityGroupWithOrganizationAndSpace{
					{
						SecurityGroup: &SecurityGroup{
							GUID: "security-group-guid-1",
							Name: "security-group-1",
						},
						Organization: &Organization{
							GUID: "org-guid-11",
							Name: "org-11",
						},
						Space: &Space{
							GUID: "space-guid-11",
							Name: "space-11",
						},
					},
					{
						SecurityGroup: &SecurityGroup{
							GUID: "security-group-guid-1",
							Name: "security-group-1",
						},
						Organization: &Organization{
							GUID: "org-guid-12",
							Name: "org-12",
						},
						Space: &Space{
							GUID: "space-guid-12",
							Name: "space-12",
						},
					},
					{
						SecurityGroup: &SecurityGroup{
							GUID: "security-group-guid-1",
							Name: "security-group-1",
						},
						Organization: &Organization{
							GUID: "org-guid-13",
							Name: "org-13",
						},
						Space: &Space{
							GUID: "space-guid-13",
							Name: "space-13",
						},
					},
					{
						SecurityGroup: &SecurityGroup{
							GUID: "security-group-guid-2",
							Name: "security-group-2",
						},
						Organization: &Organization{
							GUID: "org-guid-11",
							Name: "org-11",
						},
						Space: &Space{
							GUID: "space-guid-22",
							Name: "space-22",
						},
					},
					{
						SecurityGroup: &SecurityGroup{
							GUID: "security-group-guid-2",
							Name: "security-group-2",
						},
						Organization: &Organization{
							GUID: "org-guid-21",
							Name: "org-21",
						},
						Space: &Space{
							GUID: "space-guid-21",
							Name: "space-21",
						},
					},
					{
						SecurityGroup: &SecurityGroup{
							GUID: "security-group-guid-2",
							Name: "security-group-2",
						},
						Organization: &Organization{
							GUID: "org-guid-23",
							Name: "org-23",
						},
						Space: &Space{
							GUID: "space-guid-23",
							Name: "space-23",
						},
					},
					{
						SecurityGroup: &SecurityGroup{
							GUID: "security-group-guid-3",
							Name: "security-group-3",
						},
						Organization: &Organization{},
						Space:        &Space{},
					},
					{
						SecurityGroup: &SecurityGroup{
							GUID: "security-group-guid-4",
							Name: "security-group-4",
						},
						Organization: &Organization{
							GUID: "org-guid-11",
							Name: "org-11",
						},
						Space: &Space{
							GUID: "space-guid-32",
							Name: "space-32",
						},
					},
					{
						SecurityGroup: &SecurityGroup{
							GUID: "security-group-guid-4",
							Name: "security-group-4",
						},
						Organization: &Organization{
							GUID: "org-guid-23",
							Name: "org-23",
						},
						Space: &Space{
							GUID: "space-guid-31",
							Name: "space-31",
						},
					},
					{
						SecurityGroup: &SecurityGroup{
							GUID: "security-group-guid-4",
							Name: "security-group-4",
						},
						Organization: &Organization{
							GUID: "org-guid-33",
							Name: "org-33",
						},
						Space: &Space{
							GUID: "space-guid-33",
							Name: "space-33",
						},
					},
				}
				Expect(secGroupOrgSpaces).To(Equal(expected))
				Expect(fakeCloudControllerClient.GetSecurityGroupsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetSecurityGroupsArgsForCall(0)).To(BeNil())
				Expect(fakeCloudControllerClient.GetSpacesBySecurityGroupCallCount()).To(Equal(4))
				Expect(fakeCloudControllerClient.GetSpacesBySecurityGroupArgsForCall(0)).To(Equal("security-group-guid-1"))
				Expect(fakeCloudControllerClient.GetSpacesBySecurityGroupArgsForCall(1)).To(Equal("security-group-guid-2"))
				Expect(fakeCloudControllerClient.GetSpacesBySecurityGroupArgsForCall(2)).To(Equal("security-group-guid-3"))
				Expect(fakeCloudControllerClient.GetSpacesBySecurityGroupArgsForCall(3)).To(Equal("security-group-guid-4"))
				Expect(fakeCloudControllerClient.GetOrganizationCallCount()).To(Equal(6))
				Expect(fakeCloudControllerClient.GetOrganizationArgsForCall(0)).To(Equal("org-guid-13"))
				Expect(fakeCloudControllerClient.GetOrganizationArgsForCall(1)).To(Equal("org-guid-12"))
				Expect(fakeCloudControllerClient.GetOrganizationArgsForCall(2)).To(Equal("org-guid-11"))
				Expect(fakeCloudControllerClient.GetOrganizationArgsForCall(3)).To(Equal("org-guid-21"))
				Expect(fakeCloudControllerClient.GetOrganizationArgsForCall(4)).To(Equal("org-guid-23"))
				Expect(fakeCloudControllerClient.GetOrganizationArgsForCall(5)).To(Equal("org-guid-33"))
			})
		})
	})

	Describe("GetSecurityGroupByName", func() {
		var (
			securityGroup SecurityGroup
			warnings      Warnings
			err           error
		)
		JustBeforeEach(func() {
			securityGroup, warnings, err = actor.GetSecurityGroupByName("some-security-group")
		})
		Context("when the security group exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]ccv2.SecurityGroup{
						{
							GUID: "some-security-group-guid",
							Name: "some-security-group",
						},
					},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil,
				)
			})

			It("returns the security group and all warnings", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(securityGroup.GUID).To(Equal("some-security-group-guid"))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

				Expect(fakeCloudControllerClient.GetSecurityGroupsCallCount()).To(Equal(1))
				query := fakeCloudControllerClient.GetSecurityGroupsArgsForCall(0)
				Expect(query).To(Equal(
					[]ccv2.Query{{
						Filter:   ccv2.NameFilter,
						Operator: ccv2.EqualOperator,
						Value:    "some-security-group",
					}}))
			})
		})
		Context("when the security group does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]ccv2.SecurityGroup{},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil,
				)
			})
			It("returns a SecurityGroupNotFound error", func() {
				Expect(err).To(MatchError(SecurityGroupNotFoundError{Name: "some-security-group"}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
		Context("an error occurs", func() {
			var returnedError error

			BeforeEach(func() {
				returnedError = errors.New("get-security-groups-error")
				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]ccv2.SecurityGroup{},
					ccv2.Warnings{"warning-1", "warning-2"},
					returnedError,
				)
			})

			It("returns the error and all warnings", func() {
				Expect(err).To(MatchError(returnedError))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})

	Describe("BindSecurityGroupToSpace", func() {
		var (
			err      error
			warnings []string
		)

		JustBeforeEach(func() {
			warnings, err = actor.BindSecurityGroupToSpace("some-security-group-guid", "some-space-guid")
		})

		Context("when binding the space does not retun an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.AssociateSpaceWithSecurityGroupReturns(
					ccv2.Warnings{"warning-1", "warning-2"},
					nil,
				)
			})

			It("returns warnings and no error", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(fakeCloudControllerClient.AssociateSpaceWithSecurityGroupCallCount()).To(Equal(1))
				securityGroupGUID, spaceGUID := fakeCloudControllerClient.AssociateSpaceWithSecurityGroupArgsForCall(0)
				Expect(securityGroupGUID).To(Equal("some-security-group-guid"))
				Expect(spaceGUID).To(Equal("some-space-guid"))
			})
		})

		Context("when binding the space returns an error", func() {
			var returnedError error
			BeforeEach(func() {
				returnedError = errors.New("associate-space-error")
				fakeCloudControllerClient.AssociateSpaceWithSecurityGroupReturns(
					ccv2.Warnings{"warning-1", "warning-2"},
					returnedError,
				)
			})

			It("returns the error and warnings", func() {
				Expect(err).To(Equal(returnedError))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})

	Describe("GetSpaceRunningSecurityGroupsBySpace", func() {
		Context("when the space exists and there are no errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceReturns(
					[]ccv2.SecurityGroup{
						{
							Name: "some-shared-security-group",
						},
						{
							Name: "some-running-security-group",
						},
					},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil,
				)
			})
			It("returns the security groups and warnings", func() {
				securityGroups, warnings, err := actor.GetSpaceRunningSecurityGroupsBySpace("space-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf([]string{"warning-1", "warning-2"}))
				Expect(securityGroups).To(Equal(
					[]SecurityGroup{
						{
							Name: "some-shared-security-group",
						},
						{
							Name: "some-running-security-group",
						},
					}))

				Expect(fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceArgsForCall(0)).To(Equal("space-guid"))
			})
		})

		Context("when the space does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceReturns(
					nil,
					nil,
					ccerror.ResourceNotFoundError{})
			})

			It("returns an SpaceNotFoundError", func() {
				_, _, err := actor.GetSpaceRunningSecurityGroupsBySpace("space-guid")
				Expect(err).To(MatchError(SpaceNotFoundError{GUID: "space-guid"}))
			})
		})

		Context("when there is an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("banana")
				fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceReturns(
					nil,
					ccv2.Warnings{"warning-1", "warning-2"},
					expectedErr)
			})

			It("returns the error and warnings", func() {
				_, warnings, err := actor.GetSpaceRunningSecurityGroupsBySpace("space-guid")
				Expect(warnings).To(ConsistOf([]string{"warning-1", "warning-2"}))
				Expect(err).To(MatchError(expectedErr))
			})
		})
	})

	Describe("GetSpaceStagingSecurityGroupsBySpace", func() {
		Context("when the space exists and there are no errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceReturns(
					[]ccv2.SecurityGroup{
						{
							Name: "some-shared-security-group",
						},
						{
							Name: "some-staging-security-group",
						},
					},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil,
				)
			})

			It("returns the security groups and warnings", func() {
				securityGroups, warnings, err := actor.GetSpaceStagingSecurityGroupsBySpace("space-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf([]string{"warning-1", "warning-2"}))
				Expect(securityGroups).To(Equal(
					[]SecurityGroup{
						{
							Name: "some-shared-security-group",
						},
						{
							Name: "some-staging-security-group",
						},
					}))

				Expect(fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceArgsForCall(0)).To(Equal("space-guid"))
			})
		})

		Context("when the space does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceReturns(
					nil,
					nil,
					ccerror.ResourceNotFoundError{})
			})

			It("returns an SpaceNotFoundError", func() {
				_, _, err := actor.GetSpaceStagingSecurityGroupsBySpace("space-guid")
				Expect(err).To(MatchError(SpaceNotFoundError{GUID: "space-guid"}))
			})
		})

		Context("when there is an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("banana")
				fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceReturns(
					nil,
					ccv2.Warnings{"warning-1", "warning-2"},
					expectedErr)
			})

			It("returns the error and warnings", func() {
				_, warnings, err := actor.GetSpaceStagingSecurityGroupsBySpace("space-guid")
				Expect(warnings).To(ConsistOf([]string{"warning-1", "warning-2"}))
				Expect(err).To(MatchError(expectedErr))
			})
		})
	})

	Describe("UnbindSecurityGroupByNameAndSpace", func() {
		var (
			warnings Warnings
			err      error
		)

		JustBeforeEach(func() {
			warnings, err = actor.UnbindSecurityGroupByNameAndSpace("some-security-group", "some-space-guid")
		})

		Context("when an error is encountered getting the security group", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]ccv2.SecurityGroup{},
					ccv2.Warnings{"security-group-warning"},
					nil)
			})

			It("returns the error and all warnings", func() {
				Expect(warnings).To(ConsistOf([]string{"security-group-warning"}))
				Expect(err).To(MatchError(SecurityGroupNotFoundError{"some-security-group"}))
			})
		})

		Context("when the unbinding is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]ccv2.SecurityGroup{{
						Name: "some-security-group",
						GUID: "some-security-group-guid",
					}},
					ccv2.Warnings{"security-group-warning"},
					nil)
				fakeCloudControllerClient.RemoveSpaceFromSecurityGroupReturns(
					ccv2.Warnings{"remove-space-from-sg-warning"},
					nil)
			})

			It("returns all warnings", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf(Warnings{"security-group-warning", "remove-space-from-sg-warning"}))
			})
		})

		Context("when an error is encountered", func() {
			var returnedError error

			BeforeEach(func() {
				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]ccv2.SecurityGroup{{
						Name: "some-security-group",
						GUID: "some-security-group-guid",
					}},
					ccv2.Warnings{"security-group-warning"},
					nil)
				returnedError = errors.New("get-security-groups-error")
				fakeCloudControllerClient.RemoveSpaceFromSecurityGroupReturns(
					ccv2.Warnings{"remove-space-from-sg-warning"},
					returnedError)
			})

			It("returns all warnings", func() {
				Expect(err).To(MatchError(returnedError))
				Expect(warnings).To(ConsistOf(Warnings{"security-group-warning", "remove-space-from-sg-warning"}))
			})
		})
	})

	Describe("UnbindSecurityGroupByNameOrganizationNameAndSpaceName", func() {
		var (
			warnings []string
			err      error
		)

		JustBeforeEach(func() {
			warnings, err = actor.UnbindSecurityGroupByNameOrganizationNameAndSpaceName("some-security-group", "some-org", "some-space")
		})

		Context("when an error is encountered getting the security group", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]ccv2.SecurityGroup{},
					ccv2.Warnings{"security-group-warning"},
					nil)
			})

			It("returns the error and all warnings", func() {
				Expect(warnings).To(ConsistOf([]string{"security-group-warning"}))
				Expect(err).To(MatchError(SecurityGroupNotFoundError{"some-security-group"}))
			})
		})

		Context("when an error is encountered getting the organization", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]ccv2.SecurityGroup{{
						Name: "some-security-group",
						GUID: "some-security-group-guid",
					}},
					ccv2.Warnings{"security-group-warning"},
					nil)
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{},
					ccv2.Warnings{"org-warning"},
					nil)
			})

			It("returns the error and all warnings", func() {
				Expect(warnings).To(ConsistOf([]string{"security-group-warning", "org-warning"}))
				Expect(err).To(MatchError(OrganizationNotFoundError{Name: "some-org"}))
			})
		})

		Context("when an error is encountered getting the space", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]ccv2.SecurityGroup{{
						Name: "some-security-group",
						GUID: "some-security-group-guid",
					}},
					ccv2.Warnings{"security-group-warning"},
					nil)
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{{
						Name: "some-org",
						GUID: "some-org-guid",
					}},
					ccv2.Warnings{"org-warning"},
					nil)
				fakeCloudControllerClient.GetSpacesReturns(
					[]ccv2.Space{},
					ccv2.Warnings{"space-warning"},
					nil)
			})

			It("returns the error and all warnings", func() {
				Expect(warnings).To(ConsistOf([]string{"security-group-warning", "org-warning", "space-warning"}))
				Expect(err).To(MatchError(SpaceNotFoundError{Name: "some-space"}))
			})
		})

		Context("when an error is encountered unbinding the security group the space", func() {
			var returnedError error

			BeforeEach(func() {
				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]ccv2.SecurityGroup{{
						Name: "some-security-group",
						GUID: "some-security-group-guid",
					}},
					ccv2.Warnings{"security-group-warning"},
					nil)
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{{
						Name: "some-org",
						GUID: "some-org-guid",
					}},
					ccv2.Warnings{"org-warning"},
					nil)
				fakeCloudControllerClient.GetSpacesReturns(
					[]ccv2.Space{{
						Name: "some-space",
						GUID: "some-space-guid",
					}},
					ccv2.Warnings{"space-warning"},
					nil)
				returnedError = errors.New("associate-space-error")
				fakeCloudControllerClient.RemoveSpaceFromSecurityGroupReturns(
					ccv2.Warnings{"unbind-warning"},
					returnedError)
			})

			It("returns the error and all warnings", func() {
				Expect(warnings).To(ConsistOf([]string{
					"security-group-warning",
					"org-warning",
					"space-warning",
					"unbind-warning"}))
				Expect(err).To(MatchError(returnedError))
			})
		})

		Context("when no errors are encountered", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]ccv2.SecurityGroup{{
						Name: "some-security-group",
						GUID: "some-security-group-guid",
					}},
					ccv2.Warnings{"security-group-warning"},
					nil)
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{{
						Name: "some-org",
						GUID: "some-org-guid",
					}},
					ccv2.Warnings{"org-warning"},
					nil)
				fakeCloudControllerClient.GetSpacesReturns(
					[]ccv2.Space{{
						Name: "some-space",
						GUID: "some-space-guid",
					}},
					ccv2.Warnings{"space-warning"},
					nil)
				fakeCloudControllerClient.RemoveSpaceFromSecurityGroupReturns(
					ccv2.Warnings{"unbind-warning"},
					nil)
			})

			It("returns all warnings", func() {
				Expect(warnings).To(ConsistOf([]string{
					"security-group-warning",
					"org-warning",
					"space-warning",
					"unbind-warning"}))
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})

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
		actor                     Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil)
	})

	Describe("GetSecurityGroupByName", func() {
		var (
			security_group SecurityGroup
			warnings       Warnings
			err            error
		)
		JustBeforeEach(func() {
			security_group, warnings, err = actor.GetSecurityGroupByName("some-security-group")
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
				Expect(security_group.GUID).To(Equal("some-security-group-guid"))
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

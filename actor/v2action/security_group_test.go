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
})

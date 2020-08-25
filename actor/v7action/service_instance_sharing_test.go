package v7action_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Instance Sharing", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil, nil)
	})

	Describe("ShareServiceInstanceToSpaceAndOrg", func() {
		const (
			serviceInstanceName = "some-service-instance"
			targetedSpaceGUID   = "some-source-space-guid"
			targetedOrgGUID     = "some-org-guid"

			shareToSpaceName = "share-to-space-name"
			shareToOrgName   = "share-to-org-name"
		)

		var (
			serviceInstanceSharingParams = ServiceInstanceSharingParams{}
			warnings                     Warnings
			executionError               error
		)

		BeforeEach(func() {
			serviceInstanceSharingParams = ServiceInstanceSharingParams{
				SpaceName: shareToSpaceName,
				OrgName:   types.OptionalString{},
			}
		})

		JustBeforeEach(func() {
			warnings, executionError = actor.ShareServiceInstanceToSpaceAndOrg(
				serviceInstanceName,
				targetedSpaceGUID,
				targetedOrgGUID,
				serviceInstanceSharingParams,
			)
		})

		When("the service instance cannot be found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{},
					ccv3.IncludedResources{},
					ccv3.Warnings{"some-service-instance-warning"},
					ccerror.ServiceInstanceNotFoundError{Name: serviceInstanceName, SpaceGUID: targetedSpaceGUID},
				)
			})

			It("returns an actor error and warnings", func() {
				Expect(fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(0))
				Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(0))

				Expect(executionError).To(MatchError(actionerror.ServiceInstanceNotFoundError{Name: serviceInstanceName}))
				Expect(warnings).To(ConsistOf("some-service-instance-warning"))
			})
		})

		When("the service instance is retrieved", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{Name: serviceInstanceName},
					ccv3.IncludedResources{},
					ccv3.Warnings{"some-service-instance-warning-1"},
					nil,
				)
			})

			When("the space to share to cannot be found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpacesReturns(
						[]resources.Space{},
						ccv3.IncludedResources{},
						ccv3.Warnings{"some-space-warnings"},
						nil,
					)
				})

				It("returns an actor error and warnings", func() {
					Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetSpacesArgsForCall(0)).To(Equal(
						[]ccv3.Query{
							{
								Key:    ccv3.NameFilter,
								Values: []string{shareToSpaceName},
							},
							{
								Key:    ccv3.OrganizationGUIDFilter,
								Values: []string{targetedOrgGUID},
							},
						}))

					Expect(executionError).To(MatchError(actionerror.SpaceNotFoundError{Name: shareToSpaceName}))
					Expect(warnings).To(ConsistOf("some-service-instance-warning-1", "some-space-warnings"))
				})

				When("the org to share to is also specified", func() {
					const shareToOrgGUID = "share-to-org-guid"

					BeforeEach(func() {
						serviceInstanceSharingParams.OrgName = types.NewOptionalString(shareToOrgName)

						fakeCloudControllerClient.GetOrganizationsReturns(
							[]resources.Organization{{GUID: shareToOrgGUID}},
							ccv3.Warnings{"some-org-warning"},
							nil,
						)
					})

					It("searches for the space in the specified org", func() {
						Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
						Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
						Expect(fakeCloudControllerClient.GetSpacesArgsForCall(0)).To(Equal(
							[]ccv3.Query{
								{
									Key:    ccv3.NameFilter,
									Values: []string{shareToSpaceName},
								},
								{
									Key:    ccv3.OrganizationGUIDFilter,
									Values: []string{shareToOrgGUID},
								},
							}))
					})

					It("returns an actor error and warnings", func() {
						Expect(executionError).To(MatchError(actionerror.SpaceNotFoundError{Name: shareToSpaceName}))
						Expect(warnings).To(ConsistOf("some-service-instance-warning-1", "some-org-warning", "some-space-warnings"))
					})
				})
			})

			When("the specified org cannot be found", func() {
				BeforeEach(func() {
					serviceInstanceSharingParams.OrgName = types.NewOptionalString(shareToOrgName)

					fakeCloudControllerClient.GetOrganizationsReturns(
						[]resources.Organization{},
						ccv3.Warnings{"some-org-warning"},
						nil,
					)
				})

				It("returns an actor error and warnings", func() {
					Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetOrganizationsArgsForCall(0)).To(Equal(
						[]ccv3.Query{
							{
								Key:    ccv3.NameFilter,
								Values: []string{shareToOrgName},
							},
						}))
					Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(0))

					Expect(executionError).To(MatchError(actionerror.OrganizationNotFoundError{Name: shareToOrgName}))
					Expect(warnings).To(ConsistOf("some-service-instance-warning-1", "some-org-warning"))
				})
			})
		})
	})
})

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
			spaceGUID           = "some-source-space-guid"
			orgGUID             = "some-org-guid"
			shareToSpaceName    = "share-to-space-name"
			shareToOrgName      = "share-to-org-name"
		)

		var (
			serviceInstanceSharingParams = ServiceInstanceSharingParams{
				SpaceName: shareToSpaceName,
				OrgName:   types.OptionalString{},
			}
			warnings       Warnings
			executionError error
		)

		JustBeforeEach(func() {
			warnings, executionError = actor.ShareServiceInstanceToSpaceAndOrg(serviceInstanceName, spaceGUID, orgGUID, serviceInstanceSharingParams)
		})

		When("the service instance cannot be found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{},
					ccv3.IncludedResources{},
					ccv3.Warnings{"some-service-instance-warning"},
					ccerror.ServiceInstanceNotFoundError{Name: serviceInstanceName, SpaceGUID: spaceGUID},
				)
			})

			It("returns an actor error and warnings", func() {
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

			When("the space cannot be found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpacesReturns(
						[]resources.Space{},
						ccv3.IncludedResources{},
						ccv3.Warnings{"some-service-instance-warning-2"},
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
								Values: []string{orgGUID},
							},
						}))

					Expect(executionError).To(MatchError(actionerror.SpaceNotFoundError{Name: shareToSpaceName}))
					Expect(warnings).To(ConsistOf("some-service-instance-warning-1", "some-service-instance-warning-2"))
				})
			})

			When("the specified org cannot be found", func() {
				BeforeEach(func() {
					serviceInstanceSharingParams.OrgName = types.NewOptionalString(shareToOrgName)

					fakeCloudControllerClient.GetOrganizationsReturns(
						[]resources.Organization{},
						ccv3.Warnings{"some-service-instance-warning-2"},
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

					Expect(executionError).To(MatchError(actionerror.OrganizationNotFoundError{Name: shareToOrgName}))
					Expect(warnings).To(ConsistOf("some-service-instance-warning-1", "some-service-instance-warning-2"))
				})
			})
		})

	})
})

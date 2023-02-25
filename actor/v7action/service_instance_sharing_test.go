package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Instance Sharing", func() {
	var (
		actor                        *Actor
		fakeCloudControllerClient    *v7actionfakes.FakeCloudControllerClient
		serviceInstanceSharingParams ServiceInstanceSharingParams
		warnings                     Warnings
		executionError               error
	)

	const (
		serviceInstanceName = "some-service-instance"
		targetedSpaceGUID   = "some-source-space-guid"
		targetedOrgGUID     = "some-org-guid"

		shareToSpaceName = "share-to-space-name"
		shareToOrgName   = "share-to-org-name"
	)

	itValidatesParameters := func() {
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

		When("the specified org cannot be found", func() {
			BeforeEach(func() {
				serviceInstanceSharingParams.OrgName = types.NewOptionalString(shareToOrgName)

				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{Name: serviceInstanceName},
					ccv3.IncludedResources{},
					ccv3.Warnings{"some-service-instance-warning-1"},
					nil,
				)

				fakeCloudControllerClient.GetOrganizationsReturns(
					[]resources.Organization{},
					ccv3.Warnings{"some-org-warning"},
					nil,
				)
			})

			It("searches for the specified org", func() {
				Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetOrganizationsArgsForCall(0)).To(Equal(
					[]ccv3.Query{
						{
							Key:    ccv3.NameFilter,
							Values: []string{shareToOrgName},
						},
					}))
				Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(0))
			})

			It("returns an actor error and warnings", func() {
				Expect(executionError).To(MatchError(actionerror.OrganizationNotFoundError{Name: shareToOrgName}))
				Expect(warnings).To(ConsistOf("some-service-instance-warning-1", "some-org-warning"))
			})
		})

		When("the space cannot be found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{Name: serviceInstanceName},
					ccv3.IncludedResources{},
					ccv3.Warnings{"some-service-instance-warning-1"},
					nil,
				)

				fakeCloudControllerClient.GetSpacesReturns(
					[]resources.Space{},
					ccv3.IncludedResources{},
					ccv3.Warnings{"some-space-warnings"},
					nil,
				)
			})

			When("the org to share to is the targeted org", func() {
				It("searches for the space in the specified org", func() {
					Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(0))
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
				})

				It("returns an actor error and warnings", func() {
					Expect(executionError).To(MatchError(actionerror.SpaceNotFoundError{Name: shareToSpaceName}))
					Expect(warnings).To(ConsistOf("some-service-instance-warning-1", "some-space-warnings"))
				})
			})

			When("the org to share to is specified by user", func() {
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
	}

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil, nil)
	})

	Describe("ShareServiceInstanceToSpaceAndOrg", func() {
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

		itValidatesParameters()

		When("a successful share request is made", func() {
			spaceGUID := "fake-space-guid"
			serviceInstanceGUID := "fake-service-instance-guid"

			BeforeEach(func() {
				fakeCloudControllerClient.GetSpacesReturns(
					[]resources.Space{{GUID: spaceGUID}},
					ccv3.IncludedResources{},
					ccv3.Warnings{"some-space-warning"},
					nil,
				)

				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{GUID: serviceInstanceGUID},
					ccv3.IncludedResources{},
					ccv3.Warnings{"some-service-instance-warning"},
					nil,
				)
			})

			It("makes a request to the cloud controller", func() {
				Expect(fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.ShareServiceInstanceToSpacesCallCount()).To(Equal(1))

				actualServiceInstanceGUID, actualSpaces := fakeCloudControllerClient.ShareServiceInstanceToSpacesArgsForCall(0)
				Expect(actualServiceInstanceGUID).To(Equal(serviceInstanceGUID))
				Expect(actualSpaces[0]).To(Equal(spaceGUID))

				Expect(executionError).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-space-warning", "some-service-instance-warning"))
			})
		})

		When("sharing request returns an error", func() {
			spaceGUID := "fake-space-guid"
			serviceInstanceGUID := "fake-service-instance-guid"

			BeforeEach(func() {
				fakeCloudControllerClient.GetSpacesReturns(
					[]resources.Space{{GUID: spaceGUID}},
					ccv3.IncludedResources{},
					ccv3.Warnings{"some-space-warning"},
					nil,
				)
				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{GUID: serviceInstanceGUID},
					ccv3.IncludedResources{},
					ccv3.Warnings{"some-service-instance-warning"},
					nil,
				)
				fakeCloudControllerClient.ShareServiceInstanceToSpacesReturns(
					resources.RelationshipList{},
					ccv3.Warnings{"some-share-warning"},
					errors.New("cannot share the instance"),
				)
			})

			It("returns an error and warnings", func() {
				Expect(executionError).To(MatchError("cannot share the instance"))
				Expect(warnings).To(ConsistOf("some-space-warning", "some-service-instance-warning", "some-share-warning"))
			})
		})
	})

	Describe("UnshareServiceInstanceFromSpaceAndOrg", func() {
		BeforeEach(func() {
			serviceInstanceSharingParams = ServiceInstanceSharingParams{
				SpaceName: shareToSpaceName,
				OrgName:   types.OptionalString{},
			}
		})

		JustBeforeEach(func() {
			warnings, executionError = actor.UnshareServiceInstanceFromSpaceAndOrg(
				serviceInstanceName,
				targetedSpaceGUID,
				targetedOrgGUID,
				serviceInstanceSharingParams,
			)
		})

		itValidatesParameters()

		When("a unshare request is made", func() {
			expectedSpaceGUID := "fake-space-guid"
			expectedServiceInstanceGUID := "fake-service-instance-guid"

			BeforeEach(func() {
				fakeCloudControllerClient.GetSpacesReturns(
					[]resources.Space{{GUID: expectedSpaceGUID}},
					ccv3.IncludedResources{},
					ccv3.Warnings{"some-space-warning"},
					nil,
				)

				fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
					resources.ServiceInstance{GUID: expectedServiceInstanceGUID},
					ccv3.IncludedResources{},
					ccv3.Warnings{"some-service-instance-warning"},
					nil,
				)
			})

			It("makes the right request", func() {
				Expect(fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.UnshareServiceInstanceFromSpaceCallCount()).To(Equal(1))

				actualServiceInstanceGUID, actualSpace := fakeCloudControllerClient.UnshareServiceInstanceFromSpaceArgsForCall(0)
				Expect(actualServiceInstanceGUID).To(Equal(expectedServiceInstanceGUID))
				Expect(actualSpace).To(Equal(expectedSpaceGUID))
			})

			When("the request is successful", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.UnshareServiceInstanceFromSpaceReturns(
						ccv3.Warnings{"some-unshare-warning"},
						nil,
					)
				})

				It("returns warnings and no error", func() {
					Expect(executionError).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("some-space-warning", "some-service-instance-warning", "some-unshare-warning"))
				})
			})

			When("the request fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.UnshareServiceInstanceFromSpaceReturns(
						ccv3.Warnings{"some-unshare-warning"},
						errors.New("cannot unshare the instance"),
					)
				})

				It("returns warnings and an error", func() {
					Expect(executionError).To(MatchError("cannot unshare the instance"))
					Expect(warnings).To(ConsistOf("some-space-warning", "some-service-instance-warning", "some-unshare-warning"))

				})
			})
		})
	})
})

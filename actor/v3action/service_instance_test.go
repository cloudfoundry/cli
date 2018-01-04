package v3action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Instance Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v3actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v3actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil)
	})

	Describe("ShareServiceInstanceByOrganizationAndSpaceName", func() {
		var (
			serviceInstanceName string
			sourceSpaceGUID     string
			orgGUID             string
			spaceName           string

			warnings       Warnings
			executionError error
		)

		BeforeEach(func() {
			serviceInstanceName = "some-service-instance"
			sourceSpaceGUID = "some-source-space-guid"
			orgGUID = "some-org-guid"
			spaceName = "some-space-name"
		})

		JustBeforeEach(func() {
			warnings, executionError = actor.ShareServiceInstanceInSpaceByOrganizationAndSpaceName(serviceInstanceName, sourceSpaceGUID, orgGUID, spaceName)
		})

		Context("when the service instance name is valid", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstancesReturns([]ccv3.ServiceInstance{
					{
						Name: "some-service-instance",
						GUID: "some-service-instance-guid",
					},
				}, ccv3.Warnings{"some-service-instance-warning"}, nil)
			})

			Context("when the space name is valid", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpacesReturns([]ccv3.Space{
						{
							Name: "some-space",
							GUID: "some-space-guid",
						},
					}, ccv3.Warnings{"some-space-warning"}, nil)
				})

				Context("when the post request to the shared spaces endpoint succeeds", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.ShareServiceInstanceToSpacesReturns(ccv3.RelationshipList{}, ccv3.Warnings{"some-share-warning"}, nil)
					})

					It("calls to create a new service instance share", func() {
						Expect(fakeCloudControllerClient.GetServiceInstancesCallCount()).To(Equal(1))
						Expect(fakeCloudControllerClient.GetServiceInstancesArgsForCall(0)).To(ConsistOf(
							ccv3.Query{Key: ccv3.NameFilter, Values: []string{serviceInstanceName}},
							ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{sourceSpaceGUID}},
						))

						Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
						Expect(fakeCloudControllerClient.GetSpacesArgsForCall(0)).To(ConsistOf(
							ccv3.Query{Key: ccv3.NameFilter, Values: []string{spaceName}},
							ccv3.Query{Key: ccv3.OrganizationGUIDFilter, Values: []string{orgGUID}},
						))

						Expect(fakeCloudControllerClient.ShareServiceInstanceToSpacesCallCount()).To(Equal(1))
						si_guid, space_guids := fakeCloudControllerClient.ShareServiceInstanceToSpacesArgsForCall(0)
						Expect(si_guid).To(Equal("some-service-instance-guid"))
						Expect(space_guids).To(Equal([]string{"some-space-guid"}))
					})

					It("does not return warnings or errors", func() {
						Expect(executionError).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("some-service-instance-warning", "some-space-warning", "some-share-warning"))
					})
				})

				Context("when the post request to the shared spaces endpoint fails", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.ShareServiceInstanceToSpacesReturns(ccv3.RelationshipList{}, ccv3.Warnings{"some-share-warning"}, errors.New("Share failed"))
					})

					It("returns error", func() {
						Expect(executionError).To(MatchError("Share failed"))
						Expect(warnings).To(ConsistOf("some-service-instance-warning", "some-space-warning", "some-share-warning"))
					})
				})
			})

			Context("when resolving the space name fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpacesReturns([]ccv3.Space{}, ccv3.Warnings{"some-space-warning"}, errors.New("Space name doesn't exist"))
				})

				It("returns error", func() {
					Expect(executionError).To(MatchError("Space name doesn't exist"))
					Expect(warnings).To(ConsistOf("some-service-instance-warning", "some-space-warning"))
				})
			})
		})

		Context("when resolving the service instance name fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstancesReturns([]ccv3.ServiceInstance{}, ccv3.Warnings{"some-service-instance-warning"}, errors.New("service name doesn't exist"))
			})

			It("returns error", func() {
				Expect(executionError).To(MatchError("service name doesn't exist"))
				Expect(warnings).To(ConsistOf("some-service-instance-warning"))
			})
		})

		Context("when the service instance cannot be found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstancesReturns([]ccv3.ServiceInstance{}, ccv3.Warnings{"some-service-instance-warning"}, actionerror.ServiceInstanceNotFoundError{Name: serviceInstanceName})
			})

			It("returns error", func() {
				Expect(executionError).To(Equal(actionerror.SharedServiceInstanceNotFoundError{}))
				Expect(warnings).To(ConsistOf("some-service-instance-warning"))
			})
		})
	})

	Describe("ShareServiceInstanceByOrganizationNameAndSpaceName", func() {
		var (
			serviceInstanceName string
			sourceSpaceGUID     string
			orgName             string
			orgGUID             string
			spaceName           string

			warnings       Warnings
			executionError error
		)

		BeforeEach(func() {
			serviceInstanceName = "some-service-instance"
			sourceSpaceGUID = "some-source-space-guid"
			orgName = "some-org-name"
			orgGUID = "some-org-guid"
			spaceName = "some-space-name"
		})

		JustBeforeEach(func() {
			warnings, executionError = actor.ShareServiceInstanceInSpaceByOrganizationNameAndSpaceName(serviceInstanceName, sourceSpaceGUID, orgName, spaceName)
		})

		Context("when the org name is valid", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns([]ccv3.Organization{
					{
						Name: "some-org-name",
						GUID: "some-org-guid",
					},
				}, ccv3.Warnings{"some-org-warning"}, nil)
			})

			Context("when the service instance name is valid", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstancesReturns([]ccv3.ServiceInstance{
						{
							Name: "some-service-instance",
							GUID: "some-service-instance-guid",
						},
					}, ccv3.Warnings{"some-service-instance-warning"}, nil)
				})

				Context("when the space name is valid", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetSpacesReturns([]ccv3.Space{
							{
								Name: "some-space",
								GUID: "some-space-guid",
							},
						}, ccv3.Warnings{"some-space-warning"}, nil)
					})

					Context("when the post request to the shared spaces endpoint succeeds", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.ShareServiceInstanceToSpacesReturns(ccv3.RelationshipList{}, ccv3.Warnings{"some-share-warning"}, nil)
						})

						It("calls to create a new service instance share", func() {
							Expect(fakeCloudControllerClient.GetServiceInstancesCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.GetServiceInstancesArgsForCall(0)).To(ConsistOf(
								ccv3.Query{Key: ccv3.NameFilter, Values: []string{serviceInstanceName}},
								ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{sourceSpaceGUID}},
							))

							Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.GetOrganizationsArgsForCall(0)).To(ConsistOf(
								ccv3.Query{Key: ccv3.NameFilter, Values: []string{orgName}},
							))

							Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.GetSpacesArgsForCall(0)).To(ConsistOf(
								ccv3.Query{Key: ccv3.NameFilter, Values: []string{spaceName}},
								ccv3.Query{Key: ccv3.OrganizationGUIDFilter, Values: []string{orgGUID}},
							))

							Expect(fakeCloudControllerClient.ShareServiceInstanceToSpacesCallCount()).To(Equal(1))
							si_guid, space_guids := fakeCloudControllerClient.ShareServiceInstanceToSpacesArgsForCall(0)
							Expect(si_guid).To(Equal("some-service-instance-guid"))
							Expect(space_guids).To(Equal([]string{"some-space-guid"}))
						})

						It("does not return warnings or errors", func() {
							Expect(executionError).ToNot(HaveOccurred())
							Expect(warnings).To(ConsistOf("some-service-instance-warning", "some-org-warning", "some-space-warning", "some-share-warning"))
						})
					})

					Context("when the post request to the shared spaces endpoint fails", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.ShareServiceInstanceToSpacesReturns(ccv3.RelationshipList{}, ccv3.Warnings{"some-share-warning"}, errors.New("Share failed"))
						})

						It("returns error", func() {
							Expect(executionError).To(MatchError("Share failed"))
							Expect(warnings).To(ConsistOf("some-service-instance-warning", "some-org-warning", "some-space-warning", "some-share-warning"))
						})
					})
				})

				Context("when resolving the space name fails", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetSpacesReturns([]ccv3.Space{}, ccv3.Warnings{"some-space-warning"}, errors.New("Space name doesn't exist"))
					})

					It("returns error", func() {
						Expect(executionError).To(MatchError("Space name doesn't exist"))
						Expect(warnings).To(ConsistOf("some-org-warning", "some-service-instance-warning", "some-space-warning"))
					})
				})
			})

			Context("when resolving the service instance name fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstancesReturns([]ccv3.ServiceInstance{}, ccv3.Warnings{"some-service-instance-warning"}, errors.New("service name doesn't exist"))
				})

				It("returns error", func() {
					Expect(executionError).To(MatchError("service name doesn't exist"))
					Expect(warnings).To(ConsistOf("some-org-warning", "some-service-instance-warning"))
				})
			})
		})

		Context("when resolving the org name fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns([]ccv3.Organization{}, ccv3.Warnings{"some-org-warning"}, errors.New("Org name doesn't exist"))
			})

			It("returns error", func() {
				Expect(executionError).To(MatchError("Org name doesn't exist"))
				Expect(warnings).To(ConsistOf("some-org-warning"))
			})
		})
	})

	Describe("UnshareServiceInstanceFromSpace", func() {
		var (
			serviceInstanceName string
			sourceSpaceGUID     string
			sharedToSpaceGUID   string

			warnings       Warnings
			executionError error
		)

		BeforeEach(func() {
			serviceInstanceName = "some-service-instance"
			sourceSpaceGUID = "some-source-space-guid"
			sharedToSpaceGUID = "some-other-space-guid"
		})

		JustBeforeEach(func() {
			warnings, executionError = actor.UnshareServiceInstanceFromSpace(serviceInstanceName, sourceSpaceGUID, sharedToSpaceGUID)
		})

		Context("when the service instance name is valid", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstancesReturns([]ccv3.ServiceInstance{
					{
						Name: "some-service-instance",
						GUID: "some-service-instance-guid",
					},
				}, ccv3.Warnings{"some-service-instance-warning"}, nil)
			})

			Context("when the delete request to the shared spaces endpoint succeeds", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.UnshareServiceInstanceFromSpaceReturns(ccv3.Warnings{"some-unshare-warning"}, nil)
				})

				It("calls to delete a service instance share", func() {
					Expect(fakeCloudControllerClient.GetServiceInstancesCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetServiceInstancesArgsForCall(0)).To(ConsistOf(
						ccv3.Query{Key: ccv3.NameFilter, Values: []string{serviceInstanceName}},
						ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{sourceSpaceGUID}},
					))

					Expect(fakeCloudControllerClient.UnshareServiceInstanceFromSpaceCallCount()).To(Equal(1))
					service_instance_guid, space_guid := fakeCloudControllerClient.UnshareServiceInstanceFromSpaceArgsForCall(0)
					Expect(service_instance_guid).To(Equal("some-service-instance-guid"))
					Expect(space_guid).To(Equal("some-other-space-guid"))
				})

				It("does not return warnings or errors", func() {
					Expect(executionError).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("some-service-instance-warning", "some-unshare-warning"))
				})
			})

			Context("when the delete request to the shared spaces endpoint fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.UnshareServiceInstanceFromSpaceReturns(ccv3.Warnings{"some-unshare-warning"}, errors.New("Unshare failed"))
				})

				It("returns error", func() {
					Expect(executionError).To(MatchError("Unshare failed"))
					Expect(warnings).To(ConsistOf("some-service-instance-warning", "some-unshare-warning"))
				})
			})
		})

		Context("when resolving the service instance name fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstancesReturns([]ccv3.ServiceInstance{}, ccv3.Warnings{"some-service-instance-warning"}, errors.New("service name doesn't exist"))
			})

			It("returns error", func() {
				Expect(executionError).To(MatchError("service name doesn't exist"))
				Expect(warnings).To(ConsistOf("some-service-instance-warning"))
			})
		})

		Context("when the service instance cannot be found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstancesReturns([]ccv3.ServiceInstance{}, ccv3.Warnings{"some-service-instance-warning"}, actionerror.ServiceInstanceNotFoundError{Name: serviceInstanceName})
			})

			It("returns error", func() {
				Expect(executionError).To(Equal(actionerror.SharedServiceInstanceNotFoundError{}))
				Expect(warnings).To(ConsistOf("some-service-instance-warning"))
			})
		})
	})

	Describe("GetServiceInstanceByNameAndSpace", func() {
		var (
			serviceInstanceName string
			sourceSpaceGUID     string

			serviceInstance ServiceInstance
			warnings        Warnings
			executionError  error
		)

		BeforeEach(func() {
			serviceInstanceName = "some-service-instance"
			sourceSpaceGUID = "some-source-space-guid"
		})

		JustBeforeEach(func() {
			serviceInstance, warnings, executionError = actor.GetServiceInstanceByNameAndSpace(serviceInstanceName, sourceSpaceGUID)
		})

		Context("when the cloud controller request is successful", func() {
			Context("when the cloud controller returns one service instance", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstancesReturns([]ccv3.ServiceInstance{
						{
							Name: "some-service-instance",
							GUID: "some-service-instance-guid",
						},
					}, ccv3.Warnings{"some-service-instance-warning"}, nil)
				})

				It("returns a service instance and warnings", func() {
					Expect(executionError).NotTo(HaveOccurred())

					Expect(serviceInstance).To(Equal(ServiceInstance{Name: "some-service-instance", GUID: "some-service-instance-guid"}))
					Expect(warnings).To(ConsistOf("some-service-instance-warning"))
					Expect(fakeCloudControllerClient.GetServiceInstancesCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetServiceInstancesArgsForCall(0)).To(ConsistOf(
						ccv3.Query{Key: ccv3.NameFilter, Values: []string{serviceInstanceName}},
						ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{sourceSpaceGUID}},
					))
				})
			})

			Context("when the cloud controller returns no service instances", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstancesReturns(
						nil,
						ccv3.Warnings{"some-service-instance-warning"},
						nil)
				})

				It("returns an error and warnings", func() {
					Expect(executionError).To(MatchError(actionerror.ServiceInstanceNotFoundError{Name: serviceInstanceName}))

					Expect(warnings).To(ConsistOf("some-service-instance-warning"))
				})
			})
		})

		Context("when the cloud controller returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstancesReturns(
					nil,
					ccv3.Warnings{"some-service-instance-warning"},
					errors.New("no service instance"))
			})

			It("returns an error and warnings", func() {
				Expect(executionError).To(MatchError("no service instance"))
				Expect(warnings).To(ConsistOf("some-service-instance-warning"))
			})
		})
	})
})

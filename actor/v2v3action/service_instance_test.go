package v2v3action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	. "code.cloudfoundry.org/cli/actor/v2v3action"
	"code.cloudfoundry.org/cli/actor/v2v3action/v2v3actionfakes"
	"code.cloudfoundry.org/cli/actor/v3action"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Instance Actions", func() {
	var (
		actor       *Actor
		fakeV2Actor *v2v3actionfakes.FakeV2Actor
		fakeV3Actor *v2v3actionfakes.FakeV3Actor
	)

	BeforeEach(func() {
		fakeV2Actor = new(v2v3actionfakes.FakeV2Actor)
		fakeV3Actor = new(v2v3actionfakes.FakeV3Actor)
		actor = NewActor(fakeV2Actor, fakeV3Actor)
	})

	Describe("ShareServiceInstanceToSpaceNameByNameAndSpaceAndOrganization", func() {
		var (
			shareToSpaceName    string
			serviceInstanceName string
			sourceSpaceGUID     string
			shareToOrgGUID      string

			warnings Warnings
			shareErr error
		)

		BeforeEach(func() {
			shareToSpaceName = "some-space-name"
			serviceInstanceName = "some-service-instance"
			sourceSpaceGUID = "some-source-space-guid"
			shareToOrgGUID = "some-org-guid"
		})

		JustBeforeEach(func() {
			warnings, shareErr = actor.ShareServiceInstanceToSpaceNameByNameAndSpaceAndOrganization(shareToSpaceName, serviceInstanceName, sourceSpaceGUID, shareToOrgGUID)
		})

		Context("when no errors occur getting the service instance", func() {
			BeforeEach(func() {
				fakeV2Actor.GetServiceInstanceByNameAndSpaceReturns(
					v2action.ServiceInstance{
						GUID: "some-service-instance-guid",
					},
					v2action.Warnings{"get-service-instance-warning"},
					nil)
			})

			Context("when no errors occur getting the spaces the service instance is shared to", func() {
				BeforeEach(func() {
					fakeV2Actor.GetServiceInstanceSharedTosByServiceInstanceReturns(
						[]v2action.ServiceInstanceSharedTo{
							{SpaceGUID: "already-shared-space-guid-1"},
							{SpaceGUID: "already-shared-space-guid-2"},
						},
						v2action.Warnings{"get-service-instance-shared-tos-warning"},
						nil)
				})

				Context("when no errors occur getting the space we are sharing to", func() {
					BeforeEach(func() {
						fakeV2Actor.GetSpaceByOrganizationAndNameReturns(
							v2action.Space{
								GUID: "not-shared-space-guid",
							},
							v2action.Warnings{"get-space-warning"},
							nil)
					})

					Context("when the service instance is NOT already shared with this space", func() {
						Context("when no errors occur sharing the service instance to this space", func() {
							BeforeEach(func() {
								fakeV3Actor.ShareServiceInstanceToSpacesReturns(
									v3action.RelationshipList{
										GUIDs: []string{"some-space-guid"},
									},
									v3action.Warnings{"share-service-instance-warning"},
									nil)
							})

							It("shares the service instance to this space and returns all warnings", func() {
								Expect(shareErr).ToNot(HaveOccurred())
								Expect(warnings).To(ConsistOf(
									"get-service-instance-warning",
									"get-service-instance-shared-tos-warning",
									"get-space-warning",
									"share-service-instance-warning"))

								Expect(fakeV2Actor.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
								serviceInstanceNameArg, spaceGUIDArg := fakeV2Actor.GetServiceInstanceByNameAndSpaceArgsForCall(0)
								Expect(serviceInstanceNameArg).To(Equal(serviceInstanceName))
								Expect(spaceGUIDArg).To(Equal(sourceSpaceGUID))

								Expect(fakeV2Actor.GetServiceInstanceSharedTosByServiceInstanceCallCount()).To(Equal(1))
								Expect(fakeV2Actor.GetServiceInstanceSharedTosByServiceInstanceArgsForCall(0)).To(Equal("some-service-instance-guid"))

								Expect(fakeV2Actor.GetSpaceByOrganizationAndNameCallCount()).To(Equal(1))
								orgGUIDArg, spaceNameArg := fakeV2Actor.GetSpaceByOrganizationAndNameArgsForCall(0)
								Expect(orgGUIDArg).To(Equal(shareToOrgGUID))
								Expect(spaceNameArg).To(Equal(shareToSpaceName))

								Expect(fakeV3Actor.ShareServiceInstanceToSpacesCallCount()).To(Equal(1))
								serviceInstanceGUIDArg, spaceGUIDsArg := fakeV3Actor.ShareServiceInstanceToSpacesArgsForCall(0)
								Expect(serviceInstanceGUIDArg).To(Equal("some-service-instance-guid"))
								Expect(spaceGUIDsArg).To(Equal([]string{"not-shared-space-guid"}))
							})
						})

						Context("when an error occurs sharing the service instance to this space", func() {
							var expectedErr error

							BeforeEach(func() {
								expectedErr = errors.New("share service instance error")
								fakeV3Actor.ShareServiceInstanceToSpacesReturns(
									v3action.RelationshipList{},
									v3action.Warnings{"share-service-instance-warning"},
									expectedErr)
							})

							It("returns the error and all warnings", func() {
								Expect(shareErr).To(MatchError(expectedErr))
								Expect(warnings).To(ConsistOf(
									"get-service-instance-warning",
									"get-service-instance-shared-tos-warning",
									"get-space-warning",
									"share-service-instance-warning"))
							})
						})
					})

					Context("when the service instance IS already shared with this space", func() {
						BeforeEach(func() {
							fakeV2Actor.GetSpaceByOrganizationAndNameReturns(
								v2action.Space{
									GUID: "already-shared-space-guid-2",
								},
								v2action.Warnings{"get-space-warning"},
								nil)
						})

						It("returns a ServiceInstanceAlreadySharedError and all warnings", func() {
							Expect(shareErr).To(MatchError(actionerror.ServiceInstanceAlreadySharedError{}))
							Expect(warnings).To(ConsistOf(
								"get-service-instance-warning",
								"get-service-instance-shared-tos-warning",
								"get-space-warning",
								"Service instance some-service-instance is already shared with that space.",
							))
						})
					})
				})

				Context("when an error occurs getting the space we are sharing to", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("get space error")
						fakeV2Actor.GetSpaceByOrganizationAndNameReturns(
							v2action.Space{},
							v2action.Warnings{"get-space-warning"},
							expectedErr)
					})

					It("returns the error and all warnings", func() {
						Expect(shareErr).To(MatchError(expectedErr))
						Expect(warnings).To(ConsistOf(
							"get-service-instance-warning",
							"get-service-instance-shared-tos-warning",
							"get-space-warning"))
					})
				})
			})

			Context("when an error occurs getting the spaces the service instance is shared to", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("get shared to spaces error")
					fakeV2Actor.GetServiceInstanceSharedTosByServiceInstanceReturns(
						nil,
						v2action.Warnings{"get-service-instance-shared-tos-warning"},
						expectedErr)
				})

				It("returns the error and all warnings", func() {
					Expect(shareErr).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf(
						"get-service-instance-warning",
						"get-service-instance-shared-tos-warning"))
				})
			})
		})

		Context("when an error occurs getting the service instance", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("get service instance error")
				fakeV2Actor.GetServiceInstanceByNameAndSpaceReturns(
					v2action.ServiceInstance{},
					v2action.Warnings{"get-service-instance-warning"},
					expectedErr)
			})

			It("returns the error and all warnings", func() {
				Expect(shareErr).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("get-service-instance-warning"))
			})
		})

		Context("when the service instance does not exist", func() {
			BeforeEach(func() {
				fakeV2Actor.GetServiceInstanceByNameAndSpaceReturns(
					v2action.ServiceInstance{},
					v2action.Warnings{"get-service-instance-warning"},
					actionerror.ServiceInstanceNotFoundError{})
			})

			It("returns a SharedServiceInstanceNotFoundError and all warnings", func() {
				Expect(shareErr).To(MatchError(actionerror.SharedServiceInstanceNotFoundError{}))
				Expect(warnings).To(ConsistOf("get-service-instance-warning"))
			})
		})
	})

	Describe("ShareServiceInstanceToSpaceNameByNameAndSpaceAndOrganizationName", func() {
		var (
			shareToSpaceName    string
			serviceInstanceName string
			sourceSpaceGUID     string
			shareToOrgName      string

			warnings Warnings
			shareErr error
		)

		BeforeEach(func() {
			shareToSpaceName = "some-space-name"
			serviceInstanceName = "some-service-instance"
			sourceSpaceGUID = "some-source-space-guid"
			shareToOrgName = "some-org-name"
		})

		JustBeforeEach(func() {
			warnings, shareErr = actor.ShareServiceInstanceToSpaceNameByNameAndSpaceAndOrganizationName(shareToSpaceName, serviceInstanceName, sourceSpaceGUID, shareToOrgName)
		})

		Context("when no errors occur getting the org", func() {
			BeforeEach(func() {
				fakeV3Actor.GetOrganizationByNameReturns(
					v3action.Organization{
						GUID: "some-org-guid",
					},
					v3action.Warnings{"get-org-warning"},
					nil)

				fakeV2Actor.GetServiceInstanceByNameAndSpaceReturns(
					v2action.ServiceInstance{
						GUID: "some-service-instance-guid",
					},
					v2action.Warnings{"get-service-instance-warning"},
					nil)

				fakeV2Actor.GetServiceInstanceSharedTosByServiceInstanceReturns(
					[]v2action.ServiceInstanceSharedTo{
						{SpaceGUID: "already-shared-space-guid-1"},
						{SpaceGUID: "already-shared-space-guid-2"},
					},
					v2action.Warnings{"get-service-instance-shared-tos-warning"},
					nil)

				fakeV2Actor.GetSpaceByOrganizationAndNameReturns(
					v2action.Space{
						GUID: "not-shared-space-guid",
					},
					v2action.Warnings{"get-space-warning"},
					nil)

				fakeV3Actor.ShareServiceInstanceToSpacesReturns(
					v3action.RelationshipList{
						GUIDs: []string{"some-space-guid"},
					},
					v3action.Warnings{"share-service-instance-warning"},
					nil)
			})

			It("shares the service instance to this space and returns all warnings", func() {
				Expect(shareErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf(
					"get-org-warning",
					"get-service-instance-warning",
					"get-service-instance-shared-tos-warning",
					"get-space-warning",
					"share-service-instance-warning"))

				Expect(fakeV3Actor.GetOrganizationByNameCallCount()).To(Equal(1))
				Expect(fakeV3Actor.GetOrganizationByNameArgsForCall(0)).To(Equal(shareToOrgName))

				Expect(fakeV2Actor.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
				serviceInstanceNameArg, spaceGUIDArg := fakeV2Actor.GetServiceInstanceByNameAndSpaceArgsForCall(0)
				Expect(serviceInstanceNameArg).To(Equal(serviceInstanceName))
				Expect(spaceGUIDArg).To(Equal(sourceSpaceGUID))

				Expect(fakeV2Actor.GetServiceInstanceSharedTosByServiceInstanceCallCount()).To(Equal(1))
				Expect(fakeV2Actor.GetServiceInstanceSharedTosByServiceInstanceArgsForCall(0)).To(Equal("some-service-instance-guid"))

				Expect(fakeV2Actor.GetSpaceByOrganizationAndNameCallCount()).To(Equal(1))
				orgGUIDArg, spaceNameArg := fakeV2Actor.GetSpaceByOrganizationAndNameArgsForCall(0)
				Expect(orgGUIDArg).To(Equal("some-org-guid"))
				Expect(spaceNameArg).To(Equal(shareToSpaceName))

				Expect(fakeV3Actor.ShareServiceInstanceToSpacesCallCount()).To(Equal(1))
				serviceInstanceGUIDArg, spaceGUIDsArg := fakeV3Actor.ShareServiceInstanceToSpacesArgsForCall(0)
				Expect(serviceInstanceGUIDArg).To(Equal("some-service-instance-guid"))
				Expect(spaceGUIDsArg).To(Equal([]string{"not-shared-space-guid"}))
			})
		})

		Context("when an error occurs getting the org", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("get org error")
				fakeV3Actor.GetOrganizationByNameReturns(
					v3action.Organization{},
					v3action.Warnings{"get-org-warning"},
					expectedErr)
			})

			It("returns the error and all warnings", func() {
				Expect(shareErr).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("get-org-warning"))
			})
		})
	})
})

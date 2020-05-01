package v2v3action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/resources"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	. "code.cloudfoundry.org/cli/actor/v2v3action"
	"code.cloudfoundry.org/cli/actor/v2v3action/v2v3actionfakes"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"

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

		When("no errors occur getting the service instance", func() {
			When("no errors occur getting the space we are sharing to", func() {
				When("the service instance is a managed service instance", func() {
					BeforeEach(func() {
						fakeV2Actor.GetServiceInstanceByNameAndSpaceReturns(
							v2action.ServiceInstance{
								GUID:        "some-service-instance-guid",
								ServiceGUID: "some-service-guid",
								Type:        constant.ManagedService,
							},
							v2action.Warnings{"get-service-instance-warning"},
							nil)

						fakeV2Actor.GetSpaceByOrganizationAndNameReturns(
							v2action.Space{
								GUID: "not-shared-space-guid",
							},
							v2action.Warnings{"get-space-warning"},
							nil)
					})

					When("no errors occur getting the service", func() {
						BeforeEach(func() {
							fakeV2Actor.GetServiceReturns(
								v2action.Service{
									Extra: ccv2.ServiceExtra{
										Shareable: true,
									},
								},
								v2action.Warnings{"get-service-warning"},
								nil)
						})

						When("no errors occur getting feature flags", func() {
							BeforeEach(func() {
								fakeV2Actor.GetFeatureFlagsReturns(
									[]v2action.FeatureFlag{
										{
											Name:    "some-feature-flag",
											Enabled: true,
										},
										{
											Name:    "service_instance_sharing",
											Enabled: true,
										},
									},
									v2action.Warnings{"get-feature-flags-warning"},
									nil)
							})

							When("no errors occur getting the spaces the service instance is shared to", func() {
								BeforeEach(func() {
									fakeV2Actor.GetServiceInstanceSharedTosByServiceInstanceReturns(
										[]v2action.ServiceInstanceSharedTo{
											{SpaceGUID: "already-shared-space-guid-1"},
											{SpaceGUID: "already-shared-space-guid-2"},
										},
										v2action.Warnings{"get-service-instance-shared-tos-warning"},
										nil)
								})

								When("the service instance is NOT already shared with this space", func() {
									When("no errors occur sharing the service instance to this space", func() {
										BeforeEach(func() {
											fakeV3Actor.ShareServiceInstanceToSpacesReturns(
												resources.RelationshipList{
													GUIDs: []string{"some-space-guid"},
												},
												v3action.Warnings{"share-service-instance-warning"},
												nil)
										})

										It("shares the service instance to this space and returns all warnings", func() {
											Expect(shareErr).ToNot(HaveOccurred())
											Expect(warnings).To(ConsistOf(
												"get-service-instance-warning",
												"get-space-warning",
												"get-service-warning",
												"get-feature-flags-warning",
												"get-service-instance-shared-tos-warning",
												"share-service-instance-warning"))

											Expect(fakeV2Actor.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
											serviceInstanceNameArg, spaceGUIDArg := fakeV2Actor.GetServiceInstanceByNameAndSpaceArgsForCall(0)
											Expect(serviceInstanceNameArg).To(Equal(serviceInstanceName))
											Expect(spaceGUIDArg).To(Equal(sourceSpaceGUID))

											Expect(fakeV2Actor.GetServiceCallCount()).To(Equal(1))
											Expect(fakeV2Actor.GetServiceArgsForCall(0)).To(Equal("some-service-guid"))

											Expect(fakeV2Actor.GetFeatureFlagsCallCount()).To(Equal(1))

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

									When("an error occurs sharing the service instance to this space", func() {
										var expectedErr error

										BeforeEach(func() {
											expectedErr = errors.New("share service instance error")
											fakeV3Actor.ShareServiceInstanceToSpacesReturns(
												resources.RelationshipList{},
												v3action.Warnings{"share-service-instance-warning"},
												expectedErr)
										})

										It("returns the error and all warnings", func() {
											Expect(shareErr).To(MatchError(expectedErr))
											Expect(warnings).To(ConsistOf(
												"get-service-instance-warning",
												"get-service-warning",
												"get-feature-flags-warning",
												"get-service-instance-shared-tos-warning",
												"get-space-warning",
												"share-service-instance-warning"))
										})
									})
								})

								When("the service instance IS already shared with this space", func() {
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
											"get-service-warning",
											"get-feature-flags-warning",
											"get-service-instance-shared-tos-warning",
											"get-space-warning",
										))
									})
								})
							})

							When("an error occurs getting the spaces the service instance is shared to", func() {
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
										"get-space-warning",
										"get-service-warning",
										"get-feature-flags-warning",
										"get-service-instance-shared-tos-warning"))
								})
							})
						})

						When("an error occurs getting feature flags", func() {
							var expectedErr error

							BeforeEach(func() {
								expectedErr = errors.New("get feature flags error")
								fakeV2Actor.GetFeatureFlagsReturns(
									nil,
									v2action.Warnings{"get-feature-flag-warning"},
									expectedErr)
							})

							It("returns the error and all warnings", func() {
								Expect(shareErr).To(MatchError(expectedErr))
								Expect(warnings).To(ConsistOf(
									"get-service-instance-warning",
									"get-space-warning",
									"get-service-warning",
									"get-feature-flag-warning"))
							})
						})
					})

					When("an error occurs getting the service", func() {
						var expectedErr error

						BeforeEach(func() {
							expectedErr = errors.New("get service error")
							fakeV2Actor.GetServiceReturns(
								v2action.Service{},
								v2action.Warnings{"get-service-warning"},
								expectedErr)
						})

						It("returns the error and all warnings", func() {
							Expect(shareErr).To(MatchError(expectedErr))
							Expect(warnings).To(ConsistOf(
								"get-service-instance-warning",
								"get-space-warning",
								"get-service-warning"))
						})
					})

					When("service sharing is globally disabled, and service sharing is disabled by the service broker", func() {
						BeforeEach(func() {
							fakeV2Actor.GetServiceReturns(
								v2action.Service{
									Extra: ccv2.ServiceExtra{
										Shareable: false,
									},
								},
								v2action.Warnings{"get-service-warning"},
								nil)

							fakeV2Actor.GetFeatureFlagsReturns(
								[]v2action.FeatureFlag{
									{
										Name:    "some-feature-flag",
										Enabled: true,
									},
									{
										Name:    "service_instance_sharing",
										Enabled: false,
									},
								},
								v2action.Warnings{"get-feature-flags-warning"},
								nil)
						})

						It("returns ServiceInstanceNotShareableError and all warnings", func() {
							Expect(shareErr).To(MatchError(actionerror.ServiceInstanceNotShareableError{
								FeatureFlagEnabled:          false,
								ServiceBrokerSharingEnabled: false}))
							Expect(warnings).To(ConsistOf(
								"get-service-instance-warning",
								"get-space-warning",
								"get-service-warning",
								"get-feature-flags-warning"))
						})
					})

					When("service sharing is globally enabled, and service sharing is disabled by the service broker", func() {
						BeforeEach(func() {
							fakeV2Actor.GetServiceReturns(
								v2action.Service{
									Extra: ccv2.ServiceExtra{
										Shareable: false,
									},
								},
								v2action.Warnings{"get-service-warning"},
								nil)

							fakeV2Actor.GetFeatureFlagsReturns(
								[]v2action.FeatureFlag{
									{
										Name:    "some-feature-flag",
										Enabled: true,
									},
									{
										Name:    "service_instance_sharing",
										Enabled: true,
									},
								},
								v2action.Warnings{"get-feature-flags-warning"},
								nil)
						})

						It("returns ServiceInstanceNotShareableError and all warnings", func() {
							Expect(shareErr).To(MatchError(actionerror.ServiceInstanceNotShareableError{
								FeatureFlagEnabled:          true,
								ServiceBrokerSharingEnabled: false}))
							Expect(warnings).To(ConsistOf(
								"get-service-instance-warning",
								"get-space-warning",
								"get-service-warning",
								"get-feature-flags-warning"))
						})
					})

					When("service sharing is globally disabled, and service sharing is enabled by the service broker", func() {
						BeforeEach(func() {
							fakeV2Actor.GetServiceReturns(
								v2action.Service{
									Extra: ccv2.ServiceExtra{
										Shareable: true,
									},
								},
								v2action.Warnings{"get-service-warning"},
								nil)

							fakeV2Actor.GetFeatureFlagsReturns(
								[]v2action.FeatureFlag{
									{
										Name:    "some-feature-flag",
										Enabled: true,
									},
									{
										Name:    "service_instance_sharing",
										Enabled: false,
									},
								},
								v2action.Warnings{"get-feature-flags-warning"},
								nil)
						})

						It("returns ServiceInstanceNotShareableError and all warnings", func() {
							Expect(shareErr).To(MatchError(actionerror.ServiceInstanceNotShareableError{
								FeatureFlagEnabled:          false,
								ServiceBrokerSharingEnabled: true}))
							Expect(warnings).To(ConsistOf(
								"get-service-instance-warning",
								"get-space-warning",
								"get-service-warning",
								"get-feature-flags-warning"))
						})
					})
				})

				When("the service instance is not a managed service instance", func() {
					BeforeEach(func() {
						fakeV2Actor.GetServiceInstanceByNameAndSpaceReturns(
							v2action.ServiceInstance{
								GUID:        "some-service-instance-guid",
								ServiceGUID: "some-service-guid",
								Type:        constant.UserProvidedService,
							},
							v2action.Warnings{"get-service-instance-warning"},
							nil)

						fakeV2Actor.GetSpaceByOrganizationAndNameReturns(
							v2action.Space{
								GUID: "some-space",
							},
							v2action.Warnings{"get-space-warning"},
							nil)

						fakeV3Actor.ShareServiceInstanceToSpacesReturns(
							resources.RelationshipList{},
							v3action.Warnings{"share-service-instance-warning"},
							errors.New("User-provided services cannot be shared"))
					})

					It("always returns the error and warnings", func() {
						Expect(shareErr).To(MatchError("User-provided services cannot be shared"))
						Expect(warnings).To(ConsistOf(
							"get-service-instance-warning",
							"get-space-warning",
							"share-service-instance-warning"))

						Expect(fakeV2Actor.GetServiceCallCount()).To(Equal(0))
						Expect(fakeV2Actor.GetFeatureFlagsCallCount()).To(Equal(0))
						Expect(fakeV2Actor.GetServiceInstanceSharedTosByServiceInstanceCallCount()).To(Equal(0))
					})
				})
			})

			When("an error occurs getting the space we are sharing to", func() {
				var expectedErr error

				BeforeEach(func() {
					fakeV2Actor.GetServiceInstanceByNameAndSpaceReturns(
						v2action.ServiceInstance{},
						v2action.Warnings{"get-service-instance-warning"},
						nil)

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
						"get-space-warning"))
				})
			})
		})

		When("an error occurs getting the service instance", func() {
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

		When("the service instance does not exist", func() {
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

		When("no errors occur getting the org", func() {
			BeforeEach(func() {
				fakeV3Actor.GetOrganizationByNameReturns(
					v3action.Organization{
						GUID: "some-org-guid",
					},
					v3action.Warnings{"get-org-warning"},
					nil)

				fakeV2Actor.GetServiceInstanceByNameAndSpaceReturns(
					v2action.ServiceInstance{
						GUID:        "some-service-instance-guid",
						ServiceGUID: "some-service-guid",
						Type:        constant.ManagedService,
					},
					v2action.Warnings{"get-service-instance-warning"},
					nil)

				fakeV2Actor.GetSpaceByOrganizationAndNameReturns(
					v2action.Space{
						GUID: "not-shared-space-guid",
					},
					v2action.Warnings{"get-space-warning"},
					nil)

				fakeV2Actor.GetServiceReturns(
					v2action.Service{
						Extra: ccv2.ServiceExtra{
							Shareable: true,
						},
					},
					v2action.Warnings{"get-service-warning"},
					nil)

				fakeV2Actor.GetFeatureFlagsReturns(
					[]v2action.FeatureFlag{
						{
							Name:    "some-feature-flag",
							Enabled: true,
						},
						{
							Name:    "service_instance_sharing",
							Enabled: true,
						},
					},
					v2action.Warnings{"get-feature-flags-warning"},
					nil)

				fakeV2Actor.GetServiceInstanceSharedTosByServiceInstanceReturns(
					[]v2action.ServiceInstanceSharedTo{
						{SpaceGUID: "already-shared-space-guid-1"},
						{SpaceGUID: "already-shared-space-guid-2"},
					},
					v2action.Warnings{"get-service-instance-shared-tos-warning"},
					nil)

				fakeV3Actor.ShareServiceInstanceToSpacesReturns(
					resources.RelationshipList{
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
					"get-space-warning",
					"get-service-warning",
					"get-feature-flags-warning",
					"get-service-instance-shared-tos-warning",
					"share-service-instance-warning"))

				Expect(fakeV3Actor.GetOrganizationByNameCallCount()).To(Equal(1))
				Expect(fakeV3Actor.GetOrganizationByNameArgsForCall(0)).To(Equal(shareToOrgName))

				Expect(fakeV2Actor.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
				serviceInstanceNameArg, spaceGUIDArg := fakeV2Actor.GetServiceInstanceByNameAndSpaceArgsForCall(0)
				Expect(serviceInstanceNameArg).To(Equal(serviceInstanceName))
				Expect(spaceGUIDArg).To(Equal(sourceSpaceGUID))

				Expect(fakeV2Actor.GetSpaceByOrganizationAndNameCallCount()).To(Equal(1))
				orgGUIDArg, spaceNameArg := fakeV2Actor.GetSpaceByOrganizationAndNameArgsForCall(0)
				Expect(orgGUIDArg).To(Equal("some-org-guid"))
				Expect(spaceNameArg).To(Equal(shareToSpaceName))

				Expect(fakeV2Actor.GetServiceCallCount()).To(Equal(1))
				Expect(fakeV2Actor.GetServiceArgsForCall(0)).To(Equal("some-service-guid"))

				Expect(fakeV2Actor.GetFeatureFlagsCallCount()).To(Equal(1))

				Expect(fakeV2Actor.GetServiceInstanceSharedTosByServiceInstanceCallCount()).To(Equal(1))
				Expect(fakeV2Actor.GetServiceInstanceSharedTosByServiceInstanceArgsForCall(0)).To(Equal("some-service-instance-guid"))

				Expect(fakeV3Actor.ShareServiceInstanceToSpacesCallCount()).To(Equal(1))
				serviceInstanceGUIDArg, spaceGUIDsArg := fakeV3Actor.ShareServiceInstanceToSpacesArgsForCall(0)
				Expect(serviceInstanceGUIDArg).To(Equal("some-service-instance-guid"))
				Expect(spaceGUIDsArg).To(Equal([]string{"not-shared-space-guid"}))
			})
		})

		When("an error occurs getting the org", func() {
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

	Describe("UnshareServiceInstanceFromOrganizationNameAndSpaceNameByNameAndSpace", func() {
		var (
			shareToOrgName             string
			shareToSpaceName           string
			serviceInstanceName        string
			currentlyTargetedSpaceGUID string

			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			shareToOrgName = "shared-to-org"
			shareToSpaceName = "shared-to-space"
			serviceInstanceName = "some-service-instance"
			currentlyTargetedSpaceGUID = "currently-targeted-space-guid"
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.UnshareServiceInstanceFromOrganizationNameAndSpaceNameByNameAndSpace(shareToOrgName, shareToSpaceName, serviceInstanceName, currentlyTargetedSpaceGUID)
		})

		When("no errors occur getting the service instance", func() {
			BeforeEach(func() {
				fakeV2Actor.GetServiceInstanceByNameAndSpaceReturns(
					v2action.ServiceInstance{
						GUID: "some-service-instance-guid",
					},
					v2action.Warnings{"get-service-instance-warning"},
					nil)
			})

			When("no errors occur getting the service instance's shared to spaces", func() {
				BeforeEach(func() {
					fakeV2Actor.GetServiceInstanceSharedTosByServiceInstanceReturns(
						[]v2action.ServiceInstanceSharedTo{
							{
								SpaceGUID:        "some-other-shared-to-space-guid",
								SpaceName:        "some-other-shared-to-space",
								OrganizationName: "some-other-shared-to-org",
							},
							{
								SpaceGUID:        "shared-to-space-guid",
								SpaceName:        "shared-to-space",
								OrganizationName: "shared-to-org",
							},
						},
						v2action.Warnings{"get-shared-tos-warning"},
						nil)
				})

				When("no errors occur unsharing the service instance", func() {
					BeforeEach(func() {
						fakeV3Actor.UnshareServiceInstanceByServiceInstanceAndSpaceReturns(
							v3action.Warnings{"unshare-warning"},
							nil)
					})

					It("returns no errors and returns all warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf(
							"get-service-instance-warning",
							"get-shared-tos-warning",
							"unshare-warning"))

						Expect(fakeV2Actor.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
						serviceInstanceNameArg, spaceGUIDArg := fakeV2Actor.GetServiceInstanceByNameAndSpaceArgsForCall(0)
						Expect(serviceInstanceNameArg).To(Equal(serviceInstanceName))
						Expect(spaceGUIDArg).To(Equal(currentlyTargetedSpaceGUID))

						Expect(fakeV2Actor.GetServiceInstanceSharedTosByServiceInstanceCallCount()).To(Equal(1))
						Expect(fakeV2Actor.GetServiceInstanceSharedTosByServiceInstanceArgsForCall(0)).To(Equal("some-service-instance-guid"))

						Expect(fakeV3Actor.UnshareServiceInstanceByServiceInstanceAndSpaceCallCount()).To(Equal(1))
						serviceInstanceGUIDArg, spaceGUIDArg := fakeV3Actor.UnshareServiceInstanceByServiceInstanceAndSpaceArgsForCall(0)
						Expect(serviceInstanceGUIDArg).To(Equal("some-service-instance-guid"))
						Expect(spaceGUIDArg).To(Equal("shared-to-space-guid"))
					})
				})

				When("an error occurs unsharing the service instance", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("unshare error")
						fakeV3Actor.UnshareServiceInstanceByServiceInstanceAndSpaceReturns(
							v3action.Warnings{"unshare-warning"},
							expectedErr)
					})

					It("returns the error and all warnings", func() {
						Expect(executeErr).To(MatchError(expectedErr))
						Expect(warnings).To(ConsistOf(
							"get-service-instance-warning",
							"get-shared-tos-warning",
							"unshare-warning"))
					})
				})
			})

			When("an error occurs getting the service instance's shared to spaces", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("get shared tos error")
					fakeV2Actor.GetServiceInstanceSharedTosByServiceInstanceReturns(
						nil,
						v2action.Warnings{"get-shared-tos-warning"},
						expectedErr)
				})

				It("returns the error and all warnings", func() {
					Expect(executeErr).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf(
						"get-service-instance-warning",
						"get-shared-tos-warning"))
				})
			})

			When("the service instance is not shared to the space we want to unshare with", func() {
				BeforeEach(func() {
					fakeV2Actor.GetServiceInstanceSharedTosByServiceInstanceReturns(
						[]v2action.ServiceInstanceSharedTo{
							{
								SpaceGUID:        "some-other-shared-to-space-guid",
								SpaceName:        "some-other-shared-to-space",
								OrganizationName: "some-other-shared-to-org",
							},
						},
						v2action.Warnings{"get-shared-tos-warning"},
						nil)
				})

				It("returns a ServiceInstanceNotSharedToSpaceError and all warnings", func() {
					Expect(executeErr).To(MatchError(actionerror.ServiceInstanceNotSharedToSpaceError{
						ServiceInstanceName: "some-service-instance",
					}))
					Expect(warnings).To(ConsistOf(
						"get-service-instance-warning",
						"get-shared-tos-warning"))
				})
			})
		})

		When("an error occurs getting the service instance", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("get service instance error")
				fakeV2Actor.GetServiceInstanceByNameAndSpaceReturns(
					v2action.ServiceInstance{},
					v2action.Warnings{"get-service-instance-warning"},
					expectedErr)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("get-service-instance-warning"))
			})
		})

		When("the service instance does not exist", func() {
			BeforeEach(func() {
				fakeV2Actor.GetServiceInstanceByNameAndSpaceReturns(
					v2action.ServiceInstance{},
					v2action.Warnings{"get-service-instance-warning"},
					actionerror.ServiceInstanceNotFoundError{})
			})

			It("returns a SharedServiceInstanceNotFoundError and all warnings", func() {
				Expect(executeErr).To(MatchError(actionerror.SharedServiceInstanceNotFoundError{}))
				Expect(warnings).To(ConsistOf("get-service-instance-warning"))
			})
		})
	})
})

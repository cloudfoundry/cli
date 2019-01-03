package v2action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Instance Summary Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil)
	})

	Describe("ServiceInstanceSummary", func() {
		var summary ServiceInstanceSummary

		Describe("IsShareable", func() {
			When("the 'service_instance_sharing' feature flag is enabled", func() {
				BeforeEach(func() {
					summary.ServiceInstanceSharingFeatureFlag = true
				})

				When("the service broker has enabled sharing", func() {
					BeforeEach(func() {
						summary.Service.Extra.Shareable = true
					})

					It("returns true", func() {
						Expect(summary.IsShareable()).To(BeTrue())
					})
				})

				When("the service broker has not enabled sharing", func() {
					BeforeEach(func() {
						summary.Service.Extra.Shareable = false
					})

					It("returns true", func() {
						Expect(summary.IsShareable()).To(BeFalse())
					})
				})
			})

			When("the 'service_instance_sharing' feature flag is not enabled", func() {
				BeforeEach(func() {
					summary.ServiceInstanceSharingFeatureFlag = false
				})

				When("the service broker has enabled sharing", func() {
					BeforeEach(func() {
						summary.Service.Extra.Shareable = true
					})

					It("returns true", func() {
						Expect(summary.IsShareable()).To(BeFalse())
					})
				})

				When("the service broker has not enabled sharing", func() {
					BeforeEach(func() {
						summary.Service.Extra.Shareable = false
					})

					It("returns true", func() {
						Expect(summary.IsShareable()).To(BeFalse())
					})
				})
			})
		})
	})

	Describe("GetServiceInstanceSummaryByNameAndSpace", func() {
		var (
			summary         ServiceInstanceSummary
			summaryWarnings Warnings
			summaryErr      error
		)

		JustBeforeEach(func() {
			summary, summaryWarnings, summaryErr = actor.GetServiceInstanceSummaryByNameAndSpace("some-service-instance", "some-space-guid")
		})

		When("an error is encountered getting the service instance", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("get space service instance error")
				fakeCloudControllerClient.GetSpaceServiceInstancesReturns(
					[]ccv2.ServiceInstance{},
					ccv2.Warnings{"get-space-service-instance-warning"},
					expectedErr)
			})

			It("returns the error and all warnings", func() {
				Expect(summaryErr).To(MatchError(expectedErr))
				Expect(summaryWarnings).To(ConsistOf("get-space-service-instance-warning"))

				Expect(fakeCloudControllerClient.GetSpaceServiceInstancesCallCount()).To(Equal(1))
				spaceGUIDArg, getUserProvidedServicesArg, queriesArg := fakeCloudControllerClient.GetSpaceServiceInstancesArgsForCall(0)
				Expect(spaceGUIDArg).To(Equal("some-space-guid"))
				Expect(getUserProvidedServicesArg).To(BeTrue())
				Expect(queriesArg).To(HaveLen(1))
				Expect(queriesArg[0]).To(Equal(ccv2.Filter{
					Type:     constant.NameFilter,
					Operator: constant.EqualOperator,
					Values:   []string{"some-service-instance"},
				}))
			})
		})

		When("no errors are encountered getting the service instance", func() {
			var (
				returnedServiceInstance ccv2.ServiceInstance
				returnedFeatureFlag     ccv2.FeatureFlag
			)

			When("the service instance is a managed service instance", func() {
				BeforeEach(func() {
					returnedServiceInstance = ccv2.ServiceInstance{
						DashboardURL:    "some-dashboard",
						GUID:            "some-service-instance-guid",
						Name:            "some-service-instance",
						ServiceGUID:     "some-service-guid",
						ServicePlanGUID: "some-service-plan-guid",
						Tags:            []string{"tag-1", "tag-2"},
						Type:            constant.ServiceInstanceTypeManagedService,
					}
					fakeCloudControllerClient.GetSpaceServiceInstancesReturns(
						[]ccv2.ServiceInstance{returnedServiceInstance},
						ccv2.Warnings{"get-space-service-instance-warning"},
						nil)

					returnedFeatureFlag = ccv2.FeatureFlag{
						Name:    "service_instance_sharing",
						Enabled: true,
					}
					fakeCloudControllerClient.GetConfigFeatureFlagsReturns(
						[]ccv2.FeatureFlag{returnedFeatureFlag},
						ccv2.Warnings{"get-feature-flags-warning"},
						nil)
				})

				It("returns the service instance info and all warnings", func() {
					Expect(summaryErr).ToNot(HaveOccurred())
					Expect(summary.ServiceInstance).To(Equal(ServiceInstance(returnedServiceInstance)))
					Expect(summaryWarnings).To(ConsistOf("get-space-service-instance-warning", "get-feature-flags-warning"))

					Expect(fakeCloudControllerClient.GetSpaceServiceInstancesCallCount()).To(Equal(1))
					spaceGUIDArg, getUserProvidedServicesArg, queriesArg := fakeCloudControllerClient.GetSpaceServiceInstancesArgsForCall(0)
					Expect(spaceGUIDArg).To(Equal("some-space-guid"))
					Expect(getUserProvidedServicesArg).To(BeTrue())
					Expect(queriesArg).To(HaveLen(1))
					Expect(queriesArg[0]).To(Equal(ccv2.Filter{
						Type:     constant.NameFilter,
						Operator: constant.EqualOperator,
						Values:   []string{"some-service-instance"},
					}))
				})

				When("the service instance is shared from another space (not created in the currently targeted space)", func() {
					When("the source space of the service instance is different from the currently targeted space", func() {
						BeforeEach(func() {
							returnedServiceInstance.SpaceGUID = "not-currently-targeted-space-guid"
							fakeCloudControllerClient.GetSpaceServiceInstancesReturns(
								[]ccv2.ServiceInstance{returnedServiceInstance},
								ccv2.Warnings{"get-space-service-instance-warning"},
								nil)
						})

						When("an error is encountered getting the shared_from information", func() {
							var expectedErr error

							When("the error is generic", func() {
								BeforeEach(func() {
									expectedErr = errors.New("get-service-instance-shared-from-error")
									fakeCloudControllerClient.GetServiceInstanceSharedFromReturns(
										ccv2.ServiceInstanceSharedFrom{},
										ccv2.Warnings{"get-service-instance-shared-from-warning"},
										expectedErr,
									)
								})

								It("returns the error and all warnings", func() {
									Expect(summaryErr).To(MatchError(expectedErr))
									Expect(summaryWarnings).To(ConsistOf("get-service-instance-shared-from-warning", "get-feature-flags-warning", "get-space-service-instance-warning"))

									Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromCallCount()).To(Equal(1))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromArgsForCall(0)).To(Equal(returnedServiceInstance.GUID))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedTosCallCount()).To(Equal(0))
								})
							})

							When("the API version does not support service instance sharing", func() {
								BeforeEach(func() {
									expectedErr = ccerror.ResourceNotFoundError{}
									fakeCloudControllerClient.GetServiceInstanceSharedFromReturns(
										ccv2.ServiceInstanceSharedFrom{},
										ccv2.Warnings{"get-service-instance-shared-from-warning"},
										expectedErr,
									)
								})

								It("ignores the 404 error and continues without shared_from information", func() {
									Expect(summaryErr).ToNot(HaveOccurred())
									Expect(summaryWarnings).To(ConsistOf("get-service-instance-shared-from-warning", "get-feature-flags-warning", "get-space-service-instance-warning"))
									Expect(summary.ServiceInstanceSharedFrom).To(Equal(ServiceInstanceSharedFrom{}))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromCallCount()).To(Equal(1))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromArgsForCall(0)).To(Equal(returnedServiceInstance.GUID))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedTosCallCount()).To(Equal(0))
								})
							})
						})

						When("no errors are encountered getting the shared_from information", func() {
							When("the shared_from info is NOT empty", func() {
								var returnedServiceSharedFrom ccv2.ServiceInstanceSharedFrom

								BeforeEach(func() {
									returnedServiceSharedFrom = ccv2.ServiceInstanceSharedFrom{
										SpaceGUID:        "some-space-guid",
										SpaceName:        "some-space-name",
										OrganizationName: "some-org-name",
									}
									fakeCloudControllerClient.GetServiceInstanceSharedFromReturns(
										returnedServiceSharedFrom,
										ccv2.Warnings{"get-service-instance-shared-from-warning"},
										nil)
								})

								It("returns the service instance share type, shared_from info, and all warnings", func() {
									Expect(summaryErr).ToNot(HaveOccurred())
									Expect(summary.ServiceInstance).To(Equal(ServiceInstance(returnedServiceInstance)))
									Expect(summary.ServiceInstanceSharingFeatureFlag).To(BeTrue())
									Expect(summary.ServiceInstanceShareType).To(Equal(ServiceInstanceIsSharedFrom))
									Expect(summary.ServiceInstanceSharedFrom).To(Equal(ServiceInstanceSharedFrom(returnedServiceSharedFrom)))
									Expect(summaryWarnings).To(ConsistOf("get-service-instance-shared-from-warning", "get-feature-flags-warning", "get-space-service-instance-warning"))

									Expect(fakeCloudControllerClient.GetConfigFeatureFlagsCallCount()).To(Equal(1))

									Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromCallCount()).To(Equal(1))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromArgsForCall(0)).To(Equal(returnedServiceInstance.GUID))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedTosCallCount()).To(Equal(0))
								})
							})

							When("the shared_from info is empty", func() {
								It("sets the share type to not shared", func() {
									Expect(summaryErr).ToNot(HaveOccurred())
									Expect(summary.ServiceInstanceShareType).To(Equal(ServiceInstanceIsNotShared))
								})
							})
						})
					})

					When("the source space of the service instance is 'null'", func() {
						BeforeEach(func() {
							// API returns a json null value that is unmarshalled into the empty string
							returnedServiceInstance.SpaceGUID = ""
							fakeCloudControllerClient.GetSpaceServiceInstancesReturns(
								[]ccv2.ServiceInstance{returnedServiceInstance},
								ccv2.Warnings{"get-space-service-instance-warning"},
								nil)
						})

						When("an error is encountered getting the shared_from information", func() {
							var expectedErr error

							When("the error is generic", func() {
								BeforeEach(func() {
									expectedErr = errors.New("get-service-instance-shared-from-error")
									fakeCloudControllerClient.GetServiceInstanceSharedFromReturns(
										ccv2.ServiceInstanceSharedFrom{},
										ccv2.Warnings{"get-service-instance-shared-from-warning"},
										expectedErr,
									)
								})

								It("returns the error and all warnings", func() {
									Expect(summaryErr).To(MatchError(expectedErr))
									Expect(summaryWarnings).To(ConsistOf("get-service-instance-shared-from-warning", "get-feature-flags-warning", "get-space-service-instance-warning"))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromCallCount()).To(Equal(1))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromArgsForCall(0)).To(Equal(returnedServiceInstance.GUID))
								})
							})

							When("the API version does not support service instance sharing", func() {
								BeforeEach(func() {
									expectedErr = ccerror.ResourceNotFoundError{}
									fakeCloudControllerClient.GetServiceInstanceSharedFromReturns(
										ccv2.ServiceInstanceSharedFrom{},
										ccv2.Warnings{"get-service-instance-shared-from-warning"},
										expectedErr,
									)
								})

								It("ignores the 404 error and continues without shared_from information", func() {
									Expect(summaryErr).ToNot(HaveOccurred())
									Expect(summaryWarnings).To(ConsistOf("get-service-instance-shared-from-warning", "get-feature-flags-warning", "get-space-service-instance-warning"))
									Expect(summary.ServiceInstanceSharedFrom).To(Equal(ServiceInstanceSharedFrom{}))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromCallCount()).To(Equal(1))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromArgsForCall(0)).To(Equal(returnedServiceInstance.GUID))
								})
							})
						})

						When("no errors are encountered getting the shared_from information", func() {
							When("the shared_from info is NOT empty", func() {
								var returnedServiceSharedFrom ccv2.ServiceInstanceSharedFrom

								BeforeEach(func() {
									returnedServiceSharedFrom = ccv2.ServiceInstanceSharedFrom{
										SpaceGUID:        "some-space-guid",
										SpaceName:        "some-space-name",
										OrganizationName: "some-org-name",
									}
									fakeCloudControllerClient.GetServiceInstanceSharedFromReturns(
										returnedServiceSharedFrom,
										ccv2.Warnings{"get-service-instance-shared-from-warning"},
										nil)
								})

								It("returns the service instance share type, shared_from info, and all warnings", func() {
									Expect(summaryErr).ToNot(HaveOccurred())
									Expect(summary.ServiceInstance).To(Equal(ServiceInstance(returnedServiceInstance)))
									Expect(summary.ServiceInstanceSharingFeatureFlag).To(BeTrue())
									Expect(summary.ServiceInstanceShareType).To(Equal(ServiceInstanceIsSharedFrom))
									Expect(summary.ServiceInstanceSharedFrom).To(Equal(ServiceInstanceSharedFrom(returnedServiceSharedFrom)))
									Expect(summaryWarnings).To(ConsistOf("get-space-service-instance-warning", "get-feature-flags-warning", "get-service-instance-shared-from-warning"))

									Expect(fakeCloudControllerClient.GetConfigFeatureFlagsCallCount()).To(Equal(1))

									Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromCallCount()).To(Equal(1))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromArgsForCall(0)).To(Equal(returnedServiceInstance.GUID))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedTosCallCount()).To(Equal(0))
								})
							})

							When("the shared_from info is empty", func() {
								It("sets the share type to not shared", func() {
									Expect(summaryErr).ToNot(HaveOccurred())
									Expect(summary.ServiceInstanceShareType).To(Equal(ServiceInstanceIsNotShared))
								})
							})
						})
					})
				})

				When("the service instance is shared to other spaces", func() {
					When("the source space of the service instance is the same as the currently targeted space", func() {
						BeforeEach(func() {
							returnedServiceInstance.SpaceGUID = "some-space-guid"
							fakeCloudControllerClient.GetSpaceServiceInstancesReturns(
								[]ccv2.ServiceInstance{returnedServiceInstance},
								ccv2.Warnings{"get-space-service-instance-warning"},
								nil)
						})

						When("an error is encountered getting the shared_to information", func() {
							var expectedErr error

							When("the error is generic", func() {
								BeforeEach(func() {
									expectedErr = errors.New("get-service-instance-shared-tos-error")
									fakeCloudControllerClient.GetServiceInstanceSharedTosReturns(
										[]ccv2.ServiceInstanceSharedTo{},
										ccv2.Warnings{"get-service-instance-shared-tos-warning"},
										expectedErr,
									)
								})

								It("returns the error and all warnings", func() {
									Expect(summaryErr).To(MatchError(expectedErr))
									Expect(summaryWarnings).To(ConsistOf("get-service-instance-shared-tos-warning", "get-feature-flags-warning", "get-space-service-instance-warning"))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedTosCallCount()).To(Equal(1))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedTosArgsForCall(0)).To(Equal(returnedServiceInstance.GUID))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromCallCount()).To(Equal(0))
								})
							})

							When("the API version does not support service instance sharing", func() {
								BeforeEach(func() {
									expectedErr = ccerror.ResourceNotFoundError{}
									fakeCloudControllerClient.GetServiceInstanceSharedTosReturns(
										[]ccv2.ServiceInstanceSharedTo{},
										ccv2.Warnings{"get-service-instance-shared-tos-warning"},
										expectedErr,
									)
								})

								It("ignores the 404 error and continues without shared_to information", func() {
									Expect(summaryErr).ToNot(HaveOccurred())
									Expect(summaryWarnings).To(ConsistOf("get-service-instance-shared-tos-warning", "get-feature-flags-warning", "get-space-service-instance-warning"))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedTosCallCount()).To(Equal(1))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedTosArgsForCall(0)).To(Equal(returnedServiceInstance.GUID))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromCallCount()).To(Equal(0))
								})
							})
						})

						When("no errors are encountered getting the shared_to information", func() {
							When("the shared_to info is NOT an empty list", func() {
								var returnedServiceSharedTos []ccv2.ServiceInstanceSharedTo

								BeforeEach(func() {
									returnedServiceSharedTos = []ccv2.ServiceInstanceSharedTo{
										{
											SpaceGUID:        "some-space-guid",
											SpaceName:        "some-space-name",
											OrganizationName: "some-org-name",
										},
										{
											SpaceGUID:        "some-space-guid2",
											SpaceName:        "some-space-name2",
											OrganizationName: "some-org-name2",
										},
									}

									fakeCloudControllerClient.GetServiceInstanceSharedTosReturns(
										returnedServiceSharedTos,
										ccv2.Warnings{"get-service-instance-shared-tos-warning"},
										nil)
								})

								It("returns the service instance share type, shared_to info, and all warnings", func() {
									Expect(summaryErr).ToNot(HaveOccurred())
									Expect(summary.ServiceInstance).To(Equal(ServiceInstance(returnedServiceInstance)))
									Expect(summary.ServiceInstanceSharingFeatureFlag).To(BeTrue())
									Expect(summary.ServiceInstanceShareType).To(Equal(ServiceInstanceIsSharedTo))
									Expect(summary.ServiceInstanceSharedTos).To(ConsistOf(ServiceInstanceSharedTo(returnedServiceSharedTos[0]), ServiceInstanceSharedTo(returnedServiceSharedTos[1])))
									Expect(summaryWarnings).To(ConsistOf("get-service-instance-shared-tos-warning", "get-feature-flags-warning", "get-space-service-instance-warning"))

									Expect(fakeCloudControllerClient.GetConfigFeatureFlagsCallCount()).To(Equal(1))

									Expect(fakeCloudControllerClient.GetServiceInstanceSharedTosCallCount()).To(Equal(1))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedTosArgsForCall(0)).To(Equal(returnedServiceInstance.GUID))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromCallCount()).To(Equal(0))
								})
							})

							When("the shared_to info is an empty list", func() {
								It("sets the share type to not shared", func() {
									Expect(summaryErr).ToNot(HaveOccurred())
									Expect(summary.ServiceInstanceShareType).To(Equal(ServiceInstanceIsNotShared))
								})
							})
						})
					})
				})

				When("an error is encountered getting the service plan", func() {
					Describe("a generic error", func() {
						var expectedErr error

						BeforeEach(func() {
							expectedErr = errors.New("get-service-plan-error")
							fakeCloudControllerClient.GetServicePlanReturns(
								ccv2.ServicePlan{},
								ccv2.Warnings{"get-service-plan-warning"},
								expectedErr)
						})

						It("returns the error and all warnings", func() {
							Expect(summaryErr).To(MatchError(expectedErr))
							Expect(summaryWarnings).To(ConsistOf("get-space-service-instance-warning", "get-feature-flags-warning", "get-service-plan-warning"))
						})
					})

					Describe("a Forbidden error", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.GetServicePlanReturns(
								ccv2.ServicePlan{},
								ccv2.Warnings{"get-service-plan-warning"},
								ccerror.ForbiddenError{})
						})

						It("returns warnings and continues on", func() {
							Expect(summaryErr).ToNot(HaveOccurred())
							Expect(summaryWarnings).To(ConsistOf("get-space-service-instance-warning", "get-feature-flags-warning", "get-service-plan-warning", "This org is not authorized to view necessary data about this service plan. Contact your administrator regarding service GUID some-service-plan-guid."))

							Expect(fakeCloudControllerClient.GetServiceCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.GetServiceArgsForCall(0)).To(Equal("some-service-guid"))
						})
					})
				})

				When("no errors are encountered getting the service plan", func() {
					var returnedServicePlan ccv2.ServicePlan

					BeforeEach(func() {
						returnedServicePlan = ccv2.ServicePlan{
							GUID:        "some-service-plan-guid",
							Name:        "some-service-plan",
							ServiceGUID: "some-service-guid",
						}
						fakeCloudControllerClient.GetServicePlanReturns(
							returnedServicePlan,
							ccv2.Warnings{"get-service-plan-warning"},
							nil)
					})

					It("returns the service plan info and all warnings", func() {
						Expect(summaryErr).ToNot(HaveOccurred())
						Expect(summary.ServiceInstance).To(Equal(ServiceInstance(returnedServiceInstance)))
						Expect(summary.ServicePlan).To(Equal(ServicePlan(returnedServicePlan)))
						Expect(summaryWarnings).To(ConsistOf("get-space-service-instance-warning", "get-feature-flags-warning", "get-service-plan-warning"))

						Expect(fakeCloudControllerClient.GetServicePlanCallCount()).To(Equal(1))
						Expect(fakeCloudControllerClient.GetServicePlanArgsForCall(0)).To(Equal(returnedServiceInstance.ServicePlanGUID))
					})

					When("an error is encountered getting the service", func() {
						Describe("a generic error", func() {
							var expectedErr error

							BeforeEach(func() {
								expectedErr = errors.New("get service error")
								fakeCloudControllerClient.GetServiceReturns(
									ccv2.Service{},
									ccv2.Warnings{"get-service-warning"},
									expectedErr)
							})

							It("returns the error and all warnings", func() {
								Expect(summaryErr).To(MatchError(expectedErr))
								Expect(summaryWarnings).To(ConsistOf("get-space-service-instance-warning", "get-feature-flags-warning", "get-service-plan-warning", "get-service-warning"))

								Expect(fakeCloudControllerClient.GetServiceCallCount()).To(Equal(1))
								Expect(fakeCloudControllerClient.GetServiceArgsForCall(0)).To(Equal(returnedServicePlan.ServiceGUID))
							})
						})

						Describe("a Forbidden error", func() {
							BeforeEach(func() {
								fakeCloudControllerClient.GetServiceReturns(
									ccv2.Service{},
									ccv2.Warnings{"get-service-warning"},
									ccerror.ForbiddenError{})
							})

							It("returns warnings and continues on", func() {
								Expect(summaryErr).ToNot(HaveOccurred())
								Expect(summaryWarnings).To(ConsistOf("get-space-service-instance-warning", "get-feature-flags-warning", "get-service-plan-warning", "get-service-warning", "This org is not authorized to view necessary data about this service. Contact your administrator regarding service GUID some-service-guid."))

								Expect(fakeCloudControllerClient.GetServiceInstanceServiceBindingsCallCount()).To(BeNumerically(">=", 1))
							})
						})
					})

					When("no errors are encountered getting the service", func() {
						var returnedService ccv2.Service

						BeforeEach(func() {
							returnedService = ccv2.Service{
								GUID:             "some-service-guid",
								Label:            "some-service",
								Description:      "some-description",
								DocumentationURL: "some-url",
								Extra: ccv2.ServiceExtra{
									Shareable: true,
								},
							}

							fakeCloudControllerClient.GetServiceReturns(
								returnedService,
								ccv2.Warnings{"get-service-warning"},
								nil)
						})

						It("returns the service info and all warnings", func() {
							Expect(summaryErr).ToNot(HaveOccurred())
							Expect(summary.ServiceInstance).To(Equal(ServiceInstance(returnedServiceInstance)))
							Expect(summary.ServicePlan).To(Equal(ServicePlan(returnedServicePlan)))
							Expect(summary.Service).To(Equal(Service(returnedService)))
							Expect(summaryWarnings).To(ConsistOf("get-space-service-instance-warning", "get-feature-flags-warning", "get-service-plan-warning", "get-service-warning"))

							Expect(fakeCloudControllerClient.GetServiceCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.GetServiceArgsForCall(0)).To(Equal(returnedServicePlan.ServiceGUID))
						})

						When("an error is encountered getting the service bindings", func() {
							var expectedErr error

							BeforeEach(func() {
								expectedErr = errors.New("get service bindings error")
								fakeCloudControllerClient.GetServiceInstanceServiceBindingsReturns(
									[]ccv2.ServiceBinding{},
									ccv2.Warnings{"get-service-bindings-warning"},
									expectedErr)
							})

							It("returns the error and all warnings", func() {
								Expect(summaryErr).To(MatchError(expectedErr))
								Expect(summaryWarnings).To(ConsistOf("get-space-service-instance-warning", "get-feature-flags-warning", "get-service-plan-warning", "get-service-warning", "get-service-bindings-warning"))

								Expect(fakeCloudControllerClient.GetServiceInstanceServiceBindingsCallCount()).To(Equal(1))
								Expect(fakeCloudControllerClient.GetServiceInstanceServiceBindingsArgsForCall(0)).To(Equal(returnedServiceInstance.GUID))
							})

							When("a ResourceNotFoundError is encountered getting the service bindings", func() {
								var expectedErr error

								BeforeEach(func() {
									expectedErr = ccerror.ResourceNotFoundError{}
									fakeCloudControllerClient.GetServiceInstanceServiceBindingsReturns(
										[]ccv2.ServiceBinding{},
										ccv2.Warnings{"get-service-bindings-warning"},
										expectedErr)
								})

								It("returns no error", func() {
									Expect(summaryErr).ToNot(HaveOccurred())
									Expect(summaryWarnings).To(ConsistOf("get-space-service-instance-warning", "get-feature-flags-warning", "get-service-plan-warning", "get-service-warning", "get-service-bindings-warning"))

									Expect(fakeCloudControllerClient.GetServiceInstanceServiceBindingsCallCount()).To(Equal(1))
									Expect(fakeCloudControllerClient.GetServiceInstanceServiceBindingsArgsForCall(0)).To(Equal(returnedServiceInstance.GUID))
								})
							})

						})

						When("no errors are encountered getting the service bindings", func() {
							var returnedServiceBindings []ccv2.ServiceBinding

							BeforeEach(func() {
								returnedServiceBindings = []ccv2.ServiceBinding{
									{
										GUID:          "some-service-binding-1-guid",
										Name:          "some-service-binding-1",
										AppGUID:       "some-app-1-guid",
										LastOperation: ccv2.LastOperation{Type: "create", State: constant.LastOperationInProgress, Description: "10% complete"},
									},
									{
										GUID:          "some-service-binding-2-guid",
										Name:          "some-service-binding-2",
										AppGUID:       "some-app-2-guid",
										LastOperation: ccv2.LastOperation{Type: "delete", State: constant.LastOperationSucceeded, Description: "100% complete"},
									},
								}
								fakeCloudControllerClient.GetServiceInstanceServiceBindingsReturns(
									returnedServiceBindings,
									ccv2.Warnings{"get-service-bindings-warning"},
									nil)
							})

							When("an error is encountered getting bound application info", func() {
								var expectedErr error

								BeforeEach(func() {
									expectedErr = errors.New("get application error")
									fakeCloudControllerClient.GetApplicationReturns(
										ccv2.Application{},
										ccv2.Warnings{"get-application-warning"},
										expectedErr)
								})

								It("returns the error", func() {
									Expect(summaryErr).To(MatchError(expectedErr))
									Expect(summaryWarnings).To(ConsistOf("get-space-service-instance-warning", "get-feature-flags-warning", "get-service-plan-warning", "get-service-warning", "get-service-bindings-warning", "get-application-warning"))

									Expect(fakeCloudControllerClient.GetApplicationCallCount()).To(Equal(1))
									Expect(fakeCloudControllerClient.GetApplicationArgsForCall(0)).To(Equal(returnedServiceBindings[0].AppGUID))
								})
							})

							When("no errors are encountered getting bound application info", func() {
								BeforeEach(func() {
									fakeCloudControllerClient.GetApplicationReturnsOnCall(
										0,
										ccv2.Application{
											GUID: "some-app-1-guid",
											Name: "some-app-1",
										},
										ccv2.Warnings{"get-application-warning-1"},
										nil)
									fakeCloudControllerClient.GetApplicationReturnsOnCall(
										1,
										ccv2.Application{
											GUID: "some-app-2-guid",
											Name: "some-app-2",
										},
										ccv2.Warnings{"get-application-warning-2"},
										nil)
								})

								It("returns a list of applications bound to the service instance and all warnings", func() {
									Expect(summaryErr).ToNot(HaveOccurred())
									Expect(summary.ServiceInstance).To(Equal(ServiceInstance(returnedServiceInstance)))
									Expect(summary.ServicePlan).To(Equal(ServicePlan(returnedServicePlan)))
									Expect(summary.Service).To(Equal(Service(returnedService)))
									Expect(summary.BoundApplications).To(Equal([]BoundApplication{
										{
											AppName:            "some-app-1",
											ServiceBindingName: "some-service-binding-1",
											LastOperation: LastOperation{
												Type:        "create",
												State:       constant.LastOperationInProgress,
												Description: "10% complete",
											},
										},
										{
											AppName:            "some-app-2",
											ServiceBindingName: "some-service-binding-2",
											LastOperation: LastOperation{
												Type:        "delete",
												State:       constant.LastOperationSucceeded,
												Description: "100% complete",
											},
										},
									}))
									Expect(summaryWarnings).To(ConsistOf("get-space-service-instance-warning", "get-feature-flags-warning", "get-service-plan-warning", "get-service-warning", "get-service-bindings-warning", "get-application-warning-1", "get-application-warning-2"))

									Expect(fakeCloudControllerClient.GetServiceInstanceServiceBindingsCallCount()).To(Equal(1))
									Expect(fakeCloudControllerClient.GetServiceInstanceServiceBindingsArgsForCall(0)).To(Equal(returnedServiceInstance.GUID))

									Expect(fakeCloudControllerClient.GetApplicationCallCount()).To(Equal(2))
									Expect(fakeCloudControllerClient.GetApplicationArgsForCall(0)).To(Equal(returnedServiceBindings[0].AppGUID))
									Expect(fakeCloudControllerClient.GetApplicationArgsForCall(1)).To(Equal(returnedServiceBindings[1].AppGUID))

									Expect(fakeCloudControllerClient.GetUserProvidedServiceInstanceServiceBindingsCallCount()).To(Equal(0))
								})
							})
						})
					})
				})
			})

			When("the service instance is a user provided service instance", func() {
				BeforeEach(func() {
					returnedServiceInstance = ccv2.ServiceInstance{
						GUID: "some-user-provided-service-instance-guid",
						Name: "some-user-provided-service-instance",
						Type: constant.ServiceInstanceTypeUserProvidedService,
					}
					fakeCloudControllerClient.GetSpaceServiceInstancesReturns(
						[]ccv2.ServiceInstance{returnedServiceInstance},
						ccv2.Warnings{"get-space-service-instance-warning"},
						nil)
				})

				Context("getting the service bindings errors", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetUserProvidedServiceInstanceServiceBindingsReturns(
							nil,
							ccv2.Warnings{"some-get-user-provided-si-bindings-warnings"},
							errors.New("some-get-user-provided-si-bindings-error"))
					})

					It("should return the error and return all warnings", func() {
						Expect(summaryErr).To(MatchError("some-get-user-provided-si-bindings-error"))
						Expect(summaryWarnings).To(ConsistOf("some-get-user-provided-si-bindings-warnings",
							"get-space-service-instance-warning"))
					})
				})

				Context("getting the service bindings errors with ResourceNotFoundError", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetUserProvidedServiceInstanceServiceBindingsReturns(
							nil,
							ccv2.Warnings{"some-get-user-provided-si-bindings-warnings"},
							ccerror.ResourceNotFoundError{})
					})

					It("should return the error and return all warnings", func() {
						Expect(summaryErr).ToNot(HaveOccurred())
						Expect(summaryWarnings).To(ConsistOf("some-get-user-provided-si-bindings-warnings",
							"get-space-service-instance-warning"))
					})
				})

				When("no errors are encountered getting the service bindings", func() {
					var returnedServiceBindings []ccv2.ServiceBinding

					BeforeEach(func() {
						returnedServiceBindings = []ccv2.ServiceBinding{
							{
								GUID:    "some-service-binding-1-guid",
								Name:    "some-service-binding-1",
								AppGUID: "some-app-1-guid",
							},
							{
								GUID:    "some-service-binding-2-guid",
								Name:    "some-service-binding-2",
								AppGUID: "some-app-2-guid",
							},
						}
						fakeCloudControllerClient.GetUserProvidedServiceInstanceServiceBindingsReturns(
							returnedServiceBindings,
							ccv2.Warnings{"get-service-bindings-warning"},
							nil)
					})

					When("no errors are encountered getting bound application info", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.GetApplicationReturnsOnCall(
								0,
								ccv2.Application{
									GUID: "some-app-1-guid",
									Name: "some-app-1",
								},
								ccv2.Warnings{"get-application-warning-1"},
								nil)
							fakeCloudControllerClient.GetApplicationReturnsOnCall(
								1,
								ccv2.Application{
									GUID: "some-app-2-guid",
									Name: "some-app-2",
								},
								ccv2.Warnings{"get-application-warning-2"},
								nil)
						})

						It("returns a list of applications bound to the service instance and all warnings", func() {
							Expect(summaryErr).ToNot(HaveOccurred())
							Expect(summary).To(Equal(ServiceInstanceSummary{
								ServiceInstance: ServiceInstance(returnedServiceInstance),
								BoundApplications: []BoundApplication{
									{
										AppName:            "some-app-1",
										ServiceBindingName: "some-service-binding-1",
									},
									{
										AppName:            "some-app-2",
										ServiceBindingName: "some-service-binding-2",
									},
								},
							}))
							Expect(summaryWarnings).To(ConsistOf("get-space-service-instance-warning", "get-service-bindings-warning", "get-application-warning-1", "get-application-warning-2"))

							Expect(fakeCloudControllerClient.GetSpaceServiceInstancesCallCount()).To(Equal(1))
							spaceGUIDArg, getUserProvidedServicesArg, queriesArg := fakeCloudControllerClient.GetSpaceServiceInstancesArgsForCall(0)
							Expect(spaceGUIDArg).To(Equal("some-space-guid"))
							Expect(getUserProvidedServicesArg).To(BeTrue())
							Expect(queriesArg).To(HaveLen(1))
							Expect(queriesArg[0]).To(Equal(ccv2.Filter{
								Type:     constant.NameFilter,
								Operator: constant.EqualOperator,
								Values:   []string{"some-service-instance"},
							}))

							Expect(fakeCloudControllerClient.GetUserProvidedServiceInstanceServiceBindingsCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.GetUserProvidedServiceInstanceServiceBindingsArgsForCall(0)).To(Equal(returnedServiceInstance.GUID))

							Expect(fakeCloudControllerClient.GetApplicationCallCount()).To(Equal(2))
							Expect(fakeCloudControllerClient.GetApplicationArgsForCall(0)).To(Equal(returnedServiceBindings[0].AppGUID))
							Expect(fakeCloudControllerClient.GetApplicationArgsForCall(1)).To(Equal(returnedServiceBindings[1].AppGUID))

							Expect(fakeCloudControllerClient.GetServicePlanCallCount()).To(Equal(0))
							Expect(fakeCloudControllerClient.GetServiceCallCount()).To(Equal(0))
							Expect(fakeCloudControllerClient.GetServiceInstanceServiceBindingsCallCount()).To(Equal(0))
						})
					})
				})
			})
		})
	})

	Describe("GetServiceInstancesSummaryBySpace", func() {
		var (
			serviceInstancesSummary []ServiceInstanceSummary
			warnings                Warnings
			executeErr              error
		)

		JustBeforeEach(func() {
			serviceInstancesSummary, warnings, executeErr = actor.GetServiceInstancesSummaryBySpace("some-space-GUID")
		})

		When("an error is encountered getting a space's service instances", func() {
			var expectedErr error
			BeforeEach(func() {
				expectedErr = errors.New("get by space service instance summary error")
				fakeCloudControllerClient.GetSpaceServiceInstancesReturns(
					[]ccv2.ServiceInstance{},
					ccv2.Warnings{"get-by-space-service-instances-warning"},
					expectedErr,
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("get-by-space-service-instances-warning"))
			})
		})

		When("no errors are encountered getting a space's service instances", func() {
			var (
				serviceInstance1 ccv2.ServiceInstance
				serviceInstance2 ccv2.ServiceInstance
				application1     ccv2.Application
				application2     ccv2.Application
				bindings1        []ccv2.ServiceBinding
				bindings2        []ccv2.ServiceBinding
			)

			BeforeEach(func() {
				serviceInstance1 = ccv2.ServiceInstance{
					GUID:            "some-si-GUID-1",
					Name:            "some-si-name-1",
					ServiceGUID:     "some-service-GUID-1",
					ServicePlanGUID: "some-si-sp-GUID-1",
					Type:            constant.ServiceInstanceTypeManagedService,
					LastOperation: ccv2.LastOperation{
						Type:  "some-lo-type",
						State: "some-lo-state",
					},
				}
				serviceInstance2 = ccv2.ServiceInstance{
					GUID:            "some-si-GUID-3",
					Name:            "some-si-name-3",
					ServiceGUID:     "some-service-GUID-3",
					ServicePlanGUID: "some-si-sp-GUID-3",
					Type:            constant.ServiceInstanceTypeUserProvidedService,
					LastOperation: ccv2.LastOperation{
						Type:  "some-lo-type",
						State: "some-lo-state",
					},
				}

				fakeCloudControllerClient.GetSpaceServiceInstancesReturns(
					[]ccv2.ServiceInstance{serviceInstance2, serviceInstance1},
					ccv2.Warnings{"get-by-space-service-instances-warning"},
					nil)

				bindings1 = []ccv2.ServiceBinding{
					{
						GUID:                "some-sb-GUID-1",
						Name:                "some-service-binding-2",
						AppGUID:             "2-app-GUID", // we are testing that the app name will be sorted alphanumerically
						ServiceInstanceGUID: "some-si-GUID-1",
					},
					{
						GUID:                "some-sb-GUID-2",
						Name:                "some-service-binding-1",
						AppGUID:             "1-app-GUID", // we are testing that the app name will be sorted alphanumerically
						ServiceInstanceGUID: "some-si-GUID-3",
					},
				}
				bindings2 = []ccv2.ServiceBinding{
					{
						GUID:                "some-sb-GUID-1",
						Name:                "some-service-binding-1",
						AppGUID:             "1-app-GUID",
						ServiceInstanceGUID: "some-si-GUID-1",
					},
				}
				fakeCloudControllerClient.GetServiceInstanceServiceBindingsReturns(
					bindings1,
					ccv2.Warnings{"some-bindings-warning"},
					nil)
				fakeCloudControllerClient.GetUserProvidedServiceInstanceServiceBindingsReturns(
					bindings2,
					ccv2.Warnings{"some-bindings-warning"},
					nil)

				application1 = ccv2.Application{Name: "1-app-name", GUID: "1-app-GUID"}
				application2 = ccv2.Application{Name: "2-app-name", GUID: "2-app-GUID"}

				fakeCloudControllerClient.GetApplicationStub = func(appGUID string) (ccv2.Application, ccv2.Warnings, error) {
					switch appGUID {
					case application1.GUID:
						return application1, ccv2.Warnings{"some-get-app-warning"}, nil
					case application2.GUID:
						return application2, ccv2.Warnings{"some-get-app-warning"}, nil
					default:
						Fail("got an app guid that does not exist")
					}

					return ccv2.Application{}, nil, nil
				}
			})

			It("returns the service instances summary with bound apps in alphanumeric sorted order and all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(serviceInstancesSummary).To(Equal(
					[]ServiceInstanceSummary{
						{
							ServiceInstance: ServiceInstance(serviceInstance1),
							BoundApplications: []BoundApplication{
								{
									AppName:            "1-app-name",
									ServiceBindingName: "some-service-binding-1",
								},
								{
									AppName:            "2-app-name",
									ServiceBindingName: "some-service-binding-2",
								},
							},
						},
						{
							ServiceInstance: ServiceInstance(serviceInstance2),
							BoundApplications: []BoundApplication{
								{
									AppName:            "1-app-name",
									ServiceBindingName: "some-service-binding-1",
								},
							},
						}}))
				Expect(warnings).To(ConsistOf("get-by-space-service-instances-warning",
					"some-get-app-warning", "some-get-app-warning", "some-get-app-warning",
					"some-bindings-warning", "some-bindings-warning"))

				Expect(fakeCloudControllerClient.GetSpaceServiceInstancesCallCount()).To(Equal(1))
				passedSpaceGUID, includeUPS, filters := fakeCloudControllerClient.GetSpaceServiceInstancesArgsForCall(0)
				Expect(passedSpaceGUID).To(Equal("some-space-GUID"))
				Expect(includeUPS).To(BeTrue())
				Expect(filters).To(BeEmpty())

				Expect(fakeCloudControllerClient.GetConfigFeatureFlagsCallCount()).To(Equal(0))

				//Managed expectations
				Expect(fakeCloudControllerClient.GetServicePlanCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServiceCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServiceInstanceServiceBindingsCallCount()).To(Equal(1))
				//User-Provided expectations
				Expect(fakeCloudControllerClient.GetUserProvidedServiceInstanceServiceBindingsCallCount()).To(Equal(1))
				//Both types expectations
				Expect(fakeCloudControllerClient.GetApplicationCallCount()).To(Equal(3))
			})
		})
	})
})

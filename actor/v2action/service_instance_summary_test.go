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

var _ = Describe("Service Instance Summary Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil)
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

		Context("when an error is encountered getting the service instance", func() {
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
				Expect(queriesArg[0]).To(Equal(ccv2.QQuery{
					Filter:   ccv2.NameFilter,
					Operator: ccv2.EqualOperator,
					Values:   []string{"some-service-instance"},
				}))
			})
		})

		Context("when no errors are encountered getting the service instance", func() {
			var returnedServiceInstance ccv2.ServiceInstance

			Context("when the service instance is a managed service instance", func() {
				BeforeEach(func() {
					returnedServiceInstance = ccv2.ServiceInstance{
						GUID:            "some-service-instance-guid",
						Name:            "some-service-instance",
						Type:            ccv2.ManagedService,
						Tags:            []string{"tag-1", "tag-2"},
						DashboardURL:    "some-dashboard",
						ServicePlanGUID: "some-service-plan-guid",
					}
					fakeCloudControllerClient.GetSpaceServiceInstancesReturns(
						[]ccv2.ServiceInstance{returnedServiceInstance},
						ccv2.Warnings{"get-space-service-instance-warning"},
						nil)
				})

				It("returns the service instance info and all warnings", func() {
					Expect(summaryErr).ToNot(HaveOccurred())
					Expect(summary).To(Equal(ServiceInstanceSummary{
						ServiceInstance: ServiceInstance(returnedServiceInstance),
					}))
					Expect(summaryWarnings).To(ConsistOf("get-space-service-instance-warning"))

					Expect(fakeCloudControllerClient.GetSpaceServiceInstancesCallCount()).To(Equal(1))
					spaceGUIDArg, getUserProvidedServicesArg, queriesArg := fakeCloudControllerClient.GetSpaceServiceInstancesArgsForCall(0)
					Expect(spaceGUIDArg).To(Equal("some-space-guid"))
					Expect(getUserProvidedServicesArg).To(BeTrue())
					Expect(queriesArg).To(HaveLen(1))
					Expect(queriesArg[0]).To(Equal(ccv2.QQuery{
						Filter:   ccv2.NameFilter,
						Operator: ccv2.EqualOperator,
						Values:   []string{"some-service-instance"},
					}))
				})

				Context("when an error is encountered getting the shared_from object", func() {
					var expectedErr error

					Context("when the error is not a 404 error", func() {
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
							Expect(summaryWarnings).To(ConsistOf("get-service-instance-shared-from-warning", "get-space-service-instance-warning"))
							Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromArgsForCall(0)).To(Equal(returnedServiceInstance.GUID))
						})
					})

					Context("when the error is a http 404 error", func() {
						BeforeEach(func() {
							expectedErr = ccerror.ResourceNotFoundError{}
							fakeCloudControllerClient.GetServiceInstanceSharedFromReturns(
								ccv2.ServiceInstanceSharedFrom{},
								ccv2.Warnings{"get-service-instance-shared-from-warning"},
								expectedErr,
							)
						})

						It("ignores the error and continues without a shared_from object", func() {
							Expect(summaryErr).ToNot(HaveOccurred())
							Expect(summaryWarnings).To(ConsistOf("get-service-instance-shared-from-warning", "get-space-service-instance-warning"))
							Expect(summary.ServiceInstanceSharedFrom).To(Equal(ServiceInstanceSharedFrom{}))
							Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromArgsForCall(0)).To(Equal(returnedServiceInstance.GUID))
						})
					})
				})

				Context("when no errors are encountered getting the shared_from object", func() {
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

					It("returns the service instance shared_from info and all warnings", func() {
						Expect(summaryErr).ToNot(HaveOccurred())
						Expect(summary).To(Equal(ServiceInstanceSummary{
							ServiceInstance:           ServiceInstance(returnedServiceInstance),
							ServiceInstanceSharedFrom: ServiceInstanceSharedFrom(returnedServiceSharedFrom),
						}))
						Expect(summaryWarnings).To(ConsistOf("get-service-instance-shared-from-warning", "get-space-service-instance-warning"))

						Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromCallCount()).To(Equal(1))
						Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromArgsForCall(0)).To(Equal(returnedServiceInstance.GUID))
					})

					Context("when an error is encountered getting the service plan", func() {
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
							Expect(summaryWarnings).To(ConsistOf("get-service-instance-shared-from-warning", "get-space-service-instance-warning", "get-service-plan-warning"))

							Expect(fakeCloudControllerClient.GetServicePlanCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.GetServicePlanArgsForCall(0)).To(Equal(returnedServiceInstance.ServicePlanGUID))
						})
					})

					Context("when no errors are encountered getting the service plan", func() {
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
							Expect(summary).To(Equal(ServiceInstanceSummary{
								ServiceInstance:           ServiceInstance(returnedServiceInstance),
								ServiceInstanceSharedFrom: ServiceInstanceSharedFrom(returnedServiceSharedFrom),
								ServicePlan:               ServicePlan(returnedServicePlan),
							}))
							Expect(summaryWarnings).To(ConsistOf("get-service-instance-shared-from-warning", "get-space-service-instance-warning", "get-service-plan-warning"))

							Expect(fakeCloudControllerClient.GetServicePlanCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.GetServicePlanArgsForCall(0)).To(Equal(returnedServiceInstance.ServicePlanGUID))
						})

						Context("when an error is encountered getting the service", func() {
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
								Expect(summaryWarnings).To(ConsistOf("get-service-instance-shared-from-warning", "get-space-service-instance-warning", "get-service-plan-warning", "get-service-warning"))

								Expect(fakeCloudControllerClient.GetServiceCallCount()).To(Equal(1))
								Expect(fakeCloudControllerClient.GetServiceArgsForCall(0)).To(Equal(returnedServicePlan.ServiceGUID))
							})
						})

						Context("when no errors are encountered getting the service", func() {
							var returnedService ccv2.Service

							BeforeEach(func() {
								returnedService = ccv2.Service{
									GUID:             "some-service-guid",
									Label:            "some-service",
									Description:      "some-description",
									DocumentationURL: "some-url",
								}
								fakeCloudControllerClient.GetServiceReturns(
									returnedService,
									ccv2.Warnings{"get-service-warning"},
									nil)
							})

							It("returns the service info and all warnings", func() {
								Expect(summaryErr).ToNot(HaveOccurred())
								Expect(summary).To(Equal(ServiceInstanceSummary{
									ServiceInstance:           ServiceInstance(returnedServiceInstance),
									ServiceInstanceSharedFrom: ServiceInstanceSharedFrom(returnedServiceSharedFrom),
									ServicePlan:               ServicePlan(returnedServicePlan),
									Service:                   Service(returnedService),
								}))
								Expect(summaryWarnings).To(ConsistOf("get-service-instance-shared-from-warning", "get-space-service-instance-warning", "get-service-plan-warning", "get-service-warning"))

								Expect(fakeCloudControllerClient.GetServiceCallCount()).To(Equal(1))
								Expect(fakeCloudControllerClient.GetServiceArgsForCall(0)).To(Equal(returnedServicePlan.ServiceGUID))
							})

							Context("when an error is encountered getting the service bindings", func() {
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
									Expect(summaryWarnings).To(ConsistOf("get-service-instance-shared-from-warning", "get-space-service-instance-warning", "get-service-plan-warning", "get-service-warning", "get-service-bindings-warning"))

									Expect(fakeCloudControllerClient.GetServiceInstanceServiceBindingsCallCount()).To(Equal(1))
									Expect(fakeCloudControllerClient.GetServiceInstanceServiceBindingsArgsForCall(0)).To(Equal(returnedServiceInstance.GUID))
								})
							})

							Context("when no errors are encountered getting the service bindings", func() {
								var returnedServiceBindings []ccv2.ServiceBinding

								BeforeEach(func() {
									returnedServiceBindings = []ccv2.ServiceBinding{
										{
											GUID:    "some-service-binding-1-guid",
											AppGUID: "some-app-1-guid",
										},
										{
											GUID:    "some-service-binding-2-guid",
											AppGUID: "some-app-2-guid",
										},
									}
									fakeCloudControllerClient.GetServiceInstanceServiceBindingsReturns(
										returnedServiceBindings,
										ccv2.Warnings{"get-service-bindings-warning"},
										nil)
								})

								Context("when an error is encountered getting bound application info", func() {
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
										Expect(summaryWarnings).To(ConsistOf("get-service-instance-shared-from-warning", "get-space-service-instance-warning", "get-service-plan-warning", "get-service-warning", "get-service-bindings-warning", "get-application-warning"))

										Expect(fakeCloudControllerClient.GetApplicationCallCount()).To(Equal(1))
										Expect(fakeCloudControllerClient.GetApplicationArgsForCall(0)).To(Equal(returnedServiceBindings[0].AppGUID))
									})
								})

								Context("when no errors are encountered getting bound application info", func() {
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
											ServiceInstance:           ServiceInstance(returnedServiceInstance),
											ServiceInstanceSharedFrom: ServiceInstanceSharedFrom(returnedServiceSharedFrom),
											ServicePlan:               ServicePlan(returnedServicePlan),
											Service:                   Service(returnedService),
											BoundApplications:         []string{"some-app-1", "some-app-2"},
										}))
										Expect(summaryWarnings).To(ConsistOf("get-service-instance-shared-from-warning", "get-space-service-instance-warning", "get-service-plan-warning", "get-service-warning", "get-service-bindings-warning", "get-application-warning-1", "get-application-warning-2"))

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
			})

			Context("when the service instance is a user provided service instance", func() {
				BeforeEach(func() {
					returnedServiceInstance = ccv2.ServiceInstance{
						GUID: "some-user-provided-service-instance-guid",
						Name: "some-user-provided-service-instance",
						Type: ccv2.UserProvidedService,
					}
					fakeCloudControllerClient.GetSpaceServiceInstancesReturns(
						[]ccv2.ServiceInstance{returnedServiceInstance},
						ccv2.Warnings{"get-space-service-instance-warning"},
						nil)
				})

				Context("when no errors are encountered getting the service bindings", func() {
					var returnedServiceBindings []ccv2.ServiceBinding

					BeforeEach(func() {
						returnedServiceBindings = []ccv2.ServiceBinding{
							{
								GUID:    "some-service-binding-1-guid",
								AppGUID: "some-app-1-guid",
							},
							{
								GUID:    "some-service-binding-2-guid",
								AppGUID: "some-app-2-guid",
							},
						}
						fakeCloudControllerClient.GetUserProvidedServiceInstanceServiceBindingsReturns(
							returnedServiceBindings,
							ccv2.Warnings{"get-service-bindings-warning"},
							nil)
					})

					Context("when no errors are encountered getting bound application info", func() {
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
								ServiceInstance:   ServiceInstance(returnedServiceInstance),
								BoundApplications: []string{"some-app-1", "some-app-2"},
							}))
							Expect(summaryWarnings).To(ConsistOf("get-space-service-instance-warning", "get-service-bindings-warning", "get-application-warning-1", "get-application-warning-2"))

							Expect(fakeCloudControllerClient.GetSpaceServiceInstancesCallCount()).To(Equal(1))
							spaceGUIDArg, getUserProvidedServicesArg, queriesArg := fakeCloudControllerClient.GetSpaceServiceInstancesArgsForCall(0)
							Expect(spaceGUIDArg).To(Equal("some-space-guid"))
							Expect(getUserProvidedServicesArg).To(BeTrue())
							Expect(queriesArg).To(HaveLen(1))
							Expect(queriesArg[0]).To(Equal(ccv2.QQuery{
								Filter:   ccv2.NameFilter,
								Operator: ccv2.EqualOperator,
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
})

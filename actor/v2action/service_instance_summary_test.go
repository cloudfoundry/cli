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
					Expect(summary.ServiceInstance).To(Equal(ServiceInstance(returnedServiceInstance)))
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

				Context("when the service instance is shared from another space (not created in the currently targeted space)", func() {
					Context("when the source space of the service instance is different from the currently targeted space", func() {
						BeforeEach(func() {
							returnedServiceInstance.SpaceGUID = "not-currently-targeted-space-guid"
							fakeCloudControllerClient.GetSpaceServiceInstancesReturns(
								[]ccv2.ServiceInstance{returnedServiceInstance},
								ccv2.Warnings{"get-space-service-instance-warning"},
								nil)
						})

						Context("when an error is encountered getting the shared_from information", func() {
							var expectedErr error

							Context("when the error is generic", func() {
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
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedTosCallCount()).To(Equal(0))
								})
							})

							Context("when the API version does not support service instance sharing", func() {
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
									Expect(summaryWarnings).To(ConsistOf("get-service-instance-shared-from-warning", "get-space-service-instance-warning"))
									Expect(summary.ServiceInstanceSharedFrom).To(Equal(ServiceInstanceSharedFrom{}))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromCallCount()).To(Equal(1))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromArgsForCall(0)).To(Equal(returnedServiceInstance.GUID))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedTosCallCount()).To(Equal(0))
								})
							})
						})

						Context("when no errors are encountered getting the shared_from information", func() {
							Context("when the shared_from info is NOT empty", func() {
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
									Expect(summary.ServiceInstanceShareType).To(Equal(ServiceInstanceIsSharedFrom))
									Expect(summary.ServiceInstanceSharedFrom).To(Equal(ServiceInstanceSharedFrom(returnedServiceSharedFrom)))
									Expect(summaryWarnings).To(ConsistOf("get-service-instance-shared-from-warning", "get-space-service-instance-warning"))

									Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromCallCount()).To(Equal(1))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromArgsForCall(0)).To(Equal(returnedServiceInstance.GUID))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedTosCallCount()).To(Equal(0))
								})
							})

							Context("when the shared_from info is empty", func() {
								It("sets the share type to not shared", func() {
									Expect(summaryErr).ToNot(HaveOccurred())
									Expect(summary.ServiceInstanceShareType).To(Equal(ServiceInstanceIsNotShared))
								})
							})
						})
					})

					Context("when the source space of the service instance is 'null'", func() {
						BeforeEach(func() {
							// API returns a json null value that is unmarshalled into the empty string
							returnedServiceInstance.SpaceGUID = ""
							fakeCloudControllerClient.GetSpaceServiceInstancesReturns(
								[]ccv2.ServiceInstance{returnedServiceInstance},
								ccv2.Warnings{"get-space-service-instance-warning"},
								nil)
						})

						Context("when an error is encountered getting the shared_from information", func() {
							var expectedErr error

							Context("when the error is generic", func() {
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

							Context("when the API version does not support service instance sharing", func() {
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
									Expect(summaryWarnings).To(ConsistOf("get-service-instance-shared-from-warning", "get-space-service-instance-warning"))
									Expect(summary.ServiceInstanceSharedFrom).To(Equal(ServiceInstanceSharedFrom{}))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromCallCount()).To(Equal(1))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromArgsForCall(0)).To(Equal(returnedServiceInstance.GUID))
								})
							})
						})

						Context("when no errors are encountered getting the shared_from information", func() {
							Context("when the shared_from info is NOT empty", func() {
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
									Expect(summary.ServiceInstanceShareType).To(Equal(ServiceInstanceIsSharedFrom))
									Expect(summary.ServiceInstanceSharedFrom).To(Equal(ServiceInstanceSharedFrom(returnedServiceSharedFrom)))
									Expect(summaryWarnings).To(ConsistOf("get-service-instance-shared-from-warning", "get-space-service-instance-warning"))

									Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromCallCount()).To(Equal(1))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromArgsForCall(0)).To(Equal(returnedServiceInstance.GUID))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedTosCallCount()).To(Equal(0))
								})
							})

							Context("when the shared_from info is empty", func() {
								It("sets the share type to not shared", func() {
									Expect(summaryErr).ToNot(HaveOccurred())
									Expect(summary.ServiceInstanceShareType).To(Equal(ServiceInstanceIsNotShared))
								})
							})
						})
					})
				})

				Context("when the service instance is shared to other spaces", func() {
					Context("when the source space of the service instance is the same as the currently targeted space", func() {
						BeforeEach(func() {
							returnedServiceInstance.SpaceGUID = "some-space-guid"
							fakeCloudControllerClient.GetSpaceServiceInstancesReturns(
								[]ccv2.ServiceInstance{returnedServiceInstance},
								ccv2.Warnings{"get-space-service-instance-warning"},
								nil)
						})

						Context("when an error is encountered getting the shared_to information", func() {
							var expectedErr error

							Context("when the error is generic", func() {
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
									Expect(summaryWarnings).To(ConsistOf("get-service-instance-shared-tos-warning", "get-space-service-instance-warning"))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedTosCallCount()).To(Equal(1))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedTosArgsForCall(0)).To(Equal(returnedServiceInstance.GUID))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromCallCount()).To(Equal(0))
								})
							})

							Context("when the API version does not support service instance sharing", func() {
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
									Expect(summaryWarnings).To(ConsistOf("get-service-instance-shared-tos-warning", "get-space-service-instance-warning"))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedTosCallCount()).To(Equal(1))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedTosArgsForCall(0)).To(Equal(returnedServiceInstance.GUID))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromCallCount()).To(Equal(0))
								})
							})
						})

						Context("when no errors are encountered getting the shared_to information", func() {
							Context("when the shared_to info is NOT an empty list", func() {
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
									Expect(summary.ServiceInstanceShareType).To(Equal(ServiceInstanceIsSharedTo))
									Expect(summary.ServiceInstanceSharedTos).To(ConsistOf(ServiceInstanceSharedTo(returnedServiceSharedTos[0]), ServiceInstanceSharedTo(returnedServiceSharedTos[1])))
									Expect(summaryWarnings).To(ConsistOf("get-service-instance-shared-tos-warning", "get-space-service-instance-warning"))

									Expect(fakeCloudControllerClient.GetServiceInstanceSharedTosCallCount()).To(Equal(1))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedTosArgsForCall(0)).To(Equal(returnedServiceInstance.GUID))
									Expect(fakeCloudControllerClient.GetServiceInstanceSharedFromCallCount()).To(Equal(0))
								})
							})

							Context("when the shared_to info is an empty list", func() {
								It("sets the share type to not shared", func() {
									Expect(summaryErr).ToNot(HaveOccurred())
									Expect(summary.ServiceInstanceShareType).To(Equal(ServiceInstanceIsNotShared))
								})
							})
						})
					})
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
						Expect(summaryWarnings).To(ConsistOf("get-space-service-instance-warning", "get-service-plan-warning"))

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
						Expect(summary.ServiceInstance).To(Equal(ServiceInstance(returnedServiceInstance)))
						Expect(summary.ServicePlan).To(Equal(ServicePlan(returnedServicePlan)))
						Expect(summaryWarnings).To(ConsistOf("get-space-service-instance-warning", "get-service-plan-warning"))

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
							Expect(summaryWarnings).To(ConsistOf("get-space-service-instance-warning", "get-service-plan-warning", "get-service-warning"))

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
							Expect(summary.ServiceInstance).To(Equal(ServiceInstance(returnedServiceInstance)))
							Expect(summary.ServicePlan).To(Equal(ServicePlan(returnedServicePlan)))
							Expect(summary.Service).To(Equal(Service(returnedService)))
							Expect(summaryWarnings).To(ConsistOf("get-space-service-instance-warning", "get-service-plan-warning", "get-service-warning"))

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
								Expect(summaryWarnings).To(ConsistOf("get-space-service-instance-warning", "get-service-plan-warning", "get-service-warning", "get-service-bindings-warning"))

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
									Expect(summaryWarnings).To(ConsistOf("get-space-service-instance-warning", "get-service-plan-warning", "get-service-warning", "get-service-bindings-warning", "get-application-warning"))

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
									Expect(summary.ServiceInstance).To(Equal(ServiceInstance(returnedServiceInstance)))
									Expect(summary.ServicePlan).To(Equal(ServicePlan(returnedServicePlan)))
									Expect(summary.Service).To(Equal(Service(returnedService)))
									Expect(summary.BoundApplications).To(Equal([]string{"some-app-1", "some-app-2"}))
									Expect(summaryWarnings).To(ConsistOf("get-space-service-instance-warning", "get-service-plan-warning", "get-service-warning", "get-service-bindings-warning", "get-application-warning-1", "get-application-warning-2"))

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

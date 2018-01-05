package v2action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Instance Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil)
	})

	Describe("GetServiceInstance", func() {
		var (
			serviceInstanceGUID string

			serviceInstance ServiceInstance
			warnings        Warnings
			executeErr      error
		)

		BeforeEach(func() {
			serviceInstanceGUID = "service-instance-guid"
		})

		JustBeforeEach(func() {
			serviceInstance, warnings, executeErr = actor.GetServiceInstance(serviceInstanceGUID)
		})

		Context("when the service instance exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceReturns(ccv2.ServiceInstance{Name: "some-service-instance", GUID: "service-instance-guid"}, ccv2.Warnings{"service-instance-warnings"}, nil)
			})

			It("returns the service instance and warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(serviceInstance).To(Equal(ServiceInstance{
					GUID: "service-instance-guid",
					Name: "some-service-instance",
				}))
				Expect(warnings).To(Equal(Warnings{"service-instance-warnings"}))

				Expect(fakeCloudControllerClient.GetServiceInstanceCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServiceInstanceArgsForCall(0)).To(Equal(serviceInstanceGUID))
			})
		})

		Context("when the service instance does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceInstanceReturns(ccv2.ServiceInstance{}, ccv2.Warnings{"service-instance-warnings-1"}, ccerror.ResourceNotFoundError{})
			})

			It("returns errors and warnings", func() {
				Expect(executeErr).To(MatchError(actionerror.ServiceInstanceNotFoundError{GUID: serviceInstanceGUID}))
				Expect(warnings).To(ConsistOf("service-instance-warnings-1"))
			})
		})

		Context("when retrieving the application's bound services returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("this is indeed an error, kudos!")
				fakeCloudControllerClient.GetServiceInstanceReturns(ccv2.ServiceInstance{}, ccv2.Warnings{"service-instance-warnings-1"}, expectedErr)
			})

			It("returns errors and warnings", func() {
				Expect(executeErr).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("service-instance-warnings-1"))
			})
		})
	})

	Describe("GetServiceInstanceByNameAndSpace", func() {
		Context("when the service instance exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceServiceInstancesReturns(
					[]ccv2.ServiceInstance{
						{
							GUID: "some-service-instance-guid",
							Name: "some-service-instance",
						},
					},
					ccv2.Warnings{"foo"},
					nil,
				)
			})

			It("returns the service instance and warnings", func() {
				serviceInstance, warnings, err := actor.GetServiceInstanceByNameAndSpace("some-service-instance", "some-space-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(serviceInstance).To(Equal(ServiceInstance{
					GUID: "some-service-instance-guid",
					Name: "some-service-instance",
				}))
				Expect(warnings).To(Equal(Warnings{"foo"}))

				Expect(fakeCloudControllerClient.GetSpaceServiceInstancesCallCount()).To(Equal(1))

				spaceGUID, includeUserProvidedServices, queries := fakeCloudControllerClient.GetSpaceServiceInstancesArgsForCall(0)
				Expect(spaceGUID).To(Equal("some-space-guid"))
				Expect(includeUserProvidedServices).To(BeTrue())
				Expect(queries).To(ConsistOf([]ccv2.QQuery{
					ccv2.QQuery{
						Filter:   ccv2.NameFilter,
						Operator: ccv2.EqualOperator,
						Values:   []string{"some-service-instance"},
					},
				}))
			})
		})

		Context("when the service instance does not exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceServiceInstancesReturns([]ccv2.ServiceInstance{}, nil, nil)
			})

			It("returns a ServiceInstanceNotFoundError", func() {
				_, _, err := actor.GetServiceInstanceByNameAndSpace("some-service-instance", "some-space-guid")
				Expect(err).To(MatchError(actionerror.ServiceInstanceNotFoundError{Name: "some-service-instance"}))
			})
		})

		Context("when the cloud controller client returns an error", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.GetSpaceServiceInstancesReturns([]ccv2.ServiceInstance{}, nil, expectedError)
			})

			It("returns the error", func() {
				_, _, err := actor.GetServiceInstanceByNameAndSpace("some-service-instance", "some-space-guid")
				Expect(err).To(MatchError(expectedError))
			})
		})
	})

	Describe("GetServiceInstancesByApplication", func() {
		var (
			appGUID string

			serviceInstances []ServiceInstance
			warnings         Warnings
			executeErr       error
		)

		BeforeEach(func() {
			appGUID = "some-app-guid"
		})

		JustBeforeEach(func() {
			serviceInstances, warnings, executeErr = actor.GetServiceInstancesByApplication(appGUID)
		})

		Context("when the application has services bound", func() {
			var serviceBindings []ccv2.ServiceBinding

			BeforeEach(func() {
				serviceBindings = []ccv2.ServiceBinding{
					{ServiceInstanceGUID: "service-instance-guid-1"},
					{ServiceInstanceGUID: "service-instance-guid-2"},
					{ServiceInstanceGUID: "service-instance-guid-3"},
				}

				fakeCloudControllerClient.GetServiceBindingsReturns(serviceBindings, ccv2.Warnings{"service-bindings-warnings-1", "service-bindings-warnings-2"}, nil)
			})

			Context("when retrieving the service instances is successful", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstanceReturnsOnCall(0, ccv2.ServiceInstance{Name: "some-service-instance-1"}, ccv2.Warnings{"service-instance-warnings-1"}, nil)
					fakeCloudControllerClient.GetServiceInstanceReturnsOnCall(1, ccv2.ServiceInstance{Name: "some-service-instance-2"}, ccv2.Warnings{"service-instance-warnings-2"}, nil)
					fakeCloudControllerClient.GetServiceInstanceReturnsOnCall(2, ccv2.ServiceInstance{Name: "some-service-instance-3"}, ccv2.Warnings{"service-instance-warnings-3"}, nil)
				})

				It("returns the service instances and warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("service-bindings-warnings-1", "service-bindings-warnings-2", "service-instance-warnings-1", "service-instance-warnings-2", "service-instance-warnings-3"))
					Expect(serviceInstances).To(ConsistOf(
						ServiceInstance{Name: "some-service-instance-1"},
						ServiceInstance{Name: "some-service-instance-2"},
						ServiceInstance{Name: "some-service-instance-3"},
					))

					Expect(fakeCloudControllerClient.GetServiceInstanceCallCount()).To(Equal(3))
					Expect(fakeCloudControllerClient.GetServiceInstanceArgsForCall(0)).To(Equal("service-instance-guid-1"))
					Expect(fakeCloudControllerClient.GetServiceInstanceArgsForCall(1)).To(Equal("service-instance-guid-2"))
					Expect(fakeCloudControllerClient.GetServiceInstanceArgsForCall(2)).To(Equal("service-instance-guid-3"))
				})
			})

			Context("when retrieving the service instances returns an error", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("this is indeed an error, kudos!")
					fakeCloudControllerClient.GetServiceInstanceReturns(ccv2.ServiceInstance{}, ccv2.Warnings{"service-instance-warnings-1", "service-instance-warnings-2"}, expectedErr)
				})

				It("returns errors and warnings", func() {
					Expect(executeErr).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("service-bindings-warnings-1", "service-bindings-warnings-2", "service-instance-warnings-1", "service-instance-warnings-2"))
				})
			})
		})

		Context("when the application has no services bound", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceBindingsReturns(nil, ccv2.Warnings{"service-bindings-warnings-1", "service-bindings-warnings-2"}, nil)
			})

			It("returns an empty list and warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("service-bindings-warnings-1", "service-bindings-warnings-2"))
				Expect(serviceInstances).To(BeEmpty())
			})
		})

		Context("when retrieving the application's bound services returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("this is indeed an error, kudos!")
				fakeCloudControllerClient.GetServiceBindingsReturns(nil, ccv2.Warnings{"service-bindings-warnings-1", "service-bindings-warnings-2"}, expectedErr)
			})

			It("returns errors and warnings", func() {
				Expect(executeErr).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("service-bindings-warnings-1", "service-bindings-warnings-2"))
			})
		})
	})

	Describe("GetServiceInstancesBySpace", func() {
		Context("when there are service instances", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceServiceInstancesReturns(
					[]ccv2.ServiceInstance{
						{
							GUID: "some-service-instance-guid-1",
							Name: "some-service-instance-1",
						},
						{
							GUID: "some-service-instance-guid-2",
							Name: "some-service-instance-2",
						},
					},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil,
				)
			})

			It("returns the service instances and warnings", func() {
				serviceInstances, warnings, err := actor.GetServiceInstancesBySpace("some-space-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(serviceInstances).To(ConsistOf(
					ServiceInstance{
						GUID: "some-service-instance-guid-1",
						Name: "some-service-instance-1",
					},
					ServiceInstance{
						GUID: "some-service-instance-guid-2",
						Name: "some-service-instance-2",
					},
				))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

				Expect(fakeCloudControllerClient.GetSpaceServiceInstancesCallCount()).To(Equal(1))

				spaceGUID, includeUserProvidedServices, queries := fakeCloudControllerClient.GetSpaceServiceInstancesArgsForCall(0)
				Expect(spaceGUID).To(Equal("some-space-guid"))
				Expect(includeUserProvidedServices).To(BeTrue())
				Expect(queries).To(BeNil())
			})
		})

		Context("when the cloud controller client returns an error", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.GetSpaceServiceInstancesReturns(
					[]ccv2.ServiceInstance{},
					ccv2.Warnings{"warning-1", "warning-2"},
					expectedError)
			})

			It("returns the error and warnings", func() {
				_, warnings, err := actor.GetServiceInstancesBySpace("some-space-guid")
				Expect(err).To(MatchError(expectedError))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})

	Describe("GetSharedToSpaceGUID", func() {
		var (
			spaceGUID  string
			warnings   Warnings
			executeErr error

			serviceInstanceName string
			sourceSpaceGUID     string
			sharedToOrgName     string
			sharedToSpaceName   string
		)

		BeforeEach(func() {
			sourceSpaceGUID = "some-source-space-guid"
			serviceInstanceName = "some-service-instance"
		})

		JustBeforeEach(func() {
			spaceGUID, warnings, executeErr = actor.GetSharedToSpaceGUID(serviceInstanceName, sourceSpaceGUID, sharedToOrgName, sharedToSpaceName)
		})

		Context("when the service instance exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceServiceInstancesReturns(
					[]ccv2.ServiceInstance{
						{
							GUID: "some-service-instance-guid",
							Name: "some-service-instance",
						},
					},
					ccv2.Warnings{"get-space-service-instances-warning"},
					nil,
				)
			})

			It("calls GetSpaceServiceInstance with the correct service instance name and space guid", func() {
				Expect(fakeCloudControllerClient.GetSpaceServiceInstancesCallCount()).To(Equal(1))
				spaceGUIDArg, _, queryArg := fakeCloudControllerClient.GetSpaceServiceInstancesArgsForCall(0)

				Expect(spaceGUIDArg).To(Equal("some-source-space-guid"))
				Expect(queryArg[0].Values[0]).To(Equal("some-service-instance"))
			})

			It("calls GetServiceInstanceSharedTos with the correct service instance guid", func() {
				Expect(fakeCloudControllerClient.GetSpaceServiceInstancesCallCount()).To(Equal(1))
				serviceInstanceGUIDArg := fakeCloudControllerClient.GetServiceInstanceSharedTosArgsForCall(0)

				Expect(serviceInstanceGUIDArg).To(Equal("some-service-instance-guid"))
			})

			Context("when the service instance is shared with one other space", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstanceSharedTosReturns(
						[]ccv2.ServiceInstanceSharedTo{
							{
								SpaceGUID:        "shared-to-space-guid",
								SpaceName:        "shared-to-space-name",
								OrganizationName: "shared-to-org-name",
							},
						},
						ccv2.Warnings{"get-service-instance-shared-tos-warning"},
						nil,
					)
				})

				Context("and is shared with the specified org", func() {
					BeforeEach(func() {
						sharedToOrgName = "shared-to-org-name"
					})

					Context("and is shared with the specified space", func() {
						BeforeEach(func() {
							sharedToSpaceName = "shared-to-space-name"
						})

						It("returns the space guid and all warnings", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(spaceGUID).To(Equal("shared-to-space-guid"))
							Expect(warnings).To(Equal(Warnings{
								"get-space-service-instances-warning",
								"get-service-instance-shared-tos-warning"}))
						})
					})

					Context("and is shared with a space with the same name but different capitalization as the specified space", func() {
						BeforeEach(func() {
							sharedToSpaceName = "ShArEd-To-SpAcE-nAmE"
						})

						It("returns the space guid and all warnings", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(spaceGUID).To(Equal("shared-to-space-guid"))
							Expect(warnings).To(Equal(Warnings{
								"get-space-service-instances-warning",
								"get-service-instance-shared-tos-warning"}))
						})
					})
				})

				Context("and is shared with an org with the same name but different capitalization as the specified org", func() {
					BeforeEach(func() {
						sharedToOrgName = "Shared-To-Org-Name"
					})

					Context("and is shared with the specified space", func() {
						BeforeEach(func() {
							sharedToSpaceName = "shared-to-space-name"
						})

						It("returns the space guid and all warnings", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(spaceGUID).To(Equal("shared-to-space-guid"))
							Expect(warnings).To(Equal(Warnings{
								"get-space-service-instances-warning",
								"get-service-instance-shared-tos-warning"}))
						})
					})
				})
			})

			Context("when the service instance is shared with multiple other spaces", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstanceSharedTosReturns(
						[]ccv2.ServiceInstanceSharedTo{
							{
								SpaceGUID:        "shared-to-space-guid-first",
								SpaceName:        "shared-to-space-name-first",
								OrganizationName: "shared-to-org-name-first",
							},
							{
								SpaceGUID:        "shared-to-space-guid",
								SpaceName:        "shared-to-space-name",
								OrganizationName: "shared-to-org-name",
							},
							{
								SpaceGUID:        "shared-to-space-guid-last",
								SpaceName:        "shared-to-space-name-last",
								OrganizationName: "shared-to-org-name-last",
							},
						},
						ccv2.Warnings{"get-service-instance-shared-tos-warning"},
						nil,
					)
				})

				Context("and is shared with the specified org", func() {
					BeforeEach(func() {
						sharedToOrgName = "shared-to-org-name"
					})

					Context("and is shared with the specified space", func() {
						BeforeEach(func() {
							sharedToSpaceName = "shared-to-space-name"
						})

						It("returns the space guid and all warnings", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(spaceGUID).To(Equal("shared-to-space-guid"))
							Expect(warnings).To(Equal(Warnings{
								"get-space-service-instances-warning",
								"get-service-instance-shared-tos-warning"}))
						})
					})

					Context("and is not shared with the specified space", func() {
						BeforeEach(func() {
							sharedToSpaceName = "some-other-space-name"
						})

						It("returns an error and all warnings", func() {
							Expect(executeErr).To(MatchError(actionerror.ServiceInstanceNotSharedToSpaceError{ServiceInstanceName: serviceInstanceName}))
							Expect(warnings).To(Equal(Warnings{
								"get-space-service-instances-warning",
								"get-service-instance-shared-tos-warning"}))
						})
					})
				})

				Context("and is not shared with the specified org", func() {
					BeforeEach(func() {
						sharedToOrgName = "some-other-org-name"
					})

					Context("and is shared with a space with the same name as the specified space", func() {
						BeforeEach(func() {
							sharedToSpaceName = "shared-to-space-name"
						})

						It("returns an error and all warnings", func() {
							Expect(executeErr).To(MatchError(actionerror.ServiceInstanceNotSharedToSpaceError{ServiceInstanceName: serviceInstanceName}))
							Expect(warnings).To(Equal(Warnings{
								"get-space-service-instances-warning",
								"get-service-instance-shared-tos-warning"}))
						})
					})
				})
			})

			Context("when the service instance is not shared", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstanceSharedTosReturns(
						[]ccv2.ServiceInstanceSharedTo{},
						ccv2.Warnings{"get-service-instance-shared-tos-warning"},
						nil,
					)
				})

				It("returns an error and all warnings", func() {
					Expect(executeErr).To(MatchError(actionerror.ServiceInstanceNotSharedToSpaceError{ServiceInstanceName: serviceInstanceName}))
					Expect(warnings).To(Equal(Warnings{
						"get-space-service-instances-warning",
						"get-service-instance-shared-tos-warning"}))
				})
			})

			Context("when getting the shared-to information fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstanceSharedTosReturns(
						[]ccv2.ServiceInstanceSharedTo{},
						ccv2.Warnings{"get-service-instance-shared-tos-warning"},
						errors.New("some-shared-to-api-failure"),
					)
				})

				It("returns an error and warnings", func() {
					Expect(executeErr).To(MatchError("some-shared-to-api-failure"))
					Expect(warnings).To(Equal(Warnings{
						"get-space-service-instances-warning",
						"get-service-instance-shared-tos-warning"}))
				})
			})
		})

		Context("when the service instance does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceServiceInstancesReturns(
					[]ccv2.ServiceInstance{},
					ccv2.Warnings{"get-space-service-instances-warning"},
					nil,
				)
			})

			It("returns an error and warnings", func() {
				Expect(executeErr).To(MatchError(actionerror.ServiceInstanceNotFoundError{Name: serviceInstanceName, GUID: ""}))
				Expect(warnings).To(Equal(Warnings{"get-space-service-instances-warning"}))
			})
		})

		Context("when retrieving the service instance returns an error ", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceServiceInstancesReturns(
					[]ccv2.ServiceInstance{},
					ccv2.Warnings{"get-space-service-instances-warning"},
					errors.New("oops"),
				)
			})

			It("returns an error and warnings", func() {
				Expect(executeErr).To(MatchError(errors.New("oops")))
				Expect(warnings).To(Equal(Warnings{"get-space-service-instances-warning"}))
			})
		})
	})
})

package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service App Binding Action", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil, nil)
	})

	Describe("CreateServiceAppBinding", func() {
		const (
			serviceInstanceName = "fake-service-instance-name"
			serviceInstanceGUID = "fake-service-instance-guid"
			appName             = "fake-app-name"
			appGUID             = "fake-app-guid"
			bindingName         = "fake-binding-name"
			spaceGUID           = "fake-space-guid"
			fakeJobURL          = ccv3.JobURL("fake-job-url")
		)

		var (
			params         CreateServiceAppBindingParams
			warnings       Warnings
			executionError error
			stream         chan PollJobEvent
		)

		BeforeEach(func() {
			fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
				resources.ServiceInstance{
					Name: serviceInstanceName,
					GUID: serviceInstanceGUID,
				},
				ccv3.IncludedResources{},
				ccv3.Warnings{"get instance warning"},
				nil,
			)

			fakeCloudControllerClient.GetApplicationByNameAndSpaceReturns(
				resources.Application{
					GUID: appGUID,
					Name: appName,
				},
				ccv3.Warnings{"get app warning"},
				nil,
			)

			fakeCloudControllerClient.CreateServiceCredentialBindingReturns(
				fakeJobURL,
				ccv3.Warnings{"create binding warning"},
				nil,
			)

			fakeStream := make(chan ccv3.PollJobEvent)
			fakeCloudControllerClient.PollJobToEventStreamReturns(fakeStream)
			go func() {
				fakeStream <- ccv3.PollJobEvent{
					State:    constant.JobPolling,
					Warnings: ccv3.Warnings{"poll warning"},
				}
			}()

			params = CreateServiceAppBindingParams{
				SpaceGUID:           spaceGUID,
				ServiceInstanceName: serviceInstanceName,
				AppName:             appName,
				BindingName:         bindingName,
				Parameters: types.NewOptionalObject(map[string]interface{}{
					"foo": "bar",
				}),
			}
		})

		JustBeforeEach(func() {
			stream, warnings, executionError = actor.CreateServiceAppBinding(params)
		})

		It("returns an event stream, warnings, and no errors", func() {
			Expect(executionError).NotTo(HaveOccurred())

			Expect(warnings).To(ConsistOf(Warnings{
				"get instance warning",
				"get app warning",
				"create binding warning",
			}))

			Eventually(stream).Should(Receive(Equal(PollJobEvent{
				State:    JobPolling,
				Warnings: Warnings{"poll warning"},
				Err:      nil,
			})))
		})

		Describe("service instance lookup", func() {
			It("makes the correct call", func() {
				Expect(fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
				actualServiceInstanceName, actualSpaceGUID, actualQuery := fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceArgsForCall(0)
				Expect(actualServiceInstanceName).To(Equal(serviceInstanceName))
				Expect(actualSpaceGUID).To(Equal(spaceGUID))
				Expect(actualQuery).To(BeEmpty())
			})

			When("not found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
						resources.ServiceInstance{},
						ccv3.IncludedResources{},
						ccv3.Warnings{"get instance warning"},
						ccerror.ServiceInstanceNotFoundError{Name: serviceInstanceName},
					)
				})

				It("returns the error and warning", func() {
					Expect(warnings).To(ContainElement("get instance warning"))
					Expect(executionError).To(MatchError(actionerror.ServiceInstanceNotFoundError{Name: serviceInstanceName}))
				})
			})

			When("fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
						resources.ServiceInstance{},
						ccv3.IncludedResources{},
						ccv3.Warnings{"get instance warning"},
						errors.New("boof"),
					)
				})

				It("returns the error and warning", func() {
					Expect(warnings).To(ContainElement("get instance warning"))
					Expect(executionError).To(MatchError("boof"))
				})
			})
		})

		Describe("app lookup", func() {
			It("makes the correct call", func() {
				Expect(fakeCloudControllerClient.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
				actualAppName, actualSpaceGUID := fakeCloudControllerClient.GetApplicationByNameAndSpaceArgsForCall(0)
				Expect(actualAppName).To(Equal(appName))
				Expect(actualSpaceGUID).To(Equal(spaceGUID))
			})

			When("not found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationByNameAndSpaceReturns(
						resources.Application{},
						ccv3.Warnings{"get app warning"},
						ccerror.ApplicationNotFoundError{Name: appName},
					)
				})

				It("returns the error and warning", func() {
					Expect(warnings).To(ContainElement("get app warning"))
					Expect(executionError).To(MatchError(actionerror.ApplicationNotFoundError{Name: appName}))
				})
			})

			When("fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationByNameAndSpaceReturns(
						resources.Application{},
						ccv3.Warnings{"get app warning"},
						errors.New("boom"),
					)
				})

				It("returns the error and warning", func() {
					Expect(warnings).To(ContainElement("get app warning"))
					Expect(executionError).To(MatchError("boom"))
				})
			})
		})

		Describe("initiating the create", func() {
			It("makes the correct call", func() {
				Expect(fakeCloudControllerClient.CreateServiceCredentialBindingCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.CreateServiceCredentialBindingArgsForCall(0)).To(Equal(resources.ServiceCredentialBinding{
					Type:                resources.AppBinding,
					Name:                bindingName,
					ServiceInstanceGUID: serviceInstanceGUID,
					AppGUID:             appGUID,
					Parameters: types.NewOptionalObject(map[string]interface{}{
						"foo": "bar",
					}),
				}))
			})

			When("binding already exists", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.CreateServiceCredentialBindingReturns(
						"",
						ccv3.Warnings{"create binding warning"},
						ccerror.ResourceAlreadyExistsError{
							Message: "The app is already bound to the service instance",
						},
					)
				})

				It("returns an actionerror and warnings", func() {
					Expect(warnings).To(ContainElement("create binding warning"))
					Expect(executionError).To(MatchError(actionerror.ResourceAlreadyExistsError{
						Message: "The app is already bound to the service instance",
					}))
				})
			})

			When("fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.CreateServiceCredentialBindingReturns(
						"",
						ccv3.Warnings{"create binding warning"},
						errors.New("boop"),
					)
				})

				It("returns the error and warnings", func() {
					Expect(warnings).To(ContainElement("create binding warning"))
					Expect(executionError).To(MatchError("boop"))
				})
			})
		})

		Describe("polling the job", func() {
			It("polls the job", func() {
				Expect(fakeCloudControllerClient.PollJobToEventStreamCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.PollJobToEventStreamArgsForCall(0)).To(Equal(fakeJobURL))
			})
		})
	})

	Describe("DeleteServiceAppBinding", func() {
		const (
			serviceInstanceName = "fake-service-instance-name"
			serviceInstanceGUID = "fake-service-instance-guid"
			appName             = "fake-app-name"
			appGUID             = "fake-app-guid"
			spaceGUID           = "fake-space-guid"
			bindingGUID         = "fake-binding-guid"
			fakeJobURL          = ccv3.JobURL("fake-job-url")
		)

		var (
			params         DeleteServiceAppBindingParams
			warnings       Warnings
			executionError error
			stream         chan PollJobEvent
		)

		BeforeEach(func() {
			fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
				resources.ServiceInstance{
					Name: serviceInstanceName,
					GUID: serviceInstanceGUID,
				},
				ccv3.IncludedResources{},
				ccv3.Warnings{"get instance warning"},
				nil,
			)

			fakeCloudControllerClient.GetApplicationByNameAndSpaceReturns(
				resources.Application{
					GUID: appGUID,
					Name: appName,
				},
				ccv3.Warnings{"get app warning"},
				nil,
			)

			fakeCloudControllerClient.GetServiceCredentialBindingsReturns(
				[]resources.ServiceCredentialBinding{
					{GUID: bindingGUID},
				},
				ccv3.Warnings{"get bindings warning"},
				nil,
			)

			fakeCloudControllerClient.DeleteServiceCredentialBindingReturns(
				fakeJobURL,
				ccv3.Warnings{"delete binding warning"},
				nil,
			)

			fakeStream := make(chan ccv3.PollJobEvent)
			fakeCloudControllerClient.PollJobToEventStreamReturns(fakeStream)
			go func() {
				fakeStream <- ccv3.PollJobEvent{
					State:    constant.JobPolling,
					Warnings: ccv3.Warnings{"poll warning"},
				}
			}()

			params = DeleteServiceAppBindingParams{
				SpaceGUID:           spaceGUID,
				ServiceInstanceName: serviceInstanceName,
				AppName:             appName,
			}
		})

		JustBeforeEach(func() {
			stream, warnings, executionError = actor.DeleteServiceAppBinding(params)
		})

		It("returns an event stream, warnings, and no errors", func() {
			Expect(executionError).NotTo(HaveOccurred())

			Expect(warnings).To(ConsistOf(Warnings{
				"get instance warning",
				"get app warning",
				"get bindings warning",
				"delete binding warning",
			}))

			Eventually(stream).Should(Receive(Equal(PollJobEvent{
				State:    JobPolling,
				Warnings: Warnings{"poll warning"},
				Err:      nil,
			})))
		})

		Describe("service instance lookup", func() {
			It("makes the correct call", func() {
				Expect(fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
				actualServiceInstanceName, actualSpaceGUID, actualQuery := fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceArgsForCall(0)
				Expect(actualServiceInstanceName).To(Equal(serviceInstanceName))
				Expect(actualSpaceGUID).To(Equal(spaceGUID))
				Expect(actualQuery).To(BeEmpty())
			})

			When("not found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
						resources.ServiceInstance{},
						ccv3.IncludedResources{},
						ccv3.Warnings{"get instance warning"},
						ccerror.ServiceInstanceNotFoundError{Name: serviceInstanceName},
					)
				})

				It("returns the error and warning", func() {
					Expect(warnings).To(ContainElement("get instance warning"))
					Expect(executionError).To(MatchError(actionerror.ServiceInstanceNotFoundError{Name: serviceInstanceName}))
				})
			})

			When("fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
						resources.ServiceInstance{},
						ccv3.IncludedResources{},
						ccv3.Warnings{"get instance warning"},
						errors.New("boof"),
					)
				})

				It("returns the error and warning", func() {
					Expect(warnings).To(ContainElement("get instance warning"))
					Expect(executionError).To(MatchError("boof"))
				})
			})
		})

		Describe("app lookup", func() {
			It("makes the correct call", func() {
				Expect(fakeCloudControllerClient.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
				actualAppName, actualSpaceGUID := fakeCloudControllerClient.GetApplicationByNameAndSpaceArgsForCall(0)
				Expect(actualAppName).To(Equal(appName))
				Expect(actualSpaceGUID).To(Equal(spaceGUID))
			})

			When("not found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationByNameAndSpaceReturns(
						resources.Application{},
						ccv3.Warnings{"get app warning"},
						ccerror.ApplicationNotFoundError{Name: appName},
					)
				})

				It("returns the error and warning", func() {
					Expect(warnings).To(ContainElement("get app warning"))
					Expect(executionError).To(MatchError(actionerror.ApplicationNotFoundError{Name: appName}))
				})
			})

			When("fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationByNameAndSpaceReturns(
						resources.Application{},
						ccv3.Warnings{"get app warning"},
						errors.New("boom"),
					)
				})

				It("returns the error and warning", func() {
					Expect(warnings).To(ContainElement("get app warning"))
					Expect(executionError).To(MatchError("boom"))
				})
			})
		})

		Describe("binding lookup", func() {
			It("makes the correct call", func() {
				Expect(fakeCloudControllerClient.GetServiceCredentialBindingsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServiceCredentialBindingsArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.TypeFilter, Values: []string{"app"}},
					ccv3.Query{Key: ccv3.AppGUIDFilter, Values: []string{appGUID}},
					ccv3.Query{Key: ccv3.ServiceInstanceGUIDFilter, Values: []string{serviceInstanceGUID}},
				))
			})

			When("not found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceCredentialBindingsReturns(
						[]resources.ServiceCredentialBinding{},
						ccv3.Warnings{"get bindings warning"},
						nil,
					)
				})

				It("returns the error and warning", func() {
					Expect(warnings).To(ContainElement("get bindings warning"))
					Expect(executionError).To(MatchError(actionerror.ServiceBindingNotFoundError{
						AppGUID:             appGUID,
						ServiceInstanceGUID: serviceInstanceGUID,
					}))
				})
			})

			When("fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceCredentialBindingsReturns(
						[]resources.ServiceCredentialBinding{},
						ccv3.Warnings{"get binding warning"},
						errors.New("boom"),
					)
				})

				It("returns the error and warning", func() {
					Expect(warnings).To(ContainElement("get binding warning"))
					Expect(executionError).To(MatchError("boom"))
				})
			})
		})

		Describe("initiating the delete", func() {
			It("makes the correct call", func() {
				Expect(fakeCloudControllerClient.DeleteServiceCredentialBindingCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.DeleteServiceCredentialBindingArgsForCall(0)).To(Equal(bindingGUID))
			})

			When("fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.DeleteServiceCredentialBindingReturns(
						"",
						ccv3.Warnings{"delete binding warning"},
						errors.New("boop"),
					)
				})

				It("returns the error and warnings", func() {
					Expect(warnings).To(ContainElement("delete binding warning"))
					Expect(executionError).To(MatchError("boop"))
				})
			})
		})

		Describe("polling the job", func() {
			It("polls the job", func() {
				Expect(fakeCloudControllerClient.PollJobToEventStreamCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.PollJobToEventStreamArgsForCall(0)).To(Equal(fakeJobURL))
			})
		})
	})
})

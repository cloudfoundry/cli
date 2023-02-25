package v7action_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Route Binding Action", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil, nil)
	})

	Describe("CreateRouteBinding", func() {
		const (
			serviceInstanceName = "fake-service-instance-name"
			serviceInstanceGUID = "fake-service-instance-guid"
			domainName          = "fake-domain-name"
			domainGUID          = "fake-domain-guid"
			spaceGUID           = "fake-space-guid"
			hostname            = "fake-hostname"
			path                = "fake-path"
			routeGUID           = "fake-route-guid"
			fakeJobURL          = ccv3.JobURL("fake-job-url")
		)

		var (
			params         CreateRouteBindingParams
			warnings       Warnings
			executionError error
			stream         chan PollJobEvent
		)

		BeforeEach(func() {
			fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
				resources.ServiceInstance{
					Name: serviceInstanceName,
					GUID: serviceInstanceGUID,
					Type: resources.ManagedServiceInstance,
				},
				ccv3.IncludedResources{},
				ccv3.Warnings{"get instance warning"},
				nil,
			)

			fakeCloudControllerClient.GetDomainsReturns(
				[]resources.Domain{{
					Name: domainName,
					GUID: domainGUID,
				}},
				ccv3.Warnings{"get domain warning"},
				nil,
			)

			fakeCloudControllerClient.GetRoutesReturns(
				[]resources.Route{{
					GUID: routeGUID,
				}},
				ccv3.Warnings{"get route warning"},
				nil,
			)

			fakeCloudControllerClient.CreateRouteBindingReturns(
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

			params = CreateRouteBindingParams{
				SpaceGUID:           spaceGUID,
				ServiceInstanceName: serviceInstanceName,
				DomainName:          domainName,
				Hostname:            hostname,
				Path:                path,
				Parameters: types.NewOptionalObject(map[string]interface{}{
					"foo": "bar",
				}),
			}
		})

		JustBeforeEach(func() {
			stream, warnings, executionError = actor.CreateRouteBinding(params)
		})

		It("returns an event stream, warnings, and no errors", func() {
			Expect(executionError).NotTo(HaveOccurred())

			Expect(warnings).To(ConsistOf(Warnings{
				"get instance warning",
				"get domain warning",
				"get route warning",
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

		Describe("domain lookup", func() {
			It("makes the correct call", func() {
				Expect(fakeCloudControllerClient.GetDomainsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetDomainsArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.NameFilter, Values: []string{domainName}},
				))
			})

			When("not found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetDomainsReturns(
						[]resources.Domain{},
						ccv3.Warnings{"get domain warning"},
						nil,
					)
				})

				It("returns an error and warnings", func() {
					Expect(warnings).To(ContainElement("get domain warning"))
					Expect(executionError).To(MatchError(actionerror.DomainNotFoundError{Name: domainName}))
				})
			})

			When("fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetDomainsReturns(
						[]resources.Domain{},
						ccv3.Warnings{"get domain warning"},
						errors.New("boom"),
					)
				})

				It("returns the error and warnings", func() {
					Expect(warnings).To(ContainElement("get domain warning"))
					Expect(executionError).To(MatchError("boom"))
				})
			})
		})

		Describe("route lookup", func() {
			It("makes the correct call", func() {
				Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetRoutesArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.DomainGUIDFilter, Values: []string{domainGUID}},
					ccv3.Query{Key: ccv3.HostsFilter, Values: []string{hostname}},
					ccv3.Query{Key: ccv3.PathsFilter, Values: []string{path}},
				))
			})

			When("not found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetRoutesReturns(
						[]resources.Route{},
						ccv3.Warnings{"get route warning"},
						nil,
					)

					params.Hostname = hostname
					params.Path = path
				})

				It("returns an error and warnings", func() {
					Expect(warnings).To(ContainElement("get domain warning"))
					Expect(executionError).To(MatchError(actionerror.RouteNotFoundError{
						Host:       hostname,
						DomainName: domainName,
						Path:       path,
					}))
				})
			})

			When("fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetRoutesReturns(
						[]resources.Route{},
						ccv3.Warnings{"get route warning"},
						errors.New("pow"),
					)
				})

				It("returns the error and warnings", func() {
					Expect(warnings).To(ContainElement("get domain warning"))
					Expect(executionError).To(MatchError("pow"))
				})
			})
		})

		Describe("initiating the create", func() {
			It("makes the correct call", func() {
				Expect(fakeCloudControllerClient.CreateRouteBindingCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.CreateRouteBindingArgsForCall(0)).To(Equal(resources.RouteBinding{
					ServiceInstanceGUID: serviceInstanceGUID,
					RouteGUID:           routeGUID,
					Parameters: types.NewOptionalObject(map[string]interface{}{
						"foo": "bar",
					}),
				}))
			})

			When("binding already exists", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.CreateRouteBindingReturns(
						"",
						ccv3.Warnings{"create binding warning"},
						ccerror.ResourceAlreadyExistsError{
							Message: "The route and service instance are already bound.",
						},
					)
				})

				It("returns an actionerror and warnings", func() {
					Expect(warnings).To(ContainElement("create binding warning"))
					Expect(executionError).To(MatchError(actionerror.ResourceAlreadyExistsError{
						Message: "The route and service instance are already bound.",
					}))
				})
			})

			When("fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.CreateRouteBindingReturns(
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

	Describe("DeleteRouteBinding", func() {
		const (
			serviceInstanceName = "fake-service-instance-name"
			serviceInstanceGUID = "fake-service-instance-guid"
			domainName          = "fake-domain-name"
			domainGUID          = "fake-domain-guid"
			spaceGUID           = "fake-space-guid"
			hostname            = "fake-hostname"
			path                = "fake-path"
			routeGUID           = "fake-route-guid"
			routeBindingGUID    = "fake-route-binding-guid"
			fakeJobURL          = ccv3.JobURL("fake-job-url")
		)

		var (
			params         DeleteRouteBindingParams
			warnings       Warnings
			executionError error
			stream         chan PollJobEvent
		)

		BeforeEach(func() {
			fakeCloudControllerClient.GetServiceInstanceByNameAndSpaceReturns(
				resources.ServiceInstance{
					Name: serviceInstanceName,
					GUID: serviceInstanceGUID,
					Type: resources.ManagedServiceInstance,
				},
				ccv3.IncludedResources{},
				ccv3.Warnings{"get instance warning"},
				nil,
			)

			fakeCloudControllerClient.GetDomainsReturns(
				[]resources.Domain{{
					Name: domainName,
					GUID: domainGUID,
				}},
				ccv3.Warnings{"get domain warning"},
				nil,
			)

			fakeCloudControllerClient.GetRoutesReturns(
				[]resources.Route{{
					GUID: routeGUID,
				}},
				ccv3.Warnings{"get route warning"},
				nil,
			)

			fakeCloudControllerClient.GetRouteBindingsReturns(
				[]resources.RouteBinding{{
					GUID: routeBindingGUID,
				}},
				ccv3.IncludedResources{},
				ccv3.Warnings{"get binding warning"},
				nil,
			)

			fakeCloudControllerClient.DeleteRouteBindingReturns(
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

			params = DeleteRouteBindingParams{
				SpaceGUID:           spaceGUID,
				ServiceInstanceName: serviceInstanceName,
				DomainName:          domainName,
				Hostname:            hostname,
				Path:                path,
			}
		})

		JustBeforeEach(func() {
			stream, warnings, executionError = actor.DeleteRouteBinding(params)
		})

		It("returns an event stream, warnings, and no errors", func() {
			Expect(executionError).NotTo(HaveOccurred())

			Expect(warnings).To(ConsistOf(Warnings{
				"get instance warning",
				"get domain warning",
				"get route warning",
				"get binding warning",
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

		Describe("domain lookup", func() {
			It("makes the correct call", func() {
				Expect(fakeCloudControllerClient.GetDomainsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetDomainsArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.NameFilter, Values: []string{domainName}},
				))
			})

			When("not found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetDomainsReturns(
						[]resources.Domain{},
						ccv3.Warnings{"get domain warning"},
						nil,
					)
				})

				It("returns an error and warnings", func() {
					Expect(warnings).To(ContainElement("get domain warning"))
					Expect(executionError).To(MatchError(actionerror.DomainNotFoundError{Name: domainName}))
				})
			})

			When("fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetDomainsReturns(
						[]resources.Domain{},
						ccv3.Warnings{"get domain warning"},
						errors.New("boom"),
					)
				})

				It("returns the error and warnings", func() {
					Expect(warnings).To(ContainElement("get domain warning"))
					Expect(executionError).To(MatchError("boom"))
				})
			})
		})

		Describe("route lookup", func() {
			It("makes the correct call", func() {
				Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetRoutesArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.DomainGUIDFilter, Values: []string{domainGUID}},
					ccv3.Query{Key: ccv3.HostsFilter, Values: []string{hostname}},
					ccv3.Query{Key: ccv3.PathsFilter, Values: []string{path}},
				))
			})

			When("not found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetRoutesReturns(
						[]resources.Route{},
						ccv3.Warnings{"get route warning"},
						nil,
					)

					params.Hostname = hostname
					params.Path = path
				})

				It("returns an error and warnings", func() {
					Expect(warnings).To(ContainElement("get domain warning"))
					Expect(executionError).To(MatchError(actionerror.RouteNotFoundError{
						Host:       hostname,
						DomainName: domainName,
						Path:       path,
					}))
				})
			})

			When("fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetRoutesReturns(
						[]resources.Route{},
						ccv3.Warnings{"get route warning"},
						errors.New("pow"),
					)
				})

				It("returns the error and warnings", func() {
					Expect(warnings).To(ContainElement("get domain warning"))
					Expect(executionError).To(MatchError("pow"))
				})
			})
		})

		Describe("route binding lookup", func() {
			It("makes the correct call", func() {
				Expect(fakeCloudControllerClient.GetRouteBindingsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetRouteBindingsArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.RouteGUIDFilter, Values: []string{routeGUID}},
					ccv3.Query{Key: ccv3.ServiceInstanceGUIDFilter, Values: []string{serviceInstanceGUID}},
				))
			})

			When("not found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetRouteBindingsReturns(
						[]resources.RouteBinding{},
						ccv3.IncludedResources{},
						ccv3.Warnings{"get bindings warning"},
						nil,
					)
				})

				It("returns an error and warnings", func() {
					Expect(warnings).To(ContainElement("get bindings warning"))
					Expect(executionError).To(MatchError(actionerror.RouteBindingNotFoundError{}))
				})
			})

			When("fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetRouteBindingsReturns(
						[]resources.RouteBinding{},
						ccv3.IncludedResources{},
						ccv3.Warnings{"get bindings warning"},
						errors.New("boom"),
					)
				})

				It("returns an error and warnings", func() {
					Expect(warnings).To(ContainElement("get bindings warning"))
					Expect(executionError).To(MatchError("boom"))
				})
			})
		})

		Describe("initiating the delete", func() {
			It("makes the correct call", func() {
				Expect(fakeCloudControllerClient.DeleteRouteBindingCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.DeleteRouteBindingArgsForCall(0)).To(Equal(routeBindingGUID))
			})

			When("fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.DeleteRouteBindingReturns(
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

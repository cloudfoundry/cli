package ccv3_test

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/ccv3fakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"errors"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Service Offering", func() {
	Describe("GetServiceOfferings", func() {
		var (
			client *Client
			query  []Query

			offerings  []ServiceOffering
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			client, _ = NewTestClient()
		})

		JustBeforeEach(func() {
			offerings, warnings, executeErr = client.GetServiceOfferings(query...)
		})

		When("service offerings exist", func() {
			BeforeEach(func() {
				response1 := fmt.Sprintf(`
					{
						"pagination": {
							"next": {
								"href": "%s/v3/service_offerings?names=myServiceOffering&service_broker_names=myServiceBroker&page=2"
							}
						},
						"resources": [
							{
								"guid": "service-offering-1-guid",
								"name": "service-offering-1-name",
								"relationships": {
									"service_broker": {
										"data": {
											"name": "overview-broker"
										}
									}
								}
							},
							{
								"guid": "service-offering-2-guid",
								"name": "service-offering-2-name",
								"relationships": {
									"service_broker": {
										"data": {
											"name": "overview-broker"
										}
									}
								}
							}
						]
					}`,
					server.URL())

				response2 := `
					{
						"pagination": {
							"next": {
								"href": null
							}
						},
						"resources": [
							{
								"guid": "service-offering-3-guid",
								"name": "service-offering-3-name",
								"relationships": {
									"service_broker": {
										"data": {
											"name": "other-broker"
										}
									}
								}
							}
						]
					}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/service_offerings", "names=myServiceOffering&service_broker_names=myServiceBroker"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/service_offerings", "names=myServiceOffering&service_broker_names=myServiceBroker&page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"warning-2"}}),
					),
				)

				query = []Query{
					{
						Key:    NameFilter,
						Values: []string{"myServiceOffering"},
					},
					{
						Key:    ServiceBrokerNamesFilter,
						Values: []string{"myServiceBroker"},
					},
				}
			})

			It("returns a list of service offerings with their associated warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(offerings).To(ConsistOf(
					ServiceOffering{
						GUID:              "service-offering-1-guid",
						Name:              "service-offering-1-name",
						ServiceBrokerName: "overview-broker",
					},
					ServiceOffering{
						GUID:              "service-offering-2-guid",
						Name:              "service-offering-2-name",
						ServiceBrokerName: "overview-broker",
					},
					ServiceOffering{
						GUID:              "service-offering-3-guid",
						Name:              "service-offering-3-name",
						ServiceBrokerName: "other-broker",
					},
				))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})

		When("the cloud controller returns errors and warnings", func() {
			BeforeEach(func() {
				response := `{
					"errors": [
						{
							"code": 42424,
							"detail": "Some detailed error message",
							"title": "CF-SomeErrorTitle"
						},
						{
							"code": 11111,
							"detail": "Some other detailed error message",
							"title": "CF-SomeOtherErrorTitle"
						}
					]
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/service_offerings"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.MultiError{
					ResponseCode: http.StatusTeapot,
					Errors: []ccerror.V3Error{
						{
							Code:   42424,
							Detail: "Some detailed error message",
							Title:  "CF-SomeErrorTitle",
						},
						{
							Code:   11111,
							Detail: "Some other detailed error message",
							Title:  "CF-SomeOtherErrorTitle",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("GetServiceOfferingByNameAndBroker", func() {
		const (
			serviceOfferingName = "myServiceOffering"
		)

		var (
			client            *Client
			requester         *ccv3fakes.FakeRequester
			serviceBrokerName string
			offering          ServiceOffering
			warnings          Warnings
			executeErr        error
		)

		BeforeEach(func() {
			requester = new(ccv3fakes.FakeRequester)
			client, _ = NewFakeRequesterTestClient(requester)

			serviceBrokerName = ""
		})

		JustBeforeEach(func() {
			offering, warnings, executeErr = client.GetServiceOfferingByNameAndBroker(serviceOfferingName, serviceBrokerName)
		})

		When("there is a single match", func() {
			BeforeEach(func() {
				requester.MakeListRequestCalls(func(requestParams RequestParams) (IncludedResources, Warnings, error) {
					err := requestParams.AppendToList(ServiceOffering{GUID: "service-offering-guid-1"})
					Expect(err).NotTo(HaveOccurred())
					return IncludedResources{}, Warnings{"this is a warning"}, nil
				})
			})

			It("makes the correct request", func() {
				Expect(requester.MakeListRequestCallCount()).To(Equal(1))
				Expect(requester.MakeListRequestArgsForCall(0)).To(MatchFields(IgnoreExtras, Fields{
					"RequestName":  Equal(internal.GetServiceOfferingsRequest),
					"Query":        Equal([]Query{{Key: NameFilter, Values: []string{serviceOfferingName}}}),
					"ResponseBody": Equal(ServiceOffering{}),
				}))
			})

			It("returns the service offering and warnings", func() {
				Expect(offering).To(Equal(ServiceOffering{GUID: "service-offering-guid-1"}))
				Expect(warnings).To(ConsistOf("this is a warning"))
				Expect(executeErr).NotTo(HaveOccurred())
			})
		})

		When("there are no matches", func() {
			BeforeEach(func() {
				requester.MakeListRequestCalls(func(requestParams RequestParams) (IncludedResources, Warnings, error) {
					return IncludedResources{}, Warnings{"this is a warning"}, nil
				})

				serviceBrokerName = "myServiceBroker"
			})

			It("returns an error and warnings", func() {
				Expect(warnings).To(ConsistOf("this is a warning"))
				Expect(executeErr).To(MatchError(ccerror.ServiceOfferingNotFoundError{
					ServiceOfferingName: serviceOfferingName,
					ServiceBrokerName:   serviceBrokerName,
				}))
			})
		})

		When("there is more than one match", func() {
			BeforeEach(func() {
				requester.MakeListRequestCalls(func(requestParams RequestParams) (IncludedResources, Warnings, error) {
					err := requestParams.AppendToList(ServiceOffering{
						GUID:              "service-offering-guid-1",
						Name:              serviceOfferingName,
						ServiceBrokerName: "broker-1",
					})
					Expect(err).NotTo(HaveOccurred())
					err = requestParams.AppendToList(ServiceOffering{
						GUID:              "service-offering-guid-2",
						Name:              serviceOfferingName,
						ServiceBrokerName: "broker-2",
					})
					Expect(err).NotTo(HaveOccurred())
					return IncludedResources{}, Warnings{"this is a warning"}, nil
				})
			})

			It("returns an error and warnings", func() {
				Expect(warnings).To(ConsistOf("this is a warning"))
				Expect(executeErr).To(MatchError(ccerror.ServiceOfferingNameAmbiguityError{
					ServiceOfferingName: serviceOfferingName,
					ServiceBrokerNames:  []string{"broker-1", "broker-2"},
				}))
			})
		})

		When("the broker name is specified", func() {
			BeforeEach(func() {
				serviceBrokerName = "myServiceBroker"
			})

			It("makes the correct request", func() {
				Expect(requester.MakeListRequestCallCount()).To(Equal(1))
				Expect(requester.MakeListRequestArgsForCall(0).Query).To(ConsistOf(
					Query{Key: NameFilter, Values: []string{serviceOfferingName}},
					Query{Key: ServiceBrokerNamesFilter, Values: []string{"myServiceBroker"}},
				))
			})
		})

		When("the requester returns an error", func() {
			BeforeEach(func() {
				requester.MakeListRequestReturns(IncludedResources{}, Warnings{"this is a warning"}, errors.New("bang"))
			})

			It("returns an error and warnings", func() {
				Expect(warnings).To(ConsistOf("this is a warning"))
				Expect(executeErr).To(MatchError("bang"))
			})
		})
	})
})

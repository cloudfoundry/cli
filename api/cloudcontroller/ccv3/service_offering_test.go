package ccv3_test

import (
	"errors"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/ccv3fakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Service Offering", func() {
	Describe("GetServiceOfferings", func() {
		var (
			client *Client
			query  []Query

			offerings  []resources.ServiceOffering
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
								"href": "%s/v3/service_offerings?names=myServiceOffering&service_broker_names=myServiceBroker&fields[service_broker]=name,guid&page=2"
							}
						},
						"resources": [
							{
								"guid": "service-offering-1-guid",
								"name": "service-offering-1-name",
								"relationships": {
									"service_broker": {
										"data": {
											"guid": "overview-broker-guid"
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
											"guid": "overview-broker-guid"
										}
									}
								}
							}
						],
						"included": {
							"service_brokers": [
								{
									"guid": "overview-broker-guid",
									"name": "overview-broker"
								}
							]	
						}
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
											"guid": "other-broker-guid"
										}
									}
								}
							}
						],
						"included": {
							"service_brokers": [
								{
									"guid": "other-broker-guid",
									"name": "other-broker"
								}
							]	
						}
					}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/service_offerings", "names=myServiceOffering&service_broker_names=myServiceBroker&fields[service_broker]=name,guid"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/service_offerings", "names=myServiceOffering&service_broker_names=myServiceBroker&fields[service_broker]=name,guid&page=2"),
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
					resources.ServiceOffering{
						GUID:              "service-offering-1-guid",
						Name:              "service-offering-1-name",
						ServiceBrokerName: "overview-broker",
						ServiceBrokerGUID: "overview-broker-guid",
					},
					resources.ServiceOffering{
						GUID:              "service-offering-2-guid",
						Name:              "service-offering-2-name",
						ServiceBrokerName: "overview-broker",
						ServiceBrokerGUID: "overview-broker-guid",
					},
					resources.ServiceOffering{
						GUID:              "service-offering-3-guid",
						Name:              "service-offering-3-name",
						ServiceBrokerName: "other-broker",
						ServiceBrokerGUID: "other-broker-guid",
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

	Describe("GetServiceOfferingByGUID", func() {
		const guid = "fake-guid"

		var (
			requester *ccv3fakes.FakeRequester
			client    *Client
		)

		BeforeEach(func() {
			requester = new(ccv3fakes.FakeRequester)
			client, _ = NewFakeRequesterTestClient(requester)
		})

		When("service offering exists", func() {
			BeforeEach(func() {
				requester.MakeRequestCalls(func(params RequestParams) (JobURL, Warnings, error) {
					Expect(params.URIParams).To(BeEquivalentTo(map[string]string{"service_offering_guid": guid}))
					Expect(params.RequestName).To(Equal(internal.GetServiceOfferingRequest))
					params.ResponseBody.(*resources.ServiceOffering).GUID = guid
					return "", Warnings{"one", "two"}, nil
				})
			})

			It("returns the service offering with warnings", func() {
				offering, warnings, err := client.GetServiceOfferingByGUID(guid)
				Expect(err).ToNot(HaveOccurred())

				Expect(offering).To(Equal(resources.ServiceOffering{
					GUID: guid,
				}))
				Expect(warnings).To(ConsistOf("one", "two"))
			})
		})

		When("no guid was specified", func() {
			It("fails saying the offering was not found", func() {
				_, _, err := client.GetServiceOfferingByGUID("")
				Expect(err).To(MatchError(ccerror.ServiceOfferingNotFoundError{}))
			})
		})

		When("the cloud controller returns errors and warnings", func() {
			BeforeEach(func() {
				requester.MakeRequestReturns(
					"",
					Warnings{"one", "two"},
					ccerror.MultiError{
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
					},
				)
			})

			It("returns the error and all warnings", func() {
				_, warnings, err := client.GetServiceOfferingByGUID(guid)
				Expect(err).To(MatchError(ccerror.MultiError{
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
				Expect(warnings).To(ConsistOf("one", "two"))
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
			offering          resources.ServiceOffering
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
					err := requestParams.AppendToList(resources.ServiceOffering{GUID: "service-offering-guid-1"})
					Expect(err).NotTo(HaveOccurred())
					return IncludedResources{}, Warnings{"this is a warning"}, nil
				})
			})

			It("makes the correct request", func() {
				Expect(requester.MakeListRequestCallCount()).To(Equal(1))
				Expect(requester.MakeListRequestArgsForCall(0)).To(MatchFields(IgnoreExtras, Fields{
					"RequestName": Equal(internal.GetServiceOfferingsRequest),
					"Query": Equal([]Query{
						{Key: NameFilter, Values: []string{serviceOfferingName}},
						{Key: FieldsServiceBroker, Values: []string{"name", "guid"}},
					}),
					"ResponseBody": Equal(resources.ServiceOffering{}),
				}))
			})

			It("returns the service offering and warnings", func() {
				Expect(offering).To(Equal(resources.ServiceOffering{GUID: "service-offering-guid-1"}))
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
					err := requestParams.AppendToList(resources.ServiceOffering{
						GUID:              "service-offering-guid-1",
						Name:              serviceOfferingName,
						ServiceBrokerGUID: "broker-1-guid",
					})
					Expect(err).NotTo(HaveOccurred())
					err = requestParams.AppendToList(resources.ServiceOffering{
						GUID:              "service-offering-guid-2",
						Name:              serviceOfferingName,
						ServiceBrokerGUID: "broker-2-guid",
					})
					Expect(err).NotTo(HaveOccurred())
					return IncludedResources{
							ServiceBrokers: []resources.ServiceBroker{
								{GUID: "broker-1-guid", Name: "broker-1"},
								{GUID: "broker-2-guid", Name: "broker-2"},
							}},
						Warnings{"this is a warning"},
						nil
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
					Query{Key: FieldsServiceBroker, Values: []string{"name", "guid"}},
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

	Describe("PurgeServiceOffering", func() {
		const serviceOfferingGUID = "fake-service-offering-guid"

		var (
			client     *Client
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			client, _ = NewTestClient()
		})

		JustBeforeEach(func() {
			warnings, executeErr = client.PurgeServiceOffering(serviceOfferingGUID)
		})

		When("the Cloud Controller successfully purges the service offering", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v3/service_offerings/fake-service-offering-guid", "purge=true"),
						RespondWith(http.StatusNoContent, nil, http.Header{
							"X-Cf-Warnings": {"this is a warning"},
						}),
					),
				)
			})

			It("succeeds and returns warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		When("the Cloud Controller fails to purge the service offering", func() {
			BeforeEach(func() {
				response := `{
                               "errors": [
                                 {
                                   "code": 10008,
								   "detail": "The request is semantically invalid: command presence",
								   "title": "CF-UnprocessableEntity"
								 },
								 {
								   "code": 10010,
								   "detail": "Service offering not found",
								   "title": "CF-ResourceNotFound"
								 }
							   ]
							 }`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v3/service_offerings/fake-service-offering-guid", "purge=true"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns parsed errors and warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.MultiError{
					ResponseCode: http.StatusTeapot,
					Errors: []ccerror.V3Error{
						{
							Code:   10008,
							Detail: "The request is semantically invalid: command presence",
							Title:  "CF-UnprocessableEntity",
						},
						{
							Code:   10010,
							Detail: "Service offering not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})
})

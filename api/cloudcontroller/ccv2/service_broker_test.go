package ccv2_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Service Broker", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetServiceBrokers", func() {
		var (
			serviceBrokers []ServiceBroker
			warnings       Warnings
			executeErr     error
		)

		JustBeforeEach(func() {
			serviceBrokers, warnings, executeErr = client.GetServiceBrokers(Filter{
				Type:     constant.NameFilter,
				Operator: constant.EqualOperator,
				Values:   []string{"some-name"},
			})
		})

		Context("when the cc returns back service brokers", func() {
			BeforeEach(func() {
				response1 := `{
				"next_url": "/v2/service_brokers?q=name:some-name&page=2",
				"resources": [
					{
						"metadata": {
							"guid": "service-broker-guid-1"
						},
						"entity": {
							"name":"some-broker-name"
						}
					},
					{
						"metadata": {
							"guid": "service-broker-guid-2"
						},
						"entity": {
							"name":"other-broker-name"
						}
					}
				]
			}`
				response2 := `{
				"next_url": null,
				"resources": [
					{
						"metadata": {
							"guid": "service-broker-guid-3"
						},
						"entity": {
							"name":"some-broker-name"
						}
					},
					{
						"metadata": {
							"guid": "service-broker-guid-4"
						},
						"entity": {
							"name":"other-broker-name"
						}
					}
				]
			}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/service_brokers", "q=name:some-name"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/service_brokers", "q=name:some-name&page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)
			})

			It("returns all the queried service brokers", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(serviceBrokers).To(ConsistOf([]ServiceBroker{
					{GUID: "service-broker-guid-1", Name: "some-broker-name"},
					{GUID: "service-broker-guid-2", Name: "other-broker-name"},
					{GUID: "service-broker-guid-3", Name: "some-broker-name"},
					{GUID: "service-broker-guid-4", Name: "other-broker-name"},
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning", "this is another warning"}))
			})
		})

		Context("when the cc returns an error", func() {
			BeforeEach(func() {
				response := `{
					"description": "The broker is broken.",
					"error_code": "CF-BrokenBroker",
					"code": 90003
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/service_brokers"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns an error and warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.V2UnexpectedResponseError{
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        90003,
						Description: "The broker is broken.",
						ErrorCode:   "CF-BrokenBroker",
					},
					ResponseCode: http.StatusTeapot,
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})
})

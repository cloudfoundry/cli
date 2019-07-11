package ccv2_test

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("User-Provided Service Instance", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetUserProvidedServiceInstances", func() {
		var (
			serviceInstances []ServiceInstance
			warnings         Warnings
			executeErr       error
		)

		JustBeforeEach(func() {
			serviceInstances, warnings, executeErr = client.GetUserProvidedServiceInstances(Filter{
				Type:     constant.SpaceGUIDFilter,
				Operator: constant.EqualOperator,
				Values:   []string{"some-space-guid"},
			})
		})

		When("getting user provided service instances errors", func() {
			BeforeEach(func() {
				response := `{
					"code": 1,
					"description": "some error description",
					"error_code": "CF-SomeError"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/user_provided_service_instances", "q=space_guid:some-space-guid"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.V2UnexpectedResponseError{
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        1,
						Description: "some error description",
						ErrorCode:   "CF-SomeError",
					},
					ResponseCode: http.StatusTeapot,
				}))

				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})

		When("getting user provided service instances succeeds", func() {
			BeforeEach(func() {
				response1 := `{
				"next_url": "/v2/user_provided_service_instances?q=space_guid:some-space-guid&page=2",
				"resources": [
					{
						"metadata": {
							"guid": "some-service-guid-1"
						},
						"entity": {
							"name": "some-service-name-1",
							"route_service_url": "some-route-service-url",
							"space_guid": "some-space-guid",
							"type": "user_provided_service_instance"
						}
					},
					{
						"metadata": {
							"guid": "some-service-guid-2"
						},
						"entity": {
							"name": "some-service-name-2",
							"route_service_url": "some-route-service-url",
							"space_guid": "some-space-guid",
							"type": "user_provided_service_instance"
						}
					}
				]
			}`

				response2 := `{
				"next_url": null,
				"resources": [
					{
						"metadata": {
							"guid": "some-service-guid-3"
						},
						"entity": {
							"name": "some-service-name-3",
							"route_service_url": "some-route-service-url",
							"space_guid": "some-space-guid",
							"type": "user_provided_service_instance"
						}
					},
					{
						"metadata": {
							"guid": "some-service-guid-4"
						},
						"entity": {
							"name": "some-service-name-4",
							"route_service_url": "some-route-service-url",
							"space_guid": "some-space-guid",
							"type": "user_provided_service_instance"
						}
					}
				]
			}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/user_provided_service_instances", "q=space_guid:some-space-guid"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/user_provided_service_instances", "q=space_guid:some-space-guid&page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)
			})

			It("returns all the queried service instances", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(serviceInstances).To(ConsistOf([]ServiceInstance{
					{
						Name:            "some-service-name-1",
						GUID:            "some-service-guid-1",
						SpaceGUID:       "some-space-guid",
						RouteServiceURL: "some-route-service-url",
						Type:            constant.UserProvidedService,
					},
					{
						Name:            "some-service-name-2",
						GUID:            "some-service-guid-2",
						SpaceGUID:       "some-space-guid",
						RouteServiceURL: "some-route-service-url",
						Type:            constant.UserProvidedService,
					},
					{
						Name:            "some-service-name-3",
						GUID:            "some-service-guid-3",
						SpaceGUID:       "some-space-guid",
						RouteServiceURL: "some-route-service-url",
						Type:            constant.UserProvidedService,
					},
					{
						Name:            "some-service-name-4",
						GUID:            "some-service-guid-4",
						SpaceGUID:       "some-space-guid",
						RouteServiceURL: "some-route-service-url",
						Type:            constant.UserProvidedService,
					},
				}))

				Expect(warnings).To(ConsistOf(Warnings{"this is a warning", "this is another warning"}))
			})
		})
	})

	Describe("UpdateUserProvidedServiceInstance", func() {
		const (
			spaceGUID   = "fake-space-guid"
			serviceGUID = "fake-service-instance-guid"
		)

		DescribeTable("updating properties",
			func(body string, instance UserProvidedServiceInstance) {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPut, fmt.Sprintf("/v2/user_provided_service_instances/%s", serviceGUID)),
						VerifyJSON(body),
						RespondWith(http.StatusOK, "", http.Header{"X-Cf-Warnings": {"warning-1,warning-2"}}),
					),
				)

				previousRequests := len(server.ReceivedRequests())
				warnings, executeErr := client.UpdateUserProvidedServiceInstance(serviceGUID, instance)

				Expect(server.ReceivedRequests()).To(HaveLen(previousRequests + 1))
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf(Warnings{"warning-1", "warning-2"}))
			},
			Entry(
				"setting log URL",
				`{"syslog_drain_url": "fake-syslog-drain-url"}`,
				UserProvidedServiceInstance{}.WithSyslogDrainURL("fake-syslog-drain-url"),
			),
			Entry(
				"removing log URL",
				`{"syslog_drain_url": ""}`,
				UserProvidedServiceInstance{}.WithSyslogDrainURL(""),
			),
			Entry(
				"setting routes URL",
				`{"route_service_url": "fake-route-url"}`,
				UserProvidedServiceInstance{}.WithRouteServiceURL("fake-route-url"),
			),
			Entry(
				"removing routes URL",
				`{"route_service_url": ""}`,
				UserProvidedServiceInstance{}.WithRouteServiceURL(""),
			),
			Entry(
				"setting tags",
				`{"tags": ["tag1", "tag2"]}`,
				UserProvidedServiceInstance{}.WithTags([]string{"tag1", "tag2"}),
			),
			Entry(
				"removing tags",
				`{"tags": []}`,
				UserProvidedServiceInstance{}.WithTags(nil),
			),
			Entry(
				"setting credentials",
				`{"credentials": {"username": "super-secret-password"}}`,
				UserProvidedServiceInstance{}.WithCredentials(map[string]interface{}{"username": "super-secret-password"}),
			),
			Entry(
				"removing credentials",
				`{"credentials": {}}`,
				UserProvidedServiceInstance{}.WithCredentials(nil),
			),
			Entry(
				"setting everything",
				`{
					"syslog_drain_url":  "fake-syslog-drain-url",
					"route_service_url": "fake-route-url",
					"tags":              ["tag1", "tag2"],
					"credentials":       {"username": "super-secret-password"}
				}`,
				UserProvidedServiceInstance{}.
					WithSyslogDrainURL("fake-syslog-drain-url").
					WithRouteServiceURL("fake-route-url").
					WithTags([]string{"tag1", "tag2"}).
					WithCredentials(map[string]interface{}{"username": "super-secret-password"}),
			),
		)

		When("the endpoint returns an error", func() {
			BeforeEach(func() {
				response := `{
												"code": 10003,
												"description": "You are not authorized to perform the requested action"
			  							}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPut, fmt.Sprintf("/v2/user_provided_service_instances/%s", serviceGUID)),
						RespondWith(http.StatusForbidden, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("returns all warnings and propagates the error", func() {
				priorRequests := len(server.ReceivedRequests())
				warnings, err := client.UpdateUserProvidedServiceInstance(serviceGUID, UserProvidedServiceInstance{})

				Expect(server.ReceivedRequests()).To(HaveLen(priorRequests + 1))
				Expect(err).To(MatchError(ccerror.ForbiddenError{
					Message: "You are not authorized to perform the requested action",
				}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})
})

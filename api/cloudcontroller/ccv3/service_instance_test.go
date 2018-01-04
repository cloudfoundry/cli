package ccv3_test

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Service Instance", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetServiceInstances", func() {
		Context("when service instances exist", func() {
			BeforeEach(func() {
				response1 := fmt.Sprintf(`
					{
						 "pagination": {
								"next": {
									 "href": "%s/v3/service_instances?names=some-service-instance-name&page=2"
								}
						 },
						 "resources": [
								{
									 "guid": "service-instance-1-guid",
									 "name": "service-instance-1-name"
								},
								{
									 "guid": "service-instance-2-guid",
									 "name": "service-instance-2-name"
								}
						 ]
					}`, server.URL())

				response2 := `
					{
						 "pagination": {
								"next": {
									 "href": null
								}
						 },
						 "resources": [
								{
									 "guid": "service-instance-3-guid",
									 "name": "service-instance-3-name"
								}
						 ]
					}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/service_instances", "names=some-service-instance-name"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/service_instances", "names=some-service-instance-name&page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"warning-2"}}),
					),
				)
			})

			It("returns a list of service instances with their associated warnings", func() {
				instances, warnings, err := client.GetServiceInstances(Query{
					Key:    NameFilter,
					Values: []string{"some-service-instance-name"},
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(instances).To(ConsistOf(
					ServiceInstance{
						GUID: "service-instance-1-guid",
						Name: "service-instance-1-name",
					},
					ServiceInstance{
						GUID: "service-instance-2-guid",
						Name: "service-instance-2-name",
					},
					ServiceInstance{
						GUID: "service-instance-3-guid",
						Name: "service-instance-3-name",
					},
				))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})

	Context("when the cloud controller returns errors and warnings", func() {
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
					VerifyRequest(http.MethodGet, "/v3/service_instances"),
					RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
				),
			)
		})

		It("returns the error and all warnings", func() {
			_, warnings, err := client.GetServiceInstances()
			Expect(err).To(MatchError(ccerror.V3UnexpectedResponseError{
				ResponseCode: http.StatusTeapot,
				V3ErrorResponse: ccerror.V3ErrorResponse{
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
			}))
			Expect(warnings).To(ConsistOf("this is a warning"))
		})
	})
})

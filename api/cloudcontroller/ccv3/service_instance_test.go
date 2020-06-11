package ccv3_test

import (
	"errors"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/ccv3fakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Service Instance", func() {
	Describe("GetServiceInstances", func() {
		var (
			client     *Client
			query      Query
			instances  []resources.ServiceInstance
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			client, _ = NewTestClient()
		})

		JustBeforeEach(func() {
			instances, warnings, executeErr = client.GetServiceInstances(query)
		})

		When("service instances exist", func() {
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

				query = Query{
					Key:    NameFilter,
					Values: []string{"some-service-instance-name"},
				}
			})

			It("returns a list of service instances with their associated warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(instances).To(ConsistOf(
					resources.ServiceInstance{
						GUID: "service-instance-1-guid",
						Name: "service-instance-1-name",
					},
					resources.ServiceInstance{
						GUID: "service-instance-2-guid",
						Name: "service-instance-2-name",
					},
					resources.ServiceInstance{
						GUID: "service-instance-3-guid",
						Name: "service-instance-3-name",
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
						VerifyRequest(http.MethodGet, "/v3/service_instances"),
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

	Describe("CreateServiceInstance", func() {
		var (
			requester *ccv3fakes.FakeRequester
			client    *Client
		)

		BeforeEach(func() {
			requester = new(ccv3fakes.FakeRequester)
			client, _ = NewFakeRequesterTestClient(requester)
		})

		Context("synchronous response", func() {
			When("the request succeeds", func() {
				It("returns warnings and no errors", func() {
					requester.MakeRequestReturns("", ccv3.Warnings{"fake-warning"}, nil)

					si := resources.ServiceInstance{
						Type:      resources.UserProvidedServiceInstance,
						Name:      "fake-user-provided-service-instance",
						SpaceGUID: "fake-space-guid",
					}

					jobURL, warnings, err := client.CreateServiceInstance(si)

					Expect(jobURL).To(BeEmpty())
					Expect(warnings).To(ConsistOf("fake-warning"))
					Expect(err).NotTo(HaveOccurred())

					Expect(requester.MakeRequestCallCount()).To(Equal(1))
					Expect(requester.MakeRequestArgsForCall(0)).To(Equal(RequestParams{
						RequestName: internal.PostServiceInstanceRequest,
						RequestBody: si,
					}))
				})
			})

			When("the request fails", func() {
				It("returns errors and warnings", func() {
					requester.MakeRequestReturns("", ccv3.Warnings{"fake-warning"}, errors.New("bang"))

					si := resources.ServiceInstance{
						Type:      resources.UserProvidedServiceInstance,
						Name:      "fake-user-provided-service-instance",
						SpaceGUID: "fake-space-guid",
					}

					jobURL, warnings, err := client.CreateServiceInstance(si)

					Expect(jobURL).To(BeEmpty())
					Expect(warnings).To(ConsistOf("fake-warning"))
					Expect(err).To(MatchError("bang"))
				})
			})
		})
	})
})

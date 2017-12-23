package ccv2_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Service Instance Shared From", func() {
	var (
		client *Client

		serviceInstance ServiceInstanceSharedFrom
		warnings        Warnings
		err             error
	)

	BeforeEach(func() {
		client = NewTestClient()
	})

	JustBeforeEach(func() {
		serviceInstance, warnings, err = client.GetServiceInstanceSharedFrom("some-service-instance-guid")
	})

	Describe("GetServiceInstanceSharedFrom", func() {
		Context("when the cc api returns no errors", func() {
			Context("when the response is not an http 204", func() {
				BeforeEach(func() {
					response1 := `{
			  "space_guid": "some-space-guid",
				"space_name": "some-space-name",
			  "organization_name": "some-org-name"
		 }`

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/service_instances/some-service-instance-guid/shared_from"),
							RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("returns all the shared_from resources", func() {
					Expect(err).NotTo(HaveOccurred())

					Expect(serviceInstance).To(Equal(ServiceInstanceSharedFrom{
						SpaceGUID:        "some-space-guid",
						SpaceName:        "some-space-name",
						OrganizationName: "some-org-name",
					}))
					Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				})
			})

			Context("when the response is an http 204", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/service_instances/some-service-instance-guid/shared_from"),
							RespondWith(http.StatusNoContent, "", http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("returns an empty ServiceInstanceSharedFrom and no error", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(serviceInstance).To(Equal(ServiceInstanceSharedFrom{}))
					Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				})
			})
		})

		Context("when the cc api encounters an error", func() {
			BeforeEach(func() {
				response := `{
					"code": 10001,
					"description": "Some Error",
					"error_code": "CF-SomeError"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/service_instances/some-service-instance-guid/shared_from"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("returns an error and warnings", func() {
				Expect(err).To(MatchError(ccerror.V2UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        10001,
						Description: "Some Error",
						ErrorCode:   "CF-SomeError",
					},
				}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})
})

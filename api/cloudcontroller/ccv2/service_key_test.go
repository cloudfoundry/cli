package ccv2_test

import (
	"encoding/json"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

var _ = Describe("Service Key", func() {
	var (
		client     *Client
		warnings   Warnings
		executeErr error
		serviceKey ServiceKey
		parameters map[string]interface{}
	)

	BeforeEach(func() {
		client = NewTestClient()
		parameters = map[string]interface{}{"the-service-broker": "wants this object"}
	})

	JustBeforeEach(func() {
		serviceKey, warnings, executeErr = client.CreateServiceKey("some-service-instance-guid", "some-service-key-name", parameters)
	})

	When("the create is successful", func() {
		BeforeEach(func() {
			expectedRequestBody := map[string]interface{}{
				"service_instance_guid": "some-service-instance-guid",
				"name":                  "some-service-key-name",
				"parameters": map[string]interface{}{
					"the-service-broker": "wants this object",
				},
			}
			response := `
						{
							"metadata": {
								"guid": "some-service-key-guid"
							},
							"entity": {
								"name": "some-service-key-name",
								"service_instance_guid": "some-service-instance-guid",
								"credentials": { "some-username" : "some-password", "port": 31023 }
							}
						}`
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/v2/service_keys"),
					VerifyJSONRepresenting(expectedRequestBody),
					RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"warning"}}),
				),
			)
		})

		It("returns the created object and warnings", func() {
			Expect(executeErr).NotTo(HaveOccurred())

			Expect(serviceKey).To(BeEquivalentTo(ServiceKey{
				GUID:                "some-service-key-guid",
				Name:                "some-service-key-name",
				ServiceInstanceGUID: "some-service-instance-guid",
				Credentials:         map[string]interface{}{"some-username": "some-password", "port": json.Number("31023")},
			}))
			Expect(warnings).To(ConsistOf(Warnings{"warning"}))
		})

		When("The request cannot be serialized", func() {
			BeforeEach(func() {
				parameters = make(map[string]interface{})
				parameters["data"] = make(chan bool)
			})

			It("returns the serialization error", func() {
				Expect(executeErr).To(MatchError("json: unsupported type: chan bool"))
			})
		})
	})

	When("the create is not successful", func() {
		When("the create returns a ServiceKeyNameTaken error", func() {
			BeforeEach(func() {
				response := `
				{
					"description": "The service key name is taken: some-service-key-name",
					"error_code": "CF-ServiceKeyNameTaken",
					"code": 360001
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v2/service_keys"),
						RespondWith(http.StatusBadRequest, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.ServiceKeyTakenError{Message: "The service key name is taken: some-service-key-name"}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})

		When("the create returns a generic error", func() {
			BeforeEach(func() {
				response := `
				{
					"description": "Something went wrong",
					"error_code": "CF-SomeErrorCode",
					"code": 2108219482
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v2/service_keys"),
						RespondWith(http.StatusBadRequest, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.BadRequestError{Message: "Something went wrong"}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})

		When("the create returns a invalid JSON", func() {
			BeforeEach(func() {
				response := `{"entity": {"name": 4}}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v2/service_keys"),
						RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(HaveOccurred())
				Expect(executeErr).To(BeAssignableToTypeOf(&json.UnmarshalTypeError{}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})
	})
})

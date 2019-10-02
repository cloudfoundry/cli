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

var _ = Describe("ServiceBroker", func() {
	var client *Client

	BeforeEach(func() {
		client, _ = NewTestClient()
	})

	Describe("GetServiceBrokers", func() {
		var (
			serviceBrokers []ServiceBroker
			warnings       Warnings
			executeErr     error
		)

		JustBeforeEach(func() {
			serviceBrokers, warnings, executeErr = client.GetServiceBrokers()
		})

		When("service brokers exist", func() {
			BeforeEach(func() {
				response1 := fmt.Sprintf(`
{
	"pagination": {
		"next": {
			"href": "%s/v3/service_brokers?page=2&per_page=2"
		}
	},
	"resources": [
		{
			"name": "service-broker-name-1",
			"guid": "service-broker-guid-1",
			"url": "service-broker-url-1",
			"relationships": {}
		},
		{
			"name": "service-broker-name-2",
			"guid": "service-broker-guid-2",
			"url": "service-broker-url-2",
			"relationships": {}
		}
	]
}`, server.URL())

				response2 := `
{
	"pagination": {
		"next": null
	},
	"resources": [
		{
			"name": "service-broker-name-3",
			"guid": "service-broker-guid-3",
			"url": "service-broker-url-3",
			"relationships": {}
		}
	]
}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/service_brokers"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/service_brokers", "page=2&per_page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)
			})

			It("returns the queried service-broker and all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(serviceBrokers).To(ConsistOf(
					ServiceBroker{Name: "service-broker-name-1", GUID: "service-broker-guid-1", URL: "service-broker-url-1"},
					ServiceBroker{Name: "service-broker-name-2", GUID: "service-broker-guid-2", URL: "service-broker-url-2"},
					ServiceBroker{Name: "service-broker-name-3", GUID: "service-broker-guid-3", URL: "service-broker-url-3"},
				))
				Expect(warnings).To(ConsistOf("this is a warning", "this is another warning"))
			})
		})

		When("the cloud controller returns errors and warnings", func() {
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
	      "detail": "Isolation segment not found",
	      "title": "CF-ResourceNotFound"
	    }
	  ]
	}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/service_brokers"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
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
							Detail: "Isolation segment not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("DeleteServiceBroker", func() {
		var (
			warnings          Warnings
			executeErr        error
			serviceBrokerGUID string
		)

		BeforeEach(func() {
			serviceBrokerGUID = "some-service-broker-guid"
		})

		JustBeforeEach(func() {
			warnings, executeErr = client.DeleteServiceBroker(serviceBrokerGUID)
		})

		When("the Cloud Controller successfully deletes the broker", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v3/service_brokers/some-service-broker-guid"),
						RespondWith(http.StatusOK, "", http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("succeeds and returns warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		When("the broker is space scoped", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v3/service_brokers/some-service-broker-guid"),
						RespondWith(http.StatusOK, "", http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("succeeds and returns warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		When("the Cloud Controller fails to delete the broker", func() {
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
	      "detail": "Service broker not found",
	      "title": "CF-ResourceNotFound"
	    }
	  ]
	}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v3/service_brokers/some-service-broker-guid"),
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
							Detail: "Service broker not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("CreateServiceBroker", func() {
		const (
			name     = "name"
			url      = "url"
			username = "username"
			password = "password"
		)

		var (
			jobURL       JobURL
			warnings     Warnings
			executeErr   error
			spaceGUID    string
			expectedBody map[string]interface{}
		)

		BeforeEach(func() {
			spaceGUID = ""
			expectedBody = map[string]interface{}{
				"name": "name",
				"url":  "url",
				"credentials": map[string]interface{}{
					"type": "basic",
					"data": map[string]string{
						"username": "username",
						"password": "password",
					},
				},
			}
		})

		JustBeforeEach(func() {
			jobURL, warnings, executeErr = client.CreateServiceBroker(name, username, password, url, spaceGUID)
		})

		When("the Cloud Controller successfully creates the broker", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/service_brokers"),
						VerifyJSONRepresenting(expectedBody),
						RespondWith(http.StatusOK, "", http.Header{
							"X-Cf-Warnings": {"this is a warning"},
							"Location":      {"some-job-url"},
						}),
					),
				)
			})

			It("succeeds, returns warnings and job URL", func() {
				Expect(jobURL).To(Equal(JobURL("some-job-url")))
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		When("the broker is space scoped", func() {
			BeforeEach(func() {
				spaceGUID = "space-guid"
				expectedBody["relationships"] = map[string]interface{}{
					"space": map[string]interface{}{
						"data": map[string]string{
							"guid": "space-guid",
						},
					},
				}
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/service_brokers"),
						VerifyJSONRepresenting(expectedBody),
						RespondWith(http.StatusOK, "", http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("succeeds and returns warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		When("the Cloud Controller fails to create the broker", func() {
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
	      "detail": "Isolation segment not found",
	      "title": "CF-ResourceNotFound"
	    }
	  ]
	}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/service_brokers"),
						VerifyJSONRepresenting(expectedBody),
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
							Detail: "Isolation segment not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})
})

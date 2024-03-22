package ccv3_test

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo/v2"
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
			query          []Query
			serviceBrokers []resources.ServiceBroker
			warnings       Warnings
			executeErr     error
		)

		JustBeforeEach(func() {
			serviceBrokers, warnings, executeErr = client.GetServiceBrokers(query...)
		})

		When("there are no service brokers", func() {
			BeforeEach(func() {
				response := `
					{
						"pagination": {
							"next": null
						},
						"resources": []
					}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/service_brokers"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns an empty list", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(serviceBrokers).To(HaveLen(0))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		When("there is a service broker", func() {
			BeforeEach(func() {
				response := `
					{
						"pagination": {
							"next": null
						},
						"resources": [
							{
								"name": "service-broker-name-1",
								"guid": "service-broker-guid-1",
								"url": "service-broker-url-1"
							}
						]
					}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/service_brokers"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)
			})

			It("returns the service broker", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(serviceBrokers).To(ConsistOf(resources.ServiceBroker{
					Name:     "service-broker-name-1",
					GUID:     "service-broker-guid-1",
					URL:      "service-broker-url-1",
					Metadata: nil,
				}))
				Expect(warnings).To(ConsistOf("this is another warning"))
			})
		})

		When("the service broker has labels", func() {
			BeforeEach(func() {
				response := `
					{
						"pagination": {
							"next": null
						},
						"resources": [
							{
								"name": "service-broker-name-1",
								"guid": "service-broker-guid-1",
								"url": "service-broker-url-1",
								"metadata": {
									"labels": {
										"some-key":"some-value",
										"other-key":"other-value"
									}
								}
							}
						]
					}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/service_brokers"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)
			})

			It("returns the service broker with the labels in Metadata", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(serviceBrokers).To(ConsistOf(resources.ServiceBroker{
					Name: "service-broker-name-1",
					GUID: "service-broker-guid-1",
					URL:  "service-broker-url-1",
					Metadata: &resources.Metadata{
						Labels: map[string]types.NullString{
							"some-key":  types.NewNullString("some-value"),
							"other-key": types.NewNullString("other-value"),
						},
					},
				}))
			})
		})

		When("there is more than one page of service brokers", func() {
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
					resources.ServiceBroker{Name: "service-broker-name-1", GUID: "service-broker-guid-1", URL: "service-broker-url-1"},
					resources.ServiceBroker{Name: "service-broker-name-2", GUID: "service-broker-guid-2", URL: "service-broker-url-2"},
					resources.ServiceBroker{Name: "service-broker-name-3", GUID: "service-broker-guid-3", URL: "service-broker-url-3"},
				))
				Expect(warnings).To(ConsistOf("this is a warning", "this is another warning"))
			})
		})

		When("a filter is specified", func() {
			BeforeEach(func() {
				query = []Query{
					{
						Key:    NameFilter,
						Values: []string{"special-unicorn-broker"},
					},
				}

				response := `
					{
						"pagination": {
							"next": null
						},
						"resources": [
							{
								"name": "special-unicorn-broker",
								"guid": "service-broker-guid-1",
								"url": "service-broker-url-1"
							}
						]
					}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/service_brokers", "names=special-unicorn-broker"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)
			})

			It("passes the filter in the query and returns the result", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(serviceBrokers[0].Name).To(Equal("special-unicorn-broker"))
				Expect(warnings).To(ConsistOf("this is another warning"))
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
			jobURL            JobURL
		)

		BeforeEach(func() {
			serviceBrokerGUID = "some-service-broker-guid"
		})

		JustBeforeEach(func() {
			jobURL, warnings, executeErr = client.DeleteServiceBroker(serviceBrokerGUID)
		})

		When("the Cloud Controller successfully deletes the broker", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v3/service_brokers/some-service-broker-guid"),
						RespondWith(http.StatusOK, "", http.Header{
							"X-Cf-Warnings": {"this is a warning"},
							"Location":      {"some-job-url"},
						}),
					),
				)
			})

			It("succeeds and returns warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
				Expect(jobURL).To(Equal(JobURL("some-job-url")))
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
				"authentication": map[string]interface{}{
					"type": "basic",
					"credentials": map[string]string{
						"username": "username",
						"password": "password",
					},
				},
			}
		})

		JustBeforeEach(func() {
			serviceBrokerRequest := resources.ServiceBroker{
				Name:      name,
				URL:       url,
				Username:  username,
				Password:  password,
				SpaceGUID: spaceGUID,
			}
			jobURL, warnings, executeErr = client.CreateServiceBroker(serviceBrokerRequest)
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

	Describe("UpdateServiceBroker", func() {
		var (
			name, guid, url, username, password string
			jobURL                              JobURL
			warnings                            Warnings
			executeErr                          error
			expectedBody                        map[string]interface{}
		)

		BeforeEach(func() {
			expectedBody = map[string]interface{}{
				"url": "new-url",
				"authentication": map[string]interface{}{
					"type": "basic",
					"credentials": map[string]string{
						"username": "new-username",
						"password": "new-password",
					},
				},
			}
			name = ""
			guid = "broker-guid"
			url = "new-url"
			username = "new-username"
			password = "new-password"
		})

		When("the Cloud Controller successfully updates the broker", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPatch, "/v3/service_brokers/"+guid),
						VerifyJSONRepresenting(expectedBody),
						RespondWith(http.StatusOK, "", http.Header{
							"X-Cf-Warnings": {"this is a warning"},
							"Location":      {"some-job-url"},
						}),
					),
				)
			})

			It("succeeds, returns warnings and job URL", func() {
				jobURL, warnings, executeErr = client.UpdateServiceBroker(
					guid,
					resources.ServiceBroker{
						Name:     name,
						URL:      url,
						Username: username,
						Password: password,
					})

				Expect(jobURL).To(Equal(JobURL("some-job-url")))
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		When("the Cloud Controller fails to update the broker", func() {
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
						VerifyRequest(http.MethodPatch, "/v3/service_brokers/"+guid),
						VerifyJSONRepresenting(expectedBody),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns parsed errors and warnings", func() {
				jobURL, warnings, executeErr = client.UpdateServiceBroker(guid,
					resources.ServiceBroker{
						Name:     name,
						URL:      url,
						Username: username,
						Password: password,
					})

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

		When("only name is provided", func() {
			BeforeEach(func() {
				name = "some-name"
				username = ""
				password = ""
				url = ""

				expectedBody = map[string]interface{}{
					"name": name,
				}

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPatch, "/v3/service_brokers/"+guid),
						VerifyJSONRepresenting(expectedBody),
						RespondWith(http.StatusOK, "", http.Header{
							"X-Cf-Warnings": {"this is a warning"},
							"Location":      {"some-job-url"},
						}),
					),
				)
			})

			It("includes only the name in the request body", func() {
				jobURL, warnings, executeErr = client.UpdateServiceBroker(
					guid,
					resources.ServiceBroker{
						Name: name,
					})
				Expect(executeErr).NotTo(HaveOccurred())
			})
		})

		When("partial authentication credentials are provided", func() {
			It("errors without sending any request", func() {
				_, _, executeErr = client.UpdateServiceBroker(
					guid,
					resources.ServiceBroker{Password: password},
				)
				Expect(executeErr).To(HaveOccurred())

				_, _, executeErr = client.UpdateServiceBroker(
					guid,
					resources.ServiceBroker{Username: username},
				)
				Expect(executeErr).To(HaveOccurred())
				Expect(executeErr).To(MatchError("Incorrect usage: both username and password must be defined in order to do an update"))
			})
		})
	})
})

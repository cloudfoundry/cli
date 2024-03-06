package ccv3_test

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Isolation Segments", func() {
	var (
		client *Client
	)

	BeforeEach(func() {
		client, _ = NewTestClient()
	})

	Describe("CreateIsolationSegment", func() {
		var (
			name string

			isolationSegment resources.IsolationSegment
			warnings         Warnings
			executeErr       error
		)

		JustBeforeEach(func() {
			isolationSegment, warnings, executeErr = client.CreateIsolationSegment(resources.IsolationSegment{Name: name})
			name = "an_isolation_segment"
		})

		When("the segment does not exist", func() {
			BeforeEach(func() {
				response := `{
					"guid": "some-guid",
					"name": "an_isolation_segment"
				}`

				requestBody := map[string]string{
					"name": name,
				}

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/isolation_segments"),
						VerifyJSONRepresenting(requestBody),
						RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the created segment and all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(isolationSegment).To(Equal(resources.IsolationSegment{
					Name: name,
					GUID: "some-guid",
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
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
							"code": 10008,
							"detail": "The request is semantically invalid: command presence",
							"title": "CF-UnprocessableEntity"
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/isolation_segments"),
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
							Code:   10008,
							Detail: "The request is semantically invalid: command presence",
							Title:  "CF-UnprocessableEntity",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("GetIsolationSegments", func() {
		var (
			queries []Query

			segments   []resources.IsolationSegment
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			queries = []Query{
				{Key: OrganizationGUIDFilter, Values: []string{"some-org-guid"}},
				{Key: NameFilter, Values: []string{"iso1", "iso2", "iso3"}},
			}
			segments, warnings, executeErr = client.GetIsolationSegments(queries...)
		})

		When("the isolation segments exist", func() {
			BeforeEach(func() {
				response1 := fmt.Sprintf(`{
					"pagination": {
						"next": {
							"href": "%s/v3/isolation_segments?organization_guids=some-org-guid&names=iso1,iso2,iso3&page=2&per_page=2"
						}
					},
					"resources": [
						{
							"name": "iso-name-1",
							"guid": "iso-guid-1"
						},
						{
							"name": "iso-name-2",
							"guid": "iso-guid-2"
						}
					]
				}`, server.URL())
				response2 := `{
					"pagination": {
						"next": null
					},
					"resources": [
						{
							"name": "iso-name-3",
							"guid": "iso-guid-3"
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/isolation_segments", "organization_guids=some-org-guid&names=iso1,iso2,iso3"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/isolation_segments", "organization_guids=some-org-guid&names=iso1,iso2,iso3&page=2&per_page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)
			})

			It("returns the queried applications and all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(segments).To(ConsistOf(
					resources.IsolationSegment{Name: "iso-name-1", GUID: "iso-guid-1"},
					resources.IsolationSegment{Name: "iso-name-2", GUID: "iso-guid-2"},
					resources.IsolationSegment{Name: "iso-name-3", GUID: "iso-guid-3"},
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
							"detail": "App not found",
							"title": "CF-ResourceNotFound"
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/isolation_segments"),
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
							Detail: "App not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("GetIsolationSegment", func() {
		var (
			warnings         Warnings
			executeErr       error
			isolationSegment resources.IsolationSegment
		)

		JustBeforeEach(func() {
			isolationSegment, warnings, executeErr = client.GetIsolationSegment("some-iso-guid")
		})

		When("the isolation segment exists", func() {
			BeforeEach(func() {
				response := `{
					"guid": "some-iso-guid",
					"name": "an_isolation_segment"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/isolation_segments/some-iso-guid"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the isolation segment and all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
				Expect(isolationSegment).To(Equal(resources.IsolationSegment{
					Name: "an_isolation_segment",
					GUID: "some-iso-guid",
				}))
			})
		})

		When("the isolation segment does not exist", func() {
			BeforeEach(func() {
				response := `
				{
					  "errors": [
						    {
									  "detail": "Isolation segment not found",
							      "title": "CF-ResourceNotFound",
							      "code": 10010
						    }
					  ]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/isolation_segments/some-iso-guid"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns a ResourceNotFoundError", func() {
				Expect(warnings).To(ConsistOf("this is a warning"))
				Expect(executeErr).To(MatchError(ccerror.ResourceNotFoundError{Message: "Isolation segment not found"}))
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
							"detail": "Isolation Segment not found",
							"title": "CF-ResourceNotFound"
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/isolation_segments/some-iso-guid"),
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
							Detail: "Isolation Segment not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("DeleteIsolationSegment", func() {
		var (
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			warnings, executeErr = client.DeleteIsolationSegment("some-iso-guid")
		})

		When("the delete is successful", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v3/isolation_segments/some-iso-guid"),
						RespondWith(http.StatusOK, "", http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the queried applications and all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
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
							"detail": "App not found",
							"title": "CF-ResourceNotFound"
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v3/isolation_segments/some-iso-guid"),
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
							Detail: "App not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})
})

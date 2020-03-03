package ccv3_test

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/wrapper"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = FDescribe("shared request helpers", func() {
	var client *Client

	BeforeEach(func() {
		client, _ = NewTestClient()
	})

	Describe("MakeRequest", func() {
		var (
			requestParams RequestParams

			jobURL     JobURL
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			requestParams = RequestParams{}
		})

		JustBeforeEach(func() {
			jobURL, warnings, executeErr = client.MakeRequest(requestParams)
		})

		Context("GET single resource", func() {
			var (
				responseBody Organization
			)

			BeforeEach(func() {
				requestParams = RequestParams{
					RequestName:  internal.GetOrganizationRequest,
					URIParams:    internal.Params{"organization_guid": "some-org-guid"},
					ResponseBody: &responseBody,
				}
			})

			When("organization exists", func() {
				BeforeEach(func() {
					response := `{
					"name": "some-org-name",
					"guid": "some-org-guid",
					"relationships": {
						"quota": {
							"data": {
								"guid": "some-org-quota-guid"
							}
						}
					}
				}`

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v3/organizations/some-org-guid"),
							RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("returns the queried organization and all warnings", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(responseBody).To(Equal(Organization{
						Name:      "some-org-name",
						GUID:      "some-org-guid",
						QuotaGUID: "some-org-quota-guid",
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
					  "code": 10010,
					  "detail": "Org not found",
					  "title": "CF-ResourceNotFound"
					}
				  ]
				}`

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v3/organizations/some-org-guid"),
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
								Detail: "Org not found",
								Title:  "CF-ResourceNotFound",
							},
						},
					}))
					Expect(warnings).To(ConsistOf("this is a warning"))
				})
			})
		})

		Context("POST resource", func() {
			var (
				requestBody  Buildpack
				responseBody Buildpack
			)

			BeforeEach(func() {
				requestBody = Buildpack{
					Name:  "some-buildpack",
					Stack: "some-stack",
				}

				requestParams = RequestParams{
					RequestName:  internal.PostBuildpackRequest,
					RequestBody:  requestBody,
					ResponseBody: &responseBody,
				}
			})

			When("the resource is successfully created", func() {
				BeforeEach(func() {
					response := `{
						"guid": "some-bp-guid",
						"created_at": "2016-03-18T23:26:46Z",
						"updated_at": "2016-10-17T20:00:42Z",
						"name": "some-buildpack",
						"state": "AWAITING_UPLOAD",
						"filename": null,
						"stack": "some-stack",
						"position": 42,
						"enabled": true,
						"locked": false,
						"links": {
						  "self": {
							"href": "/v3/buildpacks/some-bp-guid"
						  },
							"upload": {
								"href": "/v3/buildpacks/some-bp-guid/upload",
								"method": "POST"
							}
						}
					}`

					expectedBody := map[string]interface{}{
						"name":  "some-buildpack",
						"stack": "some-stack",
					}

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPost, "/v3/buildpacks"),
							VerifyJSONRepresenting(expectedBody),
							RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("returns the resource and warnings", func() {
					Expect(jobURL).To(Equal(JobURL("")))
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("this is a warning"))

					expectedBuildpack := Buildpack{
						GUID:     "some-bp-guid",
						Name:     "some-buildpack",
						Stack:    "some-stack",
						Enabled:  types.NullBool{Value: true, IsSet: true},
						Filename: "",
						Locked:   types.NullBool{Value: false, IsSet: true},
						State:    constant.BuildpackAwaitingUpload,
						Position: types.NullInt{Value: 42, IsSet: true},
						Links: APILinks{
							"upload": APILink{
								Method: "POST",
								HREF:   "/v3/buildpacks/some-bp-guid/upload",
							},
							"self": APILink{
								HREF: "/v3/buildpacks/some-bp-guid",
							},
						},
					}

					Expect(responseBody).To(Equal(expectedBuildpack))
				})
			})

			When("the resource returns all errors and warnings", func() {
				BeforeEach(func() {
					response := ` {
  "errors": [
    {
      "code": 10008,
      "detail": "The request is semantically invalid: command presence",
      "title": "CF-UnprocessableEntity"
    },
    {
      "code": 10010,
      "detail": "Buildpack not found",
      "title": "CF-ResourceNotFound"
    }
  ]
}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPost, "/v3/buildpacks"),
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
								Detail: "Buildpack not found",
								Title:  "CF-ResourceNotFound",
							},
						},
					}))
					Expect(warnings).To(ConsistOf("this is a warning"))
				})
			})
		})

		Context("DELETE resource", func() {
			BeforeEach(func() {
				requestParams = RequestParams{
					RequestName: internal.DeleteSpaceRequest,
					URIParams:   internal.Params{"space_guid": "space-guid"},
				}
			})

			When("no errors are encountered", func() {
				BeforeEach(func() {

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodDelete, "/v3/spaces/space-guid"),
							RespondWith(http.StatusAccepted, nil, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}, "Location": []string{"job-url"}}),
						))
				})

				It("deletes the Space and returns all warnings", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf(Warnings{"warning-1", "warning-2"}))
					Expect(jobURL).To(Equal(JobURL("job-url")))
				})
			})

			When("an error is encountered", func() {
				BeforeEach(func() {
					response := `{
   "errors": [
      {
         "detail": "Space not found",
         "title": "CF-ResourceNotFound",
         "code": 10010
      }
   ]
}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodDelete, "/v3/spaces/space-guid"),
							RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
						))
				})

				It("returns an error and all warnings", func() {
					Expect(executeErr).To(MatchError(ccerror.ResourceNotFoundError{
						Message: "Space not found",
					}))
					Expect(warnings).To(ConsistOf(Warnings{"warning-1", "warning-2"}))
				})
			})
		})

		Context("PATCH resource", func() {
			var (
				responseBody Application
			)

			BeforeEach(func() {
				requestBody := Application{
					GUID:                "some-app-guid",
					Name:                "some-app-name",
					StackName:           "some-stack-name",
					LifecycleType:       constant.AppLifecycleTypeBuildpack,
					LifecycleBuildpacks: []string{"some-buildpack"},
					Relationships: Relationships{
						constant.RelationshipTypeSpace: Relationship{GUID: "some-space-guid"},
					},
				}
				requestParams = RequestParams{
					RequestName:  internal.PatchApplicationRequest,
					URIParams:    internal.Params{"app_guid": requestBody.GUID},
					RequestBody:  requestBody,
					ResponseBody: &responseBody,
				}

			})

			When("the application successfully is updated", func() {
				BeforeEach(func() {

					response := `{
					"guid": "some-app-guid",
					"name": "some-app-name",
					"lifecycle": {
						"type": "buildpack",
						"data": {
							"buildpacks": ["some-buildpack"],
							"stack": "some-stack-name"
						}
					}
				}`

					expectedBody := map[string]interface{}{
						"name": "some-app-name",
						"lifecycle": map[string]interface{}{
							"type": "buildpack",
							"data": map[string]interface{}{
								"buildpacks": []string{"some-buildpack"},
								"stack":      "some-stack-name",
							},
						},
						"relationships": map[string]interface{}{
							"space": map[string]interface{}{
								"data": map[string]string{
									"guid": "some-space-guid",
								},
							},
						},
					}
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPatch, "/v3/apps/some-app-guid"),
							VerifyJSONRepresenting(expectedBody),
							RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("returns the updated app and warnings", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("this is a warning"))

					Expect(responseBody).To(Equal(Application{
						GUID:                "some-app-guid",
						StackName:           "some-stack-name",
						LifecycleBuildpacks: []string{"some-buildpack"},
						LifecycleType:       constant.AppLifecycleTypeBuildpack,
						Name:                "some-app-name",
					}))
				})
			})

			When("cc returns back an error or warnings", func() {
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
							VerifyRequest(http.MethodPatch, "/v3/apps/some-app-guid"),
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

	Describe("MakeListRequest", func() {
		var (
			requestParams RequestParams

			includedResources IncludedResources
			warnings          Warnings
			executeErr        error
		)

		JustBeforeEach(func() {
			includedResources, warnings, executeErr = client.MakeListRequest(requestParams)
		})

		Context("With query params and incldued resources", func() {
			var (
				resources []Role
				query     []Query
			)

			BeforeEach(func() {
				resources = []Role{}
				query = []Query{
					{
						Key:    OrganizationGUIDFilter,
						Values: []string{"some-org-name"},
					},
					{
						Key:    Include,
						Values: []string{"users"},
					},
				}
				requestParams = RequestParams{
					RequestName:  internal.GetRolesRequest,
					Query:        query,
					ResponseBody: Role{},
					AppendToList: func(item interface{}) error {
						resources = append(resources, item.(Role))
						return nil
					},
				}
			})

			When("the request succeeds", func() {
				BeforeEach(func() {
					response1 := fmt.Sprintf(`{
	"pagination": {
		"next": {
			"href": "%s/v3/roles?organization_guids=some-org-name&page=2&per_page=1&include=users"
		}
	},
  "resources": [
    {
      "guid": "role-guid-1",
      "type": "organization_user"
    }
  ]
}`, server.URL())
					response2 := `{
							"pagination": {
								"next": null
							},
						 "resources": [
						   {
						     "guid": "role-guid-2",
						     "type": "organization_manager"
						   }
						 ]
						}`

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v3/roles", "organization_guids=some-org-name&include=users"),
							RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"warning-1"}}),
						),
					)
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v3/roles", "organization_guids=some-org-name&page=2&per_page=1&include=users"),
							RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"warning-2"}}),
						),
					)
				})

				It("returns the given resources and all warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
					Expect(resources).To(Equal([]Role{{
						GUID: "role-guid-1",
						Type: constant.OrgUserRole,
					}, {
						GUID: "role-guid-2",
						Type: constant.OrgManagerRole,
					}}))
				})
			})

			When("the request uses the `include` query key", func() {
				BeforeEach(func() {
					response1 := fmt.Sprintf(`{
						"pagination": {
							"next": {
								"href": "%s/v3/roles?organization_guids=some-org-name&page=2&per_page=1&include=users"
							}
						},
						"resources": [
							{
							  "guid": "role-guid-1",
							  "type": "organization_user",
							  "relationships": {
								"user": {
								  "data": {"guid": "user-guid-1"}
								}
							  }
							}
						],
						"included": {
							"users": [
								{
									"guid": "user-guid-1",
									"username": "user-name-1",
									"origin": "uaa"
							  	}
							]
						}
}`, server.URL())
					response2 := `{
							"pagination": {
								"next": null
							},
						 "resources": [
						   {
						     "guid": "role-guid-2",
						     "type": "organization_manager",
							  "relationships": {
								"user": {
								  "data": {"guid": "user-guid-2"}
								}
							  }
						   }
						 ],
						"included": {
							"users": [
							  {
								"guid": "user-guid-2",
								"username": "user-name-2",
								"origin": "uaa"
							  }
							]
						  }
						}`

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v3/roles", "organization_guids=some-org-name&include=users"),
							RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"warning-1"}}),
						),
					)
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v3/roles", "organization_guids=some-org-name&page=2&per_page=1&include=users"),
							RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"warning-2"}}),
						),
					)
				})

				It("returns the given route and all warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

					Expect(resources).To(Equal([]Role{{
						GUID:     "role-guid-1",
						Type:     constant.OrgUserRole,
						UserGUID: "user-guid-1",
					}, {
						GUID:     "role-guid-2",
						Type:     constant.OrgManagerRole,
						UserGUID: "user-guid-2",
					}}))

					Expect(includedResources).To(Equal(IncludedResources{
						Users: []User{
							{GUID: "user-guid-1", Username: "user-name-1", Origin: "uaa"},
							{GUID: "user-guid-2", Username: "user-name-2", Origin: "uaa"},
						},
					}))
				})
			})

			When("the request has a URI parameter", func() {
				var (
					appGUID   string
					resources []Process
				)

				BeforeEach(func() {
					appGUID = "some-app-guid"

					response1 := fmt.Sprintf(`{
						"pagination": {
							"next": {
								"href": "%s/v3/apps/%s/processes?page=2"
							}
						},
					  "resources": [
							{
							  "guid": "process-guid-1"
							}
					  	]
					}`, server.URL(), appGUID)
					response2 := `{
							"pagination": {
								"next": null
							},
							 "resources": [
							   {
								 "guid": "process-guid-2"
							   }
							 ]
						}`

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, fmt.Sprintf("/v3/apps/%s/processes", appGUID)),
							RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"warning-1"}}),
						),
					)
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, fmt.Sprintf("/v3/apps/%s/processes", appGUID), "page=2"),
							RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"warning-2"}}),
						),
					)

					requestParams = RequestParams{
						RequestName:  internal.GetApplicationProcessesRequest,
						URIParams:    internal.Params{"app_guid": appGUID},
						ResponseBody: Process{},
						AppendToList: func(item interface{}) error {
							resources = append(resources, item.(Process))
							return nil
						},
					}
				})

				It("returns the given resources and all warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
					Expect(resources).To(Equal([]Process{{
						GUID: "process-guid-1",
					}, {
						GUID: "process-guid-2",
					}}))
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
      "detail": "Org not found",
      "title": "CF-ResourceNotFound"
    }
  ]
}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v3/roles"),
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
								Detail: "Org not found",
								Title:  "CF-ResourceNotFound",
							},
						},
					}))
					Expect(warnings).To(ConsistOf("this is a warning"))
				})
			})
		})
	})

	Describe("MakeRequestReceiveRaw", func() {
		var (
			requestName string
			uriParams   internal.Params

			rawResponseBody      []byte
			warnings             Warnings
			executeErr           error
			responseBodyMimeType string
		)

		JustBeforeEach(func() {
			rawResponseBody, warnings, executeErr = client.MakeRequestReceiveRaw(requestName, uriParams, responseBodyMimeType)
		})

		Context("GET raw bytes (YAML data)", func() {
			var (
				expectedResponseBody []byte
			)

			BeforeEach(func() {
				requestName = internal.GetApplicationManifestRequest
				responseBodyMimeType = "application/x-yaml"
				uriParams = internal.Params{"app_guid": "some-app-guid"}
			})

			When("getting requested data is successful", func() {
				BeforeEach(func() {
					expectedResponseBody = []byte("---\n- banana")

					server.AppendHandlers(
						CombineHandlers(
							CombineHandlers(
								VerifyRequest(http.MethodGet, "/v3/apps/some-app-guid/manifest"),
								VerifyHeaderKV("Accept", "application/x-yaml"),
								RespondWith(
									http.StatusOK,
									expectedResponseBody,
									http.Header{
										"Content-Type":  {"application/x-yaml"},
										"X-Cf-Warnings": {"this is a warning"},
									}),
							),
						),
					)
				})

				It("returns the raw response body and all warnings", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(rawResponseBody).To(Equal(expectedResponseBody))
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
					  "detail": "Org not found",
					  "title": "CF-ResourceNotFound"
					}
				  ]
				}`

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v3/apps/some-app-guid/manifest"),
							VerifyHeaderKV("Accept", "application/x-yaml"),
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
								Detail: "Org not found",
								Title:  "CF-ResourceNotFound",
							},
						},
					}))
					Expect(warnings).To(ConsistOf("this is a warning"))
				})
			})
		})
	})

	Describe("MakeRequestSendRaw", func() {
		var (
			requestName         string
			uriParams           internal.Params
			requestBodyMimeType string

			requestBody      []byte
			responseBody     Package
			expectedJobURL   string
			responseLocation string
			warnings         Warnings
			executeErr       error
		)

		JustBeforeEach(func() {
			responseLocation, warnings, executeErr = client.MakeRequestSendRaw(requestName, uriParams, requestBody, requestBodyMimeType, &responseBody)
		})

		BeforeEach(func() {
			requestBody = []byte("fake-package-file")
			expectedJobURL = "apply-manifest-job-url"
			responseBody = Package{}

			requestName = internal.PostPackageBitsRequest
			uriParams = internal.Params{"package_guid": "package-guid"}
			requestBodyMimeType = "multipart/form-data"
		})

		When("the resource is successfully created", func() {
			BeforeEach(func() {
				response := `{
					"guid": "some-pkg-guid",
					"type": "docker",
					"state": "PROCESSING_UPLOAD",
					"links": {
						"upload": {
							"href": "some-package-upload-url",
							"method": "POST"
						}
					}
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/packages/package-guid/upload"),
						VerifyBody(requestBody),
						VerifyHeaderKV("Content-Type", "multipart/form-data"),
						RespondWith(http.StatusCreated, response, http.Header{
							"X-Cf-Warnings": {"this is a warning"},
							"Location":      {expectedJobURL},
						}),
					),
				)
			})

			It("returns the resource and warnings", func() {
				Expect(responseLocation).To(Equal(expectedJobURL))
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))

				expectedPackage := Package{
					GUID:  "some-pkg-guid",
					Type:  constant.PackageTypeDocker,
					State: constant.PackageProcessingUpload,
					Links: map[string]APILink{
						"upload": APILink{HREF: "some-package-upload-url", Method: http.MethodPost},
					},
				}
				Expect(responseBody).To(Equal(expectedPackage))
			})
		})

		When("the resource returns all errors and warnings", func() {
			BeforeEach(func() {
				response := ` {
  "errors": [
    {
      "code": 10008,
      "detail": "The request is semantically invalid: command presence",
      "title": "CF-UnprocessableEntity"
    },
    {
      "code": 10010,
      "detail": "Hamster not found",
      "title": "CF-ResourceNotFound"
    }
  ]
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/packages/package-guid/upload"),
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
							Detail: "Hamster not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("MakeRequestUploadAsync", func() {
		var (
			content				string
			requestName         string
			uriParams           internal.Params
			requestBodyMimeType string
			requestBody         io.ReadSeeker
			dataLength          int64
			writeErrors         chan error

			responseLocation string
			responseBody     Package
			warning          string
			warnings         Warnings
			executeErr       error
		)
		BeforeEach(func() {
			warning = "upload-async-warning"
			content = "I love my cats!"
			requestBody = strings.NewReader(content)
			dataLength = int64(len(content))
			writeErrors = make(chan error)

			response := `{
						"guid": "some-package-guid",
						"type": "bits",
						"state": "PROCESSING_UPLOAD"
					}`

			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/v3/packages/package-guid/upload"),
					VerifyHeaderKV("Content-Type", "multipart/form-data"),
					VerifyBody([]byte(content)),
					RespondWith(http.StatusOK, response, http.Header{
						"X-Cf-Warnings": {warning},
						"Location":      {"something"},
					}),
				),
			)
		})
		JustBeforeEach(func() {
			responseBody = Package{}
			requestName = internal.PostPackageBitsRequest
			requestBodyMimeType = "multipart/form-data"
			uriParams = internal.Params{"package_guid": "package-guid"}

			responseLocation, warnings, executeErr = client.MakeRequestUploadAsync(
				requestName,
				uriParams,
				requestBodyMimeType,
				requestBody,
				dataLength,
				&responseBody,
				writeErrors,
			)
		})
		When("there are no errors (happy path)", func() {
			BeforeEach(func() {
				go func() {
					close(writeErrors)
				}()
			})
			It("returns the location and any warnings and error", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(responseLocation).To(Equal("something"))
				Expect(responseBody).To(Equal(Package{
					GUID:  "some-package-guid",
					State: "PROCESSING_UPLOAD",
					Type:  "bits",
				}))
				Expect(warnings).To(Equal(Warnings{warning}))
			})
		})

		When("There are write errors", func() {
			BeforeEach(func() {
				go func() {
					writeErrors <- errors.New("first-error")
					writeErrors <- errors.New("second-error")
					close(writeErrors)
				}()
			})
			It("returns the first error", func() {
				Expect(executeErr).To(MatchError("first-error"))
			})
		})

		When("there are HTTP connection errors", func() {
			BeforeEach(func() {
				server.Close()
				close(writeErrors)
			})

			It("returns the first error", func() {
				_, ok := executeErr.(ccerror.RequestError)
				Expect(ok).To(BeTrue())
			})
		})

		When("a retryable error occurs", func() {
			BeforeEach(func() {
				wrapper := &wrapper.CustomWrapper{
					CustomMake: func(connection cloudcontroller.Connection, request *cloudcontroller.Request, response *cloudcontroller.Response) error {
						defer GinkgoRecover() // Since this will be running in a thread

						if strings.HasSuffix(request.URL.String(), "/v3/droplets/some-droplet-guid/upload") {
							_, err := ioutil.ReadAll(request.Body)
							Expect(err).ToNot(HaveOccurred())
							Expect(request.Body.Close()).ToNot(HaveOccurred())
							return request.ResetBody()
						}
						return connection.Make(request, response)
					},
				}

				client, _ = NewTestClient(Config{Wrappers: []ConnectionWrapper{wrapper}})
			})

			It("returns the PipeSeekError", func() {
				Expect(executeErr).To(MatchError(ccerror.PipeSeekError{}))
			})
		})

		When("an http error occurs mid-transfer", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some read error")

				wrapper := &wrapper.CustomWrapper{
					CustomMake: func(connection cloudcontroller.Connection, request *cloudcontroller.Request, response *cloudcontroller.Response) error {
						defer GinkgoRecover() // Since this will be running in a thread

						if strings.HasSuffix(request.URL.String(), "/v3/droplets/some-droplet-guid/upload") {
							defer request.Body.Close()
							readBytes, err := ioutil.ReadAll(request.Body)
							Expect(err).ToNot(HaveOccurred())
							Expect(len(readBytes)).To(BeNumerically(">", len(content)))
							return expectedErr
						}
						return connection.Make(request, response)
					},
				}

				client, _ = NewTestClient(Config{Wrappers: []ConnectionWrapper{wrapper}})
			})

			It("returns the http error", func() {
				Expect(executeErr).To(MatchError(expectedErr))
			})
		})
	})
})

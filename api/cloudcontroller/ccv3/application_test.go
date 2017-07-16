package ccv3_test

import (
	"fmt"
	"net/http"
	"net/url"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Application", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("MarshalJSON", func() {
		var (
			app      Application
			appBytes []byte
			err      error
		)

		JustBeforeEach(func() {
			appBytes, err = app.MarshalJSON()
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when no buildpacks are provided", func() {
			BeforeEach(func() {
				app = Application{}
			})

			It("omits the Lifecycle from the JSON", func() {
				Expect(string(appBytes)).To(Equal("{}"))
			})
		})

		Context("when default buildpack is provided", func() {
			BeforeEach(func() {
				app = Application{Buildpacks: []string{"default"}}
			})

			It("sets the Lifecycle buildpack to be empty in the JSON", func() {
				Expect(string(appBytes)).To(Equal(`{"lifecycle":{"data":{"buildpacks":null},"type":"buildpack"}}`))
			})
		})

		Context("when null buildpack is provided", func() {
			BeforeEach(func() {
				app = Application{Buildpacks: []string{"null"}}
			})

			It("sets the Lifecycle buildpack to be empty in the JSON", func() {
				Expect(string(appBytes)).To(Equal(`{"lifecycle":{"data":{"buildpacks":null},"type":"buildpack"}}`))
			})
		})

		Context("when other buildpacks are provided", func() {
			BeforeEach(func() {
				app = Application{Buildpacks: []string{"some-buildpack"}}
			})

			It("sets them in the JSON", func() {
				Expect(string(appBytes)).To(Equal(`{"lifecycle":{"data":{"buildpacks":["some-buildpack"]},"type":"buildpack"}}`))
			})
		})
	})

	Describe("GetApplications", func() {
		Context("when applications exist", func() {
			BeforeEach(func() {
				response1 := fmt.Sprintf(`{
	"pagination": {
		"next": {
			"href": "%s/v3/apps?space_guids=some-space-guid&names=some-app-name&page=2&per_page=2"
		}
	},
  "resources": [
    {
      "name": "app-name-1",
      "guid": "app-guid-1",
			"lifecycle": {
				"type": "buildpack",
				"data": {
					"buildpacks": ["some-buildpack"],
					"stack": "some-stack"
				}
			}
    },
    {
      "name": "app-name-2",
      "guid": "app-guid-2"
    }
  ]
}`, server.URL())
				response2 := `{
	"pagination": {
		"next": null
	},
	"resources": [
	  {
      "name": "app-name-3",
		  "guid": "app-guid-3"
		}
	]
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/apps", "space_guids=some-space-guid&names=some-app-name"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/apps", "space_guids=some-space-guid&names=some-app-name&page=2&per_page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)
			})

			It("returns the queried applications and all warnings", func() {
				apps, warnings, err := client.GetApplications(url.Values{
					SpaceGUIDFilter: []string{"some-space-guid"},
					NameFilter:      []string{"some-app-name"},
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(apps).To(ConsistOf(
					Application{
						Name:       "app-name-1",
						GUID:       "app-guid-1",
						Buildpacks: []string{"some-buildpack"},
					},
					Application{Name: "app-name-2", GUID: "app-guid-2"},
					Application{Name: "app-name-3", GUID: "app-guid-3"},
				))
				Expect(warnings).To(ConsistOf("this is a warning", "this is another warning"))
			})
		})

		Context("when the cloud controller returns errors and warnings", func() {
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
						VerifyRequest(http.MethodGet, "/v3/apps"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				_, warnings, err := client.GetApplications(nil)
				Expect(err).To(MatchError(ccerror.V3UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V3ErrorResponse: ccerror.V3ErrorResponse{
						[]ccerror.V3Error{
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
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("UpdateApplication", func() {
		Context("when the application successfully is updated", func() {
			BeforeEach(func() {
				response := `{
					"guid": "some-app-guid",
					"name": "some-app-name"
				}`

				expectedBody := map[string]interface{}{
					"name": "some-app-name",
					"lifecycle": map[string]interface{}{
						"type": "buildpack",
						"data": map[string]interface{}{
							"buildpacks": []string{"some-buildpack"},
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
				app, warnings, err := client.UpdateApplication(Application{
					GUID:       "some-app-guid",
					Name:       "some-app-name",
					Buildpacks: []string{"some-buildpack"},
					Relationships: Relationships{
						SpaceRelationship: Relationship{GUID: "some-space-guid"},
					},
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))

				Expect(app).To(Equal(Application{
					Name: "some-app-name",
					GUID: "some-app-guid",
				}))
			})
		})

		Context("when cc returns back an error or warnings", func() {
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
				_, warnings, err := client.UpdateApplication(Application{GUID: "some-app-guid"})
				Expect(err).To(MatchError(ccerror.V3UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V3ErrorResponse: ccerror.V3ErrorResponse{
						[]ccerror.V3Error{
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
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("CreateApplication", func() {
		Context("when the application successfully is created", func() {
			BeforeEach(func() {
				response := `{
					"guid": "some-app-guid",
					"name": "some-app-name"
				}`

				expectedBody := map[string]interface{}{
					"name": "some-app-name",
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
						VerifyRequest(http.MethodPost, "/v3/apps"),
						VerifyJSONRepresenting(expectedBody),
						RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the created app and warnings", func() {
				app, warnings, err := client.CreateApplication(Application{
					Name: "some-app-name",
					Relationships: Relationships{
						SpaceRelationship: Relationship{GUID: "some-space-guid"},
					},
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))

				Expect(app).To(Equal(Application{
					Name: "some-app-name",
					GUID: "some-app-guid",
				}))
			})
		})

		Context("when the caller specifies a buildpack", func() {
			BeforeEach(func() {
				response := `{
					"guid": "some-app-guid",
					"name": "some-app-name",
					"lifecycle": {
						"type": "buildpack",
						"data": {
							"buildpacks": ["some-buildpack"]
					  }
					}
				}`

				expectedBody := map[string]interface{}{
					"name": "some-app-name",
					"lifecycle": map[string]interface{}{
						"type": "buildpack",
						"data": map[string]interface{}{
							"buildpacks": []string{"some-buildpack"},
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
						VerifyRequest(http.MethodPost, "/v3/apps"),
						VerifyJSONRepresenting(expectedBody),
						RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the created app and warnings", func() {
				app, warnings, err := client.CreateApplication(Application{
					Name:       "some-app-name",
					Buildpacks: []string{"some-buildpack"},
					Relationships: Relationships{
						SpaceRelationship: Relationship{GUID: "some-space-guid"},
					},
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))

				Expect(app).To(Equal(Application{
					Name:       "some-app-name",
					GUID:       "some-app-guid",
					Buildpacks: []string{"some-buildpack"},
				}))
			})
		})

		Context("when cc returns back an error or warnings", func() {
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
						VerifyRequest(http.MethodPost, "/v3/apps"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				_, warnings, err := client.CreateApplication(Application{})
				Expect(err).To(MatchError(ccerror.V3UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V3ErrorResponse: ccerror.V3ErrorResponse{
						[]ccerror.V3Error{
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
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("SetApplicationDroplet", func() {
		Context("it sets the droplet", func() {
			BeforeEach(func() {
				response := `
{
  "data": {
    "guid": "some-droplet-guid"
  },
  "links": {
    "self": {
      "href": "https://api.example.org/v3/apps/some-app-guid/relationships/current_droplet"
    },
    "related": {
      "href": "https://api.example.org/v3/apps/some-app-guid/droplets/current"
    }
  }
}`
				requestBody := map[string]interface{}{
					"data": map[string]string{
						"guid": "some-droplet-guid",
					},
				}

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPatch, "/v3/apps/some-app-guid/relationships/current_droplet"),
						VerifyJSONRepresenting(requestBody),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns warnings and no error", func() {
				relationship, warnings, err := client.SetApplicationDroplet("some-app-guid", "some-droplet-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
				Expect(relationship.GUID).To(Equal("some-droplet-guid"))
			})
		})
	})
	Context("when setting the app to the new droplet returns errors and warnings", func() {
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
			requestBody := map[string]interface{}{
				"data": map[string]string{
					"guid": "some-droplet-guid",
				},
			}

			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPatch, "/v3/apps/no-such-app-guid/relationships/current_droplet"),
					VerifyJSONRepresenting(requestBody),
					RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
				),
			)

		})

		It("returns the error and all warnings", func() {
			_, warnings, err := client.SetApplicationDroplet("no-such-app-guid", "some-droplet-guid")
			Expect(err).To(MatchError(ccerror.V3UnexpectedResponseError{
				ResponseCode: http.StatusTeapot,
				V3ErrorResponse: ccerror.V3ErrorResponse{
					[]ccerror.V3Error{
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
				},
			}))
			Expect(warnings).To(ConsistOf("this is a warning"))
		})
	})

	Describe("StopApplication", func() {
		Context("when the response succeeds", func() {
			BeforeEach(func() {
				response := `
{
	"guid": "some-app-guid",
	"name": "some-app",
	"state": "STOPPED"
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPut, "/v3/apps/some-app-guid/stop"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns warnings and no error", func() {
				warnings, err := client.StopApplication("some-app-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Context("when stopping the app returns errors and warnings", func() {
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
					VerifyRequest(http.MethodPut, "/v3/apps/no-such-app-guid/stop"),
					RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
				),
			)

		})

		It("returns the error and all warnings", func() {
			warnings, err := client.StopApplication("no-such-app-guid")
			Expect(err).To(MatchError(ccerror.V3UnexpectedResponseError{
				ResponseCode: http.StatusTeapot,
				V3ErrorResponse: ccerror.V3ErrorResponse{
					[]ccerror.V3Error{
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
				},
			}))
			Expect(warnings).To(ConsistOf("this is a warning"))
		})
	})

	Describe("StartApplication", func() {
		Context("when the response succeeds", func() {
			BeforeEach(func() {
				response := `
{
	"guid": "some-app-guid",
	"name": "some-app"
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPut, "/v3/apps/some-app-guid/start"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns warnings and no error", func() {
				app, warnings, err := client.StartApplication("some-app-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
				Expect(app.GUID).To(Equal("some-app-guid"))
			})
		})
	})
	Context("when starting the app returns errors and warnings", func() {
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
					VerifyRequest(http.MethodPut, "/v3/apps/no-such-app-guid/start"),
					RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
				),
			)

		})

		It("returns the error and all warnings", func() {
			_, warnings, err := client.StartApplication("no-such-app-guid")
			Expect(err).To(MatchError(ccerror.V3UnexpectedResponseError{
				ResponseCode: http.StatusTeapot,
				V3ErrorResponse: ccerror.V3ErrorResponse{
					[]ccerror.V3Error{
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
				},
			}))
			Expect(warnings).To(ConsistOf("this is a warning"))
		})
	})
})

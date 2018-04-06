package ccv3_test

import (
	"encoding/json"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Application", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("Application", func() {
		Describe("MarshalJSON", func() {
			var (
				app      Application
				appBytes []byte
				err      error
			)

			BeforeEach(func() {
				app = Application{}
			})

			JustBeforeEach(func() {
				appBytes, err = app.MarshalJSON()
				Expect(err).ToNot(HaveOccurred())
			})

			Context("when no lifecycle is provided", func() {
				BeforeEach(func() {
					app = Application{}
				})

				It("omits the lifecycle from the JSON", func() {
					Expect(string(appBytes)).To(Equal("{}"))
				})
			})

			Context("when lifecycle type docker is provided", func() {
				BeforeEach(func() {
					app = Application{
						LifecycleType: constant.AppLifecycleTypeDocker,
					}
				})

				It("sets lifecycle type to docker with empty data", func() {
					Expect(string(appBytes)).To(MatchJSON(`{"lifecycle":{"type":"docker","data":{}}}`))
				})
			})

			Context("when lifecycle type buildpack is provided", func() {
				BeforeEach(func() {
					app.LifecycleType = constant.AppLifecycleTypeBuildpack
				})

				Context("when no buildpacks are provided", func() {
					It("omits the lifecycle from the JSON", func() {
						Expect(string(appBytes)).To(Equal("{}"))
					})
				})

				Context("when default buildpack is provided", func() {
					BeforeEach(func() {
						app.LifecycleBuildpacks = []string{"default"}
					})

					It("sets the lifecycle buildpack to be empty in the JSON", func() {
						Expect(string(appBytes)).To(Equal(`{"lifecycle":{"data":{"buildpacks":null},"type":"buildpack"}}`))
					})
				})

				Context("when null buildpack is provided", func() {
					BeforeEach(func() {
						app.LifecycleBuildpacks = []string{"null"}
					})

					It("sets the Lifecycle buildpack to be empty in the JSON", func() {
						Expect(string(appBytes)).To(Equal(`{"lifecycle":{"data":{"buildpacks":null},"type":"buildpack"}}`))
					})
				})

				Context("when other buildpacks are provided", func() {
					BeforeEach(func() {
						app.LifecycleBuildpacks = []string{"some-buildpack"}
					})

					It("sets them in the JSON", func() {
						Expect(string(appBytes)).To(Equal(`{"lifecycle":{"data":{"buildpacks":["some-buildpack"]},"type":"buildpack"}}`))
					})
				})
			})
		})

		Describe("UnmarshalJSON", func() {
			var (
				app      Application
				appBytes []byte
				err      error
			)

			BeforeEach(func() {
				appBytes = []byte("{}")
			})

			JustBeforeEach(func() {
				err = json.Unmarshal(appBytes, &app)
				Expect(err).ToNot(HaveOccurred())
			})

			Context("when no lifecycle is provided", func() {
				BeforeEach(func() {
					appBytes = []byte("{}")
				})

				It("omits the lifecycle from the JSON", func() {
					Expect(app).To(Equal(Application{}))
				})
			})

			Context("when lifecycle type docker is provided", func() {
				BeforeEach(func() {
					appBytes = []byte(`{"lifecycle":{"type":"docker","data":{}}}`)
				})
				It("sets the lifecycle type to docker with empty data", func() {
					Expect(app).To(Equal(Application{
						LifecycleType: constant.AppLifecycleTypeDocker,
					}))
				})
			})

			Context("when lifecycle type buildpack is provided", func() {

				Context("when other buildpacks are provided", func() {
					BeforeEach(func() {
						appBytes = []byte(`{"lifecycle":{"data":{"buildpacks":["some-buildpack"]},"type":"buildpack"}}`)
					})

					It("sets them in the JSON", func() {
						Expect(app).To(Equal(Application{
							LifecycleType:       constant.AppLifecycleTypeBuildpack,
							LifecycleBuildpacks: []string{"some-buildpack"},
						}))
					})
				})
			})
		})
	})

	Describe("CreateApplication", func() {
		var (
			appToCreate Application

			createdApp Application
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			createdApp, warnings, executeErr = client.CreateApplication(appToCreate)
		})

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

				appToCreate = Application{
					Name: "some-app-name",
					Relationships: Relationships{
						constant.RelationshipTypeSpace: Relationship{GUID: "some-space-guid"},
					},
				}
			})

			It("returns the created app and warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))

				Expect(createdApp).To(Equal(Application{
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

				appToCreate = Application{
					Name:                "some-app-name",
					LifecycleType:       constant.AppLifecycleTypeBuildpack,
					LifecycleBuildpacks: []string{"some-buildpack"},
					Relationships: Relationships{
						constant.RelationshipTypeSpace: Relationship{GUID: "some-space-guid"},
					},
				}
			})

			It("returns the created app and warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))

				Expect(createdApp).To(Equal(Application{
					Name:                "some-app-name",
					GUID:                "some-app-guid",
					LifecycleType:       constant.AppLifecycleTypeBuildpack,
					LifecycleBuildpacks: []string{"some-buildpack"},
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
				Expect(executeErr).To(MatchError(ccerror.V3UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V3ErrorResponse: ccerror.V3ErrorResponse{
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
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("GetApplications", func() {
		var (
			filters []Query

			apps       []Application
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			apps, warnings, executeErr = client.GetApplications(filters...)
		})

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

				filters = []Query{
					{Key: SpaceGUIDFilter, Values: []string{"some-space-guid"}},
					{Key: NameFilter, Values: []string{"some-app-name"}},
				}
			})

			It("returns the queried applications and all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning", "this is another warning"))

				Expect(apps).To(ConsistOf(
					Application{
						Name:                "app-name-1",
						GUID:                "app-guid-1",
						LifecycleType:       constant.AppLifecycleTypeBuildpack,
						LifecycleBuildpacks: []string{"some-buildpack"},
					},
					Application{Name: "app-name-2", GUID: "app-guid-2"},
					Application{Name: "app-name-3", GUID: "app-guid-3"},
				))
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
				Expect(executeErr).To(MatchError(ccerror.V3UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V3ErrorResponse: ccerror.V3ErrorResponse{
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
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("UpdateApplication", func() {
		var (
			appToUpdate Application

			updatedApp Application
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			updatedApp, warnings, executeErr = client.UpdateApplication(appToUpdate)
		})

		Context("when the application successfully is updated", func() {
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
						VerifyRequest(http.MethodPatch, "/v3/apps/some-app-guid"),
						VerifyJSONRepresenting(expectedBody),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)

				appToUpdate = Application{
					GUID:                "some-app-guid",
					Name:                "some-app-name",
					LifecycleType:       constant.AppLifecycleTypeBuildpack,
					LifecycleBuildpacks: []string{"some-buildpack"},
					Relationships: Relationships{
						constant.RelationshipTypeSpace: Relationship{GUID: "some-space-guid"},
					},
				}
			})

			It("returns the updated app and warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))

				Expect(updatedApp).To(Equal(Application{
					GUID:                "some-app-guid",
					LifecycleBuildpacks: []string{"some-buildpack"},
					LifecycleType:       constant.AppLifecycleTypeBuildpack,
					Name:                "some-app-name",
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

				appToUpdate = Application{
					GUID: "some-app-guid",
				}
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.V3UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V3ErrorResponse: ccerror.V3ErrorResponse{
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
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("UpdateApplicationStop", func() {
		var (
			responseApp Application
			warnings    Warnings
			executeErr  error
		)

		JustBeforeEach(func() {
			responseApp, warnings, executeErr = client.UpdateApplicationStop("some-app-guid")
		})

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
						VerifyRequest(http.MethodPost, "/v3/apps/some-app-guid/actions/stop"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the application, warnings, and no error", func() {
				Expect(responseApp).To(Equal(Application{
					GUID:  "some-app-guid",
					Name:  "some-app",
					State: constant.ApplicationStopped,
				}))
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		Context("when the CC returns an error", func() {
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
						VerifyRequest(http.MethodPost, "/v3/apps/some-app-guid/actions/stop"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns no app, the error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.V3UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V3ErrorResponse: ccerror.V3ErrorResponse{
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
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("UpdateApplicationStart", func() {
		var (
			app        Application
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			app, warnings, executeErr = client.UpdateApplicationStart("some-app-guid")
		})

		Context("when the response succeeds", func() {
			BeforeEach(func() {
				response := `
{
	"guid": "some-app-guid",
	"name": "some-app"
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/apps/some-app-guid/actions/start"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns warnings and no error", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
				Expect(app.GUID).To(Equal("some-app-guid"))
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
						VerifyRequest(http.MethodPost, "/v3/apps/some-app-guid/actions/start"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.V3UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V3ErrorResponse: ccerror.V3ErrorResponse{
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
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})
})

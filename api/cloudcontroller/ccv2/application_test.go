package ccv2_test

import (
	"net/http"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Application", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("CreateApplication", func() {
		Context("when the update is successful", func() {
			Context("when setting the minimum", func() { // are we **only** encoding the things we want
				BeforeEach(func() {
					response := `
						{
							"metadata": {
								"guid": "some-app-guid"
							},
							"entity": {
								"name": "some-app-name",
								"space_guid": "some-space-guid"
							}
						}`
					requestBody := map[string]string{
						"name":       "some-app-name",
						"space_guid": "some-space-guid",
					}
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPost, "/v2/apps"),
							VerifyJSONRepresenting(requestBody),
							RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("returns the created object and warnings", func() {
					app, warnings, err := client.CreateApplication(Application{
						Name:      "some-app-name",
						SpaceGUID: "some-space-guid",
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(app).To(Equal(Application{
						GUID: "some-app-guid",
						Name: "some-app-name",
					}))
					Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				})
			})
		})

		Context("when the create returns an error", func() {
			BeforeEach(func() {
				response := `
					{
						"description": "Request invalid due to parse error: Field: name, Error: Missing field name, Field: space_guid, Error: Missing field space_guid",
						"error_code": "CF-MessageParseError",
						"code": 1001
					}
			`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v2/apps"),
						RespondWith(http.StatusBadRequest, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and warnings", func() {
				_, warnings, err := client.CreateApplication(Application{})
				Expect(err).To(MatchError(ccerror.BadRequestError{Message: "Request invalid due to parse error: Field: name, Error: Missing field name, Field: space_guid, Error: Missing field space_guid"}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})
	})

	Describe("GetApplication", func() {
		BeforeEach(func() {
			response := `{
						"metadata": {
							"guid": "app-guid-1",
							"updated_at": null
						},
						"entity": {
							"buildpack": "ruby 1.6.29",
							"command": "some-command",
							"detected_start_command": "echo 'I am a banana'",
							"disk_quota": 586,
							"detected_buildpack": null,
							"docker_credentials": {
								"username": "docker-username",
								"password": "docker-password"
							},
							"docker_image": "some-docker-path",
							"environment_json": {
								"key1": "val1",
								"key2": 83493475092347,
								"key3": true,
								"key4": 75821.521
							},
							"health_check_timeout": 120,
							"health_check_type": "port",
							"health_check_http_endpoint": "/",
							"instances": 13,
							"memory": 1024,
							"name": "app-name-1",
							"package_state": "FAILED",
							"package_updated_at": "2015-03-10T23:11:54Z",
							"stack_guid": "some-stack-guid",
							"staging_failed_description": "some-staging-failed-description",
							"staging_failed_reason": "some-reason",
							"state": "STOPPED"
						}
			}`
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/v2/apps/app-guid-1"),
					RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
				),
			)
		})

		Context("when apps exist", func() {
			It("returns the app", func() {
				app, warnings, err := client.GetApplication("app-guid-1")
				Expect(err).NotTo(HaveOccurred())

				updatedAt, err := time.Parse(time.RFC3339, "2015-03-10T23:11:54Z")
				Expect(err).NotTo(HaveOccurred())

				Expect(app).To(Equal(Application{
					Buildpack:            types.FilteredString{IsSet: true, Value: "ruby 1.6.29"},
					Command:              types.FilteredString{IsSet: true, Value: "some-command"},
					DetectedBuildpack:    types.FilteredString{},
					DetectedStartCommand: types.FilteredString{IsSet: true, Value: "echo 'I am a banana'"},
					DiskQuota:            types.NullByteSizeInMb{IsSet: true, Value: 586},
					DockerCredentials: DockerCredentials{
						Username: "docker-username",
						Password: "docker-password",
					},
					DockerImage: "some-docker-path",
					EnvironmentVariables: map[string]string{
						"key1": "val1",
						"key2": "83493475092347",
						"key3": "true",
						"key4": "75821.521",
					},
					GUID:                     "app-guid-1",
					HealthCheckTimeout:       120,
					HealthCheckType:          "port",
					HealthCheckHTTPEndpoint:  "/",
					Instances:                types.NullInt{Value: 13, IsSet: true},
					Memory:                   types.NullByteSizeInMb{IsSet: true, Value: 1024},
					Name:                     "app-name-1",
					PackageState:             ApplicationPackageFailed,
					PackageUpdatedAt:         updatedAt,
					StackGUID:                "some-stack-guid",
					StagingFailedDescription: "some-staging-failed-description",
					StagingFailedReason:      "some-reason",
					State:                    ApplicationStopped,
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})
	})

	Describe("GetApplications", func() {
		BeforeEach(func() {
			response1 := `{
				"next_url": "/v2/apps?q=space_guid:some-space-guid&page=2",
				"resources": [
					{
						"metadata": {
							"guid": "app-guid-1",
							"updated_at": null
						},
						"entity": {
							"buildpack": "ruby 1.6.29",
							"detected_start_command": "echo 'I am a banana'",
							"disk_quota": 586,
							"detected_buildpack": null,
							"health_check_type": "port",
							"health_check_http_endpoint": "/",
							"instances": 13,
							"memory": 1024,
							"name": "app-name-1",
							"package_state": "FAILED",
							"package_updated_at": "2015-03-10T23:11:54Z",
							"stack_guid": "some-stack-guid",
							"staging_failed_reason": "some-reason",
							"state": "STOPPED"
						}
					},
					{
						"metadata": {
							"guid": "app-guid-2",
							"updated_at": null
						},
						"entity": {
							"name": "app-name-2",
							"detected_buildpack": "ruby 1.6.29",
							"package_updated_at": null
						}
					}
				]
			}`
			response2 := `{
				"next_url": null,
				"resources": [
					{
						"metadata": {
							"guid": "app-guid-3",
							"updated_at": null
						},
						"entity": {
							"name": "app-name-3"
						}
					},
					{
						"metadata": {
							"guid": "app-guid-4",
							"updated_at": null
						},
						"entity": {
							"name": "app-name-4"
						}
					}
				]
			}`
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/v2/apps", "q=space_guid:some-space-guid"),
					RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
				),
			)
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/v2/apps", "q=space_guid:some-space-guid&page=2"),
					RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
				),
			)
		})

		Context("when apps exist", func() {
			It("returns all the queried apps", func() {
				apps, warnings, err := client.GetApplications(QQuery{
					Filter:   SpaceGUIDFilter,
					Operator: EqualOperator,
					Values:   []string{"some-space-guid"},
				})
				Expect(err).NotTo(HaveOccurred())

				updatedAt, err := time.Parse(time.RFC3339, "2015-03-10T23:11:54Z")
				Expect(err).NotTo(HaveOccurred())

				Expect(apps).To(ConsistOf([]Application{
					{
						Buildpack:               types.FilteredString{IsSet: true, Value: "ruby 1.6.29"},
						DetectedBuildpack:       types.FilteredString{},
						DetectedStartCommand:    types.FilteredString{IsSet: true, Value: "echo 'I am a banana'"},
						DiskQuota:               types.NullByteSizeInMb{IsSet: true, Value: 586},
						GUID:                    "app-guid-1",
						HealthCheckType:         "port",
						HealthCheckHTTPEndpoint: "/",
						Instances:               types.NullInt{Value: 13, IsSet: true},
						Memory:                  types.NullByteSizeInMb{IsSet: true, Value: 1024},
						Name:                    "app-name-1",
						PackageState:            ApplicationPackageFailed,
						PackageUpdatedAt:        updatedAt,
						StackGUID:               "some-stack-guid",
						StagingFailedReason:     "some-reason",
						State:                   ApplicationStopped,
					},
					{
						Name:              "app-name-2",
						GUID:              "app-guid-2",
						DetectedBuildpack: types.FilteredString{IsSet: true, Value: "ruby 1.6.29"},
					},
					{Name: "app-name-3", GUID: "app-guid-3"},
					{Name: "app-name-4", GUID: "app-guid-4"},
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning", "this is another warning"}))
			})
		})
	})

	Describe("UpdateApplication", func() {
		Context("when the update is successful", func() {
			Context("when updating all fields", func() { //are we encoding everything correctly?
				BeforeEach(func() {
					response1 := `{
				"metadata": {
					"guid": "some-app-guid",
					"updated_at": null
				},
				"entity": {
					"detected_start_command": "echo 'I am a banana'",
					"disk_quota": 586,
					"detected_buildpack": null,
					"docker_credentials": {
						"username": "docker-username",
						"password": "docker-password"
					},
					"docker_image": "some-docker-path",
					"environment_json": {
						"key1": "val1",
						"key2": 83493475092347,
						"key3": true,
						"key4": 75821.521
					},
					"health_check_timeout": 120,
					"health_check_type": "some-health-check-type",
					"health_check_http_endpoint": "/anything",
					"instances": 0,
					"memory": 1024,
					"name": "app-name-1",
					"package_updated_at": "2015-03-10T23:11:54Z",
					"stack_guid": "some-stack-guid",
					"state": "STARTED"
				}
			}`
					expectedBody := map[string]interface{}{
						"buildpack":  "",
						"command":    "",
						"disk_quota": 0,
						"docker_credentials": map[string]string{
							"username": "docker-username",
							"password": "docker-password",
						},
						"docker_image": "some-docker-path",
						"environment_json": map[string]string{
							"key1": "val1",
							"key2": "83493475092347",
							"key3": "true",
							"key4": "75821.521",
						},
						"health_check_http_endpoint": "/anything",
						"health_check_type":          "some-health-check-type",
						"instances":                  0,
						"memory":                     0,
						"stack_guid":                 "some-stack-guid",
						"state":                      "STARTED",
					}

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPut, "/v2/apps/some-app-guid"),
							VerifyJSONRepresenting(expectedBody),
							RespondWith(http.StatusCreated, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("returns the updated object and warnings and sends all updated field", func() {
					app, warnings, err := client.UpdateApplication(Application{
						Buildpack: types.FilteredString{IsSet: true, Value: ""},
						Command:   types.FilteredString{IsSet: true, Value: ""},
						DiskQuota: types.NullByteSizeInMb{IsSet: true},
						DockerCredentials: DockerCredentials{
							Username: "docker-username",
							Password: "docker-password",
						},
						DockerImage: "some-docker-path",
						EnvironmentVariables: map[string]string{
							"key1": "val1",
							"key2": "83493475092347",
							"key3": "true",
							"key4": "75821.521",
						},
						GUID: "some-app-guid",
						HealthCheckHTTPEndpoint: "/anything",
						HealthCheckType:         "some-health-check-type",
						Instances:               types.NullInt{Value: 0, IsSet: true},
						Memory:                  types.NullByteSizeInMb{IsSet: true},
						StackGUID:               "some-stack-guid",
						State:                   ApplicationStarted,
					})
					Expect(err).NotTo(HaveOccurred())

					updatedAt, err := time.Parse(time.RFC3339, "2015-03-10T23:11:54Z")
					Expect(err).NotTo(HaveOccurred())

					Expect(app).To(Equal(Application{
						DetectedBuildpack:    types.FilteredString{},
						DetectedStartCommand: types.FilteredString{IsSet: true, Value: "echo 'I am a banana'"},
						DiskQuota:            types.NullByteSizeInMb{IsSet: true, Value: 586},
						DockerCredentials: DockerCredentials{
							Username: "docker-username",
							Password: "docker-password",
						},
						DockerImage: "some-docker-path",
						EnvironmentVariables: map[string]string{
							"key1": "val1",
							"key2": "83493475092347",
							"key3": "true",
							"key4": "75821.521",
						},
						GUID: "some-app-guid",
						HealthCheckHTTPEndpoint: "/anything",
						HealthCheckTimeout:      120,
						HealthCheckType:         "some-health-check-type",
						Instances:               types.NullInt{Value: 0, IsSet: true},
						Memory:                  types.NullByteSizeInMb{IsSet: true, Value: 1024},
						Name:                    "app-name-1",
						PackageUpdatedAt:        updatedAt,
						StackGUID:               "some-stack-guid",
						State:                   ApplicationStarted,
					}))
					Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				})
			})

			Context("when only updating one field", func() { // are we **only** encoding the things we want
				BeforeEach(func() {
					response1 := `{
				"metadata": {
					"guid": "some-app-guid",
					"updated_at": null
				},
				"entity": {
					"buildpack": "ruby 1.6.29",
					"detected_start_command": "echo 'I am a banana'",
					"disk_quota": 586,
					"detected_buildpack": null,
					"health_check_type": "some-health-check-type",
					"health_check_http_endpoint": "/",
					"instances": 7,
					"memory": 1024,
					"name": "app-name-1",
					"package_updated_at": "2015-03-10T23:11:54Z",
					"stack_guid": "some-stack-guid",
					"state": "STOPPED"
				}
			}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPut, "/v2/apps/some-app-guid"),
							VerifyBody([]byte(`{"instances":7}`)),
							RespondWith(http.StatusCreated, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("returns the updated object and warnings and sends only updated field", func() {
					app, warnings, err := client.UpdateApplication(Application{
						GUID:      "some-app-guid",
						Instances: types.NullInt{IsSet: true, Value: 7},
					})
					Expect(err).NotTo(HaveOccurred())

					updatedAt, err := time.Parse(time.RFC3339, "2015-03-10T23:11:54Z")
					Expect(err).NotTo(HaveOccurred())

					Expect(app).To(Equal(Application{
						Buildpack:               types.FilteredString{IsSet: true, Value: "ruby 1.6.29"},
						DetectedBuildpack:       types.FilteredString{},
						DetectedStartCommand:    types.FilteredString{IsSet: true, Value: "echo 'I am a banana'"},
						DiskQuota:               types.NullByteSizeInMb{IsSet: true, Value: 586},
						GUID:                    "some-app-guid",
						HealthCheckType:         "some-health-check-type",
						HealthCheckHTTPEndpoint: "/",
						Instances:               types.NullInt{Value: 7, IsSet: true},
						Memory:                  types.NullByteSizeInMb{IsSet: true, Value: 1024},
						Name:                    "app-name-1",
						PackageUpdatedAt:        updatedAt,
						StackGUID:               "some-stack-guid",
						State:                   ApplicationStopped,
					}))
					Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				})
			})
		})

		Context("when the update returns an error", func() {
			BeforeEach(func() {
				response := `
{
  "code": 210002,
  "description": "The app could not be found: some-app-guid",
  "error_code": "CF-AppNotFound"
}
			`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPut, "/v2/apps/some-app-guid"),
						// VerifyBody([]byte(`{"health_check_type":"some-health-check-type"}`)),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and warnings", func() {
				_, warnings, err := client.UpdateApplication(Application{
					GUID:            "some-app-guid",
					HealthCheckType: "some-health-check-type",
				})
				Expect(err).To(MatchError(ccerror.ResourceNotFoundError{Message: "The app could not be found: some-app-guid"}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})
	})

	Describe("RestageApplication", func() {
		Context("when the restage is successful", func() {
			BeforeEach(func() {
				response := `{
					"metadata": {
						"guid": "some-app-guid",
						"url": "/v2/apps/some-app-guid"
					},
					"entity": {
						"buildpack": "ruby 1.6.29",
						"detected_start_command": "echo 'I am a banana'",
						"disk_quota": 586,
						"detected_buildpack": null,
						"docker_image": "some-docker-path",
						"health_check_type": "some-health-check-type",
						"health_check_http_endpoint": "/anything",
						"instances": 13,
						"memory": 1024,
						"name": "app-name-1",
						"package_updated_at": "2015-03-10T23:11:54Z",
						"stack_guid": "some-stack-guid",
						"state": "STARTED"
					}
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v2/apps/some-app-guid/restage"),
						RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the updated object and warnings and sends all updated field", func() {
				app, warnings, err := client.RestageApplication(Application{
					DockerImage:             "some-docker-path",
					GUID:                    "some-app-guid",
					HealthCheckType:         "some-health-check-type",
					HealthCheckHTTPEndpoint: "/anything",
					State: ApplicationStarted,
				})
				Expect(err).NotTo(HaveOccurred())

				updatedAt, err := time.Parse(time.RFC3339, "2015-03-10T23:11:54Z")
				Expect(err).NotTo(HaveOccurred())

				Expect(app).To(Equal(Application{
					Buildpack:               types.FilteredString{IsSet: true, Value: "ruby 1.6.29"},
					DetectedBuildpack:       types.FilteredString{},
					DetectedStartCommand:    types.FilteredString{IsSet: true, Value: "echo 'I am a banana'"},
					DiskQuota:               types.NullByteSizeInMb{IsSet: true, Value: 586},
					DockerImage:             "some-docker-path",
					GUID:                    "some-app-guid",
					HealthCheckType:         "some-health-check-type",
					HealthCheckHTTPEndpoint: "/anything",
					Instances:               types.NullInt{Value: 13, IsSet: true},
					Memory:                  types.NullByteSizeInMb{IsSet: true, Value: 1024},
					Name:                    "app-name-1",
					PackageUpdatedAt:        updatedAt,
					StackGUID:               "some-stack-guid",
					State:                   ApplicationStarted,
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})

		Context("when the restage returns an error", func() {
			BeforeEach(func() {
				response := `
{
  "code": 210002,
  "description": "The app could not be found: some-app-guid",
  "error_code": "CF-AppNotFound"
}
			`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v2/apps/some-app-guid/restage"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and warnings", func() {
				_, warnings, err := client.RestageApplication(Application{
					GUID:            "some-app-guid",
					HealthCheckType: "some-health-check-type",
				})
				Expect(err).To(MatchError(ccerror.ResourceNotFoundError{Message: "The app could not be found: some-app-guid"}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})
	})

	Describe("GetRouteApplications", func() {
		Context("when the route guid is not found", func() {
			BeforeEach(func() {
				response := `
{
  "code": 210002,
  "description": "The route could not be found: some-route-guid",
  "error_code": "CF-RouteNotFound"
}
			`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/routes/some-route-guid/apps"),
						RespondWith(http.StatusNotFound, response),
					),
				)
			})

			It("returns an error", func() {
				_, _, err := client.GetRouteApplications("some-route-guid")
				Expect(err).To(MatchError(ccerror.ResourceNotFoundError{
					Message: "The route could not be found: some-route-guid",
				}))
			})
		})

		Context("when there are applications associated with this route", func() {
			BeforeEach(func() {
				response1 := `{
				"next_url": "/v2/routes/some-route-guid/apps?q=space_guid:some-space-guid&page=2",
				"resources": [
					{
						"metadata": {
							"guid": "app-guid-1",
							"updated_at": null
						},
						"entity": {
							"name": "app-name-1"
						}
					},
					{
						"metadata": {
							"guid": "app-guid-2",
							"updated_at": null
						},
						"entity": {
							"name": "app-name-2"
						}
					}
				]
			}`
				response2 := `{
				"next_url": null,
				"resources": [
					{
						"metadata": {
							"guid": "app-guid-3",
							"updated_at": null
						},
						"entity": {
							"name": "app-name-3"
						}
					},
					{
						"metadata": {
							"guid": "app-guid-4",
							"updated_at": null
						},
						"entity": {
							"name": "app-name-4"
						}
					}
				]
			}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/routes/some-route-guid/apps", "q=space_guid:some-space-guid"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/routes/some-route-guid/apps", "q=space_guid:some-space-guid&page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)
			})

			It("returns all the applications and all warnings", func() {
				apps, warnings, err := client.GetRouteApplications("some-route-guid", QQuery{
					Filter:   SpaceGUIDFilter,
					Operator: EqualOperator,
					Values:   []string{"some-space-guid"},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(apps).To(ConsistOf([]Application{
					{Name: "app-name-1", GUID: "app-guid-1"},
					{Name: "app-name-2", GUID: "app-guid-2"},
					{Name: "app-name-3", GUID: "app-guid-3"},
					{Name: "app-name-4", GUID: "app-guid-4"},
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning", "this is another warning"}))
			})
		})

		Context("when there are no applications associated with this route", func() {
			BeforeEach(func() {
				response := `{
				"next_url": "",
				"resources": []
			}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/routes/some-route-guid/apps"),
						RespondWith(http.StatusOK, response),
					),
				)
			})

			It("returns an empty list of applications", func() {
				apps, _, err := client.GetRouteApplications("some-route-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(apps).To(BeEmpty())
			})
		})
	})
})

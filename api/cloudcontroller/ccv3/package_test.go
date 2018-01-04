package ccv3_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Package", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetPackage", func() {
		Context("when the package exist", func() {
			BeforeEach(func() {
				response := `{
  "guid": "some-pkg-guid",
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
						VerifyRequest(http.MethodGet, "/v3/packages/some-pkg-guid"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the queried packages and all warnings", func() {
				pkg, warnings, err := client.GetPackage("some-pkg-guid")
				Expect(err).NotTo(HaveOccurred())

				expectedPackage := Package{
					GUID:  "some-pkg-guid",
					State: constant.PackageProcessingUpload,
					Links: map[string]APILink{
						"upload": APILink{HREF: "some-package-upload-url", Method: http.MethodPost},
					},
				}
				Expect(pkg).To(Equal(expectedPackage))
				Expect(warnings).To(ConsistOf("this is a warning"))
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
      "detail": "Package not found",
      "title": "CF-ResourceNotFound"
    }
  ]
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/packages/some-pkg-guid"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				_, warnings, err := client.GetPackage("some-pkg-guid")
				Expect(err).To(MatchError(ccerror.V3UnexpectedResponseError{
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
								Detail: "Package not found",
								Title:  "CF-ResourceNotFound",
							},
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("CreatePackage", func() {
		Context("when the package successfully is created", func() {
			Context("when creating a docker package", func() {
				BeforeEach(func() {
					response := `{
					"data": {
						"image": "some-docker-image",
						"username": "some-username",
						"password": "some-password"
					},
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

					expectedBody := map[string]interface{}{
						"type": "docker",
						"data": map[string]string{
							"image":    "some-docker-image",
							"username": "some-username",
							"password": "some-password",
						},
						"relationships": map[string]interface{}{
							"app": map[string]interface{}{
								"data": map[string]string{
									"guid": "some-app-guid",
								},
							},
						},
					}
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPost, "/v3/packages"),
							VerifyJSONRepresenting(expectedBody),
							RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("returns the created package and warnings", func() {
					pkg, warnings, err := client.CreatePackage(Package{
						Type: constant.PackageTypeDocker,
						Relationships: Relationships{
							ApplicationRelationship: Relationship{GUID: "some-app-guid"},
						},
						DockerImage:    "some-docker-image",
						DockerUsername: "some-username",
						DockerPassword: "some-password",
					})

					Expect(err).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("this is a warning"))

					expectedPackage := Package{
						GUID:  "some-pkg-guid",
						Type:  constant.PackageTypeDocker,
						State: constant.PackageProcessingUpload,
						Links: map[string]APILink{
							"upload": APILink{HREF: "some-package-upload-url", Method: http.MethodPost},
						},
						DockerImage:    "some-docker-image",
						DockerUsername: "some-username",
						DockerPassword: "some-password",
					}
					Expect(pkg).To(Equal(expectedPackage))
				})
			})
			Context("when creating a bits package", func() {
				BeforeEach(func() {
					response := `{
					"guid": "some-pkg-guid",
					"type": "bits",
					"state": "PROCESSING_UPLOAD",
					"links": {
						"upload": {
							"href": "some-package-upload-url",
							"method": "POST"
						}
					}
				}`

					expectedBody := map[string]interface{}{
						"type": "bits",
						"relationships": map[string]interface{}{
							"app": map[string]interface{}{
								"data": map[string]string{
									"guid": "some-app-guid",
								},
							},
						},
					}
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPost, "/v3/packages"),
							VerifyJSONRepresenting(expectedBody),
							RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("omits data, and returns the created package and warnings", func() {
					pkg, warnings, err := client.CreatePackage(Package{
						Type: constant.PackageTypeBits,
						Relationships: Relationships{
							ApplicationRelationship: Relationship{GUID: "some-app-guid"},
						},
					})

					Expect(err).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("this is a warning"))

					expectedPackage := Package{
						GUID:  "some-pkg-guid",
						Type:  constant.PackageTypeBits,
						State: constant.PackageProcessingUpload,
						Links: map[string]APILink{
							"upload": APILink{HREF: "some-package-upload-url", Method: http.MethodPost},
						},
					}
					Expect(pkg).To(Equal(expectedPackage))
				})
			})
		})

		Context("when cc returns back an error or warnings", func() {
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
      "detail": "Package not found",
      "title": "CF-ResourceNotFound"
    }
  ]
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/packages"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				_, warnings, err := client.CreatePackage(Package{})
				Expect(err).To(MatchError(ccerror.V3UnexpectedResponseError{
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
								Detail: "Package not found",
								Title:  "CF-ResourceNotFound",
							},
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("UploadPackage", func() {
		Context("when the package successfully is created", func() {
			var tempFile *os.File

			BeforeEach(func() {
				var err error
				tempFile, err = ioutil.TempFile("", "package-upload")
				Expect(err).ToNot(HaveOccurred())
				defer tempFile.Close()

				fileSize := 1024
				contents := strings.Repeat("A", fileSize)
				err = ioutil.WriteFile(tempFile.Name(), []byte(contents), 0666)
				Expect(err).NotTo(HaveOccurred())

				verifyHeaderAndBody := func(_ http.ResponseWriter, req *http.Request) {
					contentType := req.Header.Get("Content-Type")
					Expect(contentType).To(MatchRegexp("multipart/form-data; boundary=[\\w\\d]+"))

					boundary := contentType[30:]

					defer req.Body.Close()
					rawBody, err := ioutil.ReadAll(req.Body)
					Expect(err).NotTo(HaveOccurred())
					body := BufferWithBytes(rawBody)
					Expect(body).To(Say("--%s", boundary))
					Expect(body).To(Say(`name="bits"`))
					Expect(body).To(Say(contents))
					Expect(body).To(Say("--%s--", boundary))
				}

				response := `{
					"guid": "some-pkg-guid",
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
						VerifyRequest(http.MethodPost, "/v3/my-special-endpoint/some-pkg-guid/upload"),
						verifyHeaderAndBody,
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			AfterEach(func() {
				if tempFile != nil {
					Expect(os.Remove(tempFile.Name())).ToNot(HaveOccurred())
				}
			})

			It("returns the created package and warnings", func() {
				pkg, warnings, err := client.UploadPackage(Package{
					State: constant.PackageAwaitingUpload,
					Links: map[string]APILink{
						"upload": APILink{
							HREF:   fmt.Sprintf("%s/v3/my-special-endpoint/some-pkg-guid/upload", server.URL()),
							Method: http.MethodPost,
						},
					},
				}, tempFile.Name())

				Expect(err).NotTo(HaveOccurred())

				expectedPackage := Package{
					GUID:  "some-pkg-guid",
					State: constant.PackageProcessingUpload,
					Links: map[string]APILink{
						"upload": APILink{HREF: "some-package-upload-url", Method: http.MethodPost},
					},
				}
				Expect(pkg).To(Equal(expectedPackage))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		Context("when the package does not have an upload link", func() {
			It("returns an UploadLinkNotFoundError", func() {
				_, _, err := client.UploadPackage(Package{GUID: "some-pkg-guid", State: constant.PackageAwaitingUpload}, "/path/to/foo")
				Expect(err).To(MatchError(ccerror.UploadLinkNotFoundError{PackageGUID: "some-pkg-guid"}))
			})
		})
	})

	Describe("GetPackages", func() {
		Context("when cloud controller returns list of packages", func() {
			BeforeEach(func() {
				response := `{
					"resources": [
					  {
						  "guid": "some-pkg-guid-1",
							"type": "bits",
						  "state": "PROCESSING_UPLOAD",
							"created_at": "2017-08-14T21:16:12Z",
							"links": {
								"upload": {
									"href": "some-pkg-upload-url-1",
									"method": "POST"
								}
							}
					  },
					  {
						  "guid": "some-pkg-guid-2",
							"type": "bits",
						  "state": "READY",
							"created_at": "2017-08-14T21:20:13Z",
							"links": {
								"upload": {
									"href": "some-pkg-upload-url-2",
									"method": "POST"
								}
							}
					  }
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/packages", "app_guids=some-app-guid"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the queried packages and all warnings", func() {
				packages, warnings, err := client.GetPackages(Query{Key: AppGUIDFilter, Values: []string{"some-app-guid"}})
				Expect(err).NotTo(HaveOccurred())

				Expect(packages).To(Equal([]Package{
					{
						GUID:      "some-pkg-guid-1",
						Type:      constant.PackageTypeBits,
						State:     constant.PackageProcessingUpload,
						CreatedAt: "2017-08-14T21:16:12Z",
						Links: map[string]APILink{
							"upload": APILink{HREF: "some-pkg-upload-url-1", Method: http.MethodPost},
						},
					},
					{
						GUID:      "some-pkg-guid-2",
						Type:      constant.PackageTypeBits,
						State:     constant.PackageReady,
						CreatedAt: "2017-08-14T21:20:13Z",
						Links: map[string]APILink{
							"upload": APILink{HREF: "some-pkg-upload-url-2", Method: http.MethodPost},
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
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
							"detail": "Package not found",
							"title": "CF-ResourceNotFound"
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/packages", "app_guids=some-app-guid"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				_, warnings, err := client.GetPackages(Query{Key: AppGUIDFilter, Values: []string{"some-app-guid"}})
				Expect(err).To(MatchError(ccerror.V3UnexpectedResponseError{
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
								Detail: "Package not found",
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

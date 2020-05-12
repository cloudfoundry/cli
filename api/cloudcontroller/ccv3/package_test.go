package ccv3_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/ccv3fakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/wrapper"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Package", func() {
	var client *Client

	BeforeEach(func() {
		client, _ = NewTestClient()
	})

	Describe("CreatePackage", func() {
		var (
			inputPackage Package

			pkg        Package
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			pkg, warnings, executeErr = client.CreatePackage(inputPackage)
		})

		When("the package successfully is created", func() {
			When("creating a docker package", func() {
				BeforeEach(func() {
					inputPackage = Package{
						Type: constant.PackageTypeDocker,
						Relationships: resources.Relationships{
							constant.RelationshipTypeApplication: resources.Relationship{GUID: "some-app-guid"},
						},
						DockerImage:    "some-docker-image",
						DockerUsername: "some-username",
						DockerPassword: "some-password",
					}

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
					Expect(executeErr).NotTo(HaveOccurred())
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

			When("creating a bits package", func() {
				BeforeEach(func() {
					inputPackage = Package{
						Type: constant.PackageTypeBits,
						Relationships: resources.Relationships{
							constant.RelationshipTypeApplication: resources.Relationship{GUID: "some-app-guid"},
						},
					}
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
					Expect(executeErr).NotTo(HaveOccurred())
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

		When("cc returns back an error or warnings", func() {
			BeforeEach(func() {
				inputPackage = Package{}
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
							Detail: "Package not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("GetPackage", func() {
		var (
			pkg        Package
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			pkg, warnings, executeErr = client.GetPackage("some-pkg-guid")
		})

		When("the package exists", func() {
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

			It("returns the queried package and all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())

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
							Detail: "Package not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("GetPackages", func() {
		var (
			pkgs       []Package
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			pkgs, warnings, executeErr = client.GetPackages(Query{Key: AppGUIDFilter, Values: []string{"some-app-guid"}})
		})

		When("cloud controller returns list of packages", func() {
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
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(pkgs).To(Equal([]Package{
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
							Detail: "Package not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("UploadBitsPackage", func() {
		var (
			inputPackage Package
		)

		BeforeEach(func() {
			client, _ = NewTestClient()

			inputPackage = Package{
				GUID: "package-guid",
			}
		})

		When("the upload is successful", func() {
			var (
				resources           []Resource
				readerBody          []byte
				verifyHeaderAndBody func(http.ResponseWriter, *http.Request)
			)

			BeforeEach(func() {
				resources = []Resource{
					{FilePath: "foo"},
					{FilePath: "bar"},
				}

				response := `{
						"guid": "some-package-guid",
						"type": "bits",
						"state": "PROCESSING_UPLOAD"
					}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/packages/package-guid/upload"),
						func(writer http.ResponseWriter, req *http.Request) {
							verifyHeaderAndBody(writer, req)
						},
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			When("the upload has application bits to upload", func() {
				var reader io.Reader

				BeforeEach(func() {
					readerBody = []byte("hello world")
					reader = bytes.NewReader(readerBody)

					verifyHeaderAndBody = func(_ http.ResponseWriter, req *http.Request) {
						contentType := req.Header.Get("Content-Type")
						Expect(contentType).To(MatchRegexp("multipart/form-data; boundary=[\\w\\d]+"))

						defer req.Body.Close()
						requestReader := multipart.NewReader(req.Body, contentType[30:])

						// Verify that matched resources are sent properly
						resourcesPart, err := requestReader.NextPart()
						Expect(err).NotTo(HaveOccurred())

						Expect(resourcesPart.FormName()).To(Equal("resources"))

						defer resourcesPart.Close()
						expectedJSON, err := json.Marshal(resources)
						Expect(err).NotTo(HaveOccurred())
						Expect(ioutil.ReadAll(resourcesPart)).To(MatchJSON(expectedJSON))

						// Verify that the application bits are sent properly
						resourcesPart, err = requestReader.NextPart()
						Expect(err).NotTo(HaveOccurred())

						Expect(resourcesPart.FormName()).To(Equal("bits"))
						Expect(resourcesPart.FileName()).To(Equal("package.zip"))

						defer resourcesPart.Close()
						Expect(ioutil.ReadAll(resourcesPart)).To(Equal(readerBody))
					}
				})

				It("returns the created job and warnings", func() {
					pkg, warnings, err := client.UploadBitsPackage(inputPackage, resources, reader, int64(len(readerBody)))
					Expect(err).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("this is a warning"))
					Expect(pkg).To(Equal(Package{
						GUID:  "some-package-guid",
						Type:  constant.PackageTypeBits,
						State: constant.PackageProcessingUpload,
					}))
				})
			})

			When("there are no application bits to upload", func() {
				BeforeEach(func() {
					verifyHeaderAndBody = func(_ http.ResponseWriter, req *http.Request) {
						contentType := req.Header.Get("Content-Type")
						Expect(contentType).To(MatchRegexp("multipart/form-data; boundary=[\\w\\d]+"))

						defer req.Body.Close()
						requestReader := multipart.NewReader(req.Body, contentType[30:])

						// Verify that matched resources are sent properly
						resourcesPart, err := requestReader.NextPart()
						Expect(err).NotTo(HaveOccurred())

						Expect(resourcesPart.FormName()).To(Equal("resources"))

						defer resourcesPart.Close()
						expectedJSON, err := json.Marshal(resources)
						Expect(err).NotTo(HaveOccurred())
						Expect(ioutil.ReadAll(resourcesPart)).To(MatchJSON(expectedJSON))

						// Verify that the application bits are not sent
						_, err = requestReader.NextPart()
						Expect(err).To(MatchError(io.EOF))
					}
				})

				It("does not send the application bits", func() {
					pkg, warnings, err := client.UploadBitsPackage(inputPackage, resources, nil, 33513531353)
					Expect(err).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("this is a warning"))
					Expect(pkg).To(Equal(Package{
						GUID:  "some-package-guid",
						Type:  constant.PackageTypeBits,
						State: constant.PackageProcessingUpload,
					}))
				})
			})
		})

		When("the CC returns an error", func() {
			BeforeEach(func() {
				response := ` {
					"errors": [
						{
							"code": 10008,
							"detail": "Banana",
							"title": "CF-Banana"
						}
					]
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/packages/package-guid/upload"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error", func() {
				_, warnings, err := client.UploadBitsPackage(inputPackage, []Resource{}, bytes.NewReader(nil), 0)
				Expect(err).To(MatchError(ccerror.ResourceNotFoundError{Message: "Banana"}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		When("passed a nil resources", func() {
			It("returns a NilObjectError", func() {
				_, _, err := client.UploadBitsPackage(inputPackage, nil, bytes.NewReader(nil), 0)
				Expect(err).To(MatchError(ccerror.NilObjectError{Object: "matchedResources"}))
			})
		})

		When("an error is returned from the new resources reader", func() {
			var (
				fakeReader  *ccv3fakes.FakeReader
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("some read error")
				fakeReader = new(ccv3fakes.FakeReader)
				fakeReader.ReadReturns(0, expectedErr)

				server.AppendHandlers(
					VerifyRequest(http.MethodPost, "/v3/packages/package-guid/upload"),
				)
			})

			It("returns the error", func() {
				_, _, err := client.UploadBitsPackage(inputPackage, []Resource{}, fakeReader, 3)
				Expect(err).To(MatchError(expectedErr))
			})
		})

		When("a retryable error occurs", func() {
			BeforeEach(func() {
				wrapper := &wrapper.CustomWrapper{
					CustomMake: func(connection cloudcontroller.Connection, request *cloudcontroller.Request, response *cloudcontroller.Response) error {
						defer GinkgoRecover() // Since this will be running in a thread

						if strings.HasSuffix(request.URL.String(), "/v3/packages/package-guid/upload") {
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
				_, _, err := client.UploadBitsPackage(inputPackage, []Resource{}, strings.NewReader("hello world"), 3)
				Expect(err).To(MatchError(ccerror.PipeSeekError{}))
			})
		})

		When("an http error occurs mid-transfer", func() {
			var expectedErr error
			const UploadSize = 33 * 1024

			BeforeEach(func() {
				expectedErr = errors.New("some read error")

				wrapper := &wrapper.CustomWrapper{
					CustomMake: func(connection cloudcontroller.Connection, request *cloudcontroller.Request, response *cloudcontroller.Response) error {
						defer GinkgoRecover() // Since this will be running in a thread

						if strings.HasSuffix(request.URL.String(), "/v3/packages/package-guid/upload") {
							defer request.Body.Close()
							readBytes, err := ioutil.ReadAll(request.Body)
							Expect(err).ToNot(HaveOccurred())
							Expect(len(readBytes)).To(BeNumerically(">", UploadSize))
							return expectedErr
						}
						return connection.Make(request, response)
					},
				}

				client, _ = NewTestClient(Config{Wrappers: []ConnectionWrapper{wrapper}})
			})

			It("returns the http error", func() {
				_, _, err := client.UploadBitsPackage(inputPackage, []Resource{}, strings.NewReader(strings.Repeat("a", UploadSize)), 3)
				Expect(err).To(MatchError(expectedErr))
			})
		})
	})

	Describe("UploadPackage", func() {
		var (
			inputPackage Package
			fileToUpload string

			pkg        Package
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			pkg, warnings, executeErr = client.UploadPackage(inputPackage, fileToUpload)
		})

		When("the package successfully is created", func() {
			var tempFile *os.File

			BeforeEach(func() {
				var err error

				inputPackage = Package{
					State: constant.PackageAwaitingUpload,
					GUID:  "package-guid",
				}

				tempFile, err = ioutil.TempFile("", "package-upload")
				Expect(err).ToNot(HaveOccurred())
				defer tempFile.Close()

				fileToUpload = tempFile.Name()

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
						VerifyRequest(http.MethodPost, "/v3/packages/package-guid/upload"),
						verifyHeaderAndBody,
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			AfterEach(func() {
				if tempFile != nil {
					Expect(os.RemoveAll(tempFile.Name())).ToNot(HaveOccurred())
				}
			})

			It("returns the created package and warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())

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

		When("cc returns back an error or warnings", func() {
			var tempFile *os.File

			BeforeEach(func() {
				var err error

				inputPackage = Package{
					GUID:  "package-guid",
					State: constant.PackageAwaitingUpload,
				}

				tempFile, err = ioutil.TempFile("", "package-upload")
				Expect(err).ToNot(HaveOccurred())
				defer tempFile.Close()

				fileToUpload = tempFile.Name()

				fileSize := 1024
				contents := strings.Repeat("A", fileSize)
				err = ioutil.WriteFile(tempFile.Name(), []byte(contents), 0666)
				Expect(err).NotTo(HaveOccurred())

				response := ` {
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
						VerifyRequest(http.MethodPost, "/v3/packages/package-guid/upload"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			AfterEach(func() {
				if tempFile != nil {
					Expect(os.RemoveAll(tempFile.Name())).ToNot(HaveOccurred())
				}
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

	Describe("CopyPackage", func() {
		var (
			sourcePackageGUID string
			targetAppGUID     string

			targetPackage Package
			warnings      Warnings
			executeErr    error
			response      string
		)

		BeforeEach(func() {
			sourcePackageGUID = "source-package-guid"

			targetAppGUID = "target-app-guid"
			response = `{
					"guid": "some-targetPackage-guid"
				}`

			expectedBody := map[string]interface{}{
				"relationships": map[string]interface{}{
					"app": map[string]interface{}{
						"data": map[string]string{
							"guid": targetAppGUID,
						},
					},
				},
			}
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/v3/packages", "source_guid="+sourcePackageGUID),
					VerifyJSONRepresenting(expectedBody),
					RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
				),
			)

		})

		JustBeforeEach(func() {
			targetPackage, warnings, executeErr = client.CopyPackage(sourcePackageGUID, targetAppGUID)
		})

		It("returns the created target package and warnings", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(warnings).To(ConsistOf("this is a warning"))

			expectedPackage := Package{
				GUID: "some-targetPackage-guid",
			}
			Expect(targetPackage).To(Equal(expectedPackage))
		})

		When("cc returns back an error or warnings", func() {
			BeforeEach(func() {
				response = ` {
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
				server.Reset()
				time.Sleep(10 * time.Millisecond) // guards against <ccerror.RequestError>: {Err: { ... { Err: {s: "EOF"},
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/packages", "source_guid="+sourcePackageGUID),
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
							Detail: "Package not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})
})

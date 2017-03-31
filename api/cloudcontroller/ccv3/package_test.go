package ccv3_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
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
					State: PackageStateProcessingUpload,
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
				Expect(err).To(MatchError(UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					CCErrorResponse: CCErrorResponse{
						[]CCError{
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

			It("returns the created package and warnings", func() {
				pkg, warnings, err := client.CreatePackage(Package{
					Type: PackageTypeBits,
					Relationships: PackageRelationships{
						Application: Relationship{GUID: "some-app-guid"},
					},
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))

				expectedPackage := Package{
					GUID:  "some-pkg-guid",
					Type:  PackageTypeBits,
					State: PackageStateProcessingUpload,
					Links: map[string]APILink{
						"upload": APILink{HREF: "some-package-upload-url", Method: http.MethodPost},
					},
				}
				Expect(pkg).To(Equal(expectedPackage))
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
				Expect(err).To(MatchError(UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					CCErrorResponse: CCErrorResponse{
						[]CCError{
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
					os.Remove(tempFile.Name())
				}
			})

			It("returns the created package and warnings", func() {
				pkg, warnings, err := client.UploadPackage(Package{
					State: PackageStateAwaitingUpload,
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
					State: PackageStateProcessingUpload,
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
				_, _, err := client.UploadPackage(Package{GUID: "some-pkg-guid", State: PackageStateAwaitingUpload}, "/path/to/foo")
				Expect(err).To(MatchError(UploadLinkNotFoundError{PackageGUID: "some-pkg-guid"}))
			})
		})
	})
})

package ccv3_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"

	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Metadata", func() {
	var client *Client

	BeforeEach(func() {
		client, _ = NewTestClient()
	})

	Describe("UpdateResourceMetadata", func() {
		var (
			metadataToUpdate Metadata
			resourceGUID     string
			updatedMetadata  ResourceMetadata
			warnings         Warnings
			executeErr       error
		)

		When("updating metadata on an app", func() {
			JustBeforeEach(func() {
				resourceGUID = "some-guid"
				updatedMetadata, warnings, executeErr = client.UpdateResourceMetadata("app", resourceGUID, metadataToUpdate)
			})

			When("the space is updated successfully", func() {
				BeforeEach(func() {
					response := `{
					"guid": "some-guid",
					"name": "some-space-name",
					"metadata": {
						"labels": {
							"k1":"v1",
							"k2":"v2"
						}
					}
				}`

					expectedBody := map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]string{
								"k1": "v1",
								"k2": "v2",
							},
						},
					}

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPatch, "/v3/apps/some-guid"),
							VerifyJSONRepresenting(expectedBody),
							RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)

					metadataToUpdate = Metadata{
						Labels: map[string]types.NullString{
							"k1": types.NewNullString("v1"),
							"k2": types.NewNullString("v2"),
						},
					}
				})

				It("should include the labels in the JSON", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(server.ReceivedRequests()).To(HaveLen(3))
					Expect(len(warnings)).To(Equal(0))
					Expect(updatedMetadata.Metadata.Labels).To(BeEquivalentTo(
						map[string]types.NullString{
							"k1": types.NewNullString("v1"),
							"k2": types.NewNullString("v2"),
						}))
				})

			})

			When("Cloud Controller returns errors and warnings", func() {
				BeforeEach(func() {
					response := `{
  "errors": [
    {
      "code": 10008,
      "detail": "Metadata key error: label 'invalid*key' contains invalid characters",
      "title": "CF-UnprocessableEntity"
    }
  ]
}`
					expectedBody := map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]string{
								"invalid*key": "v1",
								"k2":          "v2",
							},
						},
					}

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPatch, "/v3/apps/some-guid"),
							VerifyJSONRepresenting(expectedBody),
							RespondWith(http.StatusUnprocessableEntity, response, http.Header{}),
						),
					)

					metadataToUpdate = Metadata{
						Labels: map[string]types.NullString{
							"invalid*key": types.NewNullString("v1"),
							"k2":          types.NewNullString("v2"),
						},
					}
				})

				It("returns the error and all warnings", func() {
					Expect(executeErr).To(MatchError(ccerror.UnprocessableEntityError{
						Message: "Metadata key error: label 'invalid*key' contains invalid characters",
					}))
				})
			})
		})

		When("updating metadata on an organization", func() {
			JustBeforeEach(func() {
				resourceGUID = "some-guid"
				updatedMetadata, warnings, executeErr = client.UpdateResourceMetadata("org", resourceGUID, metadataToUpdate)
			})

			When("the organization is updated successfully", func() {
				BeforeEach(func() {
					response := `{
					"guid": "some-guid",
					"name": "some-org-name",
					"metadata": {
						"labels": {
							"k1":"v1",
							"k2":"v2"
						}
					}
				}`

					expectedBody := map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]string{
								"k1": "v1",
								"k2": "v2",
							},
						},
					}

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPatch, "/v3/organizations/some-guid"),
							VerifyJSONRepresenting(expectedBody),
							RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)

					metadataToUpdate = Metadata{
						Labels: map[string]types.NullString{
							"k1": types.NewNullString("v1"),
							"k2": types.NewNullString("v2"),
						},
					}
				})

				It("should include the labels in the JSON", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(server.ReceivedRequests()).To(HaveLen(3))
					Expect(len(warnings)).To(Equal(0))
					Expect(updatedMetadata.Metadata.Labels).To(BeEquivalentTo(
						map[string]types.NullString{
							"k1": types.NewNullString("v1"),
							"k2": types.NewNullString("v2"),
						}))
				})

			})

			When("Cloud Controller returns errors and warnings", func() {
				BeforeEach(func() {
					response := `{
  "errors": [
    {
      "code": 10008,
      "detail": "Metadata key error: label 'invalid*key' contains invalid characters",
      "title": "CF-UnprocessableEntity"
    }
  ]
}`
					expectedBody := map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]string{
								"invalid*key": "v1",
								"k2":          "v2",
							},
						},
					}

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPatch, "/v3/organizations/some-guid"),
							VerifyJSONRepresenting(expectedBody),
							RespondWith(http.StatusUnprocessableEntity, response, http.Header{}),
						),
					)

					metadataToUpdate = Metadata{
						Labels: map[string]types.NullString{
							"invalid*key": types.NewNullString("v1"),
							"k2":          types.NewNullString("v2"),
						},
					}
				})

				It("returns the error and all warnings", func() {
					Expect(executeErr).To(MatchError(ccerror.UnprocessableEntityError{
						Message: "Metadata key error: label 'invalid*key' contains invalid characters",
					}))
				})
			})
		})

		When("updating metadata on a space", func() {
			JustBeforeEach(func() {
				resourceGUID = "some-guid"
				updatedMetadata, warnings, executeErr = client.UpdateResourceMetadata("space", resourceGUID, metadataToUpdate)
			})

			When("the space is updated successfully", func() {
				BeforeEach(func() {
					response := `{
					"guid": "some-guid",
					"name": "some-space-name",
					"metadata": {
						"labels": {
							"k1":"v1",
							"k2":"v2"
						}
					}
				}`

					expectedBody := map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]string{
								"k1": "v1",
								"k2": "v2",
							},
						},
					}

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPatch, "/v3/spaces/some-guid"),
							VerifyJSONRepresenting(expectedBody),
							RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)

					metadataToUpdate = Metadata{
						Labels: map[string]types.NullString{
							"k1": types.NewNullString("v1"),
							"k2": types.NewNullString("v2"),
						},
					}
				})

				It("should include the labels in the JSON", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(server.ReceivedRequests()).To(HaveLen(3))
					Expect(len(warnings)).To(Equal(0))
					Expect(updatedMetadata.Metadata.Labels).To(BeEquivalentTo(
						map[string]types.NullString{
							"k1": types.NewNullString("v1"),
							"k2": types.NewNullString("v2"),
						}))
				})

			})

			When("Cloud Controller returns errors and warnings", func() {
				BeforeEach(func() {
					response := `{
  "errors": [
    {
      "code": 10008,
      "detail": "Metadata key error: label 'invalid*key' contains invalid characters",
      "title": "CF-UnprocessableEntity"
    }
  ]
}`
					expectedBody := map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]string{
								"invalid*key": "v1",
								"k2":          "v2",
							},
						},
					}

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPatch, "/v3/spaces/some-guid"),
							VerifyJSONRepresenting(expectedBody),
							RespondWith(http.StatusUnprocessableEntity, response, http.Header{}),
						),
					)

					metadataToUpdate = Metadata{
						Labels: map[string]types.NullString{
							"invalid*key": types.NewNullString("v1"),
							"k2":          types.NewNullString("v2"),
						},
					}
				})

				It("returns the error and all warnings", func() {
					Expect(executeErr).To(MatchError(ccerror.UnprocessableEntityError{
						Message: "Metadata key error: label 'invalid*key' contains invalid characters",
					}))
				})
			})
		})
	})
})

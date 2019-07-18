package ccv3_test

import (
	"fmt"
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
		testForResourceType := func(resourceType string, resourceTypeNameForURI string) {
			if resourceTypeNameForURI == "" {
				resourceTypeNameForURI = resourceType
			}

			When(fmt.Sprintf("updating metadata on %s", resourceType), func() {

				JustBeforeEach(func() {
					resourceGUID = "some-guid"
					updatedMetadata, warnings, executeErr = client.UpdateResourceMetadata(resourceType, resourceGUID, metadataToUpdate)
				})

				When(fmt.Sprintf("the %s is updated successfully", resourceType), func() {
					BeforeEach(func() {
						response := fmt.Sprintf(`{
							"guid": "some-guid",
							"name": "some-%s-type",
							"metadata": {
								"labels": {
									"k1":"v1",
									"k2":"v2"
								}
							}
						}`, resourceType)

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
								VerifyRequest(http.MethodPatch, fmt.Sprintf("/v3/%ss/some-guid", resourceTypeNameForURI)),
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
      "detail": "Metadata label key error: 'invalid*key' contains invalid characters",
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
								VerifyRequest(http.MethodPatch, fmt.Sprintf("/v3/%ss/some-guid", resourceTypeNameForURI)),
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
							Message: "Metadata label key error: 'invalid*key' contains invalid characters",
						}))
					})
				})
			})
		}

		testForResourceType("app", "")
		testForResourceType("buildpack", "")
		testForResourceType("org", "organization")
		testForResourceType("space", "")
		testForResourceType("stack", "")
	})
})

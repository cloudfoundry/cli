package ccv3_test

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"

	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

const errorResponse = `{
  "errors": [
    {
      "code": 10008,
      "detail": "Metadata label key error: 'invalid*key' contains invalid characters",
      "title": "CF-UnprocessableEntity"
    }
  ]
}`

var _ = Describe("Metadata", func() {
	var client *Client

	BeforeEach(func() {
		client, _ = NewTestClient()
	})

	Describe("UpdateResourceMetadata", func() {
		var (
			metadataToUpdate Metadata
			resourceGUID     string
			jobURL           JobURL
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
					jobURL, warnings, executeErr = client.UpdateResourceMetadata(resourceType, resourceGUID, metadataToUpdate)
				})

				When(fmt.Sprintf("the %s is updated successfully", resourceType), func() {
					BeforeEach(func() {
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
								RespondWith(http.StatusOK, nil, http.Header{
									"X-Cf-Warnings": {"this is a warning"},
									"Location":      {"fake-job-url"},
								}),
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
						Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
						Expect(server.ReceivedRequests()).To(HaveLen(3))
						Expect(jobURL).To(Equal(JobURL("fake-job-url")))
					})
				})

				When("Cloud Controller returns errors and warnings", func() {
					BeforeEach(func() {
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
								RespondWith(http.StatusUnprocessableEntity, errorResponse, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
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
						Expect(warnings).To(ConsistOf(Warnings{"this is another warning"}))
					})
				})
			})
		}

		testForResourceType("app", "")
		testForResourceType("domain", "")
		testForResourceType("buildpack", "")
		testForResourceType("org", "organization")
		testForResourceType("route", "")
		testForResourceType("space", "")
		testForResourceType("stack", "")
		testForResourceType("service-offering", "service_offering")
		testForResourceType("service-plan", "service_plan")
		testForResourceType("service-broker", "service_broker")

		When("updating metadata on an unsupported resource", func() {
			It("returns an error", func() {
				_, _, err := client.UpdateResourceMetadata("anything", "fake-guid", Metadata{})
				Expect(err).To(MatchError("unknown resource type (anything) requested"))
			})
		})
	})
})

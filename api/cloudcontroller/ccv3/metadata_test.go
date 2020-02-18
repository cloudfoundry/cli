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
						Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
						Expect(server.ReceivedRequests()).To(HaveLen(3))
						Expect(updatedMetadata.Metadata.Labels).To(BeEquivalentTo(
							map[string]types.NullString{
								"k1": types.NewNullString("v1"),
								"k2": types.NewNullString("v2"),
							}))
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

		When("updating metadata on an unsupported resource", func() {
			It("returns an error", func() {
				_, _, err := client.UpdateResourceMetadata("anything", "fake-guid", Metadata{})
				Expect(err).To(MatchError("unknown resource type (anything) requested"))
			})
		})
	})

	Describe("UpdateResourceMetadataAsync", func() {
		When("updating metadata on service-broker", func() {
			When("the service-broker is updated successfully", func() {
				It("sends the correct data and returns the job URL", func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPatch, "/v3/service_brokers/some-guid"),
							VerifyJSON(`{"metadata":{"labels":{"k1":"v1","k2":"v2"}}}`),
							RespondWith(http.StatusAccepted, "", http.Header{"X-Cf-Warnings": {"this is a warning"}, "Location": {"fake-job-url"}}),
						),
					)

					metadataToUpdate := Metadata{
						Labels: map[string]types.NullString{
							"k1": types.NewNullString("v1"),
							"k2": types.NewNullString("v2"),
						},
					}

					jobURL, warnings, err := client.UpdateResourceMetadataAsync("service-broker", "some-guid", metadataToUpdate)
					Expect(err).NotTo(HaveOccurred())

					Expect(jobURL).To(BeEquivalentTo("fake-job-url"))
					Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				})
			})

			When("Cloud Controller returns errors and warnings", func() {
				It("returns the error and all warnings", func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPatch, "/v3/service_brokers/some-guid"),
							RespondWith(http.StatusUnprocessableEntity, errorResponse, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
						),
					)

					_, warnings, err := client.UpdateResourceMetadataAsync("service-broker", "some-guid", Metadata{})

					Expect(err).To(MatchError(ccerror.UnprocessableEntityError{
						Message: "Metadata label key error: 'invalid*key' contains invalid characters",
					}))
					Expect(warnings).To(ConsistOf(Warnings{"this is another warning"}))
				})
			})
		})

		When("updating metadata on an unsupported resource", func() {
			It("returns an error", func() {
				_, _, err := client.UpdateResourceMetadataAsync("anything", "fake-guid", Metadata{})
				Expect(err).To(MatchError("unknown async resource type (anything) requested"))
			})
		})
	})
})

package ccv3_test

import (
	"encoding/json"
	"net/http"
	"os"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Resource", func() {
	Describe("V3 formatted resource", func() {
		Describe("MarshalJSON", func() {
			It("marshals the json properly", func() {
				resource := Resource{
					FilePath:    "some-file-1",
					Mode:        os.FileMode(0744),
					Checksum:    Checksum{Value: "some-sha-1"},
					SizeInBytes: 1,
				}
				data, err := json.Marshal(resource)
				Expect(err).ToNot(HaveOccurred())
				Expect(data).To(MatchJSON(`{
				"path":   "some-file-1",
				"mode": "744",
				"checksum": {"value":"some-sha-1"},
				"size_in_bytes": 1
			}`))
			})
		})

		Describe("UnmarshalJSON", func() {
			It("Unmarshals the json properly", func() {
				raw := `{
				"path":   "some-file-1",
				"mode": "744",
				"checksum": {"value":"some-sha-1"},
				"size_in_bytes": 1
			}`

				var data Resource
				err := json.Unmarshal([]byte(raw), &data)
				Expect(err).ToNot(HaveOccurred())
				Expect(data).To(Equal(Resource{
					FilePath:    "some-file-1",
					Mode:        os.FileMode(0744),
					Checksum:    Checksum{Value: "some-sha-1"},
					SizeInBytes: 1,
				}))
			})
		})
	})

	Describe("Resource Match", func() {

		var (
			client           *Client
			allResources     []Resource
			matchedResources []Resource
			warnings         Warnings
			executeErr       error
		)

		BeforeEach(func() {
			client, _ = NewTestClient()
			allResources = []Resource{
				{FilePath: "where are you", Checksum: Checksum{Value: "value1"}, SizeInBytes: 2},
				{FilePath: "bar", Checksum: Checksum{Value: "value2"}, SizeInBytes: 3},
			}
		})

		JustBeforeEach(func() {
			matchedResources, warnings, executeErr = client.ResourceMatch(allResources)
		})

		When("the match is successful", func() {
			When("the upload has application bits to upload", func() {

				BeforeEach(func() {

					expectedBody := `{
  "resources": [
    {
      "path": "where are you",
			"checksum": {
        "value": "value1"
			},
			"size_in_bytes": 2,
			"mode": "0"
    },
    {
      "path": "bar",
			"checksum": {
        "value": "value2"
			},
			"size_in_bytes": 3,
			"mode": "0"
    }
  ]
}`

					response := `{
  "resources": [
    {
      "checksum": {
		    "value":"scooby dooby"
		  },
	    "size_in_bytes": 2,
      "path": "where are you? we got a job for you now",
			"mode": "644"
    }
  ]
}`

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPost, "/v3/resource_matches"),
							VerifyJSON(expectedBody),
							RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("returns the matched resources", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("this is a warning"))
					Expect(matchedResources).To(ConsistOf(Resource{
						Checksum:    Checksum{Value: "scooby dooby"},
						SizeInBytes: 2,
						FilePath:    "where are you? we got a job for you now",
						Mode:        os.FileMode(0644),
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
						VerifyRequest(http.MethodPost, "/v3/resource_matches"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(ccerror.ResourceNotFoundError{Message: "Banana"}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})
})

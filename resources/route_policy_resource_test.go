package resources_test

import (
	"encoding/json"
	"testing"

	"code.cloudfoundry.org/cli/v9/resources"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRoutePolicyResource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RoutePolicy Resource Suite")
}

var _ = Describe("RoutePolicy", func() {
	Describe("MarshalJSON", func() {
		It("marshals the route policy with relationships", func() {
			policy := resources.RoutePolicy{
				Source:    "cf:app:some-app-guid",
				RouteGUID: "some-route-guid",
			}

			data, err := json.Marshal(policy)
			Expect(err).NotTo(HaveOccurred())

			var result map[string]interface{}
			err = json.Unmarshal(data, &result)
			Expect(err).NotTo(HaveOccurred())

			Expect(result["source"]).To(Equal("cf:app:some-app-guid"))
			Expect(result["relationships"]).NotTo(BeNil())

			relationships := result["relationships"].(map[string]interface{})
			route := relationships["route"].(map[string]interface{})
			routeData := route["data"].(map[string]interface{})
			Expect(routeData["guid"]).To(Equal("some-route-guid"))
		})
	})

	Describe("UnmarshalJSON", func() {
		It("unmarshals the route policy from relationships", func() {
			jsonData := `{
				"guid": "some-guid",
				"source": "cf:app:app-guid",
				"relationships": {
					"route": {
						"data": {
							"guid": "route-guid-123"
						}
					}
				}
			}`

			var policy resources.RoutePolicy
			err := json.Unmarshal([]byte(jsonData), &policy)
			Expect(err).NotTo(HaveOccurred())

			Expect(policy.GUID).To(Equal("some-guid"))
			Expect(policy.Source).To(Equal("cf:app:app-guid"))
			Expect(policy.RouteGUID).To(Equal("route-guid-123"))
		})
	})
})

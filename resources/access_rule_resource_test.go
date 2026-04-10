package resources_test

import (
	"encoding/json"
	"testing"

	"code.cloudfoundry.org/cli/v9/resources"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAccessRuleResource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AccessRule Resource Suite")
}

var _ = Describe("AccessRule", func() {
	Describe("MarshalJSON", func() {
		It("marshals the access rule with relationships", func() {
			rule := resources.AccessRule{
				Name:      "allow-backend",
				Selector:  "cf:app:some-app-guid",
				RouteGUID: "some-route-guid",
			}

			data, err := json.Marshal(rule)
			Expect(err).NotTo(HaveOccurred())

			var result map[string]interface{}
			err = json.Unmarshal(data, &result)
			Expect(err).NotTo(HaveOccurred())

			Expect(result["name"]).To(Equal("allow-backend"))
			Expect(result["selector"]).To(Equal("cf:app:some-app-guid"))
			Expect(result["relationships"]).NotTo(BeNil())

			relationships := result["relationships"].(map[string]interface{})
			route := relationships["route"].(map[string]interface{})
			routeData := route["data"].(map[string]interface{})
			Expect(routeData["guid"]).To(Equal("some-route-guid"))
		})
	})

	Describe("UnmarshalJSON", func() {
		It("unmarshals the access rule from relationships", func() {
			jsonData := `{
				"guid": "some-guid",
				"name": "test-rule",
				"selector": "cf:app:app-guid",
				"relationships": {
					"route": {
						"data": {
							"guid": "route-guid-123"
						}
					}
				}
			}`

			var rule resources.AccessRule
			err := json.Unmarshal([]byte(jsonData), &rule)
			Expect(err).NotTo(HaveOccurred())

			Expect(rule.GUID).To(Equal("some-guid"))
			Expect(rule.Name).To(Equal("test-rule"))
			Expect(rule.Selector).To(Equal("cf:app:app-guid"))
			Expect(rule.RouteGUID).To(Equal("route-guid-123"))
		})
	})
})

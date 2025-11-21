package fixtures_test

import (
	"encoding/json"

	"github.com/cloudfoundry/cli/testhelpers/fixtures"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Fixtures", func() {
	Describe("GetApplicationFixture", func() {
		It("returns valid application JSON", func() {
			appJSON := fixtures.GetApplicationFixture()

			var data map[string]interface{}
			err := json.Unmarshal([]byte(appJSON), &data)

			Expect(err).NotTo(HaveOccurred())
			Expect(data).To(HaveKey("metadata"))
			Expect(data).To(HaveKey("entity"))

			entity := data["entity"].(map[string]interface{})
			Expect(entity["name"]).To(Equal("my-app"))
			Expect(entity["state"]).To(Equal("STARTED"))
		})

		It("includes all required application fields", func() {
			appJSON := fixtures.GetApplicationFixture()

			var data map[string]interface{}
			json.Unmarshal([]byte(appJSON), &data)

			entity := data["entity"].(map[string]interface{})
			Expect(entity).To(HaveKey("name"))
			Expect(entity).To(HaveKey("memory"))
			Expect(entity).To(HaveKey("instances"))
			Expect(entity).To(HaveKey("state"))
			Expect(entity).To(HaveKey("space_guid"))
		})
	})

	Describe("GetSpaceFixture", func() {
		It("returns valid space JSON", func() {
			spaceJSON := fixtures.GetSpaceFixture()

			var data map[string]interface{}
			err := json.Unmarshal([]byte(spaceJSON), &data)

			Expect(err).NotTo(HaveOccurred())

			entity := data["entity"].(map[string]interface{})
			Expect(entity["name"]).To(Equal("development"))
			Expect(entity["organization_guid"]).To(Equal("org-guid-789"))
		})
	})

	Describe("GetOrganizationFixture", func() {
		It("returns valid organization JSON", func() {
			orgJSON := fixtures.GetOrganizationFixture()

			var data map[string]interface{}
			err := json.Unmarshal([]byte(orgJSON), &data)

			Expect(err).NotTo(HaveOccurred())

			entity := data["entity"].(map[string]interface{})
			Expect(entity["name"]).To(Equal("my-org"))
			Expect(entity["status"]).To(Equal("active"))
		})
	})

	Describe("GetServiceInstanceFixture", func() {
		It("returns valid service instance JSON", func() {
			serviceJSON := fixtures.GetServiceInstanceFixture()

			var data map[string]interface{}
			err := json.Unmarshal([]byte(serviceJSON), &data)

			Expect(err).NotTo(HaveOccurred())

			entity := data["entity"].(map[string]interface{})
			Expect(entity["name"]).To(Equal("my-database"))
			Expect(entity).To(HaveKey("credentials"))

			credentials := entity["credentials"].(map[string]interface{})
			Expect(credentials["hostname"]).To(Equal("db.example.com"))
		})

		It("includes service credentials", func() {
			serviceJSON := fixtures.GetServiceInstanceFixture()

			var data map[string]interface{}
			json.Unmarshal([]byte(serviceJSON), &data)

			entity := data["entity"].(map[string]interface{})
			credentials := entity["credentials"].(map[string]interface{})

			Expect(credentials).To(HaveKey("hostname"))
			Expect(credentials).To(HaveKey("port"))
			Expect(credentials).To(HaveKey("username"))
			Expect(credentials).To(HaveKey("password"))
		})
	})

	Describe("GetRouteFixture", func() {
		It("returns valid route JSON", func() {
			routeJSON := fixtures.GetRouteFixture()

			var data map[string]interface{}
			err := json.Unmarshal([]byte(routeJSON), &data)

			Expect(err).NotTo(HaveOccurred())

			entity := data["entity"].(map[string]interface{})
			Expect(entity["host"]).To(Equal("my-app"))
			Expect(entity["domain_guid"]).To(Equal("domain-guid-456"))
		})
	})

	Describe("GetDomainFixture", func() {
		It("returns valid domain JSON", func() {
			domainJSON := fixtures.GetDomainFixture()

			var data map[string]interface{}
			err := json.Unmarshal([]byte(domainJSON), &data)

			Expect(err).NotTo(HaveOccurred())

			entity := data["entity"].(map[string]interface{})
			Expect(entity["name"]).To(Equal("example.com"))
		})
	})

	Describe("GetBuildpackFixture", func() {
		It("returns valid buildpack JSON", func() {
			buildpackJSON := fixtures.GetBuildpackFixture()

			var data map[string]interface{}
			err := json.Unmarshal([]byte(buildpackJSON), &data)

			Expect(err).NotTo(HaveOccurred())

			entity := data["entity"].(map[string]interface{})
			Expect(entity["name"]).To(Equal("ruby_buildpack"))
			Expect(entity["enabled"]).To(BeTrue())
		})
	})

	Describe("GetErrorResponseFixture", func() {
		It("returns valid error response JSON", func() {
			errorJSON := fixtures.GetErrorResponseFixture()

			var data map[string]interface{}
			err := json.Unmarshal([]byte(errorJSON), &data)

			Expect(err).NotTo(HaveOccurred())
			Expect(data).To(HaveKey("code"))
			Expect(data).To(HaveKey("description"))
			Expect(data).To(HaveKey("error_code"))
		})
	})

	Describe("GetMultipleAppsFixture", func() {
		It("returns valid paginated response", func() {
			appsJSON := fixtures.GetMultipleAppsFixture()

			var data map[string]interface{}
			err := json.Unmarshal([]byte(appsJSON), &data)

			Expect(err).NotTo(HaveOccurred())
			Expect(data).To(HaveKey("total_results"))
			Expect(data).To(HaveKey("resources"))

			resources := data["resources"].([]interface{})
			Expect(len(resources)).To(Equal(3))
		})

		It("includes pagination metadata", func() {
			appsJSON := fixtures.GetMultipleAppsFixture()

			var data map[string]interface{}
			json.Unmarshal([]byte(appsJSON), &data)

			Expect(data["total_results"]).To(Equal(float64(3)))
			Expect(data["total_pages"]).To(Equal(float64(1)))
		})
	})

	Describe("All fixtures", func() {
		It("are valid JSON", func() {
			fixtures := []string{
				fixtures.GetApplicationFixture(),
				fixtures.GetSpaceFixture(),
				fixtures.GetOrganizationFixture(),
				fixtures.GetServiceInstanceFixture(),
				fixtures.GetRouteFixture(),
				fixtures.GetDomainFixture(),
				fixtures.GetBuildpackFixture(),
				fixtures.GetErrorResponseFixture(),
				fixtures.GetMultipleAppsFixture(),
			}

			for _, fixture := range fixtures {
				var data interface{}
				err := json.Unmarshal([]byte(fixture), &data)
				Expect(err).NotTo(HaveOccurred())
			}
		})
	})
})

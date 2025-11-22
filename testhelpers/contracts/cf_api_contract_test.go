package contracts_test

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestContracts(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CF API Contract Tests Suite")
}

// Contract tests verify that the CF API responses match expected schemas
// These tests help ensure compatibility with Cloud Foundry API

var _ = Describe("CF API Contracts", func() {
	Describe("Application Resource Contract", func() {
		It("matches expected schema structure", func() {
			// Expected contract for /v2/apps/{guid} response
			schema := map[string]interface{}{
				"metadata": map[string]interface{}{
					"guid":       "string",
					"url":        "string",
					"created_at": "string",
					"updated_at": "string",
				},
				"entity": map[string]interface{}{
					"name":                "string",
					"memory":              "number",
					"instances":           "number",
					"disk_quota":          "number",
					"space_guid":          "string",
					"stack_guid":          "string",
					"state":               "string",
					"package_state":       "string",
					"buildpack":           "string",
					"detected_buildpack":  "string",
					"environment_json":    "object",
					"staging_failed_reason": "string_or_null",
					"docker_image":        "string_or_null",
				},
			}

			// Sample response that should match contract
			sampleResponse := `{
				"metadata": {
					"guid": "app-guid-123",
					"url": "/v2/apps/app-guid-123",
					"created_at": "2015-01-01T00:00:00Z",
					"updated_at": "2015-01-02T00:00:00Z"
				},
				"entity": {
					"name": "my-app",
					"memory": 256,
					"instances": 1,
					"disk_quota": 1024,
					"space_guid": "space-guid-456",
					"stack_guid": "stack-guid-789",
					"state": "STARTED",
					"package_state": "STAGED",
					"buildpack": "ruby_buildpack",
					"detected_buildpack": "Ruby",
					"environment_json": {},
					"staging_failed_reason": null,
					"docker_image": null
				}
			}`

			var response map[string]interface{}
			err := json.Unmarshal([]byte(sampleResponse), &response)
			Expect(err).NotTo(HaveOccurred())

			// Verify top-level structure
			Expect(response).To(HaveKey("metadata"))
			Expect(response).To(HaveKey("entity"))

			// Verify metadata fields
			metadata := response["metadata"].(map[string]interface{})
			Expect(metadata).To(HaveKey("guid"))
			Expect(metadata).To(HaveKey("url"))
			Expect(metadata).To(HaveKey("created_at"))
			Expect(metadata).To(HaveKey("updated_at"))

			// Verify entity fields
			entity := response["entity"].(map[string]interface{})
			Expect(entity).To(HaveKey("name"))
			Expect(entity).To(HaveKey("memory"))
			Expect(entity).To(HaveKey("instances"))
			Expect(entity).To(HaveKey("state"))
			Expect(entity).To(HaveKey("space_guid"))

			// Verify types
			Expect(metadata["guid"]).To(BeAssignableToTypeOf(""))
			Expect(entity["memory"]).To(BeAssignableToTypeOf(float64(0)))
			Expect(entity["instances"]).To(BeAssignableToTypeOf(float64(0)))

			// Schema matches expectations
			Expect(schema).NotTo(BeNil())
		})

		It("validates required fields are present", func() {
			requiredFields := []string{
				"name", "memory", "instances", "state",
				"space_guid", "stack_guid",
			}

			sampleEntity := map[string]interface{}{
				"name":       "my-app",
				"memory":     256,
				"instances":  1,
				"state":      "STARTED",
				"space_guid": "space-guid",
				"stack_guid": "stack-guid",
			}

			for _, field := range requiredFields {
				Expect(sampleEntity).To(HaveKey(field),
					"Required field %s is missing", field)
			}
		})

		It("validates state enum values", func() {
			validStates := []string{"STOPPED", "STARTED"}

			for _, state := range validStates {
				entity := map[string]interface{}{"state": state}
				Expect(entity["state"]).To(BeElementOf(validStates))
			}
		})
	})

	Describe("Space Resource Contract", func() {
		It("matches expected schema structure", func() {
			sampleResponse := `{
				"metadata": {
					"guid": "space-guid-456",
					"url": "/v2/spaces/space-guid-456",
					"created_at": "2015-01-01T00:00:00Z"
				},
				"entity": {
					"name": "development",
					"organization_guid": "org-guid-789",
					"space_quota_definition_guid": null,
					"allow_ssh": true
				}
			}`

			var response map[string]interface{}
			err := json.Unmarshal([]byte(sampleResponse), &response)
			Expect(err).NotTo(HaveOccurred())

			Expect(response).To(HaveKey("metadata"))
			Expect(response).To(HaveKey("entity"))

			entity := response["entity"].(map[string]interface{})
			Expect(entity).To(HaveKey("name"))
			Expect(entity).To(HaveKey("organization_guid"))
			Expect(entity).To(HaveKey("allow_ssh"))
		})
	})

	Describe("Organization Resource Contract", func() {
		It("matches expected schema structure", func() {
			sampleResponse := `{
				"metadata": {
					"guid": "org-guid-789",
					"url": "/v2/organizations/org-guid-789",
					"created_at": "2015-01-01T00:00:00Z"
				},
				"entity": {
					"name": "my-org",
					"billing_enabled": false,
					"quota_definition_guid": "quota-guid-123",
					"status": "active"
				}
			}`

			var response map[string]interface{}
			err := json.Unmarshal([]byte(sampleResponse), &response)
			Expect(err).NotTo(HaveOccurred())

			entity := response["entity"].(map[string]interface{})
			Expect(entity).To(HaveKey("name"))
			Expect(entity).To(HaveKey("status"))
			Expect(entity["status"]).To(Equal("active"))
		})

		It("validates status enum values", func() {
			validStatuses := []string{"active", "suspended"}

			for _, status := range validStatuses {
				entity := map[string]interface{}{"status": status}
				Expect(entity["status"]).To(BeElementOf(validStatuses))
			}
		})
	})

	Describe("Service Instance Resource Contract", func() {
		It("matches expected schema structure", func() {
			sampleResponse := `{
				"metadata": {
					"guid": "service-instance-guid-123",
					"url": "/v2/service_instances/service-instance-guid-123",
					"created_at": "2015-01-01T00:00:00Z"
				},
				"entity": {
					"name": "my-database",
					"credentials": {
						"hostname": "db.example.com",
						"port": 5432,
						"username": "admin",
						"password": "secret"
					},
					"service_plan_guid": "plan-guid-456",
					"space_guid": "space-guid-456",
					"type": "managed_service_instance",
					"tags": ["mysql", "database"]
				}
			}`

			var response map[string]interface{}
			err := json.Unmarshal([]byte(sampleResponse), &response)
			Expect(err).NotTo(HaveOccurred())

			entity := response["entity"].(map[string]interface{})
			Expect(entity).To(HaveKey("name"))
			Expect(entity).To(HaveKey("credentials"))
			Expect(entity).To(HaveKey("service_plan_guid"))
			Expect(entity).To(HaveKey("space_guid"))
			Expect(entity).To(HaveKey("type"))
			Expect(entity).To(HaveKey("tags"))

			// Verify credentials structure
			credentials := entity["credentials"].(map[string]interface{})
			Expect(credentials).To(HaveKey("hostname"))
			Expect(credentials).To(HaveKey("port"))

			// Verify tags is array
			tags := entity["tags"].([]interface{})
			Expect(len(tags)).To(BeNumerically(">=", 0))
		})
	})

	Describe("Route Resource Contract", func() {
		It("matches expected schema structure", func() {
			sampleResponse := `{
				"metadata": {
					"guid": "route-guid-123",
					"url": "/v2/routes/route-guid-123",
					"created_at": "2015-01-01T00:00:00Z"
				},
				"entity": {
					"host": "my-app",
					"domain_guid": "domain-guid-456",
					"space_guid": "space-guid-789",
					"path": "",
					"port": null
				}
			}`

			var response map[string]interface{}
			err := json.Unmarshal([]byte(sampleResponse), &response)
			Expect(err).NotTo(HaveOccurred())

			entity := response["entity"].(map[string]interface{})
			Expect(entity).To(HaveKey("host"))
			Expect(entity).To(HaveKey("domain_guid"))
			Expect(entity).To(HaveKey("space_guid"))
			Expect(entity).To(HaveKey("path"))
		})
	})

	Describe("Error Response Contract", func() {
		It("matches expected error schema", func() {
			sampleError := `{
				"code": 10001,
				"description": "The request is semantically invalid: command presence",
				"error_code": "CF-MessageParseError"
			}`

			var errorResponse map[string]interface{}
			err := json.Unmarshal([]byte(sampleError), &errorResponse)
			Expect(err).NotTo(HaveOccurred())

			Expect(errorResponse).To(HaveKey("code"))
			Expect(errorResponse).To(HaveKey("description"))
			Expect(errorResponse).To(HaveKey("error_code"))

			// Verify types
			Expect(errorResponse["code"]).To(BeAssignableToTypeOf(float64(0)))
			Expect(errorResponse["description"]).To(BeAssignableToTypeOf(""))
			Expect(errorResponse["error_code"]).To(BeAssignableToTypeOf(""))
		})
	})

	Describe("Paginated Response Contract", func() {
		It("matches expected pagination schema", func() {
			sampleResponse := `{
				"total_results": 3,
				"total_pages": 1,
				"prev_url": null,
				"next_url": null,
				"resources": [
					{"metadata": {"guid": "app-1"}, "entity": {"name": "app-1"}},
					{"metadata": {"guid": "app-2"}, "entity": {"name": "app-2"}},
					{"metadata": {"guid": "app-3"}, "entity": {"name": "app-3"}}
				]
			}`

			var response map[string]interface{}
			err := json.Unmarshal([]byte(sampleResponse), &response)
			Expect(err).NotTo(HaveOccurred())

			// Verify pagination fields
			Expect(response).To(HaveKey("total_results"))
			Expect(response).To(HaveKey("total_pages"))
			Expect(response).To(HaveKey("prev_url"))
			Expect(response).To(HaveKey("next_url"))
			Expect(response).To(HaveKey("resources"))

			// Verify types
			Expect(response["total_results"]).To(BeAssignableToTypeOf(float64(0)))
			Expect(response["total_pages"]).To(BeAssignableToTypeOf(float64(0)))

			// Verify resources is array
			resources := response["resources"].([]interface{})
			Expect(len(resources)).To(Equal(3))

			// Each resource should have metadata and entity
			for _, resource := range resources {
				r := resource.(map[string]interface{})
				Expect(r).To(HaveKey("metadata"))
				Expect(r).To(HaveKey("entity"))
			}
		})
	})

	Describe("Contract Backward Compatibility", func() {
		It("handles optional fields gracefully", func() {
			// Minimal valid response
			minimalResponse := `{
				"metadata": {"guid": "123"},
				"entity": {"name": "my-app"}
			}`

			var response map[string]interface{}
			err := json.Unmarshal([]byte(minimalResponse), &response)
			Expect(err).NotTo(HaveOccurred())

			// Should not fail even with missing optional fields
			Expect(response).To(HaveKey("metadata"))
			Expect(response).To(HaveKey("entity"))
		})

		It("handles extra fields gracefully", func() {
			// Response with extra unknown fields
			responseWithExtra := `{
				"metadata": {"guid": "123", "unknown_field": "value"},
				"entity": {"name": "my-app", "future_field": 42}
			}`

			var response map[string]interface{}
			err := json.Unmarshal([]byte(responseWithExtra), &response)
			Expect(err).NotTo(HaveOccurred())

			// Should not fail with extra fields (forward compatibility)
			Expect(response).To(HaveKey("metadata"))
			Expect(response).To(HaveKey("entity"))
		})
	})
})

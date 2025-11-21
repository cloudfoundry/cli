package snapshot_test

import (
	"testing"

	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/testhelpers/snapshot"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSnapshot(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Snapshot Testing Suite")
}

var _ = Describe("Snapshot Testing", func() {
	Describe("String Snapshots", func() {
		It("matches string output snapshots", func() {
			snap := snapshot.New("string_output_test")

			output := "Hello, World!\nThis is a test output.\nLine 3"

			snap.MatchSnapshot(output)
		})

		It("matches multiline output", func() {
			snap := snapshot.New("multiline_output_test")

			output := `
Application Information:
Name:       my-app
State:      STARTED
Instances:  3/3
Memory:     512M
Routes:     my-app.example.com
`

			snap.MatchOutputSnapshot(output)
		})
	})

	Describe("JSON Snapshots", func() {
		It("matches JSON structure snapshots", func() {
			snap := snapshot.New("json_structure_test")

			data := map[string]interface{}{
				"name":    "my-app",
				"state":   "STARTED",
				"memory":  512,
				"routes":  []string{"my-app.example.com"},
			}

			snap.MatchJSONSnapshot(data)
		})

		It("matches complex model snapshots", func() {
			snap := snapshot.New("complex_model_test")

			app := models.Application{
				ApplicationFields: models.ApplicationFields{
					Guid:      "app-guid-123",
					Name:      "my-app",
					State:     "STARTED",
					Instances: 3,
					Memory:    512,
				},
			}

			snap.MatchJSONSnapshot(app)
		})
	})

	Describe("Command Output Snapshots", func() {
		It("matches CLI command output", func() {
			snap := snapshot.New("cli_apps_output")

			output := `Getting apps in org my-org / space development as admin...
OK

name       requested state   instances   memory   disk   urls
app-1      started           1/1         256M     1G     app-1.example.com
app-2      stopped           0/1         512M     1G     app-2.example.com
my-app     started           3/3         512M     2G     my-app.example.com
`

			snap.MatchOutputSnapshot(output)
		})

		It("matches error output", func() {
			snap := snapshot.New("cli_error_output")

			output := `FAILED
Server error, status code: 500, error code: 10001, message: Internal server error`

			snap.MatchOutputSnapshot(output)
		})
	})

	Describe("API Response Snapshots", func() {
		It("matches application API response", func() {
			snap := snapshot.New("api_application_response")

			response := map[string]interface{}{
				"metadata": map[string]interface{}{
					"guid":       "app-guid-123",
					"url":        "/v2/apps/app-guid-123",
					"created_at": "2015-01-01T00:00:00Z",
					"updated_at": "2015-01-02T00:00:00Z",
				},
				"entity": map[string]interface{}{
					"name":       "my-app",
					"memory":     256,
					"instances":  1,
					"state":      "STARTED",
					"space_guid": "space-guid-456",
				},
			}

			snap.MatchJSONSnapshot(response)
		})

		It("matches paginated API response", func() {
			snap := snapshot.New("api_paginated_response")

			response := map[string]interface{}{
				"total_results": 3,
				"total_pages":   1,
				"prev_url":      nil,
				"next_url":      nil,
				"resources": []map[string]interface{}{
					{
						"metadata": map[string]interface{}{"guid": "app-1"},
						"entity":   map[string]interface{}{"name": "app-1"},
					},
					{
						"metadata": map[string]interface{}{"guid": "app-2"},
						"entity":   map[string]interface{}{"name": "app-2"},
					},
					{
						"metadata": map[string]interface{}{"guid": "app-3"},
						"entity":   map[string]interface{}{"name": "app-3"},
					},
				},
			}

			snap.MatchJSONSnapshot(response)
		})
	})

	Describe("Data Transformation Snapshots", func() {
		It("verifies data transformation output", func() {
			snap := snapshot.New("data_transformation_test")

			// Simulate some data transformation
			input := models.Application{
				ApplicationFields: models.ApplicationFields{
					Name:      "my-app",
					State:     "STARTED",
					Memory:    512,
					Instances: 3,
				},
			}

			// Transform to display format
			output := map[string]interface{}{
				"name":      input.Name,
				"state":     input.State,
				"memory":    input.Memory,
				"instances": input.Instances,
				"url":       "my-app.example.com",
			}

			snap.MatchJSONSnapshot(output)
		})
	})

	Describe("Snapshot Utilities", func() {
		It("sanitizes test names correctly", func() {
			snap := snapshot.New("Test With Spaces / And : Special * Characters")

			// The snapshot should be created with sanitized name
			snap.MatchSnapshot("test data")
		})

		It("handles byte array snapshots", func() {
			snap := snapshot.New("byte_array_test")

			data := []byte("Binary data or byte array content")

			snap.MatchSnapshot(data)
		})
	})
})

var _ = Describe("Snapshot Testing Examples", func() {
	It("demonstrates updating snapshots", func() {
		// To update snapshots, run: UPDATE_SNAPSHOTS=true ginkgo
		snap := snapshot.New("example_update_snapshot")

		currentOutput := "This is the current output"

		snap.MatchSnapshot(currentOutput)
	})

	It("demonstrates snapshot-driven development", func() {
		// 1. Write the test with expected output
		// 2. Run with UPDATE_SNAPSHOTS=true to create snapshot
		// 3. Future runs will compare against this snapshot
		// 4. If output changes, test fails showing diff
		// 5. Review changes and update snapshot if intentional

		snap := snapshot.New("snapshot_driven_development")

		expectedOutput := `Feature: User Login
Scenario: Successful login
  Given a valid username and password
  When the user attempts to login
  Then the user should be authenticated
  And redirected to the dashboard`

		snap.MatchOutputSnapshot(expectedOutput)
	})
})

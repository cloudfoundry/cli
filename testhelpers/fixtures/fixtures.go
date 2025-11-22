package fixtures

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
)

// LoadJSONFixture loads a JSON fixture file and unmarshals it into the provided interface
func LoadJSONFixture(fixtureName string, v interface{}) error {
	path := filepath.Join("testhelpers", "fixtures", "json", fixtureName)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

// LoadFixture loads a fixture file and returns its contents as a string
func LoadFixture(fixtureName string) (string, error) {
	path := filepath.Join("testhelpers", "fixtures", fixtureName)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// Common CF API response fixtures as constants for quick access
const (
	// ApplicationJSON represents a typical CF application response
	ApplicationJSON = `{
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

	// SpaceJSON represents a typical CF space response
	SpaceJSON = `{
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

	// OrganizationJSON represents a typical CF org response
	OrganizationJSON = `{
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

	// ServiceInstanceJSON represents a typical CF service instance
	ServiceInstanceJSON = `{
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

	// RouteJSON represents a typical CF route response
	RouteJSON = `{
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

	// DomainJSON represents a typical CF domain response
	DomainJSON = `{
		"metadata": {
			"guid": "domain-guid-456",
			"url": "/v2/domains/domain-guid-456",
			"created_at": "2015-01-01T00:00:00Z"
		},
		"entity": {
			"name": "example.com",
			"owning_organization_guid": null
		}
	}`

	// BuildpackJSON represents a typical CF buildpack response
	BuildpackJSON = `{
		"metadata": {
			"guid": "buildpack-guid-123",
			"url": "/v2/buildpacks/buildpack-guid-123",
			"created_at": "2015-01-01T00:00:00Z"
		},
		"entity": {
			"name": "ruby_buildpack",
			"position": 1,
			"enabled": true,
			"locked": false,
			"filename": "ruby_buildpack-v1.6.0.zip"
		}
	}`

	// ErrorResponseJSON represents a typical CF error response
	ErrorResponseJSON = `{
		"code": 10001,
		"description": "The request is semantically invalid: command presence",
		"error_code": "CF-MessageParseError"
	}`

	// MultipleAppsJSON represents a paginated list of applications
	MultipleAppsJSON = `{
		"total_results": 3,
		"total_pages": 1,
		"prev_url": null,
		"next_url": null,
		"resources": [
			{
				"metadata": {"guid": "app-1-guid"},
				"entity": {"name": "app-1", "state": "STARTED"}
			},
			{
				"metadata": {"guid": "app-2-guid"},
				"entity": {"name": "app-2", "state": "STOPPED"}
			},
			{
				"metadata": {"guid": "app-3-guid"},
				"entity": {"name": "app-3", "state": "STARTED"}
			}
		]
	}`
)

// GetApplicationFixture returns a sample application JSON
func GetApplicationFixture() string {
	return ApplicationJSON
}

// GetSpaceFixture returns a sample space JSON
func GetSpaceFixture() string {
	return SpaceJSON
}

// GetOrganizationFixture returns a sample organization JSON
func GetOrganizationFixture() string {
	return OrganizationJSON
}

// GetServiceInstanceFixture returns a sample service instance JSON
func GetServiceInstanceFixture() string {
	return ServiceInstanceJSON
}

// GetRouteFixture returns a sample route JSON
func GetRouteFixture() string {
	return RouteJSON
}

// GetDomainFixture returns a sample domain JSON
func GetDomainFixture() string {
	return DomainJSON
}

// GetBuildpackFixture returns a sample buildpack JSON
func GetBuildpackFixture() string {
	return BuildpackJSON
}

// GetErrorResponseFixture returns a sample error response JSON
func GetErrorResponseFixture() string {
	return ErrorResponseJSON
}

// GetMultipleAppsFixture returns a sample paginated apps response
func GetMultipleAppsFixture() string {
	return MultipleAppsJSON
}

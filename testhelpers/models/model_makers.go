package models

import (
	"github.com/cloudfoundry/cli/cf/models"
	"time"
)

// MakeApplication creates a test application with sensible defaults
func MakeApplication(name string, opts ...func(*models.Application)) models.Application {
	app := models.Application{
		ApplicationFields: models.ApplicationFields{
			Guid:          name + "-guid",
			Name:          name,
			State:         "STARTED",
			InstanceCount: 1,
			Memory:        256,
			DiskQuota:     1024,
		},
	}

	for _, opt := range opts {
		opt(&app)
	}

	return app
}

// WithMemory sets the application memory
func WithMemory(memory int64) func(*models.Application) {
	return func(app *models.Application) {
		app.Memory = memory
	}
}

// WithInstances sets the instance count
func WithInstances(count int) func(*models.Application) {
	return func(app *models.Application) {
		app.InstanceCount = count
	}
}

// WithState sets the application state
func WithState(state string) func(*models.Application) {
	return func(app *models.Application) {
		app.State = state
	}
}

// WithRoutes adds routes to the application
func WithRoutes(routes ...models.RouteSummary) func(*models.Application) {
	return func(app *models.Application) {
		app.Routes = routes
	}
}

// MakeRoute creates a test route with sensible defaults
func MakeRoute(host, domain string) models.Route {
	return models.Route{
		Guid: host + "-" + domain + "-guid",
		Host: host,
		Domain: models.DomainFields{
			Guid: domain + "-guid",
			Name: domain,
		},
	}
}

// MakeSpace creates a test space with sensible defaults
func MakeSpace(name string) models.Space {
	return models.Space{
		SpaceFields: models.SpaceFields{
			Guid: name + "-guid",
			Name: name,
		},
	}
}

// MakeOrganization creates a test organization with sensible defaults
func MakeOrganization(name string) models.Organization {
	return models.Organization{
		OrganizationFields: models.OrganizationFields{
			Guid: name + "-guid",
			Name: name,
		},
	}
}

// MakeServiceInstance creates a test service instance
func MakeServiceInstance(name string) models.ServiceInstance {
	return models.ServiceInstance{
		ServiceInstanceFields: models.ServiceInstanceFields{
			Guid: name + "-guid",
			Name: name,
		},
	}
}

// MakeAppInstance creates a test app instance with metrics
func MakeAppInstance(state models.InstanceState) models.AppInstanceFields {
	return models.AppInstanceFields{
		State:     state,
		Since:     time.Now(),
		CpuUsage:  50.0,
		DiskQuota: 1073741824, // 1GB
		DiskUsage: 536870912,  // 512MB
		MemQuota:  536870912,  // 512MB
		MemUsage:  268435456,  // 256MB
	}
}

// MakeQuota creates a test quota
func MakeQuota(name string, memoryMB, routes, services int) models.QuotaFields {
	return models.NewQuotaFields(
		name,
		int64(memoryMB),
		int64(memoryMB/2),
		routes,
		services,
		true,
	)
}

// MakeDomain creates a test domain
func MakeDomain(name string, shared bool) models.DomainFields {
	domain := models.DomainFields{
		Guid:   name + "-guid",
		Name:   name,
		Shared: shared,
	}

	if !shared {
		domain.OwningOrganizationGuid = "org-guid"
	}

	return domain
}

// MakeBuildpack creates a test buildpack
func MakeBuildpack(name string, position int, enabled bool) models.Buildpack {
	pos := position
	en := enabled
	locked := false

	return models.Buildpack{
		Guid:     name + "-guid",
		Name:     name,
		Position: &pos,
		Enabled:  &en,
		Locked:   &locked,
	}
}

// MakeSecurityGroup creates a test security group
func MakeSecurityGroup(name string, rules []map[string]interface{}) models.SecurityGroup {
	return models.SecurityGroup{
		SecurityGroupFields: models.SecurityGroupFields{
			Guid:  name + "-guid",
			Name:  name,
			Rules: rules,
		},
	}
}

// MakeStack creates a test stack
func MakeStack(name, description string) models.Stack {
	return models.Stack{
		Guid:        name + "-guid",
		Name:        name,
		Description: description,
	}
}

// MakeUser creates a test user
func MakeUser(username string, isAdmin bool) models.UserFields {
	return models.UserFields{
		Guid:     username + "-guid",
		Username: username,
		IsAdmin:  isAdmin,
	}
}

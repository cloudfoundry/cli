package generators

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/cloudfoundry/cli/cf/models"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// AppGenerator generates test application data
type AppGenerator struct {
	index int
}

// NewAppGenerator creates a new app generator
func NewAppGenerator() *AppGenerator {
	return &AppGenerator{index: 0}
}

// Generate creates a new application
func (g *AppGenerator) Generate() models.Application {
	g.index++

	states := []string{"STARTED", "STOPPED", "CRASHED"}
	buildpacks := []string{"ruby_buildpack", "nodejs_buildpack", "java_buildpack", "python_buildpack", "go_buildpack"}

	return models.Application{
		ApplicationFields: models.ApplicationFields{
			Guid:             fmt.Sprintf("app-guid-%d", g.index),
			Name:             fmt.Sprintf("app-%d", g.index),
			State:            states[rand.Intn(len(states))],
			Instances:        rand.Intn(10) + 1,
			Memory:           int64([]int{256, 512, 1024, 2048}[rand.Intn(4)]),
			DiskQuota:        int64([]int{512, 1024, 2048, 4096}[rand.Intn(4)]),
			BuildpackUrl:     buildpacks[rand.Intn(len(buildpacks))],
			PackageUpdatedAt: &time.Time{},
		},
	}
}

// GenerateBatch creates multiple applications
func (g *AppGenerator) GenerateBatch(count int) []models.Application {
	apps := make([]models.Application, count)
	for i := 0; i < count; i++ {
		apps[i] = g.Generate()
	}
	return apps
}

// With custom configuration
type AppConfig struct {
	Name       string
	State      string
	Instances  int
	Memory     int64
	BuildpackUrl string
}

// GenerateWithConfig creates app with specific config
func (g *AppGenerator) GenerateWithConfig(config AppConfig) models.Application {
	g.index++

	app := g.Generate()

	if config.Name != "" {
		app.Name = config.Name
	}
	if config.State != "" {
		app.State = config.State
	}
	if config.Instances > 0 {
		app.Instances = config.Instances
	}
	if config.Memory > 0 {
		app.Memory = config.Memory
	}
	if config.BuildpackUrl != "" {
		app.BuildpackUrl = config.BuildpackUrl
	}

	return app
}

// SpaceGenerator generates test space data
type SpaceGenerator struct {
	index int
}

// NewSpaceGenerator creates a new space generator
func NewSpaceGenerator() *SpaceGenerator {
	return &SpaceGenerator{index: 0}
}

// Generate creates a new space
func (g *SpaceGenerator) Generate() models.Space {
	g.index++

	return models.Space{
		SpaceFields: models.SpaceFields{
			Guid: fmt.Sprintf("space-guid-%d", g.index),
			Name: fmt.Sprintf("space-%d", g.index),
		},
		Organization: models.OrganizationFields{
			Guid: fmt.Sprintf("org-guid-%d", g.index),
			Name: fmt.Sprintf("org-%d", g.index),
		},
	}
}

// OrganizationGenerator generates test organization data
type OrganizationGenerator struct {
	index int
}

// NewOrganizationGenerator creates a new org generator
func NewOrganizationGenerator() *OrganizationGenerator {
	return &OrganizationGenerator{index: 0}
}

// Generate creates a new organization
func (g *OrganizationGenerator) Generate() models.Organization {
	g.index++

	return models.Organization{
		OrganizationFields: models.OrganizationFields{
			Guid:   fmt.Sprintf("org-guid-%d", g.index),
			Name:   fmt.Sprintf("org-%d", g.index),
			Status: "active",
		},
		Spaces: []models.SpaceFields{},
	}
}

// RouteGenerator generates test route data
type RouteGenerator struct {
	index int
}

// NewRouteGenerator creates a new route generator
func NewRouteGenerator() *RouteGenerator {
	return &RouteGenerator{index: 0}
}

// Generate creates a new route
func (g *RouteGenerator) Generate() models.Route {
	g.index++

	domains := []string{"example.com", "app.io", "cloud.com", "test.net"}
	paths := []string{"", "/api", "/v1", "/v2"}

	return models.Route{
		Guid: fmt.Sprintf("route-guid-%d", g.index),
		Host: fmt.Sprintf("app-%d", g.index),
		Domain: models.DomainFields{
			Guid: fmt.Sprintf("domain-guid-%d", g.index),
			Name: domains[rand.Intn(len(domains))],
		},
		Path: paths[rand.Intn(len(paths))],
		Port: 0,
	}
}

// ServiceInstanceGenerator generates test service instance data
type ServiceInstanceGenerator struct {
	index int
}

// NewServiceInstanceGenerator creates a new service instance generator
func NewServiceInstanceGenerator() *ServiceInstanceGenerator {
	return &ServiceInstanceGenerator{index: 0}
}

// Generate creates a new service instance
func (g *ServiceInstanceGenerator) Generate() models.ServiceInstance {
	g.index++

	serviceTypes := []string{"mysql", "postgres", "redis", "mongodb", "rabbitmq"}
	serviceType := serviceTypes[rand.Intn(len(serviceTypes))]

	return models.ServiceInstance{
		ServiceInstanceFields: models.ServiceInstanceFields{
			Guid: fmt.Sprintf("service-instance-guid-%d", g.index),
			Name: fmt.Sprintf("%s-%d", serviceType, g.index),
		},
		ServicePlan: models.ServicePlanFields{
			Guid: fmt.Sprintf("plan-guid-%d", g.index),
			Name: "standard",
		},
	}
}

// UserGenerator generates test user data
type UserGenerator struct {
	index int
}

// NewUserGenerator creates a new user generator
func NewUserGenerator() *UserGenerator {
	return &UserGenerator{index: 0}
}

// Generate creates a new user
func (g *UserGenerator) Generate() models.UserFields {
	g.index++

	firstNames := []string{"John", "Jane", "Alice", "Bob", "Charlie", "Diana"}
	lastNames := []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia"}

	firstName := firstNames[rand.Intn(len(firstNames))]
	lastName := lastNames[rand.Intn(len(lastNames))]

	return models.UserFields{
		Guid:     fmt.Sprintf("user-guid-%d", g.index),
		Username: fmt.Sprintf("%s.%s%d", firstName, lastName, g.index),
	}
}

// RealisticDataGenerator generates realistic-looking test data
type RealisticDataGenerator struct {
	appGen     *AppGenerator
	spaceGen   *SpaceGenerator
	orgGen     *OrganizationGenerator
	routeGen   *RouteGenerator
	serviceGen *ServiceInstanceGenerator
	userGen    *UserGenerator
}

// NewRealisticDataGenerator creates a new realistic data generator
func NewRealisticDataGenerator() *RealisticDataGenerator {
	return &RealisticDataGenerator{
		appGen:     NewAppGenerator(),
		spaceGen:   NewSpaceGenerator(),
		orgGen:     NewOrganizationGenerator(),
		routeGen:   NewRouteGenerator(),
		serviceGen: NewServiceInstanceGenerator(),
		userGen:    NewUserGenerator(),
	}
}

// GenerateCompleteEnvironment creates a full CF environment
func (g *RealisticDataGenerator) GenerateCompleteEnvironment() Environment {
	// Generate org
	org := g.orgGen.Generate()

	// Generate 2-4 spaces
	spaceCount := rand.Intn(3) + 2
	spaces := make([]models.Space, spaceCount)
	for i := 0; i < spaceCount; i++ {
		spaces[i] = g.spaceGen.Generate()
		spaces[i].Organization = org.OrganizationFields
	}

	// Generate 5-15 apps across spaces
	appCount := rand.Intn(11) + 5
	apps := make([]models.Application, appCount)
	for i := 0; i < appCount; i++ {
		apps[i] = g.appGen.Generate()
	}

	// Generate 3-8 routes
	routeCount := rand.Intn(6) + 3
	routes := make([]models.Route, routeCount)
	for i := 0; i < routeCount; i++ {
		routes[i] = g.routeGen.Generate()
	}

	// Generate 2-5 services
	serviceCount := rand.Intn(4) + 2
	services := make([]models.ServiceInstance, serviceCount)
	for i := 0; i < serviceCount; i++ {
		services[i] = g.serviceGen.Generate()
	}

	// Generate 3-7 users
	userCount := rand.Intn(5) + 3
	users := make([]models.UserFields, userCount)
	for i := 0; i < userCount; i++ {
		users[i] = g.userGen.Generate()
	}

	return Environment{
		Organization: org,
		Spaces:       spaces,
		Applications: apps,
		Routes:       routes,
		Services:     services,
		Users:        users,
	}
}

// Environment represents a complete CF environment
type Environment struct {
	Organization models.Organization
	Spaces       []models.Space
	Applications []models.Application
	Routes       []models.Route
	Services     []models.ServiceInstance
	Users        []models.UserFields
}

// RandomString generates a random string of given length
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// RandomGUID generates a random GUID-like string
func RandomGUID() string {
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		RandomString(8),
		RandomString(4),
		RandomString(4),
		RandomString(4),
		RandomString(12),
	)
}

// RandomChoice picks a random element from a slice
func RandomChoice(choices []string) string {
	if len(choices) == 0 {
		return ""
	}
	return choices[rand.Intn(len(choices))]
}

// RandomInt generates a random integer between min and max (inclusive)
func RandomInt(min, max int) int {
	return rand.Intn(max-min+1) + min
}

// RandomBool generates a random boolean
func RandomBool() bool {
	return rand.Intn(2) == 1
}

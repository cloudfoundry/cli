package models_test

import (
	"fmt"
	"github.com/cloudfoundry/cli/cf/models"
)

// ExampleRoute_URL demonstrates how to generate a URL from a route
func ExampleRoute_URL() {
	route := models.Route{
		Host: "my-app",
		Domain: models.DomainFields{
			Name: "example.com",
		},
	}

	fmt.Println(route.URL())
	// Output: my-app.example.com
}

// ExampleRoute_URL_noHost demonstrates URL generation without a host
func ExampleRoute_URL_noHost() {
	route := models.Route{
		Host: "",
		Domain: models.DomainFields{
			Name: "example.com",
		},
	}

	fmt.Println(route.URL())
	// Output: example.com
}

// ExampleDomainFields_UrlForHost demonstrates domain URL generation
func ExampleDomainFields_UrlForHost() {
	domain := models.DomainFields{
		Name: "cfapps.io",
	}

	fmt.Println(domain.UrlForHost("my-app"))
	// Output: my-app.cfapps.io
}

// ExampleNewQuotaFields demonstrates quota creation
func ExampleNewQuotaFields() {
	quota := models.NewQuotaFields(
		"default",  // name
		1024,       // memory limit (MB)
		512,        // instance memory limit (MB)
		10,         // routes limit
		5,          // services limit
		true,       // non-basic services allowed
	)

	fmt.Printf("Quota: %s, Memory: %dMB, Routes: %d\n",
		quota.Name, quota.MemoryLimit, quota.RoutesLimit)
	// Output: Quota: default, Memory: 1024MB, Routes: 10
}

// ExampleServiceInstance_IsUserProvided demonstrates checking user-provided services
func ExampleServiceInstance_IsUserProvided() {
	// User-provided service (no plan guid)
	userProvidedService := models.ServiceInstance{
		ServiceInstanceFields: models.ServiceInstanceFields{
			Name: "my-user-provided-service",
		},
		ServicePlan: models.ServicePlanFields{
			Guid: "",
		},
	}

	// Regular service (has plan guid)
	regularService := models.ServiceInstance{
		ServiceInstanceFields: models.ServiceInstanceFields{
			Name: "my-service",
		},
		ServicePlan: models.ServicePlanFields{
			Guid: "plan-guid",
		},
	}

	fmt.Printf("User-provided: %v\n", userProvidedService.IsUserProvided())
	fmt.Printf("Regular service: %v\n", regularService.IsUserProvided())
	// Output: User-provided: true
	// Regular service: false
}

// ExampleApplication_HasRoute demonstrates route checking
func ExampleApplication_HasRoute() {
	app := models.Application{
		Routes: []models.RouteSummary{
			{Guid: "route-1"},
			{Guid: "route-2"},
		},
	}

	route := models.Route{Guid: "route-1"}
	fmt.Println(app.HasRoute(route))
	// Output: true
}

// ExampleAppParams_Merge demonstrates merging app parameters
func ExampleAppParams_Merge() {
	name := "my-app"
	memory := int64(512)

	baseParams := models.AppParams{
		Name:   &name,
		Memory: &memory,
	}

	newMemory := int64(1024)
	updateParams := models.AppParams{
		Memory: &newMemory,
	}

	baseParams.Merge(&updateParams)

	fmt.Printf("Name: %s, Memory: %dMB\n", *baseParams.Name, *baseParams.Memory)
	// Output: Name: my-app, Memory: 1024MB
}

# Cloud Foundry CLI Testing Guide

This guide describes the testing patterns, tools, and best practices used in the Cloud Foundry CLI codebase.

## Table of Contents

- [Testing Framework](#testing-framework)
- [Running Tests](#running-tests)
- [Test Organization](#test-organization)
- [Testing Patterns](#testing-patterns)
- [Test Helpers](#test-helpers)
- [Test Fixtures](#test-fixtures)
- [Best Practices](#best-practices)
- [Coverage Analysis](#coverage-analysis)

## Testing Framework

The CF CLI uses the following testing frameworks:

### Ginkgo & Gomega

- **Ginkgo**: BDD-style testing framework
- **Gomega**: Matcher/assertion library

```go
var _ = Describe("MyComponent", func() {
    var component MyComponent

    BeforeEach(func() {
        component = NewMyComponent()
    })

    It("does something", func() {
        result := component.DoSomething()
        Expect(result).To(Equal("expected value"))
    })
})
```

### Standard Go Testing

- Used for benchmarks and property-based tests
- `testing.B` for benchmark tests
- `testing/quick` for property-based tests

## Running Tests

### Run all tests

```bash
# Using make
make test

# Using ginkgo directly
ginkgo -r

# Using go test
go test ./...
```

### Run specific package tests

```bash
ginkgo cf/actors
go test ./cf/actors/...
```

### Run with coverage

```bash
ginkgo -r -cover
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run benchmarks

```bash
go test -bench=. ./generic
go test -bench=. ./words/generator
```

## Test Organization

### File Naming

- Unit tests: `*_test.go`
- Integration tests: `*_integration_test.go`
- Benchmark tests: `*_benchmark_test.go`
- Property tests: `property_test.go`
- Test suites: `*_suite_test.go`

### Package Naming

Tests use the `_test` suffix for black-box testing:

```go
package actors_test  // Not package actors
```

This ensures tests only access exported APIs.

## Testing Patterns

### 1. Unit Tests

Standard unit tests for individual components:

```go
var _ = Describe("RouteActor", func() {
    var (
        actor     actors.RouteActor
        routeRepo *fakes.FakeRouteRepository
    )

    BeforeEach(func() {
        routeRepo = &fakes.FakeRouteRepository{}
        actor = actors.NewRouteActor(ui, routeRepo)
    })

    Describe("FindOrCreateRoute", func() {
        Context("when the route exists", func() {
            It("returns the existing route", func() {
                // Test implementation
            })
        })
    })
})
```

### 2. Table-Driven Tests

Use table-driven tests for testing multiple scenarios:

```go
type testCase struct {
    description string
    input       string
    expected    string
}

testCases := []testCase{
    {
        description: "handles empty input",
        input:       "",
        expected:    "",
    },
    {
        description: "handles normal input",
        input:       "test",
        expected:    "TEST",
    },
}

for _, tc := range testCases {
    testCase := tc  // Capture range variable
    It(testCase.description, func() {
        result := Transform(testCase.input)
        Expect(result).To(Equal(testCase.expected))
    })
}
```

### 3. Integration Tests

Test complete workflows across multiple components:

```go
Describe("Complete Route Workflow", func() {
    It("creates route, binds to app, and unbinds successfully", func() {
        // Step 1: Create route
        route := actor.FindOrCreateRoute(hostname, domain)

        // Step 2: Bind route to app
        actor.BindRoute(app, route)

        // Step 3: Verify binding
        Expect(routeRepo.BoundRouteGuid).To(Equal(route.Guid))

        // Step 4: Unbind all routes
        actor.UnbindAll(app)
    })
})
```

### 4. Benchmark Tests

Measure performance of critical operations:

```go
func BenchmarkMerge_LargeMaps(b *testing.B) {
    map1 := createLargeMap(100)
    map2 := createLargeMap(100)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        Merge(map1, map2)
    }
}

func BenchmarkParallelOperation(b *testing.B) {
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            DoOperation()
        }
    })
}
```

### 5. Property-Based Tests

Test invariants with randomly generated inputs:

```go
func TestMergeIsIdempotent(t *testing.T) {
    f := func(key, val string) bool {
        m := NewMap(map[interface{}]interface{}{key: val})

        result1 := Merge(m, m)
        result2 := Merge(m, m)

        // Merging twice should produce same result
        return result1.Get(key) == result2.Get(key)
    }

    if err := quick.Check(f, nil); err != nil {
        t.Error(err)
    }
}
```

### 6. Example Tests

Document API usage with runnable examples:

```go
// ExampleRoute_URL demonstrates how to generate a URL from a route
func ExampleRoute_URL() {
    route := models.Route{
        Host: "my-app",
        Domain: models.DomainFields{Name: "example.com"},
    }

    fmt.Println(route.URL())
    // Output: my-app.example.com
}
```

## Test Helpers

### Model Makers

Located in `testhelpers/models/`, these provide convenient functions for creating test data:

```go
// Create application with defaults
app := helpers.MakeApplication("my-app")

// Create application with custom options
app := helpers.MakeApplication("my-app",
    helpers.WithMemory(512),
    helpers.WithInstances(3),
    helpers.WithRoutes(route1, route2),
)

// Create other resources
domain := helpers.MakeDomain("example.com", true)
route := helpers.MakeRoute("my-app", "example.com")
space := helpers.MakeSpace("development")
org := helpers.MakeOrganization("my-org")
```

### Using Model Makers

The functional options pattern allows flexible test data creation:

```go
func MakeApplication(name string, opts ...func(*models.Application)) models.Application {
    app := models.Application{
        ApplicationFields: models.ApplicationFields{
            Guid:  name + "-guid",
            Name:  name,
            State: "STARTED",
        },
    }

    for _, opt := range opts {
        opt(&app)
    }

    return app
}

func WithMemory(memory int64) func(*models.Application) {
    return func(app *models.Application) {
        app.Memory = memory
    }
}
```

## Test Fixtures

### JSON Fixtures

Located in `testhelpers/fixtures/`, these provide reusable API response templates:

```go
import "github.com/cloudfoundry/cli/testhelpers/fixtures"

// Get fixture as string
appJSON := fixtures.GetApplicationFixture()
spaceJSON := fixtures.GetSpaceFixture()
orgJSON := fixtures.GetOrganizationFixture()

// Use in tests
var data map[string]interface{}
json.Unmarshal([]byte(appJSON), &data)
```

### Available Fixtures

- `GetApplicationFixture()` - CF application response
- `GetSpaceFixture()` - CF space response
- `GetOrganizationFixture()` - CF organization response
- `GetServiceInstanceFixture()` - Service instance with credentials
- `GetRouteFixture()` - Route response
- `GetDomainFixture()` - Domain response
- `GetBuildpackFixture()` - Buildpack response
- `GetErrorResponseFixture()` - Error response
- `GetMultipleAppsFixture()` - Paginated list response

## Best Practices

### 1. Test Structure

- Use `Describe` for grouping related tests
- Use `Context` for different scenarios
- Use `It` for individual test cases
- Use `BeforeEach` for test setup
- Use `AfterEach` for cleanup

```go
var _ = Describe("Component", func() {
    Describe("Method", func() {
        Context("when condition A", func() {
            It("does X", func() {
                // Test
            })
        })

        Context("when condition B", func() {
            It("does Y", func() {
                // Test
            })
        })
    })
})
```

### 2. Use Fakes, Not Mocks

The codebase uses counterfeiter-generated fakes:

```go
// Good: Using fakes
routeRepo := &fakes.FakeRouteRepository{}
routeRepo.FindByHostAndDomainReturns.Route = route
routeRepo.FindByHostAndDomainReturns.Error = nil

actor.FindOrCreateRoute(hostname, domain)

Expect(routeRepo.BoundRouteGuid).To(Equal(route.Guid))
```

### 3. Descriptive Test Names

```go
// Good
It("creates a new route when route does not exist", func() {})
It("returns existing route when route already exists", func() {})

// Bad
It("works", func() {})
It("test route creation", func() {})
```

### 4. Test One Thing

Each test should verify one behavior:

```go
// Good: Focused test
It("sets the buildpack URL", func() {
    params.BuildpackUrl = &buildpack
    Expect(*params.BuildpackUrl).To(Equal(buildpack))
})

// Bad: Testing multiple things
It("sets all the fields", func() {
    // Tests 10 different fields
})
```

### 5. Use Test Helpers

Reduce boilerplate with helpers:

```go
// Good: Using helpers
app := helpers.MakeApplication("my-app", helpers.WithMemory(512))

// Bad: Manual setup
app := models.Application{
    ApplicationFields: models.ApplicationFields{
        Guid:   "my-app-guid",
        Name:   "my-app",
        State:  "STARTED",
        Memory: 512,
        // ... many more fields
    },
}
```

### 6. Avoid Test Interdependence

Tests should be independent and runnable in any order:

```go
// Good: Independent test
BeforeEach(func() {
    repo = &fakes.FakeRepository{}
    component = NewComponent(repo)
})

// Bad: Depends on previous test
var component Component  // Shared across tests
```

### 7. Test Error Cases

Always test both success and error paths:

```go
Context("when the API returns an error", func() {
    BeforeEach(func() {
        repo.CreateReturns.Error = errors.New("API error")
    })

    It("displays the error message", func() {
        actor.Create()
        Expect(ui.Outputs).To(ContainElement(ContainSubstring("FAILED")))
    })
})
```

### 8. Use Matchers Effectively

Gomega provides many useful matchers:

```go
// Equality
Expect(value).To(Equal(expected))
Expect(value).NotTo(Equal(other))

// Strings
Expect(str).To(ContainSubstring("text"))
Expect(str).To(HavePrefix("start"))
Expect(str).To(MatchRegexp(`\d+`))

// Collections
Expect(slice).To(ContainElement(item))
Expect(slice).To(HaveLen(3))
Expect(slice).To(BeEmpty())

// Errors
Expect(err).To(HaveOccurred())
Expect(err).NotTo(HaveOccurred())

// Types
Expect(value).To(BeNil())
Expect(value).To(BeTrue())
```

## Coverage Analysis

### Generate Coverage Report

```bash
# Generate coverage profile
go test -coverprofile=coverage.out ./...

# View HTML coverage report
go tool cover -html=coverage.out

# View coverage by package
go tool cover -func=coverage.out
```

### Coverage Goals

- Critical packages (errors, actors, commands): **>80%**
- Model packages: **>70%**
- Utility packages: **>60%**

### Finding Untested Code

```bash
# Find packages with low coverage
go test -cover ./... | grep -v "100.0%"

# Generate detailed coverage
ginkgo -r -cover -coverprofile=coverage.out
```

## Additional Resources

- [Ginkgo Documentation](https://onsi.github.io/ginkgo/)
- [Gomega Documentation](https://onsi.github.io/gomega/)
- [Go Testing Package](https://golang.org/pkg/testing/)
- [Table-Driven Tests in Go](https://github.com/golang/go/wiki/TableDrivenTests)
- [Property-Based Testing](https://golang.org/pkg/testing/quick/)

## Contributing

When adding new tests:

1. Follow existing patterns in the codebase
2. Use test helpers and fixtures when appropriate
3. Write descriptive test names
4. Test both success and error cases
5. Aim for high coverage of critical paths
6. Run tests locally before committing
7. Consider adding benchmarks for performance-critical code
8. Document complex test setups

---

For questions or suggestions about testing, please open an issue or submit a pull request.

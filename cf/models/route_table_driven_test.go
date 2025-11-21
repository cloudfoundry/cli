package models_test

import (
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Route Table-Driven Tests", func() {
	Describe("Route.URL", func() {
		type routeURLTestCase struct {
			description  string
			host         string
			domainName   string
			expectedURL  string
		}

		testCases := []routeURLTestCase{
			{
				description: "full URL with host and domain",
				host:        "my-app",
				domainName:  "example.com",
				expectedURL: "my-app.example.com",
			},
			{
				description: "domain only when host is empty",
				host:        "",
				domainName:  "example.com",
				expectedURL: "example.com",
			},
			{
				description: "subdomain in host",
				host:        "api.staging",
				domainName:  "example.com",
				expectedURL: "api.staging.example.com",
			},
			{
				description: "host with dashes",
				host:        "my-app-name",
				domainName:  "cfapps.io",
				expectedURL: "my-app-name.cfapps.io",
			},
			{
				description: "numeric host",
				host:        "app123",
				domainName:  "test.com",
				expectedURL: "app123.test.com",
			},
		}

		for _, tc := range testCases {
			// Create a local copy for the closure
			testCase := tc

			It(testCase.description, func() {
				route := models.Route{
					Host: testCase.host,
					Domain: models.DomainFields{
						Name: testCase.domainName,
					},
				}

				Expect(route.URL()).To(Equal(testCase.expectedURL))
			})
		}
	})

	Describe("DomainFields.UrlForHost", func() {
		type domainURLTestCase struct {
			description string
			domainName  string
			host        string
			expectedURL string
		}

		testCases := []domainURLTestCase{
			{
				description: "with host",
				domainName:  "example.com",
				host:        "app",
				expectedURL: "app.example.com",
			},
			{
				description: "without host",
				domainName:  "example.com",
				host:        "",
				expectedURL: "example.com",
			},
			{
				description: "subdomain host",
				domainName:  "apps.internal",
				host:        "my-service",
				expectedURL: "my-service.apps.internal",
			},
			{
				description: "complex TLD",
				domainName:  "cfapps.co.uk",
				host:        "myapp",
				expectedURL: "myapp.cfapps.co.uk",
			},
		}

		for _, tc := range testCases {
			testCase := tc

			It(testCase.description, func() {
				domain := models.DomainFields{
					Name: testCase.domainName,
				}

				Expect(domain.UrlForHost(testCase.host)).To(Equal(testCase.expectedURL))
			})
		}
	})
})

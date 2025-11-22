package models_test

import (
	"github.com/cloudfoundry/cli/cf/models"
	helpers "github.com/cloudfoundry/cli/testhelpers/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Model Makers", func() {
	Describe("MakeApplication", func() {
		It("creates an application with defaults", func() {
			app := helpers.MakeApplication("my-app")

			Expect(app.Name).To(Equal("my-app"))
			Expect(app.Guid).To(Equal("my-app-guid"))
			Expect(app.State).To(Equal("STARTED"))
			Expect(app.InstanceCount).To(Equal(1))
			Expect(app.Memory).To(Equal(int64(256)))
		})

		It("applies custom options", func() {
			app := helpers.MakeApplication("my-app",
				helpers.WithMemory(512),
				helpers.WithInstances(3),
				helpers.WithState("STOPPED"),
			)

			Expect(app.Memory).To(Equal(int64(512)))
			Expect(app.InstanceCount).To(Equal(3))
			Expect(app.State).To(Equal("STOPPED"))
		})

		It("applies routes", func() {
			routes := []models.RouteSummary{
				{Guid: "route-1", Host: "host1"},
				{Guid: "route-2", Host: "host2"},
			}

			app := helpers.MakeApplication("my-app",
				helpers.WithRoutes(routes...),
			)

			Expect(len(app.Routes)).To(Equal(2))
			Expect(app.Routes[0].Guid).To(Equal("route-1"))
		})
	})

	Describe("MakeRoute", func() {
		It("creates a route", func() {
			route := helpers.MakeRoute("my-app", "example.com")

			Expect(route.Host).To(Equal("my-app"))
			Expect(route.Domain.Name).To(Equal("example.com"))
			Expect(route.URL()).To(Equal("my-app.example.com"))
		})
	})

	Describe("MakeSpace", func() {
		It("creates a space", func() {
			space := helpers.MakeSpace("my-space")

			Expect(space.Name).To(Equal("my-space"))
			Expect(space.Guid).To(Equal("my-space-guid"))
		})
	})

	Describe("MakeOrganization", func() {
		It("creates an organization", func() {
			org := helpers.MakeOrganization("my-org")

			Expect(org.Name).To(Equal("my-org"))
			Expect(org.Guid).To(Equal("my-org-guid"))
		})
	})

	Describe("MakeServiceInstance", func() {
		It("creates a service instance", func() {
			service := helpers.MakeServiceInstance("my-service")

			Expect(service.Name).To(Equal("my-service"))
			Expect(service.Guid).To(Equal("my-service-guid"))
		})
	})

	Describe("MakeAppInstance", func() {
		It("creates an app instance with metrics", func() {
			instance := helpers.MakeAppInstance(models.InstanceRunning)

			Expect(instance.State).To(Equal(models.InstanceRunning))
			Expect(instance.CpuUsage).To(BeNumerically(">", 0))
			Expect(instance.MemQuota).To(BeNumerically(">", 0))
		})
	})

	Describe("MakeQuota", func() {
		It("creates a quota", func() {
			quota := helpers.MakeQuota("default", 1024, 10, 5)

			Expect(quota.Name).To(Equal("default"))
			Expect(quota.MemoryLimit).To(Equal(int64(1024)))
			Expect(quota.RoutesLimit).To(Equal(10))
			Expect(quota.ServicesLimit).To(Equal(5))
		})
	})

	Describe("MakeDomain", func() {
		It("creates a shared domain", func() {
			domain := helpers.MakeDomain("example.com", true)

			Expect(domain.Name).To(Equal("example.com"))
			Expect(domain.Shared).To(BeTrue())
			Expect(domain.OwningOrganizationGuid).To(BeEmpty())
		})

		It("creates a private domain", func() {
			domain := helpers.MakeDomain("private.com", false)

			Expect(domain.Name).To(Equal("private.com"))
			Expect(domain.Shared).To(BeFalse())
			Expect(domain.OwningOrganizationGuid).ToNot(BeEmpty())
		})
	})

	Describe("MakeBuildpack", func() {
		It("creates a buildpack", func() {
			buildpack := helpers.MakeBuildpack("ruby_buildpack", 1, true)

			Expect(buildpack.Name).To(Equal("ruby_buildpack"))
			Expect(*buildpack.Position).To(Equal(1))
			Expect(*buildpack.Enabled).To(BeTrue())
		})
	})

	Describe("MakeSecurityGroup", func() {
		It("creates a security group with rules", func() {
			rules := []map[string]interface{}{
				{"protocol": "tcp", "destination": "10.0.0.0/8"},
			}

			sg := helpers.MakeSecurityGroup("web-sg", rules)

			Expect(sg.Name).To(Equal("web-sg"))
			Expect(len(sg.Rules)).To(Equal(1))
		})
	})

	Describe("MakeStack", func() {
		It("creates a stack", func() {
			stack := helpers.MakeStack("cflinuxfs3", "Cloud Foundry Linux-based filesystem")

			Expect(stack.Name).To(Equal("cflinuxfs3"))
			Expect(stack.Description).To(Equal("Cloud Foundry Linux-based filesystem"))
		})
	})

	Describe("MakeUser", func() {
		It("creates a regular user", func() {
			user := helpers.MakeUser("john@example.com", false)

			Expect(user.Username).To(Equal("john@example.com"))
			Expect(user.IsAdmin).To(BeFalse())
		})

		It("creates an admin user", func() {
			user := helpers.MakeUser("admin@example.com", true)

			Expect(user.Username).To(Equal("admin@example.com"))
			Expect(user.IsAdmin).To(BeTrue())
		})
	})
})

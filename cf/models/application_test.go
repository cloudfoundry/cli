package models_test

import (
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application", func() {
	Describe("HasRoute", func() {
		var app models.Application

		BeforeEach(func() {
			app = models.Application{
				Routes: []models.RouteSummary{
					{Guid: "route-1-guid", Host: "host1"},
					{Guid: "route-2-guid", Host: "host2"},
					{Guid: "route-3-guid", Host: "host3"},
				},
			}
		})

		It("returns true when the app has the route", func() {
			route := models.Route{Guid: "route-2-guid"}
			Expect(app.HasRoute(route)).To(BeTrue())
		})

		It("returns false when the app does not have the route", func() {
			route := models.Route{Guid: "nonexistent-guid"}
			Expect(app.HasRoute(route)).To(BeFalse())
		})

		It("returns false when the app has no routes", func() {
			app.Routes = []models.RouteSummary{}
			route := models.Route{Guid: "any-guid"}
			Expect(app.HasRoute(route)).To(BeFalse())
		})

		It("returns true for the first route", func() {
			route := models.Route{Guid: "route-1-guid"}
			Expect(app.HasRoute(route)).To(BeTrue())
		})

		It("returns true for the last route", func() {
			route := models.Route{Guid: "route-3-guid"}
			Expect(app.HasRoute(route)).To(BeTrue())
		})
	})

	Describe("ToParams", func() {
		It("converts application to params", func() {
			app := models.Application{
				ApplicationFields: models.ApplicationFields{
					Guid:          "app-guid",
					Name:          "my-app",
					BuildpackUrl:  "http://buildpack.url",
					Command:       "start command",
					DiskQuota:     1024,
					InstanceCount: 3,
					Memory:        512,
					State:         "started",
					SpaceGuid:     "space-guid",
					EnvironmentVars: map[string]interface{}{
						"KEY1": "value1",
						"KEY2": "value2",
					},
				},
				Stack: &models.Stack{
					Guid: "stack-guid",
					Name: "cflinuxfs3",
				},
			}

			params := app.ToParams()

			Expect(*params.Guid).To(Equal("app-guid"))
			Expect(*params.Name).To(Equal("my-app"))
			Expect(*params.BuildpackUrl).To(Equal("http://buildpack.url"))
			Expect(*params.Command).To(Equal("start command"))
			Expect(*params.DiskQuota).To(Equal(int64(1024)))
			Expect(*params.InstanceCount).To(Equal(3))
			Expect(*params.Memory).To(Equal(int64(512)))
			Expect(*params.State).To(Equal("STARTED")) // Should be uppercase
			Expect(*params.SpaceGuid).To(Equal("space-guid"))
			Expect(*params.StackGuid).To(Equal("stack-guid"))
			Expect(*params.EnvironmentVars).To(Equal(map[string]interface{}{
				"KEY1": "value1",
				"KEY2": "value2",
			}))
		})

		It("uppercases the state", func() {
			app := models.Application{
				ApplicationFields: models.ApplicationFields{
					State: "stopped",
				},
			}

			params := app.ToParams()
			Expect(*params.State).To(Equal("STOPPED"))
		})

		It("handles app without stack", func() {
			app := models.Application{
				ApplicationFields: models.ApplicationFields{
					Guid: "app-guid",
					Name: "my-app",
				},
				Stack: nil,
			}

			params := app.ToParams()
			Expect(params.StackGuid).To(BeNil())
		})

		It("handles empty environment variables", func() {
			app := models.Application{
				ApplicationFields: models.ApplicationFields{
					Guid:            "app-guid",
					EnvironmentVars: map[string]interface{}{},
				},
			}

			params := app.ToParams()
			Expect(*params.EnvironmentVars).To(Equal(map[string]interface{}{}))
		})
	})

	Describe("AppParams", func() {
		Describe("Merge", func() {
			var baseParams, otherParams *models.AppParams

			BeforeEach(func() {
				buildpackUrl := "http://original.buildpack"
				command := "original command"
				diskQuota := int64(1024)
				memory := int64(512)
				name := "original-name"
				instanceCount := 2

				baseParams = &models.AppParams{
					BuildpackUrl:  &buildpackUrl,
					Command:       &command,
					DiskQuota:     &diskQuota,
					Memory:        &memory,
					Name:          &name,
					InstanceCount: &instanceCount,
				}
			})

			It("merges buildpack URL", func() {
				newBuildpack := "http://new.buildpack"
				otherParams = &models.AppParams{
					BuildpackUrl: &newBuildpack,
				}

				baseParams.Merge(otherParams)
				Expect(*baseParams.BuildpackUrl).To(Equal("http://new.buildpack"))
			})

			It("merges command", func() {
				newCommand := "new command"
				otherParams = &models.AppParams{
					Command: &newCommand,
				}

				baseParams.Merge(otherParams)
				Expect(*baseParams.Command).To(Equal("new command"))
			})

			It("merges disk quota", func() {
				newDiskQuota := int64(2048)
				otherParams = &models.AppParams{
					DiskQuota: &newDiskQuota,
				}

				baseParams.Merge(otherParams)
				Expect(*baseParams.DiskQuota).To(Equal(int64(2048)))
			})

			It("merges memory", func() {
				newMemory := int64(1024)
				otherParams = &models.AppParams{
					Memory: &newMemory,
				}

				baseParams.Merge(otherParams)
				Expect(*baseParams.Memory).To(Equal(int64(1024)))
			})

			It("merges name", func() {
				newName := "new-name"
				otherParams = &models.AppParams{
					Name: &newName,
				}

				baseParams.Merge(otherParams)
				Expect(*baseParams.Name).To(Equal("new-name"))
			})

			It("merges instance count", func() {
				newInstanceCount := 5
				otherParams = &models.AppParams{
					InstanceCount: &newInstanceCount,
				}

				baseParams.Merge(otherParams)
				Expect(*baseParams.InstanceCount).To(Equal(5))
			})

			It("merges domains", func() {
				domains := []string{"domain1.com", "domain2.com"}
				otherParams = &models.AppParams{
					Domains: &domains,
				}

				baseParams.Merge(otherParams)
				Expect(*baseParams.Domains).To(Equal([]string{"domain1.com", "domain2.com"}))
			})

			It("merges environment variables", func() {
				envVars := map[string]interface{}{
					"KEY": "value",
				}
				otherParams = &models.AppParams{
					EnvironmentVars: &envVars,
				}

				baseParams.Merge(otherParams)
				Expect(*baseParams.EnvironmentVars).To(Equal(map[string]interface{}{
					"KEY": "value",
				}))
			})

			It("merges health check timeout", func() {
				timeout := 120
				otherParams = &models.AppParams{
					HealthCheckTimeout: &timeout,
				}

				baseParams.Merge(otherParams)
				Expect(*baseParams.HealthCheckTimeout).To(Equal(120))
			})

			It("merges hosts", func() {
				hosts := []string{"host1", "host2"}
				otherParams = &models.AppParams{
					Hosts: &hosts,
				}

				baseParams.Merge(otherParams)
				Expect(*baseParams.Hosts).To(Equal([]string{"host1", "host2"}))
			})

			It("merges guid", func() {
				guid := "new-guid"
				otherParams = &models.AppParams{
					Guid: &guid,
				}

				baseParams.Merge(otherParams)
				Expect(*baseParams.Guid).To(Equal("new-guid"))
			})

			It("merges path", func() {
				path := "/new/path"
				otherParams = &models.AppParams{
					Path: &path,
				}

				baseParams.Merge(otherParams)
				Expect(*baseParams.Path).To(Equal("/new/path"))
			})

			It("merges services to bind", func() {
				services := []string{"service1", "service2"}
				otherParams = &models.AppParams{
					ServicesToBind: &services,
				}

				baseParams.Merge(otherParams)
				Expect(*baseParams.ServicesToBind).To(Equal([]string{"service1", "service2"}))
			})

			It("merges space GUID", func() {
				spaceGuid := "new-space-guid"
				otherParams = &models.AppParams{
					SpaceGuid: &spaceGuid,
				}

				baseParams.Merge(otherParams)
				Expect(*baseParams.SpaceGuid).To(Equal("new-space-guid"))
			})

			It("merges stack GUID", func() {
				stackGuid := "new-stack-guid"
				otherParams = &models.AppParams{
					StackGuid: &stackGuid,
				}

				baseParams.Merge(otherParams)
				Expect(*baseParams.StackGuid).To(Equal("new-stack-guid"))
			})

			It("merges stack name", func() {
				stackName := "cflinuxfs4"
				otherParams = &models.AppParams{
					StackName: &stackName,
				}

				baseParams.Merge(otherParams)
				Expect(*baseParams.StackName).To(Equal("cflinuxfs4"))
			})

			It("merges state", func() {
				state := "STOPPED"
				otherParams = &models.AppParams{
					State: &state,
				}

				baseParams.Merge(otherParams)
				Expect(*baseParams.State).To(Equal("STOPPED"))
			})

			It("ORs NoRoute flag", func() {
				baseParams.NoRoute = false
				otherParams = &models.AppParams{
					NoRoute: true,
				}

				baseParams.Merge(otherParams)
				Expect(baseParams.NoRoute).To(BeTrue())
			})

			It("ORs NoHostname flag", func() {
				baseParams.NoHostname = false
				otherParams = &models.AppParams{
					NoHostname: true,
				}

				baseParams.Merge(otherParams)
				Expect(baseParams.NoHostname).To(BeTrue())
			})

			It("ORs UseRandomHostname flag", func() {
				baseParams.UseRandomHostname = false
				otherParams = &models.AppParams{
					UseRandomHostname: true,
				}

				baseParams.Merge(otherParams)
				Expect(baseParams.UseRandomHostname).To(BeTrue())
			})

			It("does not overwrite with nil values", func() {
				originalName := "original-name"
				baseParams = &models.AppParams{
					Name: &originalName,
				}
				otherParams = &models.AppParams{
					Name: nil,
				}

				baseParams.Merge(otherParams)
				Expect(*baseParams.Name).To(Equal("original-name"))
			})

			It("merges all fields at once", func() {
				newBuildpack := "http://new.buildpack"
				newCommand := "new command"
				newMemory := int64(1024)
				newName := "new-name"

				otherParams = &models.AppParams{
					BuildpackUrl: &newBuildpack,
					Command:      &newCommand,
					Memory:       &newMemory,
					Name:         &newName,
					NoRoute:      true,
				}

				baseParams.Merge(otherParams)

				Expect(*baseParams.BuildpackUrl).To(Equal("http://new.buildpack"))
				Expect(*baseParams.Command).To(Equal("new command"))
				Expect(*baseParams.Memory).To(Equal(int64(1024)))
				Expect(*baseParams.Name).To(Equal("new-name"))
				Expect(baseParams.NoRoute).To(BeTrue())
			})
		})

		Describe("IsEmpty", func() {
			It("returns true for empty params", func() {
				params := models.AppParams{}
				Expect(params.IsEmpty()).To(BeTrue())
			})

			It("returns false when Name is set", func() {
				name := "app-name"
				params := models.AppParams{
					Name: &name,
				}
				Expect(params.IsEmpty()).To(BeFalse())
			})

			It("returns false when Memory is set", func() {
				memory := int64(512)
				params := models.AppParams{
					Memory: &memory,
				}
				Expect(params.IsEmpty()).To(BeFalse())
			})

			It("returns false when NoRoute is true", func() {
				params := models.AppParams{
					NoRoute: true,
				}
				Expect(params.IsEmpty()).To(BeFalse())
			})

			It("returns false when any field is set", func() {
				command := "start-command"
				params := models.AppParams{
					Command: &command,
				}
				Expect(params.IsEmpty()).To(BeFalse())
			})
		})

		Describe("IsHostEmpty", func() {
			It("returns true when Hosts is nil", func() {
				params := models.AppParams{
					Hosts: nil,
				}
				Expect(params.IsHostEmpty()).To(BeTrue())
			})

			It("returns true when Hosts is empty slice", func() {
				hosts := []string{}
				params := models.AppParams{
					Hosts: &hosts,
				}
				Expect(params.IsHostEmpty()).To(BeTrue())
			})

			It("returns false when Hosts has one element", func() {
				hosts := []string{"host1"}
				params := models.AppParams{
					Hosts: &hosts,
				}
				Expect(params.IsHostEmpty()).To(BeFalse())
			})

			It("returns false when Hosts has multiple elements", func() {
				hosts := []string{"host1", "host2", "host3"}
				params := models.AppParams{
					Hosts: &hosts,
				}
				Expect(params.IsHostEmpty()).To(BeFalse())
			})
		})
	})
})

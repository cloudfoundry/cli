package models_test

import (
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("More Models", func() {
	Describe("AppFileFields", func() {
		It("stores app file information", func() {
			file := models.AppFileFields{
				Path: "app.rb",
				Sha1: "abc123def456",
				Size: 1024,
			}

			Expect(file.Path).To(Equal("app.rb"))
			Expect(file.Sha1).To(Equal("abc123def456"))
			Expect(file.Size).To(Equal(int64(1024)))
		})

		It("handles nested paths", func() {
			file := models.AppFileFields{
				Path: "lib/utils/helper.js",
				Sha1: "xyz789",
				Size: 2048,
			}

			Expect(file.Path).To(Equal("lib/utils/helper.js"))
		})

		It("handles large files", func() {
			file := models.AppFileFields{
				Path: "large-file.bin",
				Sha1: "sha1-hash",
				Size: 10737418240, // 10GB
			}

			Expect(file.Size).To(Equal(int64(10737418240)))
		})

		It("handles zero-byte files", func() {
			file := models.AppFileFields{
				Path: ".gitkeep",
				Sha1: "da39a3ee5e6b4b0d3255bfef95601890afd80709", // SHA1 of empty file
				Size: 0,
			}

			Expect(file.Size).To(Equal(int64(0)))
		})
	})

	Describe("ServiceKeyFields", func() {
		It("stores service key fields", func() {
			key := models.ServiceKeyFields{
				Name:                "my-service-key",
				Guid:                "key-guid",
				Url:                 "/v2/service_keys/key-guid",
				ServiceInstanceGuid: "instance-guid",
				ServiceInstanceUrl:  "/v2/service_instances/instance-guid",
			}

			Expect(key.Name).To(Equal("my-service-key"))
			Expect(key.Guid).To(Equal("key-guid"))
			Expect(key.Url).To(Equal("/v2/service_keys/key-guid"))
			Expect(key.ServiceInstanceGuid).To(Equal("instance-guid"))
			Expect(key.ServiceInstanceUrl).To(Equal("/v2/service_instances/instance-guid"))
		})
	})

	Describe("ServiceKeyRequest", func() {
		It("stores service key creation request", func() {
			params := map[string]interface{}{
				"config": "value",
			}

			request := models.ServiceKeyRequest{
				Name:                "new-key",
				ServiceInstanceGuid: "instance-guid",
				Params:              params,
			}

			Expect(request.Name).To(Equal("new-key"))
			Expect(request.ServiceInstanceGuid).To(Equal("instance-guid"))
			Expect(request.Params).To(HaveKeyWithValue("config", "value"))
		})

		It("handles nil params", func() {
			request := models.ServiceKeyRequest{
				Name:                "simple-key",
				ServiceInstanceGuid: "instance-guid",
				Params:              nil,
			}

			Expect(request.Params).To(BeNil())
		})
	})

	Describe("ServiceKey", func() {
		It("stores service key with credentials", func() {
			fields := models.ServiceKeyFields{
				Name: "my-key",
				Guid: "key-guid",
			}

			credentials := map[string]interface{}{
				"username": "admin",
				"password": "secret",
				"uri":      "postgres://localhost:5432/db",
			}

			key := models.ServiceKey{
				Fields:      fields,
				Credentials: credentials,
			}

			Expect(key.Fields.Name).To(Equal("my-key"))
			Expect(key.Credentials).To(HaveKeyWithValue("username", "admin"))
			Expect(key.Credentials).To(HaveKeyWithValue("password", "secret"))
			Expect(key.Credentials).To(HaveKeyWithValue("uri", "postgres://localhost:5432/db"))
		})

		It("handles empty credentials", func() {
			key := models.ServiceKey{
				Fields:      models.ServiceKeyFields{Name: "key"},
				Credentials: map[string]interface{}{},
			}

			Expect(len(key.Credentials)).To(Equal(0))
		})

		It("handles complex credential structures", func() {
			credentials := map[string]interface{}{
				"hostname": "db.example.com",
				"ports": map[string]interface{}{
					"http":  8080,
					"https": 8443,
				},
				"users": []interface{}{"admin", "readonly"},
			}

			key := models.ServiceKey{
				Fields:      models.ServiceKeyFields{Name: "complex-key"},
				Credentials: credentials,
			}

			Expect(key.Credentials["hostname"]).To(Equal("db.example.com"))
			ports := key.Credentials["ports"].(map[string]interface{})
			Expect(ports["http"]).To(Equal(8080))
		})
	})

	Describe("PluginRepo", func() {
		It("stores plugin repository information", func() {
			repo := models.PluginRepo{
				Name: "CF-Community",
				Url:  "https://plugins.cloudfoundry.org",
			}

			Expect(repo.Name).To(Equal("CF-Community"))
			Expect(repo.Url).To(Equal("https://plugins.cloudfoundry.org"))
		})

		It("handles different URLs", func() {
			repo1 := models.PluginRepo{
				Name: "Official",
				Url:  "https://plugins.cloudfoundry.org",
			}
			repo2 := models.PluginRepo{
				Name: "Enterprise",
				Url:  "https://enterprise-plugins.example.com",
			}
			repo3 := models.PluginRepo{
				Name: "Local",
				Url:  "http://localhost:8080/plugins",
			}

			Expect(repo1.Url).To(Equal("https://plugins.cloudfoundry.org"))
			Expect(repo2.Url).To(Equal("https://enterprise-plugins.example.com"))
			Expect(repo3.Url).To(Equal("http://localhost:8080/plugins"))
		})

		It("handles empty values", func() {
			repo := models.PluginRepo{}

			Expect(repo.Name).To(BeEmpty())
			Expect(repo.Url).To(BeEmpty())
		})
	})
})

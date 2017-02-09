package manifest_test

import (
	"gopkg.in/yaml.v2"

	"bytes"

	. "code.cloudfoundry.org/cli/cf/manifest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("generate_manifest", func() {
	Describe("Save", func() {
		var (
			m App
			f *bytes.Buffer
		)

		BeforeEach(func() {
			m = NewGenerator()
			f = &bytes.Buffer{}
		})

		Context("when each application in the manifest has all required attributes, and the save path has been set", func() {
			BeforeEach(func() {
				m.Stack("app1", "stack-name")
				m.Memory("app1", 1024)
				m.Instances("app1", 2)
				m.DiskQuota("app1", 1024)
			})

			It("creates a top-level applications key", func() {
				err := m.Save(f)
				Expect(err).NotTo(HaveOccurred())
				ymanifest := getYaml(f)
				Expect(ymanifest.Applications).To(HaveLen(1))
			})

			It("includes required attributes", func() {
				err := m.Save(f)
				Expect(err).NotTo(HaveOccurred())
				applications := getYaml(f).Applications

				Expect(applications[0].Name).To(Equal("app1"))
				Expect(applications[0].Memory).To(Equal("1024M"))
				Expect(applications[0].DiskQuota).To(Equal("1024M"))
				Expect(applications[0].Stack).To(Equal("stack-name"))
				Expect(applications[0].Instances).To(Equal(2))
			})

			It("creates entries under the given app name", func() {
				m.Stack("app2", "stack-name")
				m.Memory("app2", 2048)
				m.Instances("app2", 3)
				m.DiskQuota("app2", 2048)
				m.HealthCheckType("app2", "some-health-check-type")
				m.HealthCheckHTTPEndpoint("app2", "/some-endpoint")
				m.Save(f)

				applications := getYaml(f).Applications

				Expect(applications[1].Name).To(Equal("app2"))
				Expect(applications[1].Memory).To(Equal("2048M"))
				Expect(applications[1].DiskQuota).To(Equal("2048M"))
				Expect(applications[1].Stack).To(Equal("stack-name"))
				Expect(applications[1].Instances).To(Equal(3))
				Expect(applications[1].HealthCheckType).To(Equal("some-health-check-type"))
				Expect(applications[1].HealthCheckHTTPEndpoint).To(Equal("/some-endpoint"))
			})

			Context("when an application has app-ports", func() {
				BeforeEach(func() {
					m.AppPorts("app1", []int{1111, 2222})
				})

				It("includes app-ports for that app", func() {
					err := m.Save(f)
					Expect(err).NotTo(HaveOccurred())
					applications := getYaml(f).Applications

					Expect(applications[0].AppPorts).To(Equal([]int{1111, 2222}))
				})
			})

			Context("when an application has services", func() {
				BeforeEach(func() {
					m.Service("app1", "service1")
					m.Service("app1", "service2")
					m.Service("app1", "service3")
				})

				It("includes services for that app", func() {
					err := m.Save(f)
					Expect(err).NotTo(HaveOccurred())
					contents := getYaml(f)
					application := contents.Applications[0]
					Expect(application.Services).To(ContainElement("service1"))
					Expect(application.Services).To(ContainElement("service2"))
					Expect(application.Services).To(ContainElement("service3"))
				})
			})

			Context("when an application has a buildpack", func() {
				BeforeEach(func() {
					m.BuildpackURL("app1", "buildpack")
				})

				It("includes the buildpack url for that app", func() {
					m.Save(f)
					contents := getYaml(f)
					application := contents.Applications[0]
					Expect(application.Buildpack).To(Equal("buildpack"))
				})
			})

			Context("when an application has a non-zero health check timeout", func() {
				BeforeEach(func() {
					m.HealthCheckTimeout("app1", 5)
				})

				It("includes the healthcheck timeout for that app", func() {
					err := m.Save(f)
					Expect(err).NotTo(HaveOccurred())
					contents := getYaml(f)
					application := contents.Applications[0]
					Expect(application.Timeout).To(Equal(5))
				})
			})

			Context("when an application has a start command", func() {
				BeforeEach(func() {
					m.StartCommand("app1", "start-command")
				})

				It("includes the start command for that app", func() {
					m.Save(f)
					contents := getYaml(f)
					application := contents.Applications[0]
					Expect(application.Command).To(Equal("start-command"))
				})
			})

			It("includes no-route when the application has no routes", func() {
				m.Save(f)
				contents := getYaml(f)
				application := contents.Applications[0]
				Expect(application.NoRoute).To(BeTrue())
			})

			Context("when an application has one route with both hostname, domain path", func() {
				BeforeEach(func() {
					m.Route("app1", "host-name", "domain-name", "/path", 0)
				})

				It("includes the route", func() {
					err := m.Save(f)
					Expect(err).NotTo(HaveOccurred())
					contents := getYaml(f)
					application := contents.Applications[0]
					Expect(application.Routes[0]["route"]).To(Equal("host-name.domain-name/path"))
				})

			})

			Context("when an application has one route without a hostname", func() {
				BeforeEach(func() {
					m.Route("app1", "", "domain-name", "", 0)
				})

				It("includes the domain", func() {
					err := m.Save(f)
					Expect(err).NotTo(HaveOccurred())
					contents := getYaml(f)
					application := contents.Applications[0]
					Expect(application.Routes[0]["route"]).To(Equal("domain-name"))
				})

			})

			Context("when an application has one tcp route", func() {
				BeforeEach(func() {
					m.Route("app1", "", "domain-name", "", 123)
				})

				It("includes the route", func() {
					err := m.Save(f)
					Expect(err).NotTo(HaveOccurred())
					contents := getYaml(f)
					application := contents.Applications[0]
					Expect(application.Routes[0]["route"]).To(Equal("domain-name:123"))
				})

			})

			Context("when an application has multiple routes", func() {
				BeforeEach(func() {
					m.Route("app1", "", "http-domain", "", 0)
					m.Route("app1", "host", "http-domain", "", 0)
					m.Route("app1", "host", "http-domain", "/path", 0)
					m.Route("app1", "", "tcp-domain", "", 123)
				})

				It("includes the route", func() {
					err := m.Save(f)
					Expect(err).NotTo(HaveOccurred())
					contents := getYaml(f)
					application := contents.Applications[0]
					Expect(application.Routes[0]["route"]).To(Equal("http-domain"))
					Expect(application.Routes[1]["route"]).To(Equal("host.http-domain"))
					Expect(application.Routes[2]["route"]).To(Equal("host.http-domain/path"))
					Expect(application.Routes[3]["route"]).To(Equal("tcp-domain:123"))
				})

			})

			Context("when multiple applications have multiple routes", func() {
				BeforeEach(func() {
					m.Stack("app2", "stack-name")
					m.Memory("app2", 1024)
					m.Instances("app2", 2)
					m.DiskQuota("app2", 1024)

					m.Route("app1", "", "http-domain", "", 0)
					m.Route("app1", "host", "http-domain", "", 0)
					m.Route("app1", "host", "http-domain", "/path", 0)
					m.Route("app1", "", "tcp-domain", "", 123)
					m.Route("app2", "", "http-domain", "", 0)
					m.Route("app2", "host", "http-domain", "", 0)
					m.Route("app2", "host", "http-domain", "/path", 0)
					m.Route("app2", "", "tcp-domain", "", 123)
				})

				It("includes the route", func() {
					err := m.Save(f)
					Expect(err).NotTo(HaveOccurred())
					contents := getYaml(f)
					application := contents.Applications[0]
					Expect(application.Routes[0]["route"]).To(Equal("http-domain"))
					Expect(application.Routes[1]["route"]).To(Equal("host.http-domain"))
					Expect(application.Routes[2]["route"]).To(Equal("host.http-domain/path"))
					Expect(application.Routes[3]["route"]).To(Equal("tcp-domain:123"))

					application = contents.Applications[1]
					Expect(application.Routes[0]["route"]).To(Equal("http-domain"))
					Expect(application.Routes[1]["route"]).To(Equal("host.http-domain"))
					Expect(application.Routes[2]["route"]).To(Equal("host.http-domain/path"))
					Expect(application.Routes[3]["route"]).To(Equal("tcp-domain:123"))
				})

			})
			Context("when the application contains environment vars", func() {
				BeforeEach(func() {
					m.EnvironmentVars("app1", "foo", "foo-value")
					m.EnvironmentVars("app1", "bar", "bar-value")
				})

				It("stores each environment var", func() {
					err := m.Save(f)
					Expect(err).NotTo(HaveOccurred())
					application := getYaml(f).Applications[0]

					Expect(application.Env).To(Equal(map[string]interface{}{
						"foo": "foo-value",
						"bar": "bar-value",
					}))
				})
			})
		})

		It("returns an error when stack has not been set", func() {
			m.Memory("app1", 1024)
			m.Instances("app1", 2)
			m.DiskQuota("app1", 1024)

			err := m.Save(f)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Error saving manifest: required attribute 'stack' missing"))
		})

		It("returns an error when memory has not been set", func() {
			m.Instances("app1", 2)
			m.DiskQuota("app1", 1024)
			m.Stack("app1", "stack")

			err := m.Save(f)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Error saving manifest: required attribute 'memory' missing"))
		})

		It("returns an error when disk quota has not been set", func() {
			m.Instances("app1", 2)
			m.Memory("app1", 1024)
			m.Stack("app1", "stack")

			err := m.Save(f)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Error saving manifest: required attribute 'disk_quota' missing"))
		})

		It("returns an error when instances have not been set", func() {
			m.DiskQuota("app1", 1024)
			m.Memory("app1", 1024)
			m.Stack("app1", "stack")

			err := m.Save(f)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Error saving manifest: required attribute 'instances' missing"))
		})
	})
})

type YManifest struct {
	Applications []YApplication `yaml:"applications"`
}

type YApplication struct {
	Name                    string                 `yaml:"name"`
	Services                []string               `yaml:"services"`
	Buildpack               string                 `yaml:"buildpack"`
	Memory                  string                 `yaml:"memory"`
	Command                 string                 `yaml:"command"`
	Env                     map[string]interface{} `yaml:"env"`
	Timeout                 int                    `yaml:"timeout"`
	Instances               int                    `yaml:"instances"`
	Routes                  []map[string]string    `yaml:"routes"`
	NoRoute                 bool                   `yaml:"no-route"`
	DiskQuota               string                 `yaml:"disk_quota"`
	Stack                   string                 `yaml:"stack"`
	AppPorts                []int                  `yaml:"app-ports"`
	HealthCheckType         string                 `yaml:"health-check-type"`
	HealthCheckHTTPEndpoint string                 `yaml:"health-check-http-endpoint"`
}

func getYaml(f *bytes.Buffer) YManifest {
	var document YManifest

	err := yaml.Unmarshal([]byte(f.String()), &document)
	Expect(err).NotTo(HaveOccurred())

	return document
}

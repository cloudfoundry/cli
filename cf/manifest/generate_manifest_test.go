package manifest_test

import (
	"gopkg.in/yaml.v2"

	"bytes"

	. "github.com/cloudfoundry/cli/cf/manifest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("generate_manifest", func() {
	Describe("Save", func() {
		var (
			m AppManifest
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
				m.Save(f)

				applications := getYaml(f).Applications

				Expect(applications[1].Name).To(Equal("app2"))
				Expect(applications[1].Memory).To(Equal("2048M"))
				Expect(applications[1].DiskQuota).To(Equal("2048M"))
				Expect(applications[1].Stack).To(Equal("stack-name"))
				Expect(applications[1].Instances).To(Equal(3))
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
					m.BuildpackUrl("app1", "buildpack")
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

			It("includes no-route when the application has no domains", func() {
				m.Save(f)
				contents := getYaml(f)
				application := contents.Applications[0]
				Expect(application.NoRoute).To(BeTrue())
			})

			Context("when an application has one domain with a hostname", func() {
				BeforeEach(func() {
					m.Domain("app1", "host-name", "domain-name")
				})

				It("includes the domain", func() {
					err := m.Save(f)
					Expect(err).NotTo(HaveOccurred())
					contents := getYaml(f)
					application := contents.Applications[0]
					Expect(application.Domain).To(Equal("domain-name"))
				})

				It("does not include no-hostname", func() {
					err := m.Save(f)
					Expect(err).NotTo(HaveOccurred())
					contents := getYaml(f)
					application := contents.Applications[0]
					Expect(application.NoHostname).To(BeFalse())
				})
			})

			Context("when an application has one domain without a hostname", func() {
				BeforeEach(func() {
					m.Domain("app1", "", "domain-name")
				})

				It("includes the domain", func() {
					err := m.Save(f)
					Expect(err).NotTo(HaveOccurred())
					contents := getYaml(f)
					application := contents.Applications[0]
					Expect(application.Domain).To(Equal("domain-name"))
				})

				It("includes no-hostname", func() {
					m.Save(f)
					contents := getYaml(f)
					application := contents.Applications[0]
					Expect(application.NoHostname).To(BeTrue())
				})
			})

			It("generates a manifest containing two hosts two domains", func() {
				m.Domain("app1", "foo1", "test1.com")
				m.Domain("app1", "foo1", "test2.com")
				m.Domain("app1", "foo2", "test1.com")
				m.Domain("app1", "foo2", "test2.com")
				err := m.Save(f)
				Expect(err).NotTo(HaveOccurred())

				application := getYaml(f).Applications[0]

				Expect(application.Name).To(Equal("app1"))
				Expect(application.Hosts).To(ContainElement("foo1"))
				Expect(application.Hosts).To(ContainElement("foo2"))
				Expect(application.Domains).To(ContainElement("test1.com"))
				Expect(application.Domains).To(ContainElement("test2.com"))

				Expect(application.Host).To(Equal(""))
				Expect(application.Domain).To(Equal(""))
				Expect(application.NoRoute).To(BeFalse())
				Expect(application.NoHostname).To(BeFalse())
			})

			It("generates a manifest containing two hosts one domain", func() {
				m.Domain("app1", "foo1", "test.com")
				m.Domain("app1", "foo2", "test.com")
				err := m.Save(f)
				Expect(err).NotTo(HaveOccurred())

				application := getYaml(f).Applications[0]

				Expect(application.Name).To(Equal("app1"))
				Expect(application.Hosts).To(ContainElement("foo1"))
				Expect(application.Hosts).To(ContainElement("foo2"))
				Expect(application.Domain).To(Equal("test.com"))

				Expect(application.Host).To(Equal(""))
				Expect(application.Domains).To(BeEmpty())
				Expect(application.NoRoute).To(BeFalse())
				Expect(application.NoHostname).To(BeFalse())
			})

			It("generates a manifest containing one host two domains", func() {
				m.Domain("app1", "foo", "test1.com")
				m.Domain("app1", "foo", "test2.com")
				err := m.Save(f)
				Expect(err).NotTo(HaveOccurred())

				application := getYaml(f).Applications[0]

				Expect(application.Name).To(Equal("app1"))
				Expect(application.Host).To(Equal("foo"))
				Expect(application.Domains).To(ContainElement("test1.com"))
				Expect(application.Domains).To(ContainElement("test2.com"))

				Expect(application.Hosts).To(BeEmpty())
				Expect(application.Domain).To(Equal(""))
				Expect(application.NoRoute).To(BeFalse())
				Expect(application.NoHostname).To(BeFalse())
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
	Name       string                 `yaml:"name"`
	Services   []string               `yaml:"services"`
	Buildpack  string                 `yaml:"buildpack"`
	Memory     string                 `yaml:"memory"`
	Command    string                 `yaml:"command"`
	Env        map[string]interface{} `yaml:"env"`
	Timeout    int                    `yaml:"timeout"`
	Instances  int                    `yaml:"instances"`
	Host       string                 `yaml:"host"`
	Hosts      []string               `yaml:"hosts"`
	Domain     string                 `yaml:"domain"`
	Domains    []string               `yaml:"domains"`
	NoHostname bool                   `yaml:"no-hostname"`
	NoRoute    bool                   `yaml:"no-route"`
	DiskQuota  string                 `yaml:"disk_quota"`
	Stack      string                 `yaml:"stack"`
	AppPorts   []int                  `yaml:"app-ports"`
}

func getYaml(f *bytes.Buffer) YManifest {
	var document YManifest

	err := yaml.Unmarshal([]byte(f.String()), &document)
	Expect(err).NotTo(HaveOccurred())

	return document
}

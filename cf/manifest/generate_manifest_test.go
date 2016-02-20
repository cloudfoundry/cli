package manifest_test

import (
	"io/ioutil"
	"os"
	"strings"

	uuid "github.com/nu7hatch/gouuid"
	"gopkg.in/yaml.v2"

	. "github.com/cloudfoundry/cli/cf/manifest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("generate_manifest", func() {
	var (
		m              AppManifest
		err            error
		uniqueFilename string
	)

	BeforeEach(func() {
		guid, err := uuid.NewV4()
		Expect(err).NotTo(HaveOccurred())

		uniqueFilename = guid.String()

		m = NewGenerator()
		m.FileSavePath(uniqueFilename)
	})

	AfterEach(func() {
		err = os.Remove(uniqueFilename)
		Expect(err).NotTo(HaveOccurred())
	})

	It("creates a new file at a given path", func() {
		m.Save()

		_, err = os.Stat(uniqueFilename)
		Expect(err).NotTo(HaveOccurred())
	})

	It("has applications", func() {
		m.Save()

		contents := getYamlContent(uniqueFilename)

		Expect(contents[0]).To(Equal("applications: []"))
	})

	It("creates entry under the given app name", func() {
		m.Memory("app1", 128)
		m.Memory("app2", 64)
		m.Save()

		applications := getYaml(uniqueFilename).Applications

		Expect(applications[0].Name).To(Equal("app1"))
		Expect(applications[0].Memory).To(Equal("128M"))
		Expect(applications[0].NoRoute).To(BeTrue())

		Expect(applications[1].Name).To(Equal("app2"))
		Expect(applications[1].Memory).To(Equal("64M"))
		Expect(applications[1].NoRoute).To(BeTrue())
	})

	It("can generate services", func() {
		m.Service("app1", "service1")
		m.Service("app1", "service2")
		m.Service("app1", "service3")
		m.Save()

		contents := getYaml(uniqueFilename)

		application := contents.Applications[0]

		Expect(application.Services).To(ContainElement("service1"))
		Expect(application.Services).To(ContainElement("service2"))
		Expect(application.Services).To(ContainElement("service3"))
	})

	It("generates a manifest containing all the attributes", func() {
		m.Memory("app1", 128)
		m.StartCommand("app1", "run main.go")
		m.Service("app1", "service1")
		m.EnvironmentVars("app1", "foo", "boo")
		m.HealthCheckTimeout("app1", 100)
		m.Instances("app1", 3)
		m.Domain("app1", "foo", "blahblahblah.com")
		m.BuildpackUrl("app1", "ruby-buildpack")
		err := m.Save()
		Expect(err).NotTo(HaveOccurred())

		application := getYaml(uniqueFilename).Applications[0]

		Expect(application.Name).To(Equal("app1"))
		Expect(application.Buildpack).To(Equal("ruby-buildpack"))
		Expect(application.Memory).To(Equal("128M"))
		Expect(application.Services[0]).To(Equal("service1"))
		Expect(application.Env["foo"]).To(Equal("boo"))
		Expect(application.Timeout).To(Equal(100))
		Expect(application.Instances).To(Equal(3))
		Expect(application.Host).To(Equal("foo"))
		Expect(application.Domain).To(Equal("blahblahblah.com"))
		Expect(application.NoRoute).To(BeFalse())
	})

	Context("When there is a route with no hostname", func() {
		It("generates a manifest containing no-hostname: true", func() {
			m.Domain("app1", "", "test1.com")

			err := m.Save()
			Expect(err).NotTo(HaveOccurred())

			application := getYaml(uniqueFilename).Applications[0]

			Expect(application.Name).To(Equal("app1"))
			Expect(application.NoHostname).To(BeTrue())

			Expect(application.Host).To(Equal(""))
			Expect(application.Hosts).To(BeEmpty())
			Expect(application.NoRoute).To(BeFalse())
		})
	})

	Context("When there are multiple hosts and domains", func() {
		It("generates a manifest containing two hosts two domains", func() {
			m.Domain("app1", "foo1", "test1.com")
			m.Domain("app1", "foo1", "test2.com")
			m.Domain("app1", "foo2", "test1.com")
			m.Domain("app1", "foo2", "test2.com")
			err := m.Save()
			Expect(err).NotTo(HaveOccurred())

			application := getYaml(uniqueFilename).Applications[0]

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
	})

	Context("When there are multiple hosts and single domain", func() {
		It("generates a manifest containing two hosts one domain", func() {
			m.Domain("app1", "foo1", "test.com")
			m.Domain("app1", "foo2", "test.com")
			err := m.Save()
			Expect(err).NotTo(HaveOccurred())

			application := getYaml(uniqueFilename).Applications[0]

			Expect(application.Name).To(Equal("app1"))
			Expect(application.Hosts).To(ContainElement("foo1"))
			Expect(application.Hosts).To(ContainElement("foo2"))
			Expect(application.Domain).To(Equal("test.com"))

			Expect(application.Host).To(Equal(""))
			Expect(application.Domains).To(BeEmpty())
			Expect(application.NoRoute).To(BeFalse())
			Expect(application.NoHostname).To(BeFalse())
		})
	})

	Context("When there is single host and multiple domains", func() {
		It("generates a manifest containing one host two domains", func() {
			m.Domain("app1", "foo", "test1.com")
			m.Domain("app1", "foo", "test2.com")
			err := m.Save()
			Expect(err).NotTo(HaveOccurred())

			application := getYaml(uniqueFilename).Applications[0]

			Expect(application.Name).To(Equal("app1"))
			Expect(application.Host).To(Equal("foo"))
			Expect(application.Domains).To(ContainElement("test1.com"))
			Expect(application.Domains).To(ContainElement("test2.com"))

			Expect(application.Hosts).To(BeEmpty())
			Expect(application.Domain).To(Equal(""))
			Expect(application.NoRoute).To(BeFalse())
			Expect(application.NoHostname).To(BeFalse())
		})
	})

	It("supports setting disk quota", func() {
		m.DiskQuota("app1", 1024)
		err := m.Save()
		Expect(err).NotTo(HaveOccurred())

		application := getYaml(uniqueFilename).Applications[0]
		Expect(application.DiskQuota).To(Equal("1024M"))
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
}

func getYamlContent(path string) []string {
	b, err := ioutil.ReadFile(path)
	Expect(err).NotTo(HaveOccurred())

	return strings.Split(string(b), "\n")
}

func getYaml(path string) YManifest {
	contents, err := ioutil.ReadFile(path)
	Expect(err).NotTo(HaveOccurred())

	var document YManifest

	err = yaml.Unmarshal(contents, &document)
	Expect(err).NotTo(HaveOccurred())

	return document
}

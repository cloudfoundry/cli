package manifest_test

import (
	"io/ioutil"
	"os"
	"strings"

	. "github.com/cloudfoundry/cli/cf/manifest"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("generate_manifest", func() {

	var (
		m   AppManifest
		err error
	)
	BeforeEach(func() {
		_, err = os.Stat("./output.yml")
		Ω(err).To(HaveOccurred())

		m = NewGenerator()
		m.FileSavePath("./output.yml")
	})

	AfterEach(func() {
		err = os.Remove("./output.yml")
		Ω(err).ToNot(HaveOccurred())
	})

	It("creates a new file at a given path", func() {
		m.Save()

		_, err = os.Stat("./output.yml")
		Ω(err).ToNot(HaveOccurred())
	})

	It("starts the manifest with 3 dashes (---), followed by 'applications'", func() {
		m.Save()

		contents := getYamlContent("./output.yml")

		Ω(contents[0]).To(Equal("---"))
		Ω(contents[1]).To(Equal("applications"))
	})

	It("creates entry under the given app name", func() {
		m.Memory("app1", 128)
		m.Memory("app2", 64)
		m.Save()

		contents := getYamlContent("./output.yml")

		Ω(contents[2]).To(Equal("- name: app1"))
		Ω(contents[3]).To(Equal("  memory: 128M"))

		Ω(contents[4]).To(Equal("- name: app2"))
		Ω(contents[5]).To(Equal("  memory: 64M"))
	})

	It("uses '-' to signal the first item of fields that support multiple entries", func() {

		m.Service("app1", "service1")
		m.Service("app1", "service2")
		m.Service("app1", "service3")
		m.Save()

		contents := getYamlContent("./output.yml")

		Ω(contents[3]).To(Equal("  services:"))
		Ω(contents[4]).To(Equal("  - service1"))
		Ω(contents[5]).To(Equal("  - service2"))
		Ω(contents[6]).To(Equal("  - service3"))
	})

	It("generates a manifest containing all the attributes", func() {
		m.Memory("app1", 128)
		m.Service("app1", "service1")
		m.EnvironmentVars("app1", "foo", "boo")
		m.HealthCheckTimeout("app1", 100)
		m.Instances("app1", 3)
		m.Domain("app1", "foo", "blahblahblah.com")

		Ω(getYamlContent("./output.yml")).To(ContainSubstrings(
			[]string{"  name: app1",
				"  memory: 128M",
				"  services:",
				"  - service1",
				"  env:",
				"    foo: boo",
				"  timeout: 100",
				"  instances: 3",
				"  host: foo",
				"  domain: blahblahblah.com",
			}))
	})

})

func getYamlContent(path string) []string {
	b, err := ioutil.ReadFile(path)
	Ω(err).ToNot(HaveOccurred())

	return strings.Split(string(b), "\n")
}

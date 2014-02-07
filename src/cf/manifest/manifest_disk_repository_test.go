package manifest_test

import (
	. "cf/manifest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
	"path/filepath"
	"runtime"
)

var _ = Describe("ManifestDiskRepository", func() {
	var repo ManifestRepository

	BeforeEach(func() {
		repo = NewManifestDiskRepository()
	})

	It("can parse a manifest when provided a valid path", func() {
		m, errs := repo.ReadManifest("../../fixtures/different-manifest.yml")

		Expect(errs).To(BeEmpty())
		Expect(len(m.Applications)).To(Equal(1))
		Expect(*m.Applications[0].Name).To(Equal("goodbyte"))

		if runtime.GOOS == "windows" {
			Expect(*m.Applications[0].Path).To(Equal("..\\..\\fixtures"))
		} else {
			Expect(*m.Applications[0].Path).To(Equal("../../fixtures"))
		}
	})

	It("returns an error when the path is invalid", func() {
		_, errs := repo.ReadManifest("some/path/that/doesnt/exist/manifest.yml")
		Expect(errs).NotTo(BeEmpty())
	})

	It("returns an error when the manifest does not contain a map", func() {
		_, errs := repo.ReadManifest("../../fixtures/empty-manifest.yml")
		Expect(errs).NotTo(BeEmpty())
	})

	It("looks for manifest.yml in the current directory if the path is not given", func() {
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())

		path, filename, err := repo.ManifestPath("")

		Expect(err).NotTo(HaveOccurred())
		Expect(path).To(Equal(cwd))
		Expect(filename).To(Equal("manifest.yml"))
	})

	It("looks for manifest.yml when given a path to a directory", func() {
		dir, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())

		path, filename, err := repo.ManifestPath(dir)

		Expect(err).NotTo(HaveOccurred())
		Expect(path).To(Equal(dir))
		Expect(filename).To(Equal("manifest.yml"))
	})

	It("separates the directory from the filepath when given a file path", func() {
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())

		path, filename, err := repo.ManifestPath(filepath.Join(cwd, "example_manifest.yml"))

		Expect(err).NotTo(HaveOccurred())
		Expect(path).To(Equal(cwd))
		Expect(filename).To(Equal("example_manifest.yml"))
	})

	It("converts nested maps to generic maps", func() {
		m, errs := repo.ReadManifest("../../fixtures/different-manifest.yml")

		Expect(errs).To(BeEmpty())
		Expect(*m.Applications[0].EnvironmentVars).To(Equal(map[string]string{
			"LD_LIBRARY_PATH": "/usr/lib/somewhere",
		}))
	})

	It("merges manifests with their 'inherited' manifests", func() {
		m, errs := repo.ReadManifest("../../fixtures/inherited-manifest.yml")
		Expect(errs).To(BeEmpty())
		Expect(*m.Applications[0].Name).To(Equal("base-app"))
		Expect(*m.Applications[0].Services).To(Equal([]string{"base-service"}))
		Expect(*m.Applications[0].EnvironmentVars).To(Equal(map[string]string{
			"foo":                "bar",
			"will-be-overridden": "my-value",
		}))

		Expect(*m.Applications[1].Name).To(Equal("my-app"))

		env := *m.Applications[1].EnvironmentVars
		Expect(env["will-be-overridden"]).To(Equal("my-value"))
		Expect(env["foo"]).To(Equal("bar"))

		services := *m.Applications[1].Services
		Expect(services).To(Equal([]string{"base-service", "foo-service"}))
	})
})

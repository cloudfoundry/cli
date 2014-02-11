package manifest_test

import (
	. "cf/manifest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"path/filepath"
	"runtime"
)

var _ = Describe("ManifestDiskRepository", func() {
	var repo ManifestRepository

	BeforeEach(func() {
		repo = NewManifestDiskRepository()
	})

	Describe("given a directory containing a file called 'manifest.yml", func() {
		It("reads that file", func() {
			m, path, errs := repo.ReadManifest("../../fixtures/manifests")

			Expect(errs).To(BeEmpty())
			Expect(path).To(Equal(filepath.Clean("../../fixtures/manifests/manifest.yml")))
			Expect(*m.Applications[0].Name).To(Equal("from-default-manifest"))
		})
	})

	Describe("given a path to a file", func() {
		It("reads the file at that path", func() {
			m, path, errs := repo.ReadManifest("../../fixtures/manifests/different-manifest.yml")

			Expect(errs).To(BeEmpty())
			Expect(path).To(Equal(filepath.Clean("../../fixtures/manifests/different-manifest.yml")))
			Expect(*m.Applications[0].Name).To(Equal("from-different-manifest"))
		})

		It("passes the base directory to the manifest file", func() {
			m, _, errs := repo.ReadManifest("../../fixtures/manifests/different-manifest.yml")

			Expect(errs).To(BeEmpty())
			Expect(len(m.Applications)).To(Equal(1))
			Expect(*m.Applications[0].Name).To(Equal("from-different-manifest"))

			if runtime.GOOS == "windows" {
				Expect(*m.Applications[0].Path).To(Equal("..\\..\\fixtures\\manifests"))
			} else {
				Expect(*m.Applications[0].Path).To(Equal("../../fixtures/manifests"))
			}
		})
	})

	Describe("given a path to a file that doesn't exist", func() {
		It("returns an error", func() {
			_, _, errs := repo.ReadManifest("some/path/that/doesnt/exist/manifest.yml")
			Expect(errs).NotTo(BeEmpty())
		})

		It("returns empty string for the manifest path", func() {
			_, path, _ := repo.ReadManifest("some/path/that/doesnt/exist/manifest.yml")
			Expect(path).To(Equal(""))
		})
	})

	Describe("when the manifest is not valid", func() {
		It("returns an error", func() {
			_, _, errs := repo.ReadManifest("../../fixtures/manifests/empty-manifest.yml")
			Expect(errs).NotTo(BeEmpty())
		})

		It("returns the path to the manifest", func() {
			_, path, _ := repo.ReadManifest("../../fixtures/manifests/empty-manifest.yml")
			Expect(path).To(Equal("../../fixtures/manifests/empty-manifest.yml"))
		})
	})

	It("converts nested maps to generic maps", func() {
		m, _, errs := repo.ReadManifest("../../fixtures/manifests/different-manifest.yml")

		Expect(errs).To(BeEmpty())
		Expect(*m.Applications[0].EnvironmentVars).To(Equal(map[string]string{
			"LD_LIBRARY_PATH": "/usr/lib/somewhere",
		}))
	})

	It("merges manifests with their 'inherited' manifests", func() {
		m, _, errs := repo.ReadManifest("../../fixtures/manifests/inherited-manifest.yml")
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

package manifest_test

import (
	. "cf/manifest"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"path/filepath"
)

var _ = Describe("ManifestDiskRepository", func() {
	var repo ManifestRepository

	BeforeEach(func() {
		repo = NewManifestDiskRepository()
	})

	Describe("given a directory containing a file called 'manifest.yml'", func() {
		It("reads that file", func() {
			m, errs := repo.ReadManifest("../../fixtures/manifests")

			Expect(errs).To(BeEmpty())
			Expect(m.Path).To(Equal(filepath.Clean("../../fixtures/manifests/manifest.yml")))

			applications, errs := m.Applications()
			Expect(errs).To(BeEmpty())
			Expect(*applications[0].Name).To(Equal("from-default-manifest"))
		})
	})

	Describe("given a directory that doesn't contain a file called 'manifest.y{a}ml'", func() {
		It("returns an error", func() {
			m, errs := repo.ReadManifest("../../fixtures")

			Expect(errs).NotTo(BeEmpty())
			Expect(m.Path).To(BeEmpty())
		})
	})

	Describe("given a directory that contains a file called 'manifest.yaml'", func() {
		It("reads that file", func() {
			m, errs := repo.ReadManifest("../../fixtures/manifests/only_yaml")

			Expect(errs).To(BeEmpty())
			Expect(m.Path).To(Equal(filepath.Clean("../../fixtures/manifests/only_yaml/manifest.yaml")))

			applications, errs := m.Applications()
			Expect(errs).To(BeEmpty())
			Expect(*applications[0].Name).To(Equal("from-default-manifest"))
		})
	})

	Describe("given a directory contains files called 'manifest.yml' and 'manifest.yaml'", func() {
		It("reads the file named 'manifest.yml'", func() {
			m, errs := repo.ReadManifest("../../fixtures/manifests/both_yaml_yml")

			Expect(errs).To(BeEmpty())
			Expect(m.Path).To(Equal(filepath.Clean("../../fixtures/manifests/both_yaml_yml/manifest.yml")))

			applications, errs := m.Applications()
			Expect(errs).To(BeEmpty())
			Expect(*applications[0].Name).To(Equal("yml-extension"))
		})
	})

	Describe("given a path to a file", func() {
		var (
			inputPath string
			m         *Manifest
			errs      ManifestErrors
		)

		BeforeEach(func() {
			inputPath = filepath.Clean("../../fixtures/manifests/different-manifest.yml")
			m, errs = repo.ReadManifest(inputPath)
		})

		It("reads the file at that path", func() {
			Expect(errs).To(BeEmpty())
			Expect(m.Path).To(Equal(inputPath))

			applications, errs := m.Applications()
			Expect(errs).To(BeEmpty())
			Expect(*applications[0].Name).To(Equal("from-different-manifest"))
		})

		It("passes the base directory to the manifest file", func() {
			applications, errs := m.Applications()
			Expect(errs).To(BeEmpty())
			Expect(len(applications)).To(Equal(1))
			Expect(*applications[0].Name).To(Equal("from-different-manifest"))

			appPath := filepath.Clean("../../fixtures/manifests")
			Expect(*applications[0].Path).To(Equal(appPath))
		})
	})

	Describe("given a path to a file that doesn't exist", func() {
		It("returns an error", func() {
			_, errs := repo.ReadManifest("some/path/that/doesnt/exist/manifest.yml")
			Expect(errs).NotTo(BeEmpty())
		})

		It("returns empty string for the manifest path", func() {
			m, _ := repo.ReadManifest("some/path/that/doesnt/exist/manifest.yml")
			Expect(m.Path).To(Equal(""))
		})
	})

	Describe("when the manifest is not valid", func() {
		It("returns an error", func() {
			_, errs := repo.ReadManifest("../../fixtures/manifests/empty-manifest.yml")

			fmt.Printf("\n errors: %v", errs)

			Expect(errs).NotTo(BeEmpty())
		})

		It("returns the path to the manifest", func() {
			inputPath := filepath.Clean("../../fixtures/manifests/empty-manifest.yml")
			m, _ := repo.ReadManifest(inputPath)
			Expect(m.Path).To(Equal(inputPath))
		})
	})

	It("converts nested maps to generic maps", func() {
		m, errs := repo.ReadManifest("../../fixtures/manifests/different-manifest.yml")
		Expect(errs).To(BeEmpty())

		applications, errs := m.Applications()
		Expect(errs).To(BeEmpty())
		Expect(*applications[0].EnvironmentVars).To(Equal(map[string]string{
			"LD_LIBRARY_PATH": "/usr/lib/somewhere",
		}))
	})

	It("merges manifests with their 'inherited' manifests", func() {
		m, errs := repo.ReadManifest("../../fixtures/manifests/inherited-manifest.yml")
		Expect(errs).To(BeEmpty())

		applications, errs := m.Applications()
		Expect(errs).To(BeEmpty())
		Expect(*applications[0].Name).To(Equal("base-app"))
		Expect(*applications[0].ServicesToBind).To(Equal([]string{"base-service"}))
		Expect(*applications[0].EnvironmentVars).To(Equal(map[string]string{
			"foo":                "bar",
			"will-be-overridden": "my-value",
		}))

		Expect(*applications[1].Name).To(Equal("my-app"))

		env := *applications[1].EnvironmentVars
		Expect(env["will-be-overridden"]).To(Equal("my-value"))
		Expect(env["foo"]).To(Equal("bar"))

		services := *applications[1].ServicesToBind
		Expect(services).To(Equal([]string{"base-service", "foo-service"}))
	})
})

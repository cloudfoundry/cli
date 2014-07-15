package manifest_test

import (
	"path/filepath"

	. "github.com/cloudfoundry/cli/cf/manifest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ManifestDiskRepository", func() {
	var repo ManifestRepository

	BeforeEach(func() {
		repo = NewManifestDiskRepository()
	})

	Describe("given a directory containing a file called 'manifest.yml'", func() {
		It("reads that file", func() {
			m, err := repo.ReadManifest("../../fixtures/manifests")

			Expect(err).NotTo(HaveOccurred())
			Expect(m.Path).To(Equal(filepath.Clean("../../fixtures/manifests/manifest.yml")))

			applications, err := m.Applications()
			Expect(err).NotTo(HaveOccurred())
			Expect(*applications[0].Name).To(Equal("from-default-manifest"))
		})
	})

	Describe("given a directory that doesn't contain a file called 'manifest.y{a}ml'", func() {
		It("returns an error", func() {
			m, err := repo.ReadManifest("../../fixtures")

			Expect(err).To(HaveOccurred())
			Expect(m.Path).To(BeEmpty())
		})
	})

	Describe("given a directory that contains a file called 'manifest.yaml'", func() {
		It("reads that file", func() {
			m, err := repo.ReadManifest("../../fixtures/manifests/only_yaml")

			Expect(err).NotTo(HaveOccurred())
			Expect(m.Path).To(Equal(filepath.Clean("../../fixtures/manifests/only_yaml/manifest.yaml")))

			applications, err := m.Applications()
			Expect(err).NotTo(HaveOccurred())
			Expect(*applications[0].Name).To(Equal("from-default-manifest"))
		})
	})

	Describe("given a directory contains files called 'manifest.yml' and 'manifest.yaml'", func() {
		It("reads the file named 'manifest.yml'", func() {
			m, err := repo.ReadManifest("../../fixtures/manifests/both_yaml_yml")

			Expect(err).NotTo(HaveOccurred())
			Expect(m.Path).To(Equal(filepath.Clean("../../fixtures/manifests/both_yaml_yml/manifest.yml")))

			applications, err := m.Applications()
			Expect(err).NotTo(HaveOccurred())
			Expect(*applications[0].Name).To(Equal("yml-extension"))
		})
	})

	Describe("given a path to a file", func() {
		var (
			inputPath string
			m         *Manifest
			err       error
		)

		BeforeEach(func() {
			inputPath = filepath.Clean("../../fixtures/manifests/different-manifest.yml")
			m, err = repo.ReadManifest(inputPath)
		})

		It("reads the file at that path", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(m.Path).To(Equal(inputPath))

			applications, err := m.Applications()
			Expect(err).NotTo(HaveOccurred())
			Expect(*applications[0].Name).To(Equal("from-different-manifest"))
		})

		It("passes the base directory to the manifest file", func() {
			applications, err := m.Applications()
			Expect(err).NotTo(HaveOccurred())
			Expect(len(applications)).To(Equal(1))
			Expect(*applications[0].Name).To(Equal("from-different-manifest"))

			appPath := filepath.Clean("../../fixtures/manifests")
			Expect(*applications[0].Path).To(Equal(appPath))
		})
	})

	Describe("given a path to a file that doesn't exist", func() {
		It("returns an error", func() {
			_, err := repo.ReadManifest("some/path/that/doesnt/exist/manifest.yml")
			Expect(err).To(HaveOccurred())
		})

		It("returns empty string for the manifest path", func() {
			m, _ := repo.ReadManifest("some/path/that/doesnt/exist/manifest.yml")
			Expect(m.Path).To(Equal(""))
		})
	})

	Describe("when the manifest is not valid", func() {
		It("returns an error", func() {
			_, err := repo.ReadManifest("../../fixtures/manifests/empty-manifest.yml")
			Expect(err).To(HaveOccurred())
		})

		It("returns the path to the manifest", func() {
			inputPath := filepath.Clean("../../fixtures/manifests/empty-manifest.yml")
			m, _ := repo.ReadManifest(inputPath)
			Expect(m.Path).To(Equal(inputPath))
		})
	})

	It("converts nested maps to generic maps", func() {
		m, err := repo.ReadManifest("../../fixtures/manifests/different-manifest.yml")
		Expect(err).NotTo(HaveOccurred())

		applications, err := m.Applications()
		Expect(err).NotTo(HaveOccurred())
		Expect(*applications[0].EnvironmentVars).To(Equal(map[string]string{
			"LD_LIBRARY_PATH": "/usr/lib/somewhere",
		}))
	})

	It("merges manifests with their 'inherited' manifests", func() {
		m, err := repo.ReadManifest("../../fixtures/manifests/inherited-manifest.yml")
		Expect(err).NotTo(HaveOccurred())

		applications, err := m.Applications()
		Expect(err).NotTo(HaveOccurred())
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

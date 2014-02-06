package manifest_test

import (
	. "cf/manifest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	"os"
	"path/filepath"
	"runtime"
)

func pushingWithAbsoluteUnixPath(t mr.TestingT) {
	repo := NewManifestDiskRepository()
	m, err := repo.ReadManifest("../../fixtures/unix-manifest.yml")

	assert.NoError(t, err)
	assert.Equal(t, *m.Applications[0].Path, "/absolute/path/to/example-app")
}

func pushingWithAbsoluteWindowsPath(t mr.TestingT) {
	repo := NewManifestDiskRepository()
	m, err := repo.ReadManifest("../../fixtures/windows-manifest.yml")

	assert.NoError(t, err)
	assert.Equal(t, *m.Applications[0].Path, "C:\\path\\to\\my\\app")
}

func init() {
	Describe("ManifestDiskRepository", func() {
		It("can parse a manifest when provided a valid path", func() {
			repo := NewManifestDiskRepository()
			m, errs := repo.ReadManifest("../../fixtures/different-manifest.yml")

			assert.True(mr.T(), errs.Empty())
			assert.Equal(mr.T(), len(m.Applications), 1)
			assert.Equal(mr.T(), *m.Applications[0].Name, "goodbyte")

			if runtime.GOOS == "windows" {
				assert.Equal(mr.T(), *m.Applications[0].Path, "..\\..\\fixtures")
			} else {
				assert.Equal(mr.T(), *m.Applications[0].Path, "../../fixtures")
			}
		})

		It("TestReadManifestWithBadPath", func() {
			repo := NewManifestDiskRepository()
			_, errs := repo.ReadManifest("some/path/that/doesnt/exist/manifest.yml")

			assert.False(mr.T(), errs.Empty())
		})

		It("TestManifestPathsDefaultsToCurrentDirectory", func() {
			repo := NewManifestDiskRepository()

			cwd, err := os.Getwd()
			assert.NoError(mr.T(), err)

			path, filename, err := repo.ManifestPath("")

			assert.NoError(mr.T(), err)
			assert.Equal(mr.T(), path, cwd)
			assert.Equal(mr.T(), filename, "manifest.yml")
		})

		It("TestAppAndManifestPathsIgnoreAppPathWhenManifestPathIsSpecified", func() {
			repo := NewManifestDiskRepository()

			cwd, err := os.Getwd()
			assert.NoError(mr.T(), err)
			expectedDir := filepath.Join(cwd, "..")

			path, filename, err := repo.ManifestPath(expectedDir)

			assert.NoError(mr.T(), err)
			assert.Equal(mr.T(), path, expectedDir)
			assert.Equal(mr.T(), filename, "manifest.yml")
		})

		It("TestAppAndManifestPathsManifestFileIsDroppedFromAppPath", func() {
			repo := NewManifestDiskRepository()

			cwd, err := os.Getwd()
			assert.NoError(mr.T(), err)

			path, filename, err := repo.ManifestPath(filepath.Join(cwd, "example_manifest.yml"))

			assert.NoError(mr.T(), err)
			assert.Equal(mr.T(), path, cwd)
			assert.Equal(mr.T(), filename, "example_manifest.yml")
		})

		It("converts nested maps to generic maps", func() {
			repo := NewManifestDiskRepository()
			m, errs := repo.ReadManifest("../../fixtures/different-manifest.yml")

			Expect(errs).To(BeEmpty())
			Expect(*m.Applications[0].EnvironmentVars).To(Equal(map[string]string{
				"LD_LIBRARY_PATH": "/usr/lib/somewhere",
			}))
		})

		It("TestManifestWithInheritance", func() {
			repo := NewManifestDiskRepository()
			m, err := repo.ReadManifest("../../fixtures/inherited-manifest.yml")
			assert.NoError(mr.T(), err)
			assert.Equal(mr.T(), *m.Applications[0].Name, "base-app")
			assert.Equal(mr.T(), *m.Applications[0].Services, []string{"base-service"})
			assert.Equal(mr.T(), *m.Applications[0].EnvironmentVars, map[string]string{
				"foo":                "bar",
				"will-be-overridden": "my-value",
			})

			assert.Equal(mr.T(), *m.Applications[1].Name, "my-app")

			env := *m.Applications[1].EnvironmentVars
			assert.Equal(mr.T(), env["will-be-overridden"], "my-value")
			assert.Equal(mr.T(), env["foo"], "bar")

			services := *m.Applications[1].Services
			assert.Equal(mr.T(), services, []string{"base-service", "foo-service"})
		})

		It("TestPushingWithAbsoluteAppPathFromManifestFile", func() {
			if runtime.GOOS == "windows" {
				pushingWithAbsoluteWindowsPath(mr.T())
			} else {
				pushingWithAbsoluteUnixPath(mr.T())
			}
		})
	})
}

package manifest_test

import (
	. "cf/manifest"
	"generic"
	. "github.com/onsi/ginkgo"
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
	assert.Equal(t, m.Applications[0].Get("path"), "/absolute/path/to/example-app")
}

func pushingWithAbsoluteWindowsPath(t mr.TestingT) {
	repo := NewManifestDiskRepository()
	m, err := repo.ReadManifest("../../fixtures/windows-manifest.yml")

	assert.NoError(t, err)
	assert.Equal(t, m.Applications[0].Get("path"), "C:\\path\\to\\my\\app")
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestReadManifestWithGoodPath", func() {
			repo := NewManifestDiskRepository()
			manifest, errs := repo.ReadManifest("../../fixtures/different-manifest.yml")

			assert.True(mr.T(), errs.Empty())
			assert.Equal(mr.T(), len(manifest.Applications), 1)
			assert.Equal(mr.T(), manifest.Applications[0].Get("name").(string), "goodbyte")
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
		It("TestManifestWithInheritance", func() {

			repo := NewManifestDiskRepository()
			m, err := repo.ReadManifest("../../fixtures/inherited-manifest.yml")
			assert.NoError(mr.T(), err)
			assert.Equal(mr.T(), m.Applications[0].Get("name"), "base-app")
			assert.Equal(mr.T(), m.Applications[0].Get("services"), []string{"base-service"})
			assert.Equal(mr.T(), m.Applications[0].Get("env"), generic.NewMap(map[string]string{
				"foo":                "bar",
				"will-be-overridden": "my-value",
			}))

			assert.Equal(mr.T(), m.Applications[1].Get("name"), "my-app")

			env := generic.NewMap(m.Applications[1].Get("env"))
			assert.Equal(mr.T(), env.Get("will-be-overridden"), "my-value")
			assert.Equal(mr.T(), env.Get("foo"), "bar")

			services := m.Applications[1].Get("services")
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

package configuration_test

import (
	. "cf/configuration"
	"fileutils"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	"os"
	"path/filepath"
)

func withFakeHome(t mr.TestingT, callback func(dirPath string)) {
	fileutils.TempDir("test-config", func(dir string, err error) {
		if err != nil {
			Fail("Couldn't create tmp file")
		}
		callback(filepath.Join(dir, ".cf", "config.json"))
	})
}

func withConfigFixture(t mr.TestingT, name string, callback func(dirPath string)) {
	cwd, err := os.Getwd()
	assert.NoError(t, err)
	callback(filepath.Join(cwd, "../../fixtures/config", name, ".cf", "config.json"))
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestLoadingWithNoConfigFile", func() {
		withFakeHome(mr.T(), func(configPath string) {
			repo := NewDiskPersistor(configPath)
			configData, err := repo.Load()
			assert.NoError(mr.T(), err)

			assert.Equal(mr.T(), configData.Target, "")
			assert.Equal(mr.T(), configData.ApiVersion, "")
			assert.Equal(mr.T(), configData.AuthorizationEndpoint, "")
			assert.Equal(mr.T(), configData.AccessToken, "")
		})
	})

	It("TestSavingAndLoading", func() {
		withFakeHome(mr.T(), func(configPath string) {
			repo := NewDiskPersistor(configPath)
			configData, err := repo.Load()
			assert.NoError(mr.T(), err)

			configData.ApiVersion = "3.1.0"
			configData.Target = "https://api.target.example.com"
			configData.AuthorizationEndpoint = "https://login.target.example.com"
			configData.AccessToken = "bearer my_access_token"

			err = repo.Save(configData)
			assert.NoError(mr.T(), err)

			savedConfig, err := repo.Load()
			assert.NoError(mr.T(), err)
			assert.Equal(mr.T(), savedConfig, configData)
		})
	})

	It("TestReadingOutdatedConfigReturnsNewConfig", func() {
		withConfigFixture(mr.T(), "outdated-config", func(configPath string) {
			repo := NewDiskPersistor(configPath)
			configData, err := repo.Load()

			assert.NoError(mr.T(), err)
			assert.Equal(mr.T(), configData.Target, "")
		})
	})
})

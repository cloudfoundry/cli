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

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestLoadingWithNoConfigFile", func() {
			withFakeHome(mr.T(), func(configPath string) {
				repo := NewConfigurationDiskRepository(configPath)
				config, err := repo.Load()
				assert.NoError(mr.T(), err)

				assert.Equal(mr.T(), config.ApiEndpoint(), "")
				assert.Equal(mr.T(), config.ApiVersion(), "")
				assert.Equal(mr.T(), config.AuthorizationEndpoint(), "")
				assert.Equal(mr.T(), config.AccessToken(), "")
			})
		})

		It("TestSavingAndLoading", func() {
			withFakeHome(mr.T(), func(configPath string) {
				repo := NewConfigurationDiskRepository(configPath)
				config, err := repo.Load()
				assert.NoError(mr.T(), err)

				config.SetApiVersion("3.1.0")
				config.SetApiEndpoint("https://api.target.example.com")
				config.SetAuthorizationEndpoint("https://login.target.example.com")
				config.SetAccessToken("bearer my_access_token")

				repo.Save(config)

				savedConfig, err := repo.Load()
				assert.NoError(mr.T(), err)
				assert.Equal(mr.T(), savedConfig, config)
			})
		})

		It("TestReadingOutdatedConfigReturnsNewConfig", func() {
			withConfigFixture(mr.T(), "outdated-config", func(configPath string) {
				repo := NewConfigurationDiskRepository(configPath)
				config, err := repo.Load()

				assert.NoError(mr.T(), err)
				assert.Equal(mr.T(), config.ApiEndpoint(), "")
			})
		})
	})
}

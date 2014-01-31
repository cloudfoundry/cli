package configuration

import (
	"cf"
	"fileutils"
	"fmt"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	"os"
	"path/filepath"
	"strings"
)

func withCFHome(dirName string, callback func()) {
	defer func() {
		os.Setenv("CF_HOME", "")
	}()

	os.Setenv("CF_HOME", dirName)
	callback()
}

func withFakeHome(t mr.TestingT, callback func()) {
	oldHome := os.Getenv("HOME")
	oldHomePath := os.Getenv("HOMEPATH")
	oldHomeDrive := os.Getenv("HOMEDRIVE")
	defer func() {
		os.Setenv("HOMEDRIVE", oldHomeDrive)
		os.Setenv("HOMEPATH", oldHomePath)
		os.Setenv("HOME", oldHome)
	}()

	defer func() {
		NewConfigurationDiskRepository().Delete()
	}()

	fileutils.TempDir("test-config", func(dir string, err error) {
		os.Setenv("HOME", dir)

		volumeName := filepath.VolumeName(dir)
		if volumeName != "" {
			relativePath := strings.Replace(dir, volumeName, "", 1)

			os.Setenv("HOMEPATH", relativePath)
			os.Setenv("HOMEDRIVE", volumeName)
		}

		callback()
	})
}

func withConfigFixture(t mr.TestingT, name string, callback func()) {
	oldHome := os.Getenv("HOME")
	oldHomePath := os.Getenv("HOMEPATH")
	oldHomeDrive := os.Getenv("HOMEDRIVE")
	defer func() {
		os.Setenv("HOMEDRIVE", oldHomeDrive)
		os.Setenv("HOMEPATH", oldHomePath)
		os.Setenv("HOME", oldHome)
	}()

	defer func() {
		singleton = nil
	}()

	cwd, err := os.Getwd()
	assert.NoError(t, err)

	fixturePath := filepath.Join(cwd, fmt.Sprintf("../../fixtures/config/%s", name))
	os.Setenv("HOME", fixturePath)

	volumeName := filepath.VolumeName(fixturePath)
	if volumeName != "" {
		relativePath := strings.Replace(fixturePath, volumeName, "", 1)

		os.Setenv("HOMEPATH", relativePath)
		os.Setenv("HOMEDRIVE", volumeName)
	}

	callback()
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestLoadingWithNoConfigFile", func() {
			withFakeHome(mr.T(), func() {
				repo := NewConfigurationDiskRepository()
				config, err := repo.Get()
				assert.NoError(mr.T(), err)

				assert.Equal(mr.T(), config.Target, "")
				assert.Equal(mr.T(), config.ApiVersion, "")
				assert.Equal(mr.T(), config.AuthorizationEndpoint, "")
				assert.Equal(mr.T(), config.AccessToken, "")
			})
		})
		It("TestSavingAndLoading", func() {

			withFakeHome(mr.T(), func() {
				repo := NewConfigurationDiskRepository()
				configToSave, err := repo.Get()
				assert.NoError(mr.T(), err)

				configToSave.ApiVersion = "3.1.0"
				configToSave.Target = "https://api.target.example.com"
				configToSave.AuthorizationEndpoint = "https://login.target.example.com"
				configToSave.AccessToken = "bearer my_access_token"

				repo.Save()

				singleton = nil
				savedConfig, err := repo.Get()
				assert.NoError(mr.T(), err)
				assert.Equal(mr.T(), savedConfig, configToSave)
			})
		})
		It("TestSetOrganization", func() {

			withFakeHome(mr.T(), func() {
				repo := NewConfigurationDiskRepository()
				config, err := repo.Get()
				assert.NoError(mr.T(), err)

				config.OrganizationFields = cf.OrganizationFields{}

				org := cf.OrganizationFields{}
				org.Name = "my-org"
				org.Guid = "my-org-guid"
				err = repo.SetOrganization(org)
				assert.NoError(mr.T(), err)

				repo.Save()

				savedConfig, err := repo.Get()
				assert.NoError(mr.T(), err)
				assert.Equal(mr.T(), savedConfig.OrganizationFields, org)
				assert.Equal(mr.T(), savedConfig.SpaceFields, cf.SpaceFields{})
			})
		})
		It("TestSetSpace", func() {

			withFakeHome(mr.T(), func() {
				repo := NewConfigurationDiskRepository()
				repo.Get()
				space := cf.SpaceFields{}
				space.Name = "my-space"
				space.Guid = "my-space-guid"

				err := repo.SetSpace(space)
				assert.NoError(mr.T(), err)

				repo.Save()

				savedConfig, err := repo.Get()
				assert.NoError(mr.T(), err)
				assert.Equal(mr.T(), savedConfig.SpaceFields, space)
			})
		})
		It("TestClearTokens", func() {

			withFakeHome(mr.T(), func() {
				org := cf.OrganizationFields{}
				org.Name = "my-org"
				space := cf.SpaceFields{}
				space.Name = "my-space"

				repo := NewConfigurationDiskRepository()
				config, err := repo.Get()
				assert.NoError(mr.T(), err)

				config.Target = "http://api.example.com"
				config.RefreshToken = "some old refresh token"
				config.AccessToken = "some old access token"
				config.OrganizationFields = org
				config.SpaceFields = space
				repo.Save()

				err = repo.ClearTokens()
				assert.NoError(mr.T(), err)

				repo.Save()

				savedConfig, err := repo.Get()
				assert.NoError(mr.T(), err)
				assert.Equal(mr.T(), savedConfig.Target, "http://api.example.com")
				assert.Empty(mr.T(), savedConfig.AccessToken)
				assert.Empty(mr.T(), savedConfig.RefreshToken)
				assert.Equal(mr.T(), savedConfig.OrganizationFields, org)
				assert.Equal(mr.T(), savedConfig.SpaceFields, space)
			})
		})
		It("TestClearSession", func() {

			withFakeHome(mr.T(), func() {
				repo := NewConfigurationDiskRepository()
				config, err := repo.Get()
				assert.NoError(mr.T(), err)

				config.Target = "http://api.example.com"
				config.RefreshToken = "some old refresh token"
				config.AccessToken = "some old access token"
				org := cf.OrganizationFields{}
				org.Name = "my-org"
				space := cf.SpaceFields{}
				space.Name = "my-space"
				repo.Save()

				err = repo.ClearSession()
				assert.NoError(mr.T(), err)

				repo.Save()

				savedConfig, err := repo.Get()
				assert.NoError(mr.T(), err)
				assert.Equal(mr.T(), savedConfig.Target, "http://api.example.com")
				assert.Empty(mr.T(), savedConfig.AccessToken)
				assert.Empty(mr.T(), savedConfig.RefreshToken)
				assert.Equal(mr.T(), savedConfig.OrganizationFields, cf.OrganizationFields{})
				assert.Equal(mr.T(), savedConfig.SpaceFields, cf.SpaceFields{})
			})
		})
		It("TestNewConfigGivesYouCurrentVersionedConfig", func() {

			withFakeHome(mr.T(), func() {
				repo := NewConfigurationDiskRepository()
				config, err := repo.Get()
				assert.NoError(mr.T(), err)
				assert.Equal(mr.T(), config.ConfigVersion, 2)
			})
		})
		It("TestReadingOutdatedConfigReturnsNewConfig", func() {

			withConfigFixture(mr.T(), "outdated-config", func() {
				repo := NewConfigurationDiskRepository()
				config, err := repo.Get()

				assert.NoError(mr.T(), err)
				assert.Equal(mr.T(), config.Target, "")
			})
		})
		It("TestReadingVersionNumberFromExistingConfig", func() {

			withConfigFixture(mr.T(), "versioned-config", func() {
				repo := NewConfigurationDiskRepository()
				config, err := repo.Get()
				assert.NoError(mr.T(), err)
				assert.Equal(mr.T(), config.ConfigVersion, 9001)
			})
		})
		It("TestWithCFHome", func() {

			fileutils.TempDir("cf-home-test", func(dir string, err error) {
				assert.NoError(mr.T(), err)

				withCFHome(dir, func() {
					repo := NewConfigurationDiskRepository()
					_, err := repo.Get()
					assert.NoError(mr.T(), err)

					fileInfo, err := os.Lstat(filepath.Join(dir, ".cf"))
					assert.NoError(mr.T(), err)
					assert.True(mr.T(), fileInfo.IsDir())

					_, err = os.Lstat(filepath.Join(dir, ".cf", "config.json"))
					assert.NoError(mr.T(), err)
				})
			})
		})
	})
}

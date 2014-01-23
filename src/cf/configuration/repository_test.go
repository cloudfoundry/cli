package configuration

import (
	"cf"
	"fileutils"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadingWithNoConfigFile(t *testing.T) {
	withFakeHome(t, func() {
		repo := NewConfigurationDiskRepository()
		config, err := repo.Get()
		assert.NoError(t, err)

		assert.Equal(t, config.Target, "")
		assert.Equal(t, config.ApiVersion, "")
		assert.Equal(t, config.AuthorizationEndpoint, "")
		assert.Equal(t, config.AccessToken, "")
	})
}

func TestSavingAndLoading(t *testing.T) {
	withFakeHome(t, func() {
		repo := NewConfigurationDiskRepository()
		configToSave, err := repo.Get()
		assert.NoError(t, err)

		configToSave.ApiVersion = "3.1.0"
		configToSave.Target = "https://api.target.example.com"
		configToSave.AuthorizationEndpoint = "https://login.target.example.com"
		configToSave.AccessToken = "bearer my_access_token"

		repo.Save()

		singleton = nil
		savedConfig, err := repo.Get()
		assert.NoError(t, err)
		assert.Equal(t, savedConfig, configToSave)
	})
}

func TestSetOrganization(t *testing.T) {
	withFakeHome(t, func() {
		repo := NewConfigurationDiskRepository()
		config, err := repo.Get()
		assert.NoError(t, err)

		config.OrganizationFields = cf.OrganizationFields{}

		org := cf.OrganizationFields{}
		org.Name = "my-org"
		org.Guid = "my-org-guid"
		err = repo.SetOrganization(org)
		assert.NoError(t, err)

		repo.Save()

		savedConfig, err := repo.Get()
		assert.NoError(t, err)
		assert.Equal(t, savedConfig.OrganizationFields, org)
		assert.Equal(t, savedConfig.SpaceFields, cf.SpaceFields{})
	})
}

func TestSetSpace(t *testing.T) {
	withFakeHome(t, func() {
		repo := NewConfigurationDiskRepository()
		repo.Get()
		space := cf.SpaceFields{}
		space.Name = "my-space"
		space.Guid = "my-space-guid"

		err := repo.SetSpace(space)
		assert.NoError(t, err)

		repo.Save()

		savedConfig, err := repo.Get()
		assert.NoError(t, err)
		assert.Equal(t, savedConfig.SpaceFields, space)
	})
}

func TestClearTokens(t *testing.T) {
	withFakeHome(t, func() {
		org := cf.OrganizationFields{}
		org.Name = "my-org"
		space := cf.SpaceFields{}
		space.Name = "my-space"

		repo := NewConfigurationDiskRepository()
		config, err := repo.Get()
		assert.NoError(t, err)

		config.Target = "http://api.example.com"
		config.RefreshToken = "some old refresh token"
		config.AccessToken = "some old access token"
		config.OrganizationFields = org
		config.SpaceFields = space
		repo.Save()

		err = repo.ClearTokens()
		assert.NoError(t, err)

		repo.Save()

		savedConfig, err := repo.Get()
		assert.NoError(t, err)
		assert.Equal(t, savedConfig.Target, "http://api.example.com")
		assert.Empty(t, savedConfig.AccessToken)
		assert.Empty(t, savedConfig.RefreshToken)
		assert.Equal(t, savedConfig.OrganizationFields, org)
		assert.Equal(t, savedConfig.SpaceFields, space)
	})
}

func TestClearSession(t *testing.T) {
	withFakeHome(t, func() {
		repo := NewConfigurationDiskRepository()
		config, err := repo.Get()
		assert.NoError(t, err)

		config.Target = "http://api.example.com"
		config.RefreshToken = "some old refresh token"
		config.AccessToken = "some old access token"
		org := cf.OrganizationFields{}
		org.Name = "my-org"
		space := cf.SpaceFields{}
		space.Name = "my-space"
		repo.Save()

		err = repo.ClearSession()
		assert.NoError(t, err)

		repo.Save()

		savedConfig, err := repo.Get()
		assert.NoError(t, err)
		assert.Equal(t, savedConfig.Target, "http://api.example.com")
		assert.Empty(t, savedConfig.AccessToken)
		assert.Empty(t, savedConfig.RefreshToken)
		assert.Equal(t, savedConfig.OrganizationFields, cf.OrganizationFields{})
		assert.Equal(t, savedConfig.SpaceFields, cf.SpaceFields{})
	})
}

func TestNewConfigGivesYouCurrentVersionedConfig(t *testing.T) {
	withFakeHome(t, func() {
		repo := NewConfigurationDiskRepository()
		config, err := repo.Get()
		assert.NoError(t, err)
		assert.Equal(t, config.ConfigVersion, 1)
	})
}

func TestReadingOutdatedConfigReturnsNewConfig(t *testing.T) {
	withConfigFixture(t, "outdated-config", func() {
		repo := NewConfigurationDiskRepository()
		config, err := repo.Get()

		assert.NoError(t, err)
		assert.Equal(t, config.ConfigVersion, 1)
		assert.Equal(t, config.Target, "")
	})
}

func TestReadingVersionNumberFromExistingConfig(t *testing.T) {
	withConfigFixture(t, "versioned-config", func() {
		repo := NewConfigurationDiskRepository()
		config, err := repo.Get()
		assert.NoError(t, err)
		assert.Equal(t, config.ConfigVersion, 9001)
	})
}

func withFakeHome(t *testing.T, callback func()) {
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

func withConfigFixture(t *testing.T, name string, callback func()) {
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

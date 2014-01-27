package configuration

import (
	"cf"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
)

const (
	filePermissions      = 0600
	dirPermissions       = 0700
	currentConfigVersion = 2
)

var singleton *Configuration

type ConfigurationRepository interface {
	Get() (config *Configuration, err error)
	Delete()
	Save() (err error)
	ClearTokens() (err error)
	ClearSession() (err error)
	SetOrganization(org cf.OrganizationFields) (err error)
	SetSpace(space cf.SpaceFields) (err error)
}

type ConfigurationDiskRepository struct{}

func NewConfigurationDiskRepository() (repo ConfigurationDiskRepository) {
	return ConfigurationDiskRepository{}
}

func (repo ConfigurationDiskRepository) SetOrganization(org cf.OrganizationFields) (err error) {
	config, err := repo.Get()
	if err != nil {
		return
	}

	config.OrganizationFields = org
	config.SpaceFields = cf.SpaceFields{}

	return saveConfiguration(config)
}

func (repo ConfigurationDiskRepository) SetSpace(space cf.SpaceFields) (err error) {
	config, err := repo.Get()
	if err != nil {
		return
	}

	config.SpaceFields = space

	return saveConfiguration(config)
}

func (repo ConfigurationDiskRepository) Get() (c *Configuration, err error) {
	if singleton == nil {
		singleton, err = repo.load()

		if err != nil {
			return
		}
	}

	return singleton, nil
}

func (repo ConfigurationDiskRepository) Delete() {
	file, err := ConfigFile()

	if err != nil {
		return
	}

	os.Remove(file)
	singleton = nil
}

func (repo ConfigurationDiskRepository) Save() (err error) {
	c, err := repo.Get()
	if err != nil {
		return
	}
	return saveConfiguration(c)
}

func (repo ConfigurationDiskRepository) ClearTokens() (err error) {
	c, err := repo.Get()
	if err != nil {
		return
	}
	c.AccessToken = ""
	c.RefreshToken = ""
	return
}

func (repo ConfigurationDiskRepository) ClearSession() (err error) {
	err = repo.ClearTokens()
	if err != nil {
		return
	}

	c, err := repo.Get()
	if err != nil {
		return
	}
	c.OrganizationFields = cf.OrganizationFields{}
	c.SpaceFields = cf.SpaceFields{}

	return saveConfiguration(c)
}

// Keep this one public for configtest/configuration.go
func ConfigFile() (file string, err error) {
	var configDir string

	if os.Getenv("CF_HOME") != "" {
		configDir = os.Getenv("CF_HOME")
	} else {
		configDir = filepath.Join(userHomeDir(), ".cf")
	}

	err = os.MkdirAll(configDir, dirPermissions)

	if err != nil {
		return
	}

	file = filepath.Join(configDir, "config.json")
	return
}

// See: http://stackoverflow.com/questions/7922270/obtain-users-home-directory
// we can't cross compile using cgo and use user.Current()
func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}

	return os.Getenv("HOME")
}

func defaultConfig() (c *Configuration) {
	c = new(Configuration)
	c.Target = ""
	c.ApiVersion = ""
	c.AuthorizationEndpoint = ""
	c.ApplicationStartTimeout = 30 // seconds
	c.ConfigVersion = currentConfigVersion

	return
}

func (repo ConfigurationDiskRepository) load() (c *Configuration, parseError error) {
	file, readError := ConfigFile()
	c = new(Configuration)

	if readError != nil {
		c := defaultConfig()
		return c, saveConfiguration(c)
	}

	data, readError := ioutil.ReadFile(file)

	if readError != nil {
		c := defaultConfig()
		return c, saveConfiguration(c)
	}

	parseError = json.Unmarshal(data, c)
	if parseError != nil {
		return
	}

	if c.ConfigVersion < currentConfigVersion {
		c = defaultConfig()
		return c, nil
	}

	return
}

func saveConfiguration(config *Configuration) (err error) {
	bytes, err := json.Marshal(config)
	if err != nil {
		return
	}

	file, err := ConfigFile()

	if err != nil {
		return
	}
	err = ioutil.WriteFile(file, bytes, filePermissions)

	return
}

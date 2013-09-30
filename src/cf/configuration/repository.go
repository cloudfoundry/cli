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
	filePermissions = 0644
	dirPermissions  = 0700
)

type ConfigurationRepository interface {
	Get() (config Configuration, err error)
	Delete()
	Save(config Configuration) (err error)
	ClearSession() (err error)
}

type ConfigurationDiskRepository struct {
}

func NewConfigurationDiskRepository() (repo ConfigurationDiskRepository) {
	return ConfigurationDiskRepository{}
}

func (repo ConfigurationDiskRepository) Get() (c Configuration, err error) {
	data, err := repo.readConfigFileContents()
	if err != nil {
		return
	}

	c = Configuration{}

	err = json.Unmarshal(data, &c)
	if err != nil {
		repo.InitializeConfigFile()
		return
	}

	return
}

func (repo ConfigurationDiskRepository) Delete() {
	file, err := ConfigFile()

	if err != nil {
		return
	}

	os.Remove(file)
}

func (repo ConfigurationDiskRepository) Save(c Configuration) (err error) {
	bytes, err := json.Marshal(c)
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

func (repo ConfigurationDiskRepository) ClearSession() (err error) {
	c, err := repo.Get()
	if err != nil {
		return
	}
	c.AccessToken = ""
	c.Organization = cf.Organization{}
	c.Space = cf.Space{}

	return repo.Save(c)
}

func (repo ConfigurationDiskRepository) InitializeConfigFile() (err error) {
	err = repo.Save(defaultConfig())
	return err
}

// Keep this one public for configtest/configuration.go
func ConfigFile() (file string, err error) {

	configDir := filepath.Join(userHomeDir(), ".cf")

	err = os.MkdirAll(configDir, dirPermissions)
	if err != nil {
		return
	}

	file = filepath.Join(configDir, "config.json")
	return
}

func (repo ConfigurationDiskRepository) readConfigFileContents() (data []byte, err error) {
	file, err := ConfigFile()
	if err != nil {
		return
	}
	data, err = ioutil.ReadFile(file)
	if err != nil {
		err = repo.InitializeConfigFile()
		if err != nil {
			return
		}
		data, err = ioutil.ReadFile(file)
		return
	}
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

func defaultConfig() (c Configuration) {
	c = Configuration{}
	c.ApplicationStartTimeout = 30 // seconds
	return
}

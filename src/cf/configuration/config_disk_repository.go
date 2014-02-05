package configuration

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	filePermissions = 0600
	dirPermissions  = 0700
)

type ConfigurationRepository interface {
	Load() (config ConfigReadWriteCloser, err error)
	Delete()
	Save(config ConfigReadWriteCloser) (err error)
}

type ConfigurationDiskRepository struct {
	filePath string
	config   ConfigReadWriteCloser
}

func NewConfigurationDiskRepository(path string) (repo ConfigurationDiskRepository) {
	return ConfigurationDiskRepository{filePath: path}
}

func (repo ConfigurationDiskRepository) Load() (c ConfigReadWriteCloser, err error) {
	err = os.MkdirAll(filepath.Dir(repo.filePath), dirPermissions)
	if err != nil {
		return
	}

	data, err := ioutil.ReadFile(repo.filePath)
	if err != nil {
		c = NewConfigReadWriteCloser(newConfiguration())
		err = repo.saveConfiguration(c)
		return
	}

	return ConfigFromJsonV2(data)
}

func (repo ConfigurationDiskRepository) Delete() {
	os.Remove(repo.filePath)
}

func (repo ConfigurationDiskRepository) Save(config ConfigReadWriteCloser) (err error) {
	return repo.saveConfiguration(config)
}

func (repo ConfigurationDiskRepository) saveConfiguration(config ConfigReader) (err error) {
	bytes, err := ConfigToJsonV2(config)
	if err != nil {
		return
	}

	err = ioutil.WriteFile(repo.filePath, bytes, filePermissions)
	return
}

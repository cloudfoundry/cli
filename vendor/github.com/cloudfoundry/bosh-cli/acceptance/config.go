package acceptance

import (
	"encoding/json"
	"errors"
	"os"

	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type Config struct {
	StemcellURL              string `json:"stemcell_url"`
	StemcellSHA1             string `json:"stemcell_sha1"`
	StemcellPath             string `json:"stemcell_path"`
	CPIReleaseURL            string `json:"cpi_release_url"`
	CPIReleaseSHA1           string `json:"cpi_release_sha1"`
	CPIReleasePath           string `json:"cpi_release_path"`
	DummyReleasePath         string `json:"dummy_release_path"`
	DummyTooReleasePath      string `json:"dummy_too_release_path"`
	DummyCompiledReleasePath string `json:"dummy_compiled_release_path"`
}

func NewConfig(fs boshsys.FileSystem) (*Config, error) {
	path := os.Getenv("BOSH_INIT_CONFIG_PATH")
	if path == "" {
		return &Config{}, errors.New("Must provide config file via BOSH_INIT_CONFIG_PATH environment variable")
	}

	configContents, err := fs.ReadFile(path)
	if err != nil {
		return &Config{}, err
	}

	var config Config
	err = json.Unmarshal(configContents, &config)
	if err != nil {
		return &Config{}, err
	}

	return &config, nil
}

func (c *Config) IsLocalCPIRelease() bool {
	return c.CPIReleasePath != ""
}

func (c *Config) IsLocalStemcell() bool {
	return c.StemcellPath != ""
}

func (c *Config) Validate() error {
	if c.StemcellURL == "" && c.StemcellPath == "" {
		return errors.New("Must provide 'stemcell_url' or 'stemcell_path' in config")
	}

	if c.CPIReleaseURL == "" && c.CPIReleasePath == "" {
		return errors.New("Must provide 'cpi_release_url' or 'cpi_release_path' in config")
	}

	if c.DummyTooReleasePath == "" {
		return errors.New("Must provide 'dummy_too_release_path' in config")
	}

	if c.DummyCompiledReleasePath == "" {
		return errors.New("Must provide 'dummy_compiled_release_path' in config")
	}

	return nil
}

package configuration

import (
	"cf"
	"encoding/json"
	"io/ioutil"
	"os"
	"runtime"
	"time"
)

const (
	filePermissions = 0644
	dirPermissions  = 0700
)

var singleton *Configuration

type Configuration struct {
	Target                  string
	ApiVersion              string
	AuthorizationEndpoint   string
	AccessToken             string
	Organization            cf.Organization
	Space                   cf.Space
	ApplicationStartTimeout time.Duration // will be used as seconds
}

func Get() *Configuration {
	if singleton == nil {
		var err error
		singleton, err = load()

		if err != nil {
			println("Error loading configuration")
			os.Exit(-1)
		}
	}

	return singleton
}

func defaultConfig() (c *Configuration) {
	c = new(Configuration)
	c.Target = "https://api.run.pivotal.io"
	c.ApiVersion = "2"
	c.AuthorizationEndpoint = "https://login.run.pivotal.io"
	c.ApplicationStartTimeout = 30 // seconds

	return
}

func Delete() {
	file, err := ConfigFile()

	if err != nil {
		return
	}

	os.Remove(file)
	singleton = nil
}

func load() (c *Configuration, parseError error) {
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

	return
}

func Save() (err error) {
	return saveConfiguration(Get())
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

func ClearSession() (err error) {
	c := Get()
	c.AccessToken = ""
	c.Organization = cf.Organization{}
	c.Space = cf.Space{}

	return saveConfiguration(c)
}

func (c Configuration) UserEmail() (email string) {
	clearInfo, err := DecodeTokenInfo(c.AccessToken)

	if err != nil {
		return
	}

	type TokenInfo struct {
		UserName string `json:"user_name"`
		Email    string `json:"email"`
	}

	tokenInfo := new(TokenInfo)
	err = json.Unmarshal(clearInfo, &tokenInfo)

	if err != nil {
		return
	}

	return tokenInfo.Email
}

func (c Configuration) IsLoggedIn() bool {
	return c.AccessToken != ""
}

func (c Configuration) HasOrganization() bool {
	return c.Organization.Guid != "" && c.Organization.Name != ""
}

func (c Configuration) HasSpace() bool {
	return c.Space.Guid != "" && c.Space.Name != ""
}

// Keep this one public for configtest/configuration.go
func ConfigFile() (file string, err error) {
	configDir := userHomeDir() + "/.cf"

	err = os.MkdirAll(configDir, dirPermissions)

	if err != nil {
		return
	}

	file = configDir + "/config.json"
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

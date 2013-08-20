package configuration

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
)

const (
	filePermissions = 0644
	dirPermissions  = 0700
)

type Configuration struct {
	Target                string
	ApiVersion            string
	AuthorizationEndpoint string
	AccessToken           string
}

func Default() (c Configuration) {
	c.Target = "https://api.run.pivotal.io"
	c.ApiVersion = "2"
	c.AuthorizationEndpoint = "https://login.run.pivotal.io"
	return
}

func Delete() {
	file, err := configFile()

	if err != nil {
		return
	}

	os.Remove(file)
}

func Load() (c Configuration, err error) {
	file, err := configFile()

	if err != nil {
		return
	}

	data, err := ioutil.ReadFile(file)

	if err != nil {
		return
	}

	err = json.Unmarshal(data, &c)

	return
}

func (c Configuration) Save() (err error) {
	bytes, err := json.Marshal(c)
	if err != nil {
		return
	}

	file, err := configFile()

	if err != nil {
		return
	}
	err = ioutil.WriteFile(file, bytes, filePermissions)

	return
}

func configFile() (file string, err error) {
	currentUser, err := user.Current()

	if err != nil {
		return
	}

	configDir := currentUser.HomeDir + "/.cf"

	err = os.MkdirAll(configDir, dirPermissions)

	if err != nil {
		return
	}

	file = configDir + "/config.json"
	return
}

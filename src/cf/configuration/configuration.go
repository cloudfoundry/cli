package configuration

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"strings"
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

func (c Configuration) UserEmail() (email string) {
	tokenParts := strings.Split(c.AccessToken, " ")

	if len(tokenParts) < 2 {
		return
	}

	token := tokenParts[1]
	encodedInfoParts := strings.Split(token, ".")

	if len(encodedInfoParts) < 3 {
		return
	}

	encodedInfo := encodedInfoParts[1]
	clearInfoPart, err := base64.StdEncoding.DecodeString(encodedInfo)

	if err != nil {
		return
	}

	type TokenInfo struct {
		UserName string `json:"user_name"`
		Email    string `json:"email"`
	}

	tokenInfo := new(TokenInfo)
	err = json.Unmarshal(clearInfoPart, &tokenInfo)

	if err != nil {
		return
	}

	return tokenInfo.Email
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

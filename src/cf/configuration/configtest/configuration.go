package configtest

import (
	"cf/configuration"
	"io/ioutil"
	"encoding/json"
)

func GetSavedConfig() (config configuration.Configuration, err error) {
	file, err := configuration.ConfigFile()

	if err != nil {
		return
	}

	data, err := ioutil.ReadFile(file)

	if err != nil {
		return
	}

	err = json.Unmarshal(data, &config)

	return
}

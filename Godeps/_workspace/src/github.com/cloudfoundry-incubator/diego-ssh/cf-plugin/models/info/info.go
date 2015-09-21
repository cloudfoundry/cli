package info

import (
	"encoding/json"
	"errors"

	"github.com/cloudfoundry/cli/plugin"
)

type InfoFactory interface {
	Get() (Info, error)
}

type infoFactory struct {
	cli plugin.CliConnection
}

func NewInfoFactory(cli plugin.CliConnection) InfoFactory {
	return &infoFactory{cli: cli}
}

type Info struct {
	SSHEndpoint            string `json:"app_ssh_endpoint"`
	SSHEndpointFingerprint string `json:"app_ssh_host_key_fingerprint"`
}

func (ifactory *infoFactory) Get() (Info, error) {
	var info Info

	output, err := ifactory.cli.CliCommandWithoutTerminalOutput("curl", "/v2/info")
	if err != nil {
		return info, errors.New("Failed to acquire SSH endpoint info")
	}

	response := []byte(output[0])

	err = json.Unmarshal(response, &info)
	if err != nil {
		return info, errors.New("Failed to acquire SSH endpoint info")
	}

	return info, nil
}

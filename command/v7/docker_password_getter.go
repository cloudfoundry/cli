package v7

import "code.cloudfoundry.org/cli/command"

func GetDockerPassword(ui command.UI, config command.Config, dockerUsername string) (string, error) {
	if dockerUsername == "" { // no need for a password without a username
		return "", nil
	}

	if config.DockerPassword() == "" {
		ui.DisplayText("Environment variable CF_DOCKER_PASSWORD not set.")
		return ui.DisplayPasswordPrompt("Docker password")
	}

	ui.DisplayText("Using docker repository password from environment variable CF_DOCKER_PASSWORD.")
	return config.DockerPassword(), nil
}

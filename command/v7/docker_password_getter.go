package v7

func (cmd PushCommand) GetDockerPassword(dockerUsername string, containsPrivateDockerImages bool) (string, error) {
	if dockerUsername == "" && !containsPrivateDockerImages { // no need for a password without a username
		return "", nil
	}

	if cmd.Config.DockerPassword() == "" {
		cmd.UI.DisplayText("Environment variable CF_DOCKER_PASSWORD not set.")
		return cmd.UI.DisplayPasswordPrompt("Docker password")
	}

	cmd.UI.DisplayText("Using docker repository password from environment variable CF_DOCKER_PASSWORD.")
	return cmd.Config.DockerPassword(), nil
}

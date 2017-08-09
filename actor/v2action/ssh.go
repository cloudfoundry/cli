package v2action

func (actor Actor) GetSSHPasscode() (string, error) {
	return actor.UAAClient.GetSSHPasscode(actor.Config.AccessToken(), actor.Config.SSHOAuthClient())
}

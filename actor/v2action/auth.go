package v2action

import "fmt"

// Authenticate authenticates the user in UAA and sets the returned tokens in
// the config.
//
// It unsets the currently targeted org and space whether authentication
// succeeds or not.
func (actor Actor) Authenticate(config Config, username string, password string) error {
	config.UnsetOrganizationInformation()
	config.UnsetSpaceInformation()

	accessToken, refreshToken, err := actor.UAAClient.Authenticate(username, password)
	if err != nil {
		config.SetTokenInformation("", "", "")
		return err
	}

	accessToken = fmt.Sprintf("bearer %s", accessToken)
	config.SetTokenInformation(accessToken, refreshToken, "")
	return nil
}

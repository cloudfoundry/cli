package v2action

import "fmt"

// Authenticate authenticates the user in UAA and sets the returned tokens in
// the config.
//
// It unsets the currently targeted org and space whether authentication
// succeeds or not.
func (actor Actor) Authenticate(username string, password string) error {
	actor.Config.UnsetOrganizationInformation()
	actor.Config.UnsetSpaceInformation()

	accessToken, refreshToken, err := actor.UAAClient.Authenticate(username, password)
	if err != nil {
		actor.Config.SetTokenInformation("", "", "")
		return err
	}

	accessToken = fmt.Sprintf("bearer %s", accessToken)
	actor.Config.SetTokenInformation(accessToken, refreshToken, "")
	return nil
}

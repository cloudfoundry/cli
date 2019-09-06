package v2action

import (
	"fmt"

	"code.cloudfoundry.org/cli/api/uaa/constant"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
)

// Authenticate authenticates the user in UAA and sets the returned tokens in
// the config.
//
// It unsets the currently targeted org and space whether authentication
// succeeds or not.
func (actor Actor) Authenticate(ID string, secret string, origin string, grantType constant.GrantType) error {

	actor.Config.UnsetOrganizationAndSpaceInformation()
	credentials := make(map[string]string)

	if grantType == constant.GrantTypePassword {
		credentials["username"] = ID
		credentials["password"] = secret
	} else if grantType == constant.GrantTypeClientCredentials {
		credentials["client_id"] = ID
		credentials["client_secret"] = secret
	}

	accessToken, refreshToken, err := actor.UAAClient.Authenticate(credentials, origin, grantType)
	if err != nil {
		actor.Config.SetTokenInformation("", "", "")
		return err
	}

	accessToken = fmt.Sprintf("bearer %s", accessToken)
	actor.Config.SetTokenInformation(accessToken, refreshToken, "")

	if grantType == constant.GrantTypePassword {
		actor.Config.SetUAAGrantType("")
	} else {
		actor.Config.SetUAAGrantType(string(grantType))
	}

	if grantType == constant.GrantTypeClientCredentials {
		actor.Config.SetUAAClientCredentials(ID, "")
	}

	return nil
}

func (actor Actor) GetLoginPrompts() map[string]coreconfig.AuthPrompt {
	rawPrompts := actor.UAAClient.LoginPrompts()
	prompts := make(map[string]coreconfig.AuthPrompt)
	for key, val := range rawPrompts {
		prompts[key] = coreconfig.AuthPrompt{
			Type:        knownAuthPromptTypes[val[0]],
			DisplayName: val[1],
		}
	}
	return prompts
}

var knownAuthPromptTypes = map[string]coreconfig.AuthPromptType{
	"text":     coreconfig.AuthPromptTypeText,
	"password": coreconfig.AuthPromptTypePassword,
}

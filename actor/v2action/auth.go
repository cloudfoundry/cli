package v2action

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

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

// TODO: error check this in future stories
func (actor Actor) Revoke() error {
	accessToken := actor.Config.AccessToken()
	if actor.isTokenRevocable(accessToken) {
		refreshToken := actor.Config.RefreshToken()
		_ = actor.UAAClient.Revoke(refreshToken)
		_ = actor.UAAClient.Revoke(accessToken)
	}
	return nil
}

func (actor Actor) isTokenRevocable(token string) bool {
	segments := strings.Split(token, ".")

	if len(segments) < 2 {
		return false
	}

	jsonPayload, err := base64.RawURLEncoding.DecodeString(segments[1])

	if err != nil {
		return false
	}

	payload := make(map[string]interface{})
	json.Unmarshal(jsonPayload, &payload)
	revocable, ok := payload["revocable"].(bool)

	if !ok {
		return false
	}

	return revocable
}

package v7action

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/uaa/constant"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
)

func (actor Actor) Authenticate(credentials map[string]string, origin string, grantType constant.GrantType) error {
	if grantType == constant.GrantTypePassword && actor.Config.UAAGrantType() == string(constant.GrantTypeClientCredentials) {
		return actionerror.PasswordGrantTypeLogoutRequiredError{}
	}

	actor.Config.UnsetOrganizationAndSpaceInformation()
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
		actor.Config.SetUAAClientCredentials(credentials["client_id"], "")
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

// TODO: error check this in future stories
func (actor Actor) RevokeAccessAndRefreshTokens() error {
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
	err = json.Unmarshal(jsonPayload, &payload)
	if err != nil {
		return false
	}

	revocable, ok := payload["revocable"].(bool)

	if !ok {
		return false
	}

	return revocable
}

var knownAuthPromptTypes = map[string]coreconfig.AuthPromptType{
	"text":     coreconfig.AuthPromptTypeText,
	"password": coreconfig.AuthPromptTypePassword,
}

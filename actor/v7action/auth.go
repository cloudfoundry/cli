package v7action

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/api/uaa/constant"
	"code.cloudfoundry.org/cli/v9/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/v9/util/configv3"
)

type defaultAuthActor struct {
	config    Config
	uaaClient UAAClient
}

func NewDefaultAuthActor(config Config, uaaClient UAAClient) AuthActor {
	return &defaultAuthActor{
		config:    config,
		uaaClient: uaaClient,
	}
}

func (actor defaultAuthActor) Authenticate(credentials map[string]string, origin string, grantType constant.GrantType) error {
	if (grantType == constant.GrantTypePassword || grantType == constant.GrantTypeJwtBearer) && actor.config.UAAGrantType() == string(constant.GrantTypeClientCredentials) {
		return actionerror.PasswordGrantTypeLogoutRequiredError{}
	}

	actor.config.UnsetOrganizationAndSpaceInformation()
	accessToken, refreshToken, err := actor.uaaClient.Authenticate(credentials, origin, grantType)
	if err != nil {
		actor.config.SetTokenInformation("", "", "")
		return err
	}

	accessToken = fmt.Sprintf("bearer %s", accessToken)
	actor.config.SetTokenInformation(accessToken, refreshToken, "")

	if grantType == constant.GrantTypePassword {
		actor.config.SetUAAGrantType("")
	} else {
		actor.config.SetUAAGrantType(string(grantType))
	}

	if (grantType == constant.GrantTypeClientCredentials || grantType == constant.GrantTypeJwtBearer) && credentials["client_id"] != "" {
		actor.config.SetUAAClientCredentials(credentials["client_id"], "")
	}

	return nil
}

func (actor defaultAuthActor) GetLoginPrompts() (map[string]coreconfig.AuthPrompt, error) {
	rawPrompts, err := actor.uaaClient.GetLoginPrompts()
	if err != nil {
		return nil, err
	}

	prompts := make(map[string]coreconfig.AuthPrompt)
	for key, val := range rawPrompts {
		prompts[key] = coreconfig.AuthPrompt{
			Type:        knownAuthPromptTypes[val[0]],
			DisplayName: val[1],
		}
	}

	return prompts, nil
}

func (actor defaultAuthActor) GetCurrentUser() (configv3.User, error) {
	return actor.config.CurrentUser()
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

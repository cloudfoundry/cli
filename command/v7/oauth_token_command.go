package v7

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/api/uaa/constant"
)

type OauthTokenCommand struct {
	BaseCommand

	usage           interface{} `usage:"CF_NAME oauth-token"`
	relatedCommands interface{} `related_commands:"curl"`
}

func (cmd OauthTokenCommand) Execute(_ []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	if cmd.Config.UAAGrantType() == string(constant.GrantTypeClientCredentials) && cmd.Config.UAAOAuthClientSecret() == "" {
		token, err := cmd.Actor.ParseAccessToken(cmd.Config.AccessToken())
		if err != nil {
			return errors.New(cmd.UI.TranslateText("Access token is invalid."))
		}

		expiration, success := token.Claims().Expiration()
		if !success {
			return errors.New(cmd.UI.TranslateText("Access token is missing expiration claim."))
		}

		if expiration.Before(time.Now()) {
			return errors.New(cmd.UI.TranslateText("Access token has expired."))
		}

		cmd.UI.DisplayText(cmd.Config.AccessToken())
		return nil
	}

	accessToken, err := cmd.Actor.RefreshAccessToken()
	if err != nil {
		return err
	}

	cmd.UI.DisplayText(accessToken)
	return nil
}

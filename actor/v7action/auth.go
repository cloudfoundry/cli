package v7action

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/uaa/constant"
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

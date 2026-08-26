package configv3

import utiljwt "code.cloudfoundry.org/cli/v9/util/jwt"

type DefaultUserConfig struct {
	// ConfigFile stores the configuration from the .cf/config
	ConfigFile *JSONConfig
}

// CurrentUser returns user information decoded from the JWT access token in
// .cf/config.json.
func (config DefaultUserConfig) CurrentUser() (User, error) {
	return decodeUserFromJWT(config.ConfigFile.AccessToken)
}

// CurrentUserName returns the name of a user as returned by CurrentUser()
func (config DefaultUserConfig) CurrentUserName() (string, error) {
	user, err := config.CurrentUser()
	if err != nil {
		return "", err
	}
	return user.Name, nil
}

func decodeUserFromJWT(accessToken string) (User, error) {
	if accessToken == "" {
		return User{}, nil
	}

	claims, err := utiljwt.ParseUnverified(accessToken[7:])
	if err != nil {
		return User{}, err
	}

	var name, GUID, origin string
	var isClient bool
	if value, ok := claims["user_name"].(string); ok {
		name = value
		GUID, _ = claims["user_id"].(string)
		origin, _ = claims["origin"].(string)
		isClient = false
	} else {
		name, _ = claims["client_id"].(string)
		GUID = name
		isClient = true
	}

	return User{
		Name:     name,
		GUID:     GUID,
		Origin:   origin,
		IsClient: isClient,
	}, nil
}

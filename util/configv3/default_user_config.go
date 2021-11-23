package configv3

import (
	"github.com/SermoDigital/jose/jws"
)

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

	token, err := jws.ParseJWT([]byte(accessToken[7:]))
	if err != nil {
		return User{}, err
	}

	claims := token.Claims()

	var name, GUID, origin string
	var isClient bool
	if claims.Has("user_name") {
		name = claims.Get("user_name").(string)
		GUID = claims.Get("user_id").(string)
		origin = claims.Get("origin").(string)
		isClient = false
	} else {
		name = claims.Get("client_id").(string)
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

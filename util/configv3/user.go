package configv3

import (
	"fmt"

	"github.com/SermoDigital/jose/jws"
)

// User represents the user information provided by the JWT access token
type User struct {
	Name string
}

// CurrentUser returns user information decoded from the JWT access token in
// .cf/config.json
func (config *Config) CurrentUser() (User, error) {
	return decodeUserFromJWT(config.ConfigFile.AccessToken)
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
	ID := ""
	if claims.Has("user_name") {
		ID = claims.Get("user_name").(string)
	} else if claims.Has("client_id") {
		ID = claims.Get("client_id").(string)
	} else {
		//TODO
		fmt.Printf("No user found in token")
	}
	return User{
		Name: ID,
	}, nil
}

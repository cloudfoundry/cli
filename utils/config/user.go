package config

import "github.com/SermoDigital/jose/jws"

// User represents the user information provided by the JWT access token
type User struct {
	Name string
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
	return User{
		Name: claims.Get("user_name").(string),
	}, nil
}

package jwt

import (
	"strings"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

// ParseUnverified parses JWT claims without validating the access token signature.
// The returned claims must not be used for authentication or authorization.
func ParseUnverified(accessToken string) (jwtv5.MapClaims, error) {
	tokenString := strings.TrimPrefix(accessToken, "bearer ")
	claims := jwtv5.MapClaims{}
	_, _, err := new(jwtv5.Parser).ParseUnverified(tokenString, claims)
	return claims, err
}

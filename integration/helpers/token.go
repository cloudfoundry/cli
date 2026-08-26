package helpers

import (
	"strings"
	"time"

	utiljwt "code.cloudfoundry.org/cli/v9/util/jwt"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	. "github.com/onsi/gomega"
)

// BuildTokenString returns a string typed JSON web token with the specified expiration time
func BuildTokenString(expiration time.Time) string {
	claims := jwtv5.MapClaims{
		"exp":       jwtv5.NewNumericDate(expiration),
		"user_name": "some-user",
		"user_id":   "some-guid",
		"origin":    "uaa",
	}
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodNone, claims)
	tokenBytes, err := token.SignedString(jwtv5.UnsafeAllowNoneSignatureType)
	Expect(err).NotTo(HaveOccurred())
	return string(tokenBytes)
}

// ParseTokenString takes a string typed token and returns a jwt.JWT struct representation of that token
func ParseTokenString(token string) jwtv5.MapClaims {
	strippedToken := strings.TrimPrefix(token, "bearer ")
	claims, err := utiljwt.ParseUnverified(strippedToken)
	Expect(err).NotTo(HaveOccurred())
	return claims
}

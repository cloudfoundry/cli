package helpers

import (
	"strings"
	"time"

	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"github.com/SermoDigital/jose/jwt"
	. "github.com/onsi/gomega"
)

func BuildTokenString(expiration time.Time) string {
	c := jws.Claims{}
	c.SetExpiration(expiration)
	c.Set("user_name", "some-user")
	token := jws.NewJWT(c, crypto.Unsecured)
	tokenBytes, err := token.Serialize(nil)
	Expect(err).NotTo(HaveOccurred())
	return string(tokenBytes)
}

func ParseTokenString(token string) jwt.JWT {
	strippedToken := strings.TrimPrefix(token, "bearer ")
	jwt, err := jws.ParseJWT([]byte(strippedToken))
	Expect(err).NotTo(HaveOccurred())
	return jwt
}

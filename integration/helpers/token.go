package helpers

import (
	"time"

	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
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

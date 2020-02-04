package v7_test

import (
	uuid "github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	"testing"
)

func TestV3(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "V7 Command Suite")
}

var _ = BeforeEach(func() {
	log.SetLevel(log.PanicLevel)
})

// RandomString provides a random string
func RandomString(prefix string) string {
	guid, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}

	return prefix + "-" + guid.String()
}

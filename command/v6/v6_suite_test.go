package v6_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"

	log "github.com/sirupsen/logrus"
)

func TestV6(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "V6 Command Suite")
}

var _ = BeforeEach(func() {
	log.SetLevel(log.PanicLevel)
})

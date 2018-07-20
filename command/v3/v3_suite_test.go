package v3_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	"testing"
)

func TestV3(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "V3 Command Suite")
}

var _ = BeforeEach(func() {
	log.SetLevel(log.PanicLevel)
})

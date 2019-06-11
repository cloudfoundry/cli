package loggingaction_test

import (
	log "github.com/sirupsen/logrus"

	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestLoggingaction(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Loggingaction Suite")
}

var _ = BeforeEach(func() {
	log.SetLevel(log.PanicLevel)
})

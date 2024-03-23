package command_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"

	log "github.com/sirupsen/logrus"
)

func TestCommand(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Command Suite")
}

var _ = BeforeEach(func() {
	log.SetLevel(log.PanicLevel)
})

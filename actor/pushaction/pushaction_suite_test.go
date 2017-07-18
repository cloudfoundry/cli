package pushaction_test

import (
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"

	log "github.com/sirupsen/logrus"
)

func TestPushAction(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Push Actions Suite")
}

var _ = BeforeEach(func() {
	SetDefaultEventuallyTimeout(3 * time.Second)
	log.SetLevel(log.PanicLevel)
})

func getCurrentDir() string {
	pwd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())
	return pwd
}

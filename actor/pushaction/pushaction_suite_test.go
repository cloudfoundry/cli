package pushaction_test

import (
	"os"
	"time"

	. "code.cloudfoundry.org/cli/actor/pushaction"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/types"

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

func EqualEither(events ...Event) GomegaMatcher {
	var equals []GomegaMatcher
	for _, event := range events {
		equals = append(equals, Equal(event))
	}

	return Or(equals...)
}

func getCurrentDir() string {
	pwd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())
	return pwd
}

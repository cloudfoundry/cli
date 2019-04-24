package v7pushaction_test

import (
	"os"
	"time"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/actor/v7pushaction/v7pushactionfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/types"

	"testing"

	log "github.com/sirupsen/logrus"
)

func TestPushAction(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "V7 Push Actions Suite")
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

func getTestPushActor() (*Actor, *v7pushactionfakes.FakeV2Actor, *v7pushactionfakes.FakeV7Actor, *v7pushactionfakes.FakeSharedActor) {
	fakeV2Actor := new(v7pushactionfakes.FakeV2Actor)
	fakeV7Actor := new(v7pushactionfakes.FakeV7Actor)
	fakeSharedActor := new(v7pushactionfakes.FakeSharedActor)
	actor := NewActor(fakeV2Actor, fakeV7Actor, fakeSharedActor)
	return actor, fakeV2Actor, fakeV7Actor, fakeSharedActor
}

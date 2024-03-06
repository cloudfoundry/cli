package pushaction_test

import (
	"os"
	"time"

	. "code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/pushaction/pushactionfakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

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

func EqualEither(events ...Event) types.GomegaMatcher {
	var equals []types.GomegaMatcher
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

func getTestPushActor() (*Actor, *pushactionfakes.FakeV2Actor, *pushactionfakes.FakeV3Actor, *pushactionfakes.FakeSharedActor) {
	fakeV2Actor := new(pushactionfakes.FakeV2Actor)
	fakeV3Actor := new(pushactionfakes.FakeV3Actor)
	fakeSharedActor := new(pushactionfakes.FakeSharedActor)
	actor := NewActor(fakeV2Actor, fakeV3Actor, fakeSharedActor)
	return actor, fakeV2Actor, fakeV3Actor, fakeSharedActor
}

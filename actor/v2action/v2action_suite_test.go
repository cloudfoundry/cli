package v2action_test

import (
	"testing"
	"time"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

func TestV2Action(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "V2 Actions Suite")
}

var _ = BeforeEach(func() {
	SetDefaultEventuallyTimeout(3 * time.Second)
	log.SetLevel(log.PanicLevel)
})

func NewTestActor() (*Actor, *v2actionfakes.FakeCloudControllerClient, *v2actionfakes.FakeUAAClient, *v2actionfakes.FakeConfig) {
	fakeCloudControllerClient := new(v2actionfakes.FakeCloudControllerClient)
	fakeUAAClient := new(v2actionfakes.FakeUAAClient)
	fakeConfig := new(v2actionfakes.FakeConfig)
	actor := NewActor(fakeCloudControllerClient, fakeUAAClient, fakeConfig)

	return actor, fakeCloudControllerClient, fakeUAAClient, fakeConfig
}

package v7action_test

import (
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	"testing"
)

func TestV3Action(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "V7 Actions Suite")
}

var _ = BeforeEach(func() {
	log.SetLevel(log.PanicLevel)
})

func NewTestActor() (*Actor, *v7actionfakes.FakeCloudControllerClient, *v7actionfakes.FakeConfig, *v7actionfakes.FakeSharedActor, *v7actionfakes.FakeUAAClient) {
	fakeCloudControllerClient := new(v7actionfakes.FakeCloudControllerClient)
	fakeConfig := new(v7actionfakes.FakeConfig)
	fakeSharedActor := new(v7actionfakes.FakeSharedActor)
	fakeUAAClient := new(v7actionfakes.FakeUAAClient)
	actor := NewActor(fakeCloudControllerClient, fakeConfig, fakeSharedActor, fakeUAAClient)

	return actor, fakeCloudControllerClient, fakeConfig, fakeSharedActor, fakeUAAClient
}

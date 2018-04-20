package pushaction_test

import (
	. "code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/pushaction/pushactionfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version Check Actions", func() {
	var (
		actor       *Actor
		fakeV2Actor *pushactionfakes.FakeV2Actor
	)

	BeforeEach(func() {
		fakeV2Actor = new(pushactionfakes.FakeV2Actor)
		actor = NewActor(fakeV2Actor, nil, nil)
	})

	Describe("CloudControllerAPIVersion", func() {
		It("returns the V2 CC API version", func() {
			expectedVersion := "2.75.0"
			fakeV2Actor.CloudControllerAPIVersionReturns(expectedVersion)
			Expect(actor.CloudControllerV2APIVersion()).To(Equal(expectedVersion))
		})
	})
})

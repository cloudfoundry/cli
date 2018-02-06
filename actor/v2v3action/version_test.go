package v2v3action_test

import (
	. "code.cloudfoundry.org/cli/actor/v2v3action"
	"code.cloudfoundry.org/cli/actor/v2v3action/v2v3actionfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version Check Actions", func() {
	var (
		actor       *Actor
		fakeV3Actor *v2v3actionfakes.FakeV3Actor
	)

	BeforeEach(func() {
		fakeV3Actor = new(v2v3actionfakes.FakeV3Actor)
		actor = NewActor(nil, fakeV3Actor)
	})

	Describe("CloudControllerV3APIVersion", func() {
		It("returns the V3 CC API version", func() {
			expectedVersion := "3.0.0-alpha.5"
			fakeV3Actor.CloudControllerAPIVersionReturns(expectedVersion)
			Expect(actor.CloudControllerV3APIVersion()).To(Equal(expectedVersion))
		})
	})
})

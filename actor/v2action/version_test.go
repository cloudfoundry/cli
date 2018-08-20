package v2action_test

import (
	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version Check Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
		fakeUAAClient             *v2actionfakes.FakeUAAClient
	)

	BeforeEach(func() {
		actor, fakeCloudControllerClient, fakeUAAClient, _ = NewTestActor()
	})

	Describe("CloudControllerAPIVersion", func() {
		It("returns the V2 CC API version", func() {
			expectedVersion := "2.75.0"
			fakeCloudControllerClient.APIVersionReturns(expectedVersion)
			Expect(actor.CloudControllerAPIVersion()).To(Equal(expectedVersion))
		})
	})

	Describe("UAAAPIVersion", func() {
		It("returns the V2 CC API version", func() {
			expectedVersion := "1.96.0"
			fakeUAAClient.APIVersionReturns(expectedVersion)
			Expect(actor.UAAAPIVersion()).To(Equal(expectedVersion))
		})
	})
})

package v7action_test

import (
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version Check Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
		fakeUAAClient             *v7actionfakes.FakeUAAClient
	)

	BeforeEach(func() {
		actor, fakeCloudControllerClient, _, _, fakeUAAClient, _, _ = NewTestActor()
	})

	Describe("CloudControllerAPIVersion", func() {
		It("returns the CC API version", func() {
			expectedVersion := "3.75.0"
			fakeCloudControllerClient.CloudControllerAPIVersionReturns(expectedVersion)
			Expect(actor.CloudControllerAPIVersion()).To(Equal(expectedVersion))
		})
	})

	Describe("UAAAPIVersion", func() {
		It("returns the UAA API version", func() {
			expectedVersion := "1.96.0"
			fakeUAAClient.APIVersionReturns(expectedVersion)
			Expect(actor.UAAAPIVersion()).To(Equal(expectedVersion))
		})
	})
})

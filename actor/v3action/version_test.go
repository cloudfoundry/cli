package v3action_test

import (
	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Task Actions", func() {
	var (
		actor                     Actor
		fakeCloudControllerClient *v3actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v3actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient)
	})

	Describe("APIVersion", func() {
		It("returns back the CC API version", func() {
			expectedVersion := "3.0.0-alpha.5"
			fakeCloudControllerClient.CloudControllerAPIVersionReturns(expectedVersion)
			Expect(actor.CloudControllerAPIVersion()).To(Equal(expectedVersion))
		})
	})
})

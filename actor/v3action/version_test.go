package v3action_test

import (
	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version Check Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v3actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v3actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil)
	})

	Describe("CloudControllerAPIVersion", func() {
		It("returns the V3 CC API version", func() {
			expectedVersion := "3.0.0-alpha.5"
			fakeCloudControllerClient.CloudControllerAPIVersionReturns(expectedVersion)
			Expect(actor.CloudControllerAPIVersion()).To(Equal(expectedVersion))
		})
	})
})

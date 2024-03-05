package uaa_test

import (
	. "code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/api/uaa/uaafakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SetupResources", func() {
	var (
		client     *Client
		fakeConfig *uaafakes.FakeConfig

		setupResourcesErr error
	)

	BeforeEach(func() {
		fakeConfig = NewTestConfig()
		client = NewClient(fakeConfig)
	})

	JustBeforeEach(func() {
		setupResourcesErr = client.SetupResources(uaaServer.URL(), server.URL())
	})

	It("populates client.info", func() {
		Expect(setupResourcesErr).ToNot(HaveOccurred())
		Expect(client.Info.Links.UAA).To(Equal(uaaServer.URL()))
		Expect(client.Info.Links.Login).To(Equal(server.URL()))

		Expect(fakeConfig.SetUAAEndpointCallCount()).To(Equal(1))
		Expect(fakeConfig.SetUAAEndpointArgsForCall(0)).To(Equal(uaaServer.URL()))
	})
})

package ccv2_test

import (
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/cloudcontrollerv2fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cloud Controller Client", func() {
	var (
		client *CloudControllerClient
	)

	BeforeEach(func() {
		client = NewCloudControllerClient()
	})

	Describe("WrapConnection", func() {
		var fakeConnectionWrapper *cloudcontrollerv2fakes.FakeConnectionWrapper

		BeforeEach(func() {
			fakeConnectionWrapper = new(cloudcontrollerv2fakes.FakeConnectionWrapper)
			fakeConnectionWrapper.WrapReturns(fakeConnectionWrapper)
		})

		It("wraps the existing connection in the provided wrapper", func() {
			client.WrapConnection(fakeConnectionWrapper)
			Expect(fakeConnectionWrapper.WrapCallCount()).To(Equal(1))

			client.DeleteServiceBinding("does-not-matter")
			Expect(fakeConnectionWrapper.MakeCallCount()).To(Equal(1))
		})
	})
})

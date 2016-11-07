package uaa_test

import (
	. "code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/api/uaa/uaafakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AccessToken", func() {
	var (
		client    *Client
		fakeStore *uaafakes.FakeAuthenticationStore
	)

	BeforeEach(func() {
		client, fakeStore = NewTestUAAClientAndStore()
		fakeStore.AccessTokenReturns("access-token")
	})

	It("returns an access token", func() {
		Expect(client.AccessToken()).To(Equal("access-token"))
	})
})

package uaa_test

import (
	. "code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/api/uaa/uaafakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UAA Client", func() {
	var (
		fakeStore *uaafakes.FakeAuthenticationStore
		client    *Client
	)

	BeforeEach(func() {
		fakeStore = new(uaafakes.FakeAuthenticationStore)
		fakeStore.SkipSSLValidationReturns(true)

		client = NewClient(server.URL(), fakeStore)
	})

	Describe("AccessToken", func() {
		BeforeEach(func() {
			fakeStore.AccessTokenReturns("access-token")
		})

		It("returns an access token", func() {
			Expect(client.AccessToken()).To(Equal("access-token"))
		})
	})
})

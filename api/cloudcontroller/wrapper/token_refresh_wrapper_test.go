package wrapper_test

import (
	"errors"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/cloudcontrollerfakes"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/wrapper"
	"code.cloudfoundry.org/cli/api/cloudcontroller/wrapper/wrapperfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Token Refresh Wrapper", func() {
	var (
		fakeConnection *cloudcontrollerfakes.FakeConnection
		fakeStore      *wrapperfakes.FakeAuthenticationStore

		wrapper cloudcontroller.Connection
	)

	BeforeEach(func() {
		fakeConnection = new(cloudcontrollerfakes.FakeConnection)
		fakeStore = new(wrapperfakes.FakeAuthenticationStore)
		fakeStore.AccessTokenReturns("foobar")

		inner := NewTokenRefreshWrapper(fakeStore)
		wrapper = inner.Wrap(fakeConnection)
	})

	Describe("Make", func() {
		It("adds authentication headers", func() {
			request := cloudcontroller.Request{}
			wrapper.Make(request, nil)

			Expect(fakeConnection.MakeCallCount()).To(Equal(1))
			authenticatedRequest, _ := fakeConnection.MakeArgsForCall(0)
			headers := authenticatedRequest.Header
			Expect(headers["Authorization"]).To(ConsistOf([]string{"foobar"}))
		})

		Context("when the request already has headers", func() {
			It("preserves existing headers", func() {
				header := http.Header{}
				header.Add("Existing", "header")

				request := cloudcontroller.Request{
					Header: header,
				}
				wrapper.Make(request, nil)

				Expect(fakeConnection.MakeCallCount()).To(Equal(1))
				authenticatedRequest, _ := fakeConnection.MakeArgsForCall(0)
				headers := authenticatedRequest.Header
				Expect(headers["Existing"]).To(ConsistOf([]string{"header"}))
			})
		})

		Context("when the wrapped connection returns nil", func() {
			It("returns nil", func() {
				fakeConnection.MakeReturns(nil)

				request := cloudcontroller.Request{}
				err := wrapper.Make(request, nil)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when the wrapped connection returns an error", func() {
			It("returns the error", func() {
				innerError := errors.New("inner error")
				fakeConnection.MakeReturns(innerError)

				request := cloudcontroller.Request{}
				err := wrapper.Make(request, nil)
				Expect(err).To(Equal(innerError))
			})
		})
	})
})

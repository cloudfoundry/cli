package ccv2_test

import (
	"errors"
	"net/http"

	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/cloudcontrollerv2fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Token Refresh Wrapper", func() {
	var (
		fakeConnection *cloudcontrollerv2fakes.FakeConnection
		fakeStore      *cloudcontrollerv2fakes.FakeAuthenticationStore

		wrapper Connection
	)

	BeforeEach(func() {
		fakeConnection = new(cloudcontrollerv2fakes.FakeConnection)
		fakeStore = new(cloudcontrollerv2fakes.FakeAuthenticationStore)
		fakeStore.AccessTokenReturns("foobar")

		inner := NewTokenRefreshWrapper(fakeStore)
		wrapper = inner.Wrap(fakeConnection)
	})

	Describe("Make", func() {
		It("adds authentication headers", func() {
			request := Request{}
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

				request := Request{
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

				request := Request{}
				err := wrapper.Make(request, nil)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when the wrapped connection returns an error", func() {
			It("returns the error", func() {
				innerError := errors.New("inner error")
				fakeConnection.MakeReturns(innerError)

				request := Request{}
				err := wrapper.Make(request, nil)
				Expect(err).To(Equal(innerError))
			})
		})
	})
})

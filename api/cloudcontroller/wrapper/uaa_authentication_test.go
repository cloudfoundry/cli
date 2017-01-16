package wrapper_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/cloudcontrollerfakes"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/wrapper"
	"code.cloudfoundry.org/cli/api/cloudcontroller/wrapper/wrapperfakes"
	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/api/uaa/wrapper/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UAA Authentication", func() {
	var (
		fakeConnection *cloudcontrollerfakes.FakeConnection
		fakeClient     *wrapperfakes.FakeUAAClient
		inMemoryCache  *util.InMemoryCache

		wrapper cloudcontroller.Connection
		request *http.Request
	)

	BeforeEach(func() {
		fakeConnection = new(cloudcontrollerfakes.FakeConnection)
		fakeClient = new(wrapperfakes.FakeUAAClient)
		inMemoryCache = util.NewInMemoryTokenCache()
		inMemoryCache.SetAccessToken("a-ok")

		inner := NewUAAAuthentication(fakeClient, inMemoryCache)
		wrapper = inner.Wrap(fakeConnection)

		request = &http.Request{
			Header: http.Header{},
		}
	})

	Describe("Make", func() {
		Context("when the token is valid", func() {
			It("adds authentication headers", func() {
				wrapper.Make(request, nil)

				Expect(fakeConnection.MakeCallCount()).To(Equal(1))
				authenticatedRequest, _ := fakeConnection.MakeArgsForCall(0)
				headers := authenticatedRequest.Header
				Expect(headers["Authorization"]).To(ConsistOf([]string{"a-ok"}))
			})

			Context("when the request already has headers", func() {
				It("preserves existing headers", func() {
					request.Header.Add("Existing", "header")
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

					err := wrapper.Make(request, nil)
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Context("when the wrapped connection returns an error", func() {
				It("returns the error", func() {
					innerError := errors.New("inner error")
					fakeConnection.MakeReturns(innerError)

					err := wrapper.Make(request, nil)
					Expect(err).To(Equal(innerError))
				})
			})
		})

		Context("when the token is invalid", func() {
			var expectedBody string

			BeforeEach(func() {
				expectedBody = "this body content should be preserved"
				request.Body = ioutil.NopCloser(strings.NewReader(expectedBody))

				makeCount := 0
				fakeConnection.MakeStub = func(request *http.Request, response *cloudcontroller.Response) error {
					body, err := ioutil.ReadAll(request.Body)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(body)).To(Equal(expectedBody))

					if makeCount == 0 {
						makeCount += 1
						return cloudcontroller.InvalidAuthTokenError{}
					} else {
						return nil
					}
				}

				inMemoryCache.SetAccessToken("what")

				fakeClient.RefreshAccessTokenReturns(
					uaa.RefreshToken{
						AccessToken:  "foobar-2",
						RefreshToken: "bananananananana",
						Type:         "bearer",
					},
					nil,
				)

				err := wrapper.Make(request, nil)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should refresh the token", func() {
				Expect(fakeClient.RefreshAccessTokenCallCount()).To(Equal(1))
			})

			It("should resend the request", func() {
				Expect(fakeConnection.MakeCallCount()).To(Equal(2))

				request, _ := fakeConnection.MakeArgsForCall(1)
				Expect(request.Header.Get("Authorization")).To(Equal("bearer foobar-2"))
			})

			It("should save the refresh token", func() {
				Expect(inMemoryCache.RefreshToken()).To(Equal("bananananananana"))
			})
		})
	})
})

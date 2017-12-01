package wrapper_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/api/uaa/uaafakes"
	. "code.cloudfoundry.org/cli/api/uaa/wrapper"
	"code.cloudfoundry.org/cli/api/uaa/wrapper/util"
	"code.cloudfoundry.org/cli/api/uaa/wrapper/wrapperfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UAA Authentication", func() {
	var (
		fakeConnection *uaafakes.FakeConnection
		fakeClient     *wrapperfakes.FakeUAAClient
		inMemoryCache  *util.InMemoryCache

		wrapper uaa.Connection
		request *http.Request
		inner   *UAAAuthentication
	)

	BeforeEach(func() {
		fakeConnection = new(uaafakes.FakeConnection)
		fakeClient = new(wrapperfakes.FakeUAAClient)
		inMemoryCache = util.NewInMemoryTokenCache()

		inner = NewUAAAuthentication(fakeClient, inMemoryCache)
		wrapper = inner.Wrap(fakeConnection)
	})

	Describe("Make", func() {
		Context("when the client is nil", func() {
			BeforeEach(func() {
				inner.SetClient(nil)

				fakeConnection.MakeReturns(uaa.InvalidAuthTokenError{})
			})

			It("calls the connection without any side effects", func() {
				err := wrapper.Make(request, nil)
				Expect(err).To(MatchError(uaa.InvalidAuthTokenError{}))

				Expect(fakeClient.RefreshAccessTokenCallCount()).To(Equal(0))
				Expect(fakeConnection.MakeCallCount()).To(Equal(1))
			})
		})

		Context("when the token is valid", func() {
			BeforeEach(func() {
				request = &http.Request{
					Header: http.Header{},
				}
				inMemoryCache.SetAccessToken("a-ok")
			})

			It("adds authentication headers", func() {
				err := wrapper.Make(request, nil)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeConnection.MakeCallCount()).To(Equal(1))
				authenticatedRequest, _ := fakeConnection.MakeArgsForCall(0)
				headers := authenticatedRequest.Header
				Expect(headers["Authorization"]).To(ConsistOf([]string{"a-ok"}))
			})

			Context("when the request already has headers", func() {
				It("preserves existing headers", func() {
					request.Header.Add("Existing", "header")
					err := wrapper.Make(request, nil)
					Expect(err).ToNot(HaveOccurred())

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
				request, err := http.NewRequest(
					http.MethodGet,
					server.URL(),
					ioutil.NopCloser(strings.NewReader(expectedBody)),
				)
				Expect(err).NotTo(HaveOccurred())

				makeCount := 0
				fakeConnection.MakeStub = func(request *http.Request, response *uaa.Response) error {
					body, readErr := ioutil.ReadAll(request.Body)
					Expect(readErr).NotTo(HaveOccurred())
					Expect(string(body)).To(Equal(expectedBody))

					if makeCount == 0 {
						makeCount++
						return uaa.InvalidAuthTokenError{}
					} else {
						return nil
					}
				}

				fakeClient.RefreshAccessTokenReturns(
					uaa.RefreshedTokens{
						AccessToken:  "foobar-2",
						RefreshToken: "bananananananana",
						Type:         "bearer",
					},
					nil,
				)

				inMemoryCache.SetAccessToken("what")

				err = wrapper.Make(request, nil)
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

		Context("when refreshing the token", func() {
			var originalAuthHeader string
			BeforeEach(func() {
				body := strings.NewReader(url.Values{
					"grant_type": {"refresh_token"},
				}.Encode())

				request, err := http.NewRequest("POST", fmt.Sprintf("%s/oauth/token", server.URL()), body)
				Expect(err).NotTo(HaveOccurred())
				request.SetBasicAuth("some-user", "some-password")
				originalAuthHeader = request.Header.Get("Authorization")

				inMemoryCache.SetAccessToken("some-access-token")

				err = wrapper.Make(request, nil)
				Expect(err).ToNot(HaveOccurred())
			})

			It("does not change the 'Authorization' header", func() {
				Expect(fakeConnection.MakeCallCount()).To(Equal(1))

				request, _ := fakeConnection.MakeArgsForCall(0)
				Expect(request.Header.Get("Authorization")).To(Equal(originalAuthHeader))
			})
		})

		Context("when logging in", func() {
			var originalAuthHeader string
			BeforeEach(func() {
				body := strings.NewReader(url.Values{
					"grant_type": {"password"},
				}.Encode())

				request, err := http.NewRequest("POST", fmt.Sprintf("%s/oauth/token", server.URL()), body)
				Expect(err).NotTo(HaveOccurred())
				request.SetBasicAuth("some-user", "some-password")
				originalAuthHeader = request.Header.Get("Authorization")

				inMemoryCache.SetAccessToken("some-access-token")

				err = wrapper.Make(request, nil)
				Expect(err).ToNot(HaveOccurred())
			})

			It("does not change the 'Authorization' header", func() {
				Expect(fakeConnection.MakeCallCount()).To(Equal(1))

				request, _ := fakeConnection.MakeArgsForCall(0)
				Expect(request.Header.Get("Authorization")).To(Equal(originalAuthHeader))
			})
		})
	})
})

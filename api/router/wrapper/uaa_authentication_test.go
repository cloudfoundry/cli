package wrapper_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"code.cloudfoundry.org/cli/api/router"
	"code.cloudfoundry.org/cli/api/router/routererror"
	"code.cloudfoundry.org/cli/api/router/routerfakes"
	. "code.cloudfoundry.org/cli/api/router/wrapper"
	"code.cloudfoundry.org/cli/api/router/wrapper/wrapperfakes"
	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/api/uaa/wrapper/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UAA Authentication", func() {
	var (
		fakeConnection *routerfakes.FakeConnection
		fakeClient     *wrapperfakes.FakeUAAClient
		inMemoryCache  *util.InMemoryCache

		wrapper router.Connection
		request *router.Request
		inner   *UAAAuthentication
	)

	BeforeEach(func() {
		fakeConnection = new(routerfakes.FakeConnection)
		fakeClient = new(wrapperfakes.FakeUAAClient)
		inMemoryCache = util.NewInMemoryTokenCache()
		inMemoryCache.SetAccessToken("a-ok")

		inner = NewUAAAuthentication(fakeClient, inMemoryCache)
		wrapper = inner.Wrap(fakeConnection)

		request = &router.Request{
			Request: &http.Request{
				Header: http.Header{},
			},
		}
	})

	Describe("Make", func() {
		When("the client is nil", func() {
			BeforeEach(func() {
				inner.SetClient(nil)

				fakeConnection.MakeReturns(routererror.InvalidAuthTokenError{})
			})

			It("calls the connection without any side effects", func() {
				err := wrapper.Make(request, nil)
				Expect(err).To(MatchError(routererror.InvalidAuthTokenError{}))

				Expect(fakeClient.RefreshAccessTokenCallCount()).To(Equal(0))
				Expect(fakeConnection.MakeCallCount()).To(Equal(1))
			})
		})

		When("the token is valid", func() {
			It("adds authentication headers", func() {
				err := wrapper.Make(request, nil)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeConnection.MakeCallCount()).To(Equal(1))
				authenticatedRequest, _ := fakeConnection.MakeArgsForCall(0)
				headers := authenticatedRequest.Header
				Expect(headers["Authorization"]).To(ConsistOf([]string{"a-ok"}))
			})

			When("the request already has headers", func() {
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

			When("the wrapped connection returns nil", func() {
				It("returns nil", func() {
					fakeConnection.MakeReturns(nil)

					err := wrapper.Make(request, nil)
					Expect(err).ToNot(HaveOccurred())
				})
			})

			When("the wrapped connection returns an error", func() {
				It("returns the error", func() {
					innerError := errors.New("inner error")
					fakeConnection.MakeReturns(innerError)

					err := wrapper.Make(request, nil)
					Expect(err).To(Equal(innerError))
				})
			})
		})

		When("the token is invalid", func() {
			var (
				expectedBody string
				request      *router.Request
				executeErr   error
			)

			BeforeEach(func() {
				expectedBody = "this body content should be preserved"
				body := strings.NewReader(expectedBody)
				request = router.NewRequest(&http.Request{
					Header: http.Header{},
					Body:   ioutil.NopCloser(body),
				}, body)

				makeCount := 0
				fakeConnection.MakeStub = func(request *router.Request, response *router.Response) error {
					body, err := ioutil.ReadAll(request.Body)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(body)).To(Equal(expectedBody))

					if makeCount == 0 {
						makeCount++
						return routererror.InvalidAuthTokenError{}
					} else {
						return nil
					}
				}

				inMemoryCache.SetAccessToken("what")

				fakeClient.RefreshAccessTokenReturns(
					uaa.RefreshedTokens{
						AccessToken:  "foobar-2",
						RefreshToken: "bananananananana",
						Type:         "bearer",
					},
					nil,
				)
			})

			JustBeforeEach(func() {
				executeErr = wrapper.Make(request, nil)
			})

			It("should refresh the token", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeClient.RefreshAccessTokenCallCount()).To(Equal(1))
			})

			It("should resend the request", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeConnection.MakeCallCount()).To(Equal(2))

				requestArg, _ := fakeConnection.MakeArgsForCall(1)
				Expect(requestArg.Header.Get("Authorization")).To(Equal("bearer foobar-2"))
			})

			It("should save the refresh token", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(inMemoryCache.RefreshToken()).To(Equal("bananananananana"))
			})

			When("a PipeSeekError is returned from ResetBody", func() {
				BeforeEach(func() {
					body, writer := router.NewPipeBomb()
					req, err := http.NewRequest(http.MethodGet, "https://foo.bar.com/banana", body)
					Expect(err).NotTo(HaveOccurred())
					request = router.NewRequest(req, body)

					go func() {
						defer GinkgoRecover()

						_, err := writer.Write([]byte(expectedBody))
						Expect(err).NotTo(HaveOccurred())
						err = writer.Close()
						Expect(err).NotTo(HaveOccurred())
					}()
				})

				It("set the err on PipeSeekError", func() {
					Expect(executeErr).To(MatchError(routererror.PipeSeekError{Err: routererror.InvalidAuthTokenError{}}))
				})
			})
		})
	})
})

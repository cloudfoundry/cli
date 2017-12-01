package wrapper_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
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
		request *cloudcontroller.Request
		inner   *UAAAuthentication
	)

	BeforeEach(func() {
		fakeConnection = new(cloudcontrollerfakes.FakeConnection)
		fakeClient = new(wrapperfakes.FakeUAAClient)
		inMemoryCache = util.NewInMemoryTokenCache()
		inMemoryCache.SetAccessToken("a-ok")

		inner = NewUAAAuthentication(fakeClient, inMemoryCache)
		wrapper = inner.Wrap(fakeConnection)

		request = &cloudcontroller.Request{
			Request: &http.Request{
				Header: http.Header{},
			},
		}
	})

	Describe("Make", func() {
		Context("when the client is nil", func() {
			BeforeEach(func() {
				inner.SetClient(nil)

				fakeConnection.MakeReturns(ccerror.InvalidAuthTokenError{})
			})

			It("calls the connection without any side effects", func() {
				err := wrapper.Make(request, nil)
				Expect(err).To(MatchError(ccerror.InvalidAuthTokenError{}))

				Expect(fakeClient.RefreshAccessTokenCallCount()).To(Equal(0))
				Expect(fakeConnection.MakeCallCount()).To(Equal(1))
			})
		})

		Context("when the token is valid", func() {
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
			var (
				expectedBody string
				request      *cloudcontroller.Request
				executeErr   error
			)

			BeforeEach(func() {
				expectedBody = "this body content should be preserved"
				body := strings.NewReader(expectedBody)
				request = cloudcontroller.NewRequest(&http.Request{
					Header: http.Header{},
					Body:   ioutil.NopCloser(body),
				}, body)

				makeCount := 0
				fakeConnection.MakeStub = func(request *cloudcontroller.Request, response *cloudcontroller.Response) error {
					body, err := ioutil.ReadAll(request.Body)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(body)).To(Equal(expectedBody))

					if makeCount == 0 {
						makeCount++
						return ccerror.InvalidAuthTokenError{}
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

			Context("when a PipeSeekError is returned from ResetBody", func() {
				BeforeEach(func() {
					body, writer := cloudcontroller.NewPipeBomb()
					req, err := http.NewRequest(http.MethodGet, "https://foo.bar.com/banana", body)
					Expect(err).NotTo(HaveOccurred())
					request = cloudcontroller.NewRequest(req, body)

					go func() {
						defer GinkgoRecover()

						_, err := writer.Write([]byte(expectedBody))
						Expect(err).NotTo(HaveOccurred())
						err = writer.Close()
						Expect(err).NotTo(HaveOccurred())
					}()
				})

				It("set the err on PipeSeekError", func() {
					Expect(executeErr).To(MatchError(ccerror.PipeSeekError{Err: ccerror.InvalidAuthTokenError{}}))
				})
			})
		})
	})
})

package wrapper_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking"
	"code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking/cfnetworkingfakes"
	. "code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking/wrapper"
	"code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking/wrapper/util"
	"code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking/wrapper/wrapperfakes"
	"code.cloudfoundry.org/cli/api/uaa"

	"code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking/networkerror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UAA Authentication", func() {
	var (
		fakeConnection *cfnetworkingfakes.FakeConnection
		fakeClient     *wrapperfakes.FakeUAAClient
		inMemoryCache  *util.InMemoryCache

		wrapper cfnetworking.Connection
		request *cfnetworking.Request
		inner   *UAAAuthentication
	)

	BeforeEach(func() {
		fakeConnection = new(cfnetworkingfakes.FakeConnection)
		fakeClient = new(wrapperfakes.FakeUAAClient)
		inMemoryCache = util.NewInMemoryTokenCache()
		inMemoryCache.SetAccessToken("a-ok")

		inner = NewUAAAuthentication(fakeClient, inMemoryCache)
		wrapper = inner.Wrap(fakeConnection)

		request = &cfnetworking.Request{
			Request: &http.Request{
				Header: http.Header{},
			},
		}
	})

	Describe("Make", func() {
		It("adds authentication headers", func() {
			err := wrapper.Make(request, nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeConnection.MakeCallCount()).To(Equal(1))
			authenticatedRequest, _ := fakeConnection.MakeArgsForCall(0)
			headers := authenticatedRequest.Header
			Expect(headers["Authorization"]).To(ConsistOf([]string{"a-ok"}))
		})

		Context("when the token is valid", func() {
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
				request      *cfnetworking.Request
				executeErr   error
			)

			BeforeEach(func() {
				expectedBody = "this body content should be preserved"
				body := strings.NewReader(expectedBody)
				request = cfnetworking.NewRequest(&http.Request{
					Header: http.Header{},
					Body:   ioutil.NopCloser(body),
				}, body)

				makeCount := 0
				fakeConnection.MakeStub = func(request *cfnetworking.Request, response *cfnetworking.Response) error {
					body, err := ioutil.ReadAll(request.Body)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(body)).To(Equal(expectedBody))

					if makeCount == 0 {
						makeCount += 1
						return networkerror.InvalidAuthTokenError{}
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

			Context("when the reseting the request body fails", func() {
				BeforeEach(func() {
					fakeConnection.MakeReturnsOnCall(0, networkerror.InvalidAuthTokenError{})

					fakeReadSeeker := new(cfnetworkingfakes.FakeReadSeeker)
					fakeReadSeeker.SeekReturns(0, errors.New("oh noes"))

					req, err := http.NewRequest(http.MethodGet, "https://foo.bar.com/banana", fakeReadSeeker)
					Expect(err).NotTo(HaveOccurred())
					request = cfnetworking.NewRequest(req, fakeReadSeeker)
				})

				It("returns error on seek", func() {
					Expect(executeErr).To(MatchError("oh noes"))
				})
			})
		})
	})
})

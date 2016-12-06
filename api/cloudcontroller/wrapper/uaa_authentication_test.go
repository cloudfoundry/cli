package wrapper_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/cloudcontrollerfakes"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/wrapper"
	"code.cloudfoundry.org/cli/api/cloudcontroller/wrapper/wrapperfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UAA Authentication", func() {
	var (
		fakeConnection *cloudcontrollerfakes.FakeConnection
		fakeClient     *wrapperfakes.FakeUAAClient

		wrapper cloudcontroller.Connection
		request *http.Request
	)

	BeforeEach(func() {
		fakeConnection = new(cloudcontrollerfakes.FakeConnection)
		fakeClient = new(wrapperfakes.FakeUAAClient)
		fakeClient.AccessTokenReturns("foobar")

		inner := NewUAAAuthentication(fakeClient)
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
				Expect(headers["Authorization"]).To(ConsistOf([]string{"foobar"}))
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

				count := 0
				fakeClient.AccessTokenStub = func() string {
					count = count + 1
					return fmt.Sprintf("foobar-%d", count)
				}

				err := wrapper.Make(request, nil)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should refresh the token", func() {
				Expect(fakeClient.RefreshTokenCallCount()).To(Equal(1))
			})

			It("should resend the request", func() {
				Expect(fakeConnection.MakeCallCount()).To(Equal(2))

				request, _ := fakeConnection.MakeArgsForCall(1)
				Expect(request.Header.Get("Authorization")).To(Equal("foobar-2"))
			})
		})
	})
})

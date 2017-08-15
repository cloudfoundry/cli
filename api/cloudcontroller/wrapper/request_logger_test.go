package wrapper_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/cloudcontrollerfakes"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/wrapper"
	"code.cloudfoundry.org/cli/api/cloudcontroller/wrapper/wrapperfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Request Logger", func() {
	var (
		fakeConnection *cloudcontrollerfakes.FakeConnection
		fakeOutput     *wrapperfakes.FakeRequestLoggerOutput

		wrapper cloudcontroller.Connection

		request  *cloudcontroller.Request
		response *cloudcontroller.Response
		makeErr  error
	)

	BeforeEach(func() {
		fakeConnection = new(cloudcontrollerfakes.FakeConnection)
		fakeOutput = new(wrapperfakes.FakeRequestLoggerOutput)

		wrapper = NewRequestLogger(fakeOutput).Wrap(fakeConnection)

		body := bytes.NewReader([]byte("foo"))

		req, err := http.NewRequest(http.MethodGet, "https://foo.bar.com/banana", body)
		Expect(err).NotTo(HaveOccurred())

		req.URL.RawQuery = url.Values{
			"query1": {"a"},
			"query2": {"b"},
		}.Encode()

		headers := http.Header{}
		headers.Add("Aghi", "bar")
		headers.Add("Abc", "json")
		headers.Add("Adef", "application/json")
		req.Header = headers

		response = &cloudcontroller.Response{
			RawResponse:  []byte("some-response-body"),
			HTTPResponse: &http.Response{},
		}
		request = cloudcontroller.NewRequest(req, body)
	})

	JustBeforeEach(func() {
		makeErr = wrapper.Make(request, response)
	})

	Describe("Make", func() {
		It("outputs the request", func() {
			Expect(makeErr).NotTo(HaveOccurred())

			Expect(fakeOutput.DisplayTypeCallCount()).To(BeNumerically(">=", 1))
			name, date := fakeOutput.DisplayTypeArgsForCall(0)
			Expect(name).To(Equal("REQUEST"))
			Expect(date).To(BeTemporally("~", time.Now(), time.Second))

			Expect(fakeOutput.DisplayRequestHeaderCallCount()).To(Equal(1))
			method, uri, protocol := fakeOutput.DisplayRequestHeaderArgsForCall(0)
			Expect(method).To(Equal(http.MethodGet))
			Expect(uri).To(MatchRegexp("/banana\\?(?:query1=a&query2=b|query2=b&query1=a)"))
			Expect(protocol).To(Equal("HTTP/1.1"))

			Expect(fakeOutput.DisplayHostCallCount()).To(Equal(1))
			host := fakeOutput.DisplayHostArgsForCall(0)
			Expect(host).To(Equal("foo.bar.com"))

			Expect(fakeOutput.DisplayHeaderCallCount()).To(BeNumerically(">=", 3))
			name, value := fakeOutput.DisplayHeaderArgsForCall(0)
			Expect(name).To(Equal("Abc"))
			Expect(value).To(Equal("json"))
			name, value = fakeOutput.DisplayHeaderArgsForCall(1)
			Expect(name).To(Equal("Adef"))
			Expect(value).To(Equal("application/json"))
			name, value = fakeOutput.DisplayHeaderArgsForCall(2)
			Expect(name).To(Equal("Aghi"))
			Expect(value).To(Equal("bar"))

			Expect(fakeOutput.DisplayMessageCallCount()).To(Equal(0))
		})

		Context("when an authorization header is in the request", func() {
			BeforeEach(func() {
				request.Header = http.Header{"Authorization": []string{"should not be shown"}}
			})

			It("redacts the contents of the authorization header", func() {
				Expect(makeErr).NotTo(HaveOccurred())
				Expect(fakeOutput.DisplayHeaderCallCount()).To(Equal(1))
				key, value := fakeOutput.DisplayHeaderArgsForCall(0)
				Expect(key).To(Equal("Authorization"))
				Expect(value).To(Equal("[PRIVATE DATA HIDDEN]"))
			})
		})

		Context("when passed a body", func() {
			Context("when the request's Content-Type is application/json", func() {
				BeforeEach(func() {
					request.Header.Set("Content-Type", "application/json")
				})

				It("outputs the body", func() {
					Expect(makeErr).NotTo(HaveOccurred())

					Expect(fakeOutput.DisplayJSONBodyCallCount()).To(BeNumerically(">=", 1))
					Expect(fakeOutput.DisplayJSONBodyArgsForCall(0)).To(Equal([]byte("foo")))

					bytes, err := ioutil.ReadAll(request.Body)
					Expect(err).NotTo(HaveOccurred())
					Expect(bytes).To(Equal([]byte("foo")))
				})
			})

			Context("when the request's Content-Type is application/x-www-form-urlencoded", func() {
				BeforeEach(func() {
					request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				})

				It("outputs the body", func() {
					Expect(makeErr).NotTo(HaveOccurred())

					bytes, err := ioutil.ReadAll(request.Body)
					Expect(err).NotTo(HaveOccurred())
					Expect(bytes).To(Equal([]byte("foo")))
					Expect(fakeOutput.DisplayMessageCallCount()).To(Equal(1))
					Expect(fakeOutput.DisplayMessageArgsForCall(0)).To(Equal("[application/x-www-form-urlencoded foo]"))
				})
			})

			Context("when request's Content-Type is anything else", func() {
				BeforeEach(func() {
					request.Header.Set("Content-Type", "banana;rama")
				})

				It("does not display the body", func() {
					Expect(makeErr).NotTo(HaveOccurred())
					Expect(fakeOutput.DisplayJSONBodyCallCount()).To(Equal(1)) // Once for response body only
					Expect(fakeOutput.DisplayMessageCallCount()).To(Equal(1))
					Expect(fakeOutput.DisplayMessageArgsForCall(0)).To(Equal("[banana Content Hidden]"))
				})
			})
		})

		Context("when an error occures while trying to log the request", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("this should never block the request")

				calledOnce := false
				fakeOutput.StartStub = func() error {
					if !calledOnce {
						calledOnce = true
						return expectedErr
					}
					return nil
				}
			})

			It("should display the error and continue on", func() {
				Expect(makeErr).NotTo(HaveOccurred())

				Expect(fakeOutput.HandleInternalErrorCallCount()).To(Equal(1))
				Expect(fakeOutput.HandleInternalErrorArgsForCall(0)).To(MatchError(expectedErr))
			})
		})

		Context("when the request is successful", func() {
			BeforeEach(func() {
				response = &cloudcontroller.Response{
					RawResponse: []byte("some-response-body"),
					HTTPResponse: &http.Response{
						Proto:  "HTTP/1.1",
						Status: "200 OK",
						Header: http.Header{
							"BBBBB": {"second"},
							"AAAAA": {"first"},
							"CCCCC": {"third"},
						},
					},
				}
			})

			It("outputs the response", func() {
				Expect(makeErr).NotTo(HaveOccurred())

				Expect(fakeOutput.DisplayTypeCallCount()).To(Equal(2))
				name, date := fakeOutput.DisplayTypeArgsForCall(1)
				Expect(name).To(Equal("RESPONSE"))
				Expect(date).To(BeTemporally("~", time.Now(), time.Second))

				Expect(fakeOutput.DisplayResponseHeaderCallCount()).To(Equal(1))
				protocol, status := fakeOutput.DisplayResponseHeaderArgsForCall(0)
				Expect(protocol).To(Equal("HTTP/1.1"))
				Expect(status).To(Equal("200 OK"))

				Expect(fakeOutput.DisplayHeaderCallCount()).To(BeNumerically(">=", 6))
				name, value := fakeOutput.DisplayHeaderArgsForCall(3)
				Expect(name).To(Equal("AAAAA"))
				Expect(value).To(Equal("first"))
				name, value = fakeOutput.DisplayHeaderArgsForCall(4)
				Expect(name).To(Equal("BBBBB"))
				Expect(value).To(Equal("second"))
				name, value = fakeOutput.DisplayHeaderArgsForCall(5)
				Expect(name).To(Equal("CCCCC"))
				Expect(value).To(Equal("third"))

				Expect(fakeOutput.DisplayJSONBodyCallCount()).To(BeNumerically(">=", 1))
				Expect(fakeOutput.DisplayJSONBodyArgsForCall(0)).To(Equal([]byte("some-response-body")))
			})
		})

		Context("when the request is unsuccessful", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("banana")
				fakeConnection.MakeReturns(expectedErr)
			})

			Context("when the http response is not set", func() {
				BeforeEach(func() {
					response = &cloudcontroller.Response{}
				})

				It("outputs nothing", func() {
					Expect(makeErr).To(MatchError(expectedErr))
					Expect(fakeOutput.DisplayResponseHeaderCallCount()).To(Equal(0))
				})
			})

			Context("when the http response is set", func() {
				BeforeEach(func() {
					response = &cloudcontroller.Response{
						RawResponse: []byte("some-error-body"),
						HTTPResponse: &http.Response{
							Proto:  "HTTP/1.1",
							Status: "200 OK",
							Header: http.Header{
								"BBBBB": {"second"},
								"AAAAA": {"first"},
								"CCCCC": {"third"},
							},
						},
					}
				})

				It("outputs the response", func() {
					Expect(makeErr).To(MatchError(expectedErr))

					Expect(fakeOutput.DisplayTypeCallCount()).To(Equal(2))
					name, date := fakeOutput.DisplayTypeArgsForCall(1)
					Expect(name).To(Equal("RESPONSE"))
					Expect(date).To(BeTemporally("~", time.Now(), time.Second))

					Expect(fakeOutput.DisplayResponseHeaderCallCount()).To(Equal(1))
					protocol, status := fakeOutput.DisplayResponseHeaderArgsForCall(0)
					Expect(protocol).To(Equal("HTTP/1.1"))
					Expect(status).To(Equal("200 OK"))

					Expect(fakeOutput.DisplayHeaderCallCount()).To(BeNumerically(">=", 6))
					name, value := fakeOutput.DisplayHeaderArgsForCall(3)
					Expect(name).To(Equal("AAAAA"))
					Expect(value).To(Equal("first"))
					name, value = fakeOutput.DisplayHeaderArgsForCall(4)
					Expect(name).To(Equal("BBBBB"))
					Expect(value).To(Equal("second"))
					name, value = fakeOutput.DisplayHeaderArgsForCall(5)
					Expect(name).To(Equal("CCCCC"))
					Expect(value).To(Equal("third"))

					Expect(fakeOutput.DisplayJSONBodyCallCount()).To(BeNumerically(">=", 1))
					Expect(fakeOutput.DisplayJSONBodyArgsForCall(0)).To(Equal([]byte("some-error-body")))
				})
			})
		})

		Context("when an error occures while trying to log the response", func() {
			var (
				originalErr error
				expectedErr error
			)

			BeforeEach(func() {
				originalErr = errors.New("this error should not be overwritten")
				fakeConnection.MakeReturns(originalErr)

				expectedErr = errors.New("this should never block the request")

				calledOnce := false
				fakeOutput.StartStub = func() error {
					if !calledOnce {
						calledOnce = true
						return nil
					}
					return expectedErr
				}
			})

			It("should display the error and continue on", func() {
				Expect(makeErr).To(MatchError(originalErr))

				Expect(fakeOutput.HandleInternalErrorCallCount()).To(Equal(1))
				Expect(fakeOutput.HandleInternalErrorArgsForCall(0)).To(MatchError(expectedErr))
			})
		})

		It("starts and stops the output", func() {
			Expect(fakeOutput.StartCallCount()).To(Equal(2))
			Expect(fakeOutput.StopCallCount()).To(Equal(2))
		})

		Context("when displaying the logs have an error", func() {
			var expectedErr error
			BeforeEach(func() {
				expectedErr = errors.New("Display error on request")
				fakeOutput.StartReturns(expectedErr)
			})

			It("calls handle internal error", func() {
				Expect(makeErr).ToNot(HaveOccurred())

				Expect(fakeOutput.HandleInternalErrorCallCount()).To(Equal(2))
				Expect(fakeOutput.HandleInternalErrorArgsForCall(0)).To(MatchError(expectedErr))
				Expect(fakeOutput.HandleInternalErrorArgsForCall(1)).To(MatchError(expectedErr))
			})
		})
	})
})

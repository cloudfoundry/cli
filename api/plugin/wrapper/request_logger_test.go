package wrapper_test

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"code.cloudfoundry.org/cli/api/plugin"
	"code.cloudfoundry.org/cli/api/plugin/pluginfakes"
	. "code.cloudfoundry.org/cli/api/plugin/wrapper"
	"code.cloudfoundry.org/cli/api/plugin/wrapper/wrapperfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Request Logger", func() {
	var (
		fakeConnection  *pluginfakes.FakeConnection
		fakeOutput      *wrapperfakes.FakeRequestLoggerOutput
		fakeProxyReader *pluginfakes.FakeProxyReader

		wrapper plugin.Connection

		request  *http.Request
		response *plugin.Response
		err      error
	)

	BeforeEach(func() {
		fakeConnection = new(pluginfakes.FakeConnection)
		fakeOutput = new(wrapperfakes.FakeRequestLoggerOutput)
		fakeProxyReader = new(pluginfakes.FakeProxyReader)

		wrapper = NewRequestLogger(fakeOutput).Wrap(fakeConnection)

		var err error
		request, err = http.NewRequest(http.MethodGet, "https://foo.bar.com/banana", nil)
		Expect(err).NotTo(HaveOccurred())

		request.URL.RawQuery = url.Values{
			"query1": {"a"},
			"query2": {"b"},
		}.Encode()

		headers := http.Header{}
		headers.Add("Aghi", "bar")
		headers.Add("Abc", "json")
		headers.Add("Adef", "application/json")
		request.Header = headers

		response = &plugin.Response{
			RawResponse:  []byte("some-response-body"),
			HTTPResponse: &http.Response{},
		}
	})

	JustBeforeEach(func() {
		err = wrapper.Make(request, response, fakeProxyReader)
	})

	Describe("Make", func() {
		It("outputs the request", func() {
			Expect(err).NotTo(HaveOccurred())

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

			Expect(err).ToNot(HaveOccurred())
			Expect(fakeConnection.MakeCallCount()).To(Equal(1))
			_, _, proxyReader := fakeConnection.MakeArgsForCall(0)
			Expect(proxyReader).To(Equal(fakeProxyReader))
		})

		Context("when an authorization header is in the request", func() {
			BeforeEach(func() {
				request.Header = http.Header{"Authorization": []string{"should not be shown"}}
			})

			It("redacts the contents of the authorization header", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeOutput.DisplayHeaderCallCount()).To(Equal(1))
				key, value := fakeOutput.DisplayHeaderArgsForCall(0)
				Expect(key).To(Equal("Authorization"))
				Expect(value).To(Equal("[PRIVATE DATA HIDDEN]"))
			})
		})

		Context("when passed a body", func() {
			Context("when the request's Content-Type is application/json", func() {
				var originalBody io.ReadCloser
				BeforeEach(func() {
					request.Header.Set("Content-Type", "application/json")
					originalBody = ioutil.NopCloser(bytes.NewReader([]byte("foo")))
					request.Body = originalBody
				})

				It("outputs the body", func() {
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeOutput.DisplayJSONBodyCallCount()).To(BeNumerically(">=", 1))
					Expect(fakeOutput.DisplayJSONBodyArgsForCall(0)).To(Equal([]byte("foo")))

					bytes, err := ioutil.ReadAll(request.Body)
					Expect(err).NotTo(HaveOccurred())
					Expect(bytes).To(Equal([]byte("foo")))
				})
			})

			Context("when request's Content-Type is anything else", func() {
				BeforeEach(func() {
					request.Header.Set("Content-Type", "banana")
				})

				It("does not display the request body", func() {
					Expect(fakeOutput.DisplayJSONBodyCallCount()).To(Equal(0))
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
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeOutput.HandleInternalErrorCallCount()).To(Equal(1))
				Expect(fakeOutput.HandleInternalErrorArgsForCall(0)).To(MatchError(expectedErr))
			})
		})

		Context("when the request is successful", func() {
			Context("when the response is JSON", func() {
				BeforeEach(func() {
					response = &plugin.Response{
						RawResponse: []byte(`{"some-key":"some-value"}`),
						HTTPResponse: &http.Response{
							Proto:  "HTTP/1.1",
							Status: "200 OK",
							Header: http.Header{
								"Content-Type": {"application/json"},
								"BBBBB":        {"second"},
								"AAAAA":        {"first"},
								"CCCCC":        {"third"},
							},
							Body: ioutil.NopCloser(bytes.NewReader([]byte(`{"some-key":"some-value"}`))),
						},
					}
				})

				It("outputs the response", func() {
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeOutput.DisplayTypeCallCount()).To(Equal(2))
					name, date := fakeOutput.DisplayTypeArgsForCall(1)
					Expect(name).To(Equal("RESPONSE"))
					Expect(date).To(BeTemporally("~", time.Now(), time.Second))

					Expect(fakeOutput.DisplayResponseHeaderCallCount()).To(Equal(1))
					protocol, status := fakeOutput.DisplayResponseHeaderArgsForCall(0)
					Expect(protocol).To(Equal("HTTP/1.1"))
					Expect(status).To(Equal("200 OK"))

					Expect(fakeOutput.DisplayHeaderCallCount()).To(BeNumerically(">=", 7))
					name, value := fakeOutput.DisplayHeaderArgsForCall(3)
					Expect(name).To(Equal("AAAAA"))
					Expect(value).To(Equal("first"))
					name, value = fakeOutput.DisplayHeaderArgsForCall(4)
					Expect(name).To(Equal("BBBBB"))
					Expect(value).To(Equal("second"))
					name, value = fakeOutput.DisplayHeaderArgsForCall(5)
					Expect(name).To(Equal("CCCCC"))
					Expect(value).To(Equal("third"))
					name, value = fakeOutput.DisplayHeaderArgsForCall(6)
					Expect(name).To(Equal("Content-Type"))
					Expect(value).To(Equal("application/json"))

					Expect(fakeOutput.DisplayJSONBodyCallCount()).To(BeNumerically(">=", 1))
					Expect(fakeOutput.DisplayJSONBodyArgsForCall(0)).To(Equal(response.RawResponse))

					Expect(fakeOutput.DisplayDumpCallCount()).To(Equal(0))
				})
			})

			Context("when the response is not JSON", func() {
				BeforeEach(func() {
					response = &plugin.Response{
						RawResponse: []byte(`not JSON`),
						HTTPResponse: &http.Response{
							Proto:  "HTTP/1.1",
							Status: "200 OK",
							Header: http.Header{
								"BBBBB": {"second"},
								"AAAAA": {"first"},
								"CCCCC": {"third"},
							},
							Body: ioutil.NopCloser(bytes.NewReader([]byte(`not JSON`))),
						},
					}
				})

				It("outputs the response", func() {
					Expect(err).NotTo(HaveOccurred())

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

					Expect(fakeOutput.DisplayDumpCallCount()).To(Equal(1))
					text := fakeOutput.DisplayDumpArgsForCall(0)
					Expect(text).To(Equal("[NON-JSON BODY CONTENT HIDDEN]"))

					Expect(fakeOutput.DisplayJSONBodyCallCount()).To(Equal(0))
				})
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
					response = &plugin.Response{}
				})

				It("outputs nothing", func() {
					Expect(err).To(MatchError(expectedErr))
					Expect(fakeOutput.DisplayResponseHeaderCallCount()).To(Equal(0))
				})
			})

			Context("when the http response body is nil", func() {
				BeforeEach(func() {
					response = &plugin.Response{
						HTTPResponse: &http.Response{Body: nil},
					}
				})

				It("does not output the response body", func() {
					Expect(err).To(MatchError(expectedErr))
					Expect(fakeOutput.DisplayResponseHeaderCallCount()).To(Equal(1))

					Expect(fakeOutput.DisplayJSONBodyCallCount()).To(Equal(0))
					Expect(fakeOutput.DisplayDumpCallCount()).To(Equal(0))
				})
			})

			Context("when the http response is set", func() {
				BeforeEach(func() {
					response = &plugin.Response{
						RawResponse: []byte("some-error-body"),
						HTTPResponse: &http.Response{
							Proto:  "HTTP/1.1",
							Status: "200 OK",
							Header: http.Header{
								"Content-Type": {"application/json"},
								"BBBBB":        {"second"},
								"AAAAA":        {"first"},
								"CCCCC":        {"third"},
							},
							Body: ioutil.NopCloser(bytes.NewReader([]byte(`some-error-body`))),
						},
					}
				})

				It("outputs the response", func() {
					Expect(err).To(MatchError(expectedErr))

					Expect(fakeOutput.DisplayTypeCallCount()).To(Equal(2))
					name, date := fakeOutput.DisplayTypeArgsForCall(1)
					Expect(name).To(Equal("RESPONSE"))
					Expect(date).To(BeTemporally("~", time.Now(), time.Second))

					Expect(fakeOutput.DisplayResponseHeaderCallCount()).To(Equal(1))
					protocol, status := fakeOutput.DisplayResponseHeaderArgsForCall(0)
					Expect(protocol).To(Equal("HTTP/1.1"))
					Expect(status).To(Equal("200 OK"))

					Expect(fakeOutput.DisplayHeaderCallCount()).To(BeNumerically(">=", 7))
					name, value := fakeOutput.DisplayHeaderArgsForCall(3)
					Expect(name).To(Equal("AAAAA"))
					Expect(value).To(Equal("first"))
					name, value = fakeOutput.DisplayHeaderArgsForCall(4)
					Expect(name).To(Equal("BBBBB"))
					Expect(value).To(Equal("second"))
					name, value = fakeOutput.DisplayHeaderArgsForCall(5)
					Expect(name).To(Equal("CCCCC"))
					Expect(value).To(Equal("third"))
					name, value = fakeOutput.DisplayHeaderArgsForCall(6)
					Expect(name).To(Equal("Content-Type"))
					Expect(value).To(Equal("application/json"))

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
				Expect(err).To(MatchError(originalErr))

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
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeOutput.HandleInternalErrorCallCount()).To(Equal(2))
				Expect(fakeOutput.HandleInternalErrorArgsForCall(0)).To(MatchError(expectedErr))
				Expect(fakeOutput.HandleInternalErrorArgsForCall(1)).To(MatchError(expectedErr))
			})
		})
	})
})

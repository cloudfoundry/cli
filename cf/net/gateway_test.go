package net_test

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"runtime"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/cf/api/authentication"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	. "code.cloudfoundry.org/cli/cf/net"
	"code.cloudfoundry.org/cli/cf/net/netfakes"
	"code.cloudfoundry.org/cli/cf/terminal/terminalfakes"
	"code.cloudfoundry.org/cli/cf/trace/tracefakes"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testnet "code.cloudfoundry.org/cli/util/testhelpers/net"
	"code.cloudfoundry.org/cli/version"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Gateway", func() {
	var (
		ccServer    *ghttp.Server
		ccGateway   Gateway
		uaaGateway  Gateway
		config      coreconfig.ReadWriter
		authRepo    authentication.Repository
		currentTime time.Time
		clock       func() time.Time

		client *netfakes.FakeHTTPClientInterface
	)

	BeforeEach(func() {
		currentTime = time.Unix(0, 0)
		clock = func() time.Time { return currentTime }
		config = testconfig.NewRepository()

		ccGateway = NewCloudControllerGateway(config, clock, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), "")
		ccGateway.PollingThrottle = 3 * time.Millisecond
		uaaGateway = NewUAAGateway(config, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), "")
	})

	Describe("async timeout", func() {
		Context("when the config has a positive async timeout", func() {
			It("inherits the async timeout from the config", func() {
				config.SetAsyncTimeout(9001)
				ccGateway = NewCloudControllerGateway(config, time.Now, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), "")
				Expect(ccGateway.AsyncTimeout()).To(Equal(9001 * time.Minute))
			})
		})
	})

	Describe("Connection errors", func() {
		var oldNewHTTPClient func(tr *http.Transport, dumper RequestDumper) HTTPClientInterface

		BeforeEach(func() {
			client = new(netfakes.FakeHTTPClientInterface)

			oldNewHTTPClient = NewHTTPClient
			NewHTTPClient = func(tr *http.Transport, dumper RequestDumper) HTTPClientInterface {
				return client
			}
		})

		AfterEach(func() {
			NewHTTPClient = oldNewHTTPClient
		})

		It("only retry when response body is nil and error occurred", func() {
			client.DoReturns(&http.Response{Status: "internal error", StatusCode: 500}, errors.New("internal error"))
			request, apiErr := ccGateway.NewRequest("GET", "https://example.com/v2/apps", "BEARER my-access-token", nil)
			Expect(apiErr).ToNot(HaveOccurred())

			_, apiErr = ccGateway.PerformRequest(request)
			Expect(client.DoCallCount()).To(Equal(1))
			Expect(apiErr).To(HaveOccurred())
		})

		It("Retries 3 times if we cannot contact the server", func() {
			client.DoReturns(nil, errors.New("Connection refused"))
			request, apiErr := ccGateway.NewRequest("GET", "https://example.com/v2/apps", "BEARER my-access-token", nil)
			Expect(apiErr).ToNot(HaveOccurred())

			_, apiErr = ccGateway.PerformRequest(request)
			Expect(apiErr).To(HaveOccurred())
			Expect(client.DoCallCount()).To(Equal(3))
		})
	})

	Describe("NewRequest", func() {
		var (
			request *Request
			apiErr  error
		)

		Context("when the body is nil", func() {
			BeforeEach(func() {
				request, apiErr = ccGateway.NewRequest("GET", "https://example.com/v2/apps", "BEARER my-access-token", nil)
				Expect(apiErr).NotTo(HaveOccurred())
			})

			It("does not use a ProgressReader as the SeekableBody", func() {
				Expect(reflect.TypeOf(request.SeekableBody)).To(BeNil())
			})

			It("sets the Authorization header", func() {
				Expect(request.HTTPReq.Header.Get("Authorization")).To(Equal("BEARER my-access-token"))
			})

			It("sets the accept header to application/json", func() {
				Expect(request.HTTPReq.Header.Get("accept")).To(Equal("application/json"))
			})

			It("sets the user agent header", func() {
				Expect(request.HTTPReq.Header.Get("User-Agent")).To(Equal("go-cli " + version.BinaryVersion + " / " + runtime.GOOS))
			})
		})

		Context("when the body is a file", func() {
			BeforeEach(func() {
				f, _ := os.Open("../../fixtures/test.file")
				request, apiErr = ccGateway.NewRequestForFile("PUT", "https://example.com/v2/apps", "BEARER my-access-token", f)
				Expect(apiErr).NotTo(HaveOccurred())
			})

			It("Uses a ProgressReader as the SeekableBody", func() {
				Expect(reflect.TypeOf(request.SeekableBody).String()).To(ContainSubstring("ProgressReader"))
			})

		})

	})

	Describe("PerformRequestForJSONResponse()", func() {
		BeforeEach(func() {
			ccServer = ghttp.NewServer()
			config.SetAPIEndpoint(ccServer.URL())
		})

		AfterEach(func() {
			ccServer.Close()
		})

		Context("When CC response with an api error", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/some-endpoint"),
						ghttp.VerifyHeader(http.Header{
							"accept": []string{"application/json"},
						}),
						ghttp.RespondWith(http.StatusUnauthorized, `{
  "code": 10003,
  "description": "You are not authorized to perform the requested action",
  "error_code": "CF-NotAuthorized"
}`),
					),
				)
			})

			It("tries to unmarshal error response into provided resource", func() {
				type apiErrResponse struct {
					Code        int    `json:"code,omitempty"`
					Description string `json:"description,omitempty"`
				}

				errResponse := new(apiErrResponse)
				request, _ := ccGateway.NewRequest("GET", config.APIEndpoint()+"/v2/some-endpoint", config.AccessToken(), nil)
				_, apiErr := ccGateway.PerformRequestForJSONResponse(request, errResponse)

				Expect(apiErr).To(HaveOccurred())
				Expect(errResponse.Code).To(Equal(10003))
			})

			It("ignores any unmarshal error and does not alter the api err response", func() {
				request, _ := ccGateway.NewRequest("GET", config.APIEndpoint()+"/v2/some-endpoint", config.AccessToken(), nil)
				_, apiErr := ccGateway.PerformRequestForJSONResponse(request, nil)

				Expect(apiErr.Error()).To(Equal("Server error, status code: 401, error code: 10003, message: You are not authorized to perform the requested action"))
			})

		})

	})

	Describe("CRUD methods", func() {
		Describe("Delete", func() {
			var apiServer *httptest.Server

			Describe("DeleteResourceSynchronously", func() {
				var queryParams string
				BeforeEach(func() {
					apiServer = httptest.NewTLSServer(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
						queryParams = request.URL.RawQuery
					}))
					ccGateway.SetTrustedCerts(apiServer.TLS.Certificates)
				})

				It("does not send the async=true flag", func() {
					err := ccGateway.DeleteResourceSynchronously(apiServer.URL, "/v2/foobars/SOME_GUID")
					Expect(err).NotTo(HaveOccurred())
					Expect(queryParams).ToNot(ContainSubstring("async=true"))
				})

				It("deletes a resource", func() {
					err := ccGateway.DeleteResource(apiServer.URL, "/v2/foobars/SOME_GUID")
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Context("when the config has an async timeout", func() {
				BeforeEach(func() {
					count := 0
					apiServer = httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
						switch request.URL.Path {
						case "/v2/foobars/SOME_GUID":
							writer.WriteHeader(http.StatusNoContent)
						case "/v2/foobars/TIMEOUT":
							currentTime = currentTime.Add(time.Minute * 31)
							fmt.Fprintln(writer, `
{
  "metadata": {
    "guid": "8438916f-5c00-4d44-a19b-1df65abe9d52",
    "created_at": "2014-05-15T19:15:01+00:00",
    "url": "/v2/jobs/8438916f-5c00-4d44-a19b-1df65abe9d52"
  },
  "entity": {
    "guid": "8438916f-5c00-4d44-a19b-1df65abe9d52",
    "status": "queued"
  }
}`)
							writer.WriteHeader(http.StatusAccepted)
						case "/v2/jobs/8438916f-5c00-4d44-a19b-1df65abe9d52":
							if count == 0 {
								count++
								currentTime = currentTime.Add(time.Minute * 31)

								writer.WriteHeader(http.StatusOK)
								fmt.Fprintln(writer, `
{
  "entity": {
    "guid": "8438916f-5c00-4d44-a19b-1df65abe9d52",
    "status": "queued"
  }
}`)
							} else {
								panic("FAIL")
							}
						default:
							panic("shouldn't have made call to this URL: " + request.URL.Path)
						}
					}))

					config.SetAsyncTimeout(30)
					ccGateway.SetTrustedCerts(apiServer.TLS.Certificates)
				})

				AfterEach(func() {
					apiServer.Close()
				})

				It("deletes a resource", func() {
					err := ccGateway.DeleteResource(apiServer.URL, "/v2/foobars/SOME_GUID")
					Expect(err).ToNot(HaveOccurred())
				})

				Context("when the request would take longer than the async timeout", func() {
					It("returns an error", func() {
						apiErr := ccGateway.DeleteResource(apiServer.URL, "/v2/foobars/TIMEOUT")
						Expect(apiErr).To(HaveOccurred())
						Expect(apiErr).To(BeAssignableToTypeOf(errors.NewAsyncTimeoutError("http://some.url")))
					})
				})
			})
		})
	})

	Describe("making an async request", func() {
		var (
			jobStatus     string
			apiServer     *httptest.Server
			authServer    *httptest.Server
			statusChannel chan string
		)

		BeforeEach(func() {
			jobStatus = "queued"
			statusChannel = make(chan string, 10)

			apiServer = httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				currentTime = currentTime.Add(time.Millisecond * 11)

				updateStatus, ok := <-statusChannel
				if ok {
					jobStatus = updateStatus
				}

				switch request.URL.Path {
				case "/v2/foo":
					fmt.Fprintln(writer, `{ "metadata": { "url": "/v2/jobs/the-job-guid" } }`)
				case "/v2/jobs/the-job-guid":
					fmt.Fprintf(writer, `
					{
						"entity": {
							"status": "%s",
							"error_details": {
								"description": "he's dead, Jim"
							}
						}
					}`, jobStatus)
				default:
					writer.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(writer, `"Unexpected request path '%s'"`, request.URL.Path)
				}
			}))

			authServer, _ = testnet.NewTLSServer([]testnet.TestRequest{})

			config, authRepo = createAuthenticationRepository(apiServer, authServer)
			ccGateway.SetTokenRefresher(authRepo)

			ccGateway.SetTrustedCerts(apiServer.TLS.Certificates)
		})

		AfterEach(func() {
			apiServer.Close()
			authServer.Close()
		})

		It("returns the last response if the job completes before the timeout", func() {
			go func() {
				statusChannel <- "queued"
				statusChannel <- "finished"
			}()

			request, _ := ccGateway.NewRequest("GET", config.APIEndpoint()+"/v2/foo", config.AccessToken(), nil)
			_, apiErr := ccGateway.PerformPollingRequestForJSONResponse(config.APIEndpoint(), request, new(struct{}), 500*time.Millisecond)
			Expect(apiErr).NotTo(HaveOccurred())
		})

		It("returns an error with the right message when the job fails", func() {
			go func() {
				statusChannel <- "queued"
				statusChannel <- "failed"
			}()

			request, _ := ccGateway.NewRequest("GET", config.APIEndpoint()+"/v2/foo", config.AccessToken(), nil)
			_, apiErr := ccGateway.PerformPollingRequestForJSONResponse(config.APIEndpoint(), request, new(struct{}), 500*time.Millisecond)
			Expect(apiErr.Error()).To(ContainSubstring("he's dead, Jim"))
		})

		It("returns an error if jobs takes longer than the timeout", func() {
			go func() {
				statusChannel <- "queued"
				statusChannel <- "OHNOES"
			}()
			request, _ := ccGateway.NewRequest("GET", config.APIEndpoint()+"/v2/foo", config.AccessToken(), nil)
			_, apiErr := ccGateway.PerformPollingRequestForJSONResponse(config.APIEndpoint(), request, new(struct{}), 10*time.Millisecond)
			Expect(apiErr).To(HaveOccurred())
			Expect(apiErr).To(BeAssignableToTypeOf(errors.NewAsyncTimeoutError("http://some.url")))
		})
	})

	Describe("when uploading a file", func() {
		var (
			err          error
			request      *Request
			apiErr       error
			apiServer    *httptest.Server
			authServer   *httptest.Server
			fileToUpload *os.File
		)

		BeforeEach(func() {
			apiServer = httptest.NewTLSServer(refreshTokenAPIEndPoint(
				`{ "code": 1000, "description": "Auth token is invalid" }`,
				testnet.TestResponse{Status: http.StatusOK},
			))

			authServer = httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				fmt.Fprintln(
					writer,
					`{ "access_token": "new-access-token", "token_type": "bearer", "refresh_token": "new-refresh-token"}`)
			}))

			fileToUpload, err = ioutil.TempFile("", "test-gateway")
			strings.NewReader("expected body").WriteTo(fileToUpload)

			config, auth := createAuthenticationRepository(apiServer, authServer)
			ccGateway.SetTokenRefresher(auth)
			ccGateway.SetTrustedCerts(apiServer.TLS.Certificates)

			request, apiErr = ccGateway.NewRequestForFile("POST", config.APIEndpoint()+"/v2/foo", config.AccessToken(), fileToUpload)
		})

		AfterEach(func() {
			apiServer.Close()
			authServer.Close()
			fileToUpload.Close()
			os.Remove(fileToUpload.Name())
		})

		It("sets the content length to the size of the file", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(apiErr).NotTo(HaveOccurred())
			Expect(request.HTTPReq.ContentLength).To(Equal(int64(13)))
		})

		Describe("when the access token expires during the upload", func() {
			It("successfully re-sends the file on the second request", func() {
				_, apiErr = ccGateway.PerformRequest(request)
				Expect(apiErr).NotTo(HaveOccurred())
			})
		})
	})

	Describe("refreshing the auth token", func() {
		var authServer *httptest.Server

		BeforeEach(func() {
			authServer = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, `{
				 	"access_token": "new-access-token",
				 	"token_type": "bearer",
					"refresh_token": "new-refresh-token"
				}`)
			}))

			uaaGateway.SetTrustedCerts(authServer.TLS.Certificates)
		})

		AfterEach(func() {
			authServer.Close()
		})

		It("refreshes the token when UAA requests fail", func() {
			apiServer := httptest.NewTLSServer(refreshTokenAPIEndPoint(
				`{ "error": "invalid_token", "error_description": "Auth token is invalid" }`,
				testnet.TestResponse{Status: http.StatusOK},
			))
			defer apiServer.Close()
			ccGateway.SetTrustedCerts(apiServer.TLS.Certificates)

			config, auth := createAuthenticationRepository(apiServer, authServer)
			uaaGateway.SetTokenRefresher(auth)
			request, apiErr := uaaGateway.NewRequest("POST", config.APIEndpoint()+"/v2/foo", config.AccessToken(), strings.NewReader("expected body"))
			_, apiErr = uaaGateway.PerformRequest(request)

			Expect(apiErr).NotTo(HaveOccurred())
			Expect(config.AccessToken()).To(Equal("bearer new-access-token"))
			Expect(config.RefreshToken()).To(Equal("new-refresh-token"))
		})

		It("refreshes the token when CC requests fail", func() {
			apiServer := httptest.NewTLSServer(refreshTokenAPIEndPoint(
				`{ "code": 1000, "description": "Auth token is invalid" }`,
				testnet.TestResponse{Status: http.StatusOK}))
			defer apiServer.Close()
			ccGateway.SetTrustedCerts(apiServer.TLS.Certificates)

			config, auth := createAuthenticationRepository(apiServer, authServer)
			ccGateway.SetTokenRefresher(auth)
			request, apiErr := ccGateway.NewRequest("POST", config.APIEndpoint()+"/v2/foo", config.AccessToken(), strings.NewReader("expected body"))
			_, apiErr = ccGateway.PerformRequest(request)

			Expect(apiErr).NotTo(HaveOccurred())
			Expect(config.AccessToken()).To(Equal("bearer new-access-token"))
			Expect(config.RefreshToken()).To(Equal("new-refresh-token"))
		})

		It("returns a failure response when token refresh fails after a UAA request", func() {
			apiServer := httptest.NewTLSServer(refreshTokenAPIEndPoint(
				`{ "error": "invalid_token", "error_description": "Auth token is invalid" }`,
				testnet.TestResponse{Status: http.StatusBadRequest, Body: `{
					"error": "333", "error_description": "bad request"
				}`}))
			defer apiServer.Close()
			ccGateway.SetTrustedCerts(apiServer.TLS.Certificates)

			config, auth := createAuthenticationRepository(apiServer, authServer)
			uaaGateway.SetTokenRefresher(auth)
			request, apiErr := uaaGateway.NewRequest("POST", config.APIEndpoint()+"/v2/foo", config.AccessToken(), strings.NewReader("expected body"))
			_, apiErr = uaaGateway.PerformRequest(request)

			Expect(apiErr).To(HaveOccurred())
			Expect(apiErr.(errors.HTTPError).ErrorCode()).To(Equal("333"))
		})

		It("returns a failure response when token refresh fails after a CC request", func() {
			apiServer := httptest.NewTLSServer(refreshTokenAPIEndPoint(
				`{ "code": 1000, "description": "Auth token is invalid" }`,
				testnet.TestResponse{Status: http.StatusBadRequest, Body: `{
					"code": 333, "description": "bad request"
				}`}))
			defer apiServer.Close()
			ccGateway.SetTrustedCerts(apiServer.TLS.Certificates)

			config, auth := createAuthenticationRepository(apiServer, authServer)
			ccGateway.SetTokenRefresher(auth)
			request, apiErr := ccGateway.NewRequest("POST", config.APIEndpoint()+"/v2/foo", config.AccessToken(), strings.NewReader("expected body"))
			_, apiErr = ccGateway.PerformRequest(request)

			Expect(apiErr).To(HaveOccurred())
			Expect(apiErr.(errors.HTTPError).ErrorCode()).To(Equal("333"))
		})
	})

	Describe("SSL certificate validation errors", func() {
		var (
			request   *Request
			apiServer *httptest.Server
		)

		BeforeEach(func() {
			apiServer = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				fmt.Fprintln(w, `{}`)
			}))
			request, _ = ccGateway.NewRequest("POST", apiServer.URL+"/v2/foo", "the-access-token", nil)
		})

		AfterEach(func() {
			apiServer.Close()
		})

		Context("when SSL validation is enabled", func() {
			It("returns an invalid cert error if the server's CA is unknown (e.g. cert is self-signed)", func() {
				apiServer.TLS.Certificates = []tls.Certificate{testnet.MakeSelfSignedTLSCert()}

				_, apiErr := ccGateway.PerformRequest(request)
				certErr, ok := apiErr.(*errors.InvalidSSLCert)
				Expect(ok).To(BeTrue())
				Expect(certErr.URL).To(Equal(getHost(apiServer.URL)))
				Expect(certErr.Reason).To(Equal("unknown authority"))
			})

			It("returns an invalid cert error if the server's cert doesn't match its host", func() {
				apiServer.TLS.Certificates = []tls.Certificate{testnet.MakeTLSCertWithInvalidHost()}

				_, apiErr := ccGateway.PerformRequest(request)
				certErr, ok := apiErr.(*errors.InvalidSSLCert)
				Expect(ok).To(BeTrue())
				Expect(certErr.URL).To(Equal(getHost(apiServer.URL)))
				if runtime.GOOS != "windows" {
					Expect(certErr.Reason).To(Equal("not valid for the requested host"))
				}
			})

			It("returns an invalid cert error if the server's cert has expired", func() {
				apiServer.TLS.Certificates = []tls.Certificate{testnet.MakeExpiredTLSCert()}

				_, apiErr := ccGateway.PerformRequest(request)
				certErr, ok := apiErr.(*errors.InvalidSSLCert)
				Expect(ok).To(BeTrue())
				Expect(certErr.URL).To(Equal(getHost(apiServer.URL)))
				if runtime.GOOS != "windows" {
					Expect(certErr.Reason).To(Equal(""))
				}
			})
		})

		Context("when SSL validation is disabled", func() {
			BeforeEach(func() {
				apiServer.TLS.Certificates = []tls.Certificate{testnet.MakeExpiredTLSCert()}
				config.SetSSLDisabled(true)
			})

			It("succeeds", func() {
				_, apiErr := ccGateway.PerformRequest(request)
				Expect(apiErr).NotTo(HaveOccurred())
			})
		})

	})

	Describe("collecting warnings", func() {
		var (
			apiServer  *httptest.Server
			authServer *httptest.Server
		)

		BeforeEach(func() {
			apiServer = httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				switch request.URL.Path {
				case "/v2/happy":
					fmt.Fprintln(writer, `{ "metadata": { "url": "/v2/jobs/the-job-guid" } }`)
				case "/v2/warning1":
					writer.Header().Add("X-Cf-Warnings", url.QueryEscape("Something not too awful has happened"))
					fmt.Fprintln(writer, `{ "metadata": { "url": "/v2/jobs/the-job-guid" } }`)
				case "/v2/warning2":
					writer.Header().Add("X-Cf-Warnings", url.QueryEscape("Something a little awful"))
					writer.Header().Add("X-Cf-Warnings", url.QueryEscape("Don't worry, but be careful"))
					writer.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(writer, `{ "key": "value" }`)
				}
			}))

			authServer, _ = testnet.NewTLSServer([]testnet.TestRequest{})

			config, authRepo = createAuthenticationRepository(apiServer, authServer)
			ccGateway.SetTokenRefresher(authRepo)

			ccGateway.SetTrustedCerts(apiServer.TLS.Certificates)

			config, authRepo = createAuthenticationRepository(apiServer, authServer)
		})

		AfterEach(func() {
			apiServer.Close()
			authServer.Close()
		})

		It("saves all X-Cf-Warnings headers and exposes them", func() {
			request, _ := ccGateway.NewRequest("GET", config.APIEndpoint()+"/v2/happy", config.AccessToken(), nil)
			ccGateway.PerformRequest(request)
			request, _ = ccGateway.NewRequest("GET", config.APIEndpoint()+"/v2/warning1", config.AccessToken(), nil)
			ccGateway.PerformRequest(request)
			request, _ = ccGateway.NewRequest("GET", config.APIEndpoint()+"/v2/warning2", config.AccessToken(), nil)
			ccGateway.PerformRequest(request)

			Expect(ccGateway.Warnings()).To(Equal(
				[]string{"Something not too awful has happened", "Something a little awful", "Don't worry, but be careful"},
			))
		})

		It("defaults warnings to an empty slice", func() {
			Expect(ccGateway.Warnings()).ToNot(BeNil())
		})
	})
})

func getHost(urlString string) string {
	url, err := url.Parse(urlString)
	Expect(err).NotTo(HaveOccurred())
	return url.Host
}

func refreshTokenAPIEndPoint(unauthorizedBody string, secondReqResp testnet.TestResponse) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		var jsonResponse string

		bodyBytes, err := ioutil.ReadAll(request.Body)
		if err != nil || string(bodyBytes) != "expected body" {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		switch request.Header.Get("Authorization") {
		case "bearer initial-access-token":
			writer.WriteHeader(http.StatusUnauthorized)
			jsonResponse = unauthorizedBody
		case "bearer new-access-token":
			writer.WriteHeader(secondReqResp.Status)
			jsonResponse = secondReqResp.Body
		default:
			writer.WriteHeader(http.StatusInternalServerError)
		}

		fmt.Fprintln(writer, jsonResponse)
	}
}

func createAuthenticationRepository(apiServer *httptest.Server, authServer *httptest.Server) (coreconfig.ReadWriter, authentication.Repository) {
	config := testconfig.NewRepository()
	config.SetAuthenticationEndpoint(authServer.URL)
	config.SetAPIEndpoint(apiServer.URL)
	config.SetAccessToken("bearer initial-access-token")
	config.SetRefreshToken("initial-refresh-token")

	authGateway := NewUAAGateway(config, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), "")
	authGateway.SetTrustedCerts(authServer.TLS.Certificates)

	fakePrinter := new(tracefakes.FakePrinter)
	dumper := NewRequestDumper(fakePrinter)
	authenticator := authentication.NewUAARepository(authGateway, config, dumper)

	return config, authenticator
}

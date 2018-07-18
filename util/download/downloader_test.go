package download_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "code.cloudfoundry.org/cli/util/download"
	"code.cloudfoundry.org/cli/util/download/downloadfakes"
)

var _ = Describe("Downloader", func() {
	var (
		fakeHTTPClient *downloadfakes.FakeHTTPClient
		downloader     *Downloader
	)

	BeforeEach(func() {
		fakeHTTPClient = new(downloadfakes.FakeHTTPClient)
		downloader = &Downloader{
			HTTPClient: fakeHTTPClient,
		}
	})

	Describe("Download", func() {
		var (
			url string

			file       string
			executeErr error
		)

		BeforeEach(func() {
			url = "https://some.url"
		})

		AfterEach(func() {
			Expect(os.RemoveAll(file)).ToNot(HaveOccurred())
		})

		JustBeforeEach(func() {
			file, executeErr = downloader.Download(url)
		})

		Context("successful download", func() {
			var responseBody string

			BeforeEach(func() {
				responseBody = "some response body"
				response := &http.Response{
					Body:          ioutil.NopCloser(strings.NewReader(responseBody)),
					ContentLength: int64(len(responseBody)),
					StatusCode:    http.StatusOK,
				}
				fakeHTTPClient.GetReturns(response, nil)
			})

			It("returns correct path to the downloaded file", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				raw, err := ioutil.ReadFile(file)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(raw)).To(Equal(responseBody))
			})

			It("downloads the file from the provided URL", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeHTTPClient.GetCallCount()).To(Equal(1))
				Expect(fakeHTTPClient.GetArgsForCall(0)).To(Equal(url))
			})
		})

		Context("when the client returns an error", func() {
			BeforeEach(func() {
				fakeHTTPClient.GetReturns(nil, errors.New("stop all the downloading"))
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("stop all the downloading"))
			})
		})

		Context("when HTTP request returns 4xx or 5xx error", func() {
			var responseBody string

			BeforeEach(func() {
				responseBody = "not found"
				response := &http.Response{
					Body:       ioutil.NopCloser(strings.NewReader(responseBody)),
					StatusCode: http.StatusNotFound,
					Status:     "404 Not Found",
				}
				fakeHTTPClient.GetReturns(response, nil)
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(RawHTTPStatusError{Status: "404 Not Found", RawResponse: []byte(responseBody)}))
			})
		})
	})
})

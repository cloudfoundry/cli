package plugin_test

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	. "code.cloudfoundry.org/cli/api/plugin"
	"code.cloudfoundry.org/cli/api/plugin/pluginerror"
	"code.cloudfoundry.org/cli/api/plugin/pluginfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("DownloadPlugin", func() {
	var (
		client   *Client
		tempPath string
	)

	BeforeEach(func() {
		client = NewTestClient()

		tempFile, err := ioutil.TempFile("", "")
		Expect(err).NotTo(HaveOccurred())
		tempPath = tempFile.Name()

		err = tempFile.Close()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := os.Remove(tempPath)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when there are no errors", func() {
		var (
			data []byte
		)

		BeforeEach(func() {
			data = []byte("some test data")
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/"),
					RespondWith(http.StatusOK, data),
				),
			)
		})

		It("downloads the plugin, and writes the plugin file to the specified path", func() {
			fakeProxyReader := new(pluginfakes.FakeProxyReader)

			fakeProxyReader.WrapStub = func(reader io.Reader) io.ReadCloser {
				return ioutil.NopCloser(reader)
			}
			err := client.DownloadPlugin(server.URL(), tempPath, fakeProxyReader)
			Expect(err).ToNot(HaveOccurred())

			fileData, err := ioutil.ReadFile(tempPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(fileData).To(Equal(data))

			Expect(fakeProxyReader.WrapCallCount()).To(Equal(1))
		})
	})

	Context("when the URL is invalid", func() {
		It("returns an URL error", func() {
			err := client.DownloadPlugin("://", tempPath, nil)
			_, isURLError := err.(*url.Error)
			Expect(isURLError).To(BeTrue())
		})
	})

	Context("when downloading the plugin errors", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/"),
					RespondWith(http.StatusTeapot, nil),
				),
			)
		})

		It("returns a RawHTTPStatusError", func() {
			err := client.DownloadPlugin(server.URL(), tempPath, nil)
			Expect(err).To(MatchError(pluginerror.RawHTTPStatusError{Status: "418 I'm a teapot", RawResponse: []byte("")}))
		})
	})

	Context("when the path is not writeable", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/"),
					RespondWith(http.StatusOK, nil),
				),
			)
		})

		It("returns some error", func() {
			err := client.DownloadPlugin(server.URL(), "/a/path/that/does/not/exist", nil)
			_, isPathError := err.(*os.PathError)
			Expect(isPathError).To(BeTrue())
		})
	})
})

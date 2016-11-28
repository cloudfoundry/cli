package downloader_test

import (
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"code.cloudfoundry.org/cli/util/downloader"

	"github.com/onsi/gomega/ghttp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Downloader", func() {
	var (
		d       downloader.Downloader
		tempDir string
	)

	BeforeEach(func() {
		var err error
		tempDir, err = ioutil.TempDir("", "file-download-test")
		Expect(err).NotTo(HaveOccurred())
		d = downloader.NewDownloader(tempDir)
	})

	AfterEach(func() {
		os.RemoveAll(tempDir)
	})

	Describe("DownloadFile", func() {
		var server *ghttp.Server

		BeforeEach(func() {
			server = ghttp.NewServer()
		})

		AfterEach(func() {
			server.Close()

			err := d.RemoveFile()
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the server responds with the file", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/abc.zip"),
						ghttp.RespondWith(http.StatusOK, "abc123"),
					),
				)
			})

			It("saves file with name found in URL in provided dir", func() {
				_, _, err := d.DownloadFile(server.URL() + "/abc.zip")
				Expect(err).NotTo(HaveOccurred())

				_, err = os.Stat(path.Join(tempDir, "abc.zip"))
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns the number of bytes written to the file", func() {
				n, _, err := d.DownloadFile(server.URL() + "/abc.zip")
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(int64(len("abc123"))))
			})

			It("returns the name of the file that was downloaded", func() {
				_, name, err := d.DownloadFile(server.URL() + "/abc.zip")
				Expect(err).NotTo(HaveOccurred())
				Expect(name).To(Equal("abc.zip"))
			})
		})

		Context("when the server responds with the filename in the header", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/abc.zip"),
						ghttp.RespondWith(http.StatusOK, "abc123", http.Header{
							"Content-Disposition": []string{"attachment; filename=header.zip"},
						}),
					),
				)
			})

			It("downloads the file named in the header to the provided dir, and trims spaces appropriately", func() {
				_, _, err := d.DownloadFile(server.URL() + "/abc.zip")
				Expect(err).NotTo(HaveOccurred())

				_, err = os.Stat(path.Join(tempDir, "header.zip"))
				Expect(err).NotTo(HaveOccurred())

				_, err = os.Stat(path.Join(tempDir, "abc.zip"))
				Expect(err).To(HaveOccurred())
			})

			It("returns the number of bytes written to the file", func() {
				n, _, err := d.DownloadFile(server.URL() + "/abc.zip")
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(int64(len("abc123"))))
			})

			It("returns the name of the file that was downloaded", func() {
				_, name, err := d.DownloadFile(server.URL() + "/abc.zip")
				Expect(err).NotTo(HaveOccurred())
				Expect(name).To(Equal("header.zip"))
			})
		})

		Context("when the server returns a redirect to a file", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/abc.zip"),
						ghttp.RespondWith(http.StatusFound, "", http.Header{
							"Location": []string{server.URL() + "/redirect.zip"},
						}),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/redirect.zip"),
						ghttp.RespondWith(http.StatusOK, "abc123"),
					),
				)
			})

			It("downloads the file from the redirect to the provided dir", func() {
				_, _, err := d.DownloadFile(server.URL() + "/abc.zip")
				Expect(err).NotTo(HaveOccurred())

				_, err = os.Stat(path.Join(tempDir, "redirect.zip"))
				Expect(err).NotTo(HaveOccurred())

				_, err = os.Stat(path.Join(tempDir, "abc.zip"))
				Expect(err).To(HaveOccurred())
			})

			It("returns the number of bytes written to the file", func() {
				n, _, err := d.DownloadFile(server.URL() + "/abc.zip")
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(int64(len("abc123"))))
			})

			It("returns the name of the file that was downloaded", func() {
				_, name, err := d.DownloadFile(server.URL() + "/abc.zip")
				Expect(err).NotTo(HaveOccurred())
				Expect(name).To(Equal("redirect.zip"))
			})
		})

		Context("when the URL is invalid", func() {
			It("returns an error", func() {
				_, _, err := d.DownloadFile("http://going.nowwhere/abc.zip")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("RemoveFile", func() {
		var server *ghttp.Server

		BeforeEach(func() {
			server = ghttp.NewServer()
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/abc.zip"),
					ghttp.RespondWith(http.StatusOK, "abc123"),
				),
			)
		})

		AfterEach(func() {
			server.Close()
		})

		Context("when a file has been downloaded", func() {
			BeforeEach(func() {
				_, _, err := d.DownloadFile(server.URL() + "/abc.zip")
				Expect(err).NotTo(HaveOccurred())
			})

			It("removes the downloaded file", func() {
				_, err := os.Stat(path.Join(tempDir, "abc.zip"))
				Expect(err).NotTo(HaveOccurred())

				err = d.RemoveFile()
				Expect(err).NotTo(HaveOccurred())

				_, err = os.Stat(path.Join(tempDir, "abc.zip"))
				Expect(err).To(HaveOccurred())
			})
		})

		It("does not return an error when a file has not been downloaded", func() {
			err := d.RemoveFile()
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

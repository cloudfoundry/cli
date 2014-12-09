package fileutils_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	. "github.com/cloudfoundry/cli/fileutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Downloader", func() {
	var (
		ts         *httptest.Server
		downloader Downloader
	)

	Describe("DownloadFile", func() {
		Context("URL contains filename", func() {
			BeforeEach(func() {
				ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprintln(w, "abc123")
				}))
				downloader = NewDownloader("./")
			})

			AfterEach(func() {
				ts.Close()

				err := downloader.RemoveFile()
				Ω(err).ToNot(HaveOccurred())
			})

			It("saves file with name found in URL in provided dir", func() {
				_, name, err := downloader.DownloadFile(ts.URL + "/abc.zip")
				Ω(err).ToNot(HaveOccurred())
				Ω(name).To(Equal("abc.zip"))

				_, err = os.Stat("./abc.zip")
				Ω(err).ToNot(HaveOccurred())
			})

			It("returns the number of bytes written to the file", func() {
				n, _, err := downloader.DownloadFile(ts.URL + "/abc.zip")
				Ω(err).ToNot(HaveOccurred())
				Ω(n).To(Equal(int64(len("abc123") + 1)))
			})
		})

		Context("header contains filename", func() {
			BeforeEach(func() {
				ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Disposition", "attachment;filename=header.zip")
					fmt.Fprintln(w, "abc123")
				}))

				downloader = NewDownloader("./")
			})

			AfterEach(func() {
				ts.Close()

				err := downloader.RemoveFile()
				Ω(err).ToNot(HaveOccurred())
			})

			It("saves file with name found in header, ignore filename in url", func() {
				_, name, err := downloader.DownloadFile(ts.URL + "/abc.zip")
				Ω(err).ToNot(HaveOccurred())
				Ω(name).To(Equal("header.zip"))

				_, err = os.Stat("./abc.zip")
				Ω(err).To(HaveOccurred())

				_, err = os.Stat("./header.zip")
				Ω(err).ToNot(HaveOccurred())
			})

		})

		Context("When URL redirects", func() {
			BeforeEach(func() {
				ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

					if !strings.Contains(r.URL.String(), "redirect.zip") {
						http.Redirect(w, r, ts.URL+"/redirect.zip", http.StatusMovedPermanently)
					} else {
						fmt.Fprintln(w, "abc123")
					}
				}))

				downloader = NewDownloader("./")
			})

			AfterEach(func() {
				ts.Close()

				err := downloader.RemoveFile()
				Ω(err).ToNot(HaveOccurred())
			})

			It("follows redirects and download file", func() {
				downloader = NewDownloader("./")

				_, _, err := downloader.DownloadFile(ts.URL)
				Ω(err).ToNot(HaveOccurred())

				_, err = os.Stat("./redirect.zip")
				Ω(err).ToNot(HaveOccurred())
			})

		})

		Context("When URL is invalid", func() {
			It("returns an error message", func() {
				downloader = NewDownloader("./")

				_, name, err := downloader.DownloadFile("http://going.nowwhere/abc.zip")

				Ω(err).To(HaveOccurred())
				Ω(name).To(Equal(""))
			})
		})

	})

	Describe("RemoveFile", func() {
		BeforeEach(func() {
			ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, "abc123")
			}))
		})

		AfterEach(func() {
			ts.Close()
		})

		It("removes the downloaded file", func() {
			_, _, err := downloader.DownloadFile(ts.URL + "/abc.zip")
			Ω(err).ToNot(HaveOccurred())

			_, err = os.Stat("./abc.zip")
			Ω(err).ToNot(HaveOccurred())

			err = downloader.RemoveFile()
			Ω(err).ToNot(HaveOccurred())

			_, err = os.Stat("./abc.zip")
			Ω(err).To(HaveOccurred())
		})
	})

})

package ui_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "code.cloudfoundry.org/cli/util/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("Request Logger File Writer", func() {
	var (
		testUI   *UI
		display  *RequestLoggerFileWriter
		tmpdir   string
		logFile1 string
		logFile2 string
	)

	BeforeEach(func() {
		testUI = NewTestUI(NewBuffer(), NewBuffer(), NewBuffer())
	})

	Describe("Valid file paths", func() {
		BeforeEach(func() {
			var err error
			tmpdir, err = ioutil.TempDir("", "request_logger")
			Expect(err).ToNot(HaveOccurred())

			logFile1 = filepath.Join(tmpdir, "tmp_sub_dir", "tmpfile1")
			logFile2 = filepath.Join(tmpdir, "tmp", "sub", "dir", ".", "tmpfile2")
			display = testUI.RequestLoggerFileWriter([]string{logFile1, logFile2})
			err = display.Start()
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			Expect(os.RemoveAll(tmpdir)).NotTo(HaveOccurred())
		})

		Describe("DisplayBody", func() {
			It("writes the redacted value", func() {
				err := display.DisplayBody([]byte("this is a body"))
				Expect(err).ToNot(HaveOccurred())

				err = display.Stop()
				Expect(err).ToNot(HaveOccurred())

				contents, err := ioutil.ReadFile(logFile1)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(contents)).To(Equal(RedactedValue + "\n"))

				contents, err = ioutil.ReadFile(logFile2)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(contents)).To(Equal(RedactedValue + "\n"))
			})
		})

		Describe("DisplayDump", func() {
			It("creates the intermediate dirs and writes the dump to file", func() {
				err := display.DisplayDump("this is a dump of stuff")
				Expect(err).ToNot(HaveOccurred())

				err = display.Stop()
				Expect(err).ToNot(HaveOccurred())

				contents, err := ioutil.ReadFile(logFile1)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(contents)).To(Equal("this is a dump of stuff\n"))

				contents, err = ioutil.ReadFile(logFile2)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(contents)).To(Equal("this is a dump of stuff\n"))
			})
		})

		Describe("DisplayHeader", func() {
			It("writes the header key and value", func() {
				err := display.DisplayHeader("Header", "Value")
				Expect(err).ToNot(HaveOccurred())

				err = display.Stop()
				Expect(err).ToNot(HaveOccurred())

				contents, err := ioutil.ReadFile(logFile1)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(contents)).To(Equal("Header: Value\n\n"))

				contents, err = ioutil.ReadFile(logFile2)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(contents)).To(Equal("Header: Value\n\n"))
			})
		})

		Describe("DisplayHost", func() {
			It("writes the host", func() {
				err := display.DisplayHost("banana")
				Expect(err).ToNot(HaveOccurred())

				err = display.Stop()
				Expect(err).ToNot(HaveOccurred())

				contents, err := ioutil.ReadFile(logFile1)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(contents)).To(Equal("Host: banana\n\n"))

				contents, err = ioutil.ReadFile(logFile2)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(contents)).To(Equal("Host: banana\n\n"))
			})
		})

		Describe("DisplayJSONBody", func() {
			Context("when provided well formed JSON", func() {
				It("writes a formated output", func() {
					raw := `{"a":"b", "c":"d", "don't escape HTML":"<&>"}`
					formatted := `{
  "a": "b",
  "c": "d",
  "don't escape HTML": "<&>"
}

` // Additonal spaces required
					err := display.DisplayJSONBody([]byte(raw))
					Expect(err).ToNot(HaveOccurred())

					err = display.Stop()
					Expect(err).ToNot(HaveOccurred())

					contents, err := ioutil.ReadFile(logFile1)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(contents)).To(Equal(formatted))

					contents, err = ioutil.ReadFile(logFile2)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(contents)).To(Equal(formatted))
				})
			})

			Context("when the body is empty", func() {
				It("does not write the body", func() {
					err := display.DisplayJSONBody(nil)
					Expect(err).ToNot(HaveOccurred())

					err = display.Stop()
					Expect(err).ToNot(HaveOccurred())

					contents, err := ioutil.ReadFile(logFile1)
					Expect(err).ToNot(HaveOccurred())
					// display.Stop() writes "\n" to the file
					Expect(string(contents)).To(Equal("\n"))

					contents, err = ioutil.ReadFile(logFile2)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(contents)).To(Equal("\n"))
				})
			})
		})

		Describe("DisplayRequestHeader", func() {
			It("writes the method, uri and http protocal", func() {
				err := display.DisplayRequestHeader("GET", "/v2/spaces/guid/summary", "HTTP/1.1")
				Expect(err).ToNot(HaveOccurred())

				err = display.Stop()
				Expect(err).ToNot(HaveOccurred())

				contents, err := ioutil.ReadFile(logFile1)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(contents)).To(Equal("GET /v2/spaces/guid/summary HTTP/1.1\n\n"))

				contents, err = ioutil.ReadFile(logFile2)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(contents)).To(Equal("GET /v2/spaces/guid/summary HTTP/1.1\n\n"))
			})
		})

		Describe("DisplayResponseHeader", func() {
			It("writes the method, uri and http protocal", func() {
				err := display.DisplayResponseHeader("HTTP/1.1", "200 OK")
				Expect(err).ToNot(HaveOccurred())

				err = display.Stop()
				Expect(err).ToNot(HaveOccurred())

				contents, err := ioutil.ReadFile(logFile1)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(contents)).To(Equal("HTTP/1.1 200 OK\n\n"))

				contents, err = ioutil.ReadFile(logFile2)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(contents)).To(Equal("HTTP/1.1 200 OK\n\n"))
			})
		})

		Describe("DisplayType", func() {
			It("writes the passed type and time in localized ISO 8601", func() {
				passedTime := time.Now()
				err := display.DisplayType("banana", passedTime)
				Expect(err).ToNot(HaveOccurred())

				err = display.Stop()
				Expect(err).ToNot(HaveOccurred())

				contents, err := ioutil.ReadFile(logFile1)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(contents)).To(Equal(fmt.Sprintf("banana: [%s]\n\n", passedTime.Format(time.RFC3339))))

				contents, err = ioutil.ReadFile(logFile2)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(contents)).To(Equal(fmt.Sprintf("banana: [%s]\n\n", passedTime.Format(time.RFC3339))))
			})
		})

		Describe("HandleInternalError", func() {
			It("sends error to standard error", func() {
				err := errors.New("foobar")
				display.HandleInternalError(err)
				Expect(testUI.Err).To(Say("foobar"))
				Expect(display.Stop()).NotTo(HaveOccurred())
			})
		})

		Describe("Start and Stop", func() {
			It("locks and then unlocks the mutex properly", func() { // and creates the intermediate dirs
				c := make(chan bool)
				go func() {
					Expect(display.Start()).ToNot(HaveOccurred())
					c <- true
				}()
				Consistently(c).ShouldNot(Receive())
				Expect(display.Stop()).NotTo(HaveOccurred())
				Eventually(c).Should(Receive())
			})
		})
	})

	Describe("when the log file path is invalid", func() {
		var pathName string

		BeforeEach(func() {
			tmpdir, err := ioutil.TempDir("", "request_logger")
			Expect(err).ToNot(HaveOccurred())

			pathName = filepath.Join(tmpdir, "foo")
		})

		AfterEach(func() {
			Expect(display.Stop()).NotTo(HaveOccurred())
			Expect(os.RemoveAll(tmpdir)).NotTo(HaveOccurred())
		})

		It("returns the os error when we unsuccessfully try to write to a file", func() {
			Expect(os.Mkdir(pathName, os.ModeDir|os.ModePerm)).NotTo(HaveOccurred())
			display = testUI.RequestLoggerFileWriter([]string{pathName})
			err := display.Start()

			Expect(err).To(MatchError(fmt.Sprintf("open %s: is a directory", pathName)))
		})

		It("returns the os error when the parent directory for the log file is in the root directory", func() {
			file, err := os.OpenFile(pathName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
			Expect(err).ToNot(HaveOccurred())
			_, err = file.WriteString("hello world")
			Expect(err).ToNot(HaveOccurred())
			err = file.Close()
			Expect(err).ToNot(HaveOccurred())

			display = testUI.RequestLoggerFileWriter([]string{filepath.Join(pathName, "bar")})
			err = display.Start()

			pathName = strings.Replace(pathName, `\`, `\\`, -1)
			Expect(err).To(MatchError(MatchRegexp(fmt.Sprintf("mkdir %s", pathName))))
		})
	})
})

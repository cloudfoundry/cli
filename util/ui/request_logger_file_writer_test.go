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
			Expect(display.Start()).ToNot(HaveOccurred())
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

			It("redacts auth tokens", func() {
				dump := `GET /apps/ce03a2e2-95c0-4f3b-abb9-32718d408c8b/stream HTTP/1.1
Host: wss://doppler.bosh-lite.com:443
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Version: 13
Sec-WebSocket-Key: [HIDDEN]
Authorization: bearer eyJhbGciOiJSUzI1NiIsImtpZCI6ImtleS0xIiwidHlwIjoiSldUIn0.eyJqdGkiOiI3YzRmYWEyZjI5MmQ0MTQ5ODM5NGE3OTU0Y2E3ZWNlMCIsInN1YiI6IjIyMjNiM2IzLTE3ZTktNDJkNi1iNzQzLThjZjcyZWIwOWRlNSIsInNjb3BlIjpbInJvdXRpbmcucm91dGVyX2dyb3Vwcy5yZWFkIiwiY2xvdWRfY29udHJvbGxlci5yZWFkIiwicGFzc3dvcmQud3JpdGUiLCJjbG91ZF9jb250cm9sbGVyLndyaXRlIiwib3BlbmlkIiwicm91dGluZy5yb3V0ZXJfZ3JvdXBzLndyaXRlIiwiZG9wcGxlci5maXJlaG9zZSIsInNjaW0ud3JpdGUiLCJzY2ltLnJlYWQiLCJjbG91ZF9jb250cm9sbGVyLmFkbWluIiwidWFhLnVzZXIiXSwiY2xpZW50X2lkIjoiY2YiLCJjaWQiOiJjZiIsImF6cCI6ImNmIiwiZ3JhbnRfdHlwZSI6InBhc3N3b3JkIiwidXNlcl9pZCI6IjIyMjNiM2IzLTE3ZTktNDJkNi1iNzQzLThjZjcyZWIwOWRlNSIsIm9yaWdpbiI6InVhYSIsInVzZXJfbmFtZSI6ImFkbWluIiwiZW1haWwiOiJhZG1pbiIsInJldl9zaWciOiI4NDBiMDBhMyIsImlhdCI6MTQ5NjQyNTU5NiwiZXhwIjoxNDk2NDI2MTk2LCJpc3MiOiJodHRwczovL3VhYS5ib3NoLWxpdGUuY29tL29hdXRoL3Rva2VuIiwiemlkIjoidWFhIiwiYXVkIjpbInNjaW0iLCJjbG91ZF9jb250cm9sbGVyIiwicGFzc3dvcmQiLCJjZiIsInVhYSIsIm9wZW5pZCIsImRvcHBsZXIiLCJyb3V0aW5nLnJvdXRlcl9ncm91cHMiXX0.TFDmHviKcs-eeNoz79dVwOl-k_dHTdqHkyztont2qnBDchNSpWvR5Yba54MMG8uTUHM72YbCopxdyaLY-g8s5wJFGLaBocrDgqswUh3mQRvynQG6_zne1h_0oHXnm0U-ZPnTyV8qjtHUoLvks4GOuktXc6ZE3NriWODpKIU5WdMgEbvyhuTnUEn88rQnmGJbKvHOIilulb6avSkZfTEq1o8w4VLCeRDlVLNh5JzCUtGzLfImNb31ks_Wv6HuI8kFjQZ5PQiTYjlhkuDQOcNSaAyWxQ_7425hiA7x8omBgEr-uST7GsxLvgoHqQaDH0JSTgMmO_GaN_QD52JVuru9og
Origin: wss://doppler.bosh-lite.com:443`
				err := display.DisplayDump(dump)
				Expect(err).ToNot(HaveOccurred())

				err = display.Stop()
				Expect(err).ToNot(HaveOccurred())

				raw, err := ioutil.ReadFile(logFile1)
				Expect(err).ToNot(HaveOccurred())
				contents := string(raw)

				Expect(contents).To(MatchRegexp("Connection: Upgrade"))
				Expect(contents).To(MatchRegexp("Authorization: \\[PRIVATE DATA HIDDEN\\]"))
				Expect(contents).To(MatchRegexp("Origin: wss://doppler.bosh-lite.com:443"))

				raw, err = ioutil.ReadFile(logFile2)
				Expect(err).ToNot(HaveOccurred())
				contents = string(raw)

				Expect(contents).To(MatchRegexp("Connection: Upgrade"))
				Expect(contents).To(MatchRegexp("Authorization: \\[PRIVATE DATA HIDDEN\\]"))
				Expect(contents).To(MatchRegexp("Origin: wss://doppler.bosh-lite.com:443"))
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
				It("writes a formatted output", func() {
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
					Expect(string(contents)).To(Equal("\n"))

					contents, err = ioutil.ReadFile(logFile2)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(contents)).To(Equal("\n"))
				})
			})

			Context("when provided malformed JSON", func() {
				It("displays the raw body", func() {
					raw := `[{"data":1, "banana": 2}]`
					err := display.DisplayJSONBody([]byte(raw))
					Expect(err).ToNot(HaveOccurred())

					err = display.Stop()
					Expect(err).ToNot(HaveOccurred())

					contents, err := ioutil.ReadFile(logFile1)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(contents)).To(Equal(raw + "\n\n"))

					contents, err = ioutil.ReadFile(logFile2)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(contents)).To(Equal(raw + "\n\n"))
				})
			})
		})

		Describe("DisplayMessage", func() {
			It("writes the message", func() {
				msg := "i am a message!!!!"
				err := display.DisplayMessage(msg)
				Expect(err).ToNot(HaveOccurred())

				err = display.Stop()
				Expect(err).ToNot(HaveOccurred())

				contents, err := ioutil.ReadFile(logFile1)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(contents)).To(ContainSubstring(msg))

				contents, err = ioutil.ReadFile(logFile2)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(contents)).To(ContainSubstring(msg))
			})
		})

		Describe("DisplayRequestHeader", func() {
			It("writes the method, uri and http protocol", func() {
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
			It("writes the method, uri and http protocol", func() {
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
			BeforeEach(func() {
				// Cleanup old display output directory
				Expect(display.Stop()).NotTo(HaveOccurred())
				logFile1 = filepath.Join(tmpdir, "tmp_sub_dir", "tmpfile3")
				logFile2 = filepath.Join(tmpdir, "tmp", "sub", "dir", ".", "tmpfile4")
				display = testUI.RequestLoggerFileWriter([]string{logFile1, logFile2})
			})

			It("locks and then unlocks the mutex properly", func() { // and creates the intermediate dirs
				c := make(chan bool)
				go func() {
					Expect(display.Start()).ToNot(HaveOccurred())
					c <- true
				}()
				Eventually(c).Should(Receive())
				Expect(display.Stop()).NotTo(HaveOccurred())
			})
		})
	})

	Describe("when the log file path is invalid", func() {
		var pathName string

		BeforeEach(func() {
			var err error
			tmpdir, err = ioutil.TempDir("", "request_logger")
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

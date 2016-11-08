package common_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	. "code.cloudfoundry.org/cli/commands/v2/common"
	"code.cloudfoundry.org/cli/utils/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("Request Logger File Writer", func() {
	var (
		fakeUI  *ui.UI
		display *RequestLoggerFileWriter
		tmpdir  string
		tmpfn   string
	)

	BeforeEach(func() {
		fakeUI = ui.NewTestUI(NewBuffer(), NewBuffer(), NewBuffer())
		var err error
		tmpdir, err = ioutil.TempDir("", "request_logger")
		Expect(err).ToNot(HaveOccurred())

		tmpfn = filepath.Join(tmpdir, "tmpfile")
		display = NewRequestLoggerFileWriter(fakeUI, tmpfn)
		err = display.Start()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		os.RemoveAll(tmpdir)
	})

	Describe("DisplayBody", func() {
		Context("when provided well formed JSON", func() {
			It("writes a formated output", func() {
				raw := `{"a":"b", "c":"d"}`
				formatted := `{
  "a": "b",
  "c": "d"
}

` // Additonal spaces required
				err := display.DisplayBody([]byte(raw))
				Expect(err).ToNot(HaveOccurred())

				err = display.Stop()
				Expect(err).ToNot(HaveOccurred())

				contents, err := ioutil.ReadFile(tmpfn)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(contents)).To(Equal(formatted))
			})
		})
	})

	Describe("DisplayHeader", func() {
		It("writes the header key and value", func() {
			err := display.DisplayHeader("Header", "Value")
			Expect(err).ToNot(HaveOccurred())

			err = display.Stop()
			Expect(err).ToNot(HaveOccurred())

			contents, err := ioutil.ReadFile(tmpfn)
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

			contents, err := ioutil.ReadFile(tmpfn)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(contents)).To(Equal("Host: banana\n\n"))
		})
	})

	Describe("DisplayRequestHeader", func() {
		It("writes the method, uri and http protocal", func() {
			err := display.DisplayRequestHeader("GET", "/v2/spaces/guid/summary", "HTTP/1.1")
			Expect(err).ToNot(HaveOccurred())

			err = display.Stop()
			Expect(err).ToNot(HaveOccurred())

			contents, err := ioutil.ReadFile(tmpfn)
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

			contents, err := ioutil.ReadFile(tmpfn)
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

			contents, err := ioutil.ReadFile(tmpfn)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(contents)).To(Equal(fmt.Sprintf("banana: [%s]\n\n", passedTime.Format(time.RFC3339))))
		})
	})

	Describe("HandleInternalError", func() {
		It("sends error to standard error", func() {
			err := errors.New("foobar")
			display.HandleInternalError(err)
			Expect(fakeUI.Err).To(Say("foobar"))
		})
	})
})

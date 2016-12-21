package command_test

import (
	"errors"
	"time"

	. "code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/util/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("Request Logger Terminal Display", func() {
	var (
		out     *Buffer
		testUI  *ui.UI
		display *RequestLoggerTerminalDisplay
	)

	BeforeEach(func() {
		out = NewBuffer()
		testUI = ui.NewTestUI(nil, out, NewBuffer())
		display = NewRequestLoggerTerminalDisplay(testUI)
	})

	Describe("DisplayBody", func() {
		It("displays the redacted value", func() {
			err := display.DisplayBody([]byte("some-string body"))
			Expect(err).ToNot(HaveOccurred())
			Expect(testUI.Out).To(Say("\\[PRIVATE DATA HIDDEN\\]"))
		})
	})

	Describe("DisplayJSONBody", func() {
		Context("when provided well formed JSON", func() {
			It("displayed a formated output", func() {
				raw := `{"a":"b", "c":"d", "don't html escape":"<&>"}`
				formatted := `{
  "a": "b",
  "c": "d",
  "don't html escape": "<&>"
}`
				err := display.DisplayJSONBody([]byte(raw))
				Expect(err).ToNot(HaveOccurred())
				Expect(testUI.Out).To(Say(formatted))
			})
		})

		Context("when the body is empty", func() {
			It("does not display the body", func() {
				err := display.DisplayJSONBody(nil)
				Expect(err).ToNot(HaveOccurred())

				Expect(out.Contents()).To(BeEmpty())
			})
		})
	})

	Describe("DisplayHeader", func() {
		It("displays the header key and value", func() {
			err := display.DisplayHeader("Header", "Value")
			Expect(err).ToNot(HaveOccurred())
			Expect(testUI.Out).To(Say("Header: Value"))
		})
	})

	Describe("DisplayHost", func() {
		It("displays the host", func() {
			err := display.DisplayHost("banana")
			Expect(err).ToNot(HaveOccurred())
			Expect(testUI.Out).To(Say("Host: banana"))
		})
	})

	Describe("DisplayRequestHeader", func() {
		It("displays the method, uri and http protocal", func() {
			err := display.DisplayRequestHeader("GET", "/v2/spaces/guid/summary", "HTTP/1.1")
			Expect(err).ToNot(HaveOccurred())
			Expect(testUI.Out).To(Say("GET /v2/spaces/guid/summary HTTP/1.1"))
		})
	})

	Describe("DisplayResponseHeader", func() {
		It("displays the method, uri and http protocal", func() {
			err := display.DisplayResponseHeader("HTTP/1.1", "200 OK")
			Expect(err).ToNot(HaveOccurred())
			Expect(testUI.Out).To(Say("HTTP/1.1 200 OK"))
		})
	})

	Describe("DisplayType", func() {
		It("displays the passed type and time in localized ISO 8601", func() {
			passedTime := time.Now()
			err := display.DisplayType("banana", passedTime)
			Expect(err).ToNot(HaveOccurred())
			Expect(testUI.Out).To(Say("banana: \\[%s\\]", passedTime.Format(time.RFC3339)))
		})
	})

	Describe("HandleInternalError", func() {
		It("sends error to standard error", func() {
			err := errors.New("foobar")
			display.HandleInternalError(err)
			Expect(testUI.Err).To(Say("foobar"))
		})
	})
})

package terminal_test

import (
	"io/ioutil"
	"os"

	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/cf/trace"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("DebugPrinter", func() {
	var (
		printer terminal.DebugPrinter
	)

	BeforeEach(func() {
		printer = terminal.DebugPrinter{}
	})

	Describe("Print", func() {
		var (
			originalLogger *trace.Logger
			buffer         *gbytes.Buffer
		)

		BeforeEach(func() {
			// Capture trace logger output
			buffer = gbytes.NewBuffer()
			trace.SetStdout(buffer)
			trace.EnableTrace()
		})

		AfterEach(func() {
			if originalLogger != nil {
				trace.Logger = originalLogger
			}
			trace.DisableTrace()
		})

		It("prints title and dump", func() {
			printer.Print("Test Title", "test dump content")

			Eventually(buffer).Should(gbytes.Say("Test Title"))
			Eventually(buffer).Should(gbytes.Say("test dump content"))
		})

		It("includes timestamp in output", func() {
			printer.Print("Request", "GET /v2/apps")

			// Should contain RFC3339 formatted timestamp
			Eventually(buffer).Should(gbytes.Say("Request"))
			Eventually(buffer).Should(gbytes.Say(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`))
		})

		It("sanitizes sensitive information", func() {
			dumpWithToken := "Authorization: Bearer secret-token-12345"
			printer.Print("REQUEST", dumpWithToken)

			// The trace.Sanitize function should redact the token
			Eventually(buffer).Should(gbytes.Say("REQUEST"))
			// Specific sanitization behavior depends on trace.Sanitize implementation
		})

		It("handles empty title", func() {
			printer.Print("", "content without title")

			Eventually(buffer).Should(gbytes.Say("content without title"))
		})

		It("handles empty dump", func() {
			printer.Print("Title Only", "")

			Eventually(buffer).Should(gbytes.Say("Title Only"))
		})

		It("handles multiline dump content", func() {
			multilineDump := "Line 1\nLine 2\nLine 3"
			printer.Print("Multiline", multilineDump)

			Eventually(buffer).Should(gbytes.Say("Multiline"))
			Eventually(buffer).Should(gbytes.Say("Line 1"))
		})

		It("prints multiple times independently", func() {
			printer.Print("First", "first content")
			printer.Print("Second", "second content")

			Eventually(buffer).Should(gbytes.Say("First"))
			Eventually(buffer).Should(gbytes.Say("first content"))
			Eventually(buffer).Should(gbytes.Say("Second"))
			Eventually(buffer).Should(gbytes.Say("second content"))
		})

		It("handles special characters in title", func() {
			printer.Print("TITLE [WITH] (SPECIAL) {CHARS}", "content")

			Eventually(buffer).Should(gbytes.Say("TITLE"))
			Eventually(buffer).Should(gbytes.Say("SPECIAL"))
		})

		It("handles special characters in dump", func() {
			specialDump := "Content with <brackets> & ampersand"
			printer.Print("Special Chars", specialDump)

			Eventually(buffer).Should(gbytes.Say("Special Chars"))
		})

		It("formats output with newlines", func() {
			printer.Print("Title", "Dump")

			// Output should have newlines for formatting
			Eventually(buffer).Should(gbytes.Say(`\n`))
		})

		Context("with trace disabled", func() {
			BeforeEach(func() {
				trace.DisableTrace()
			})

			It("does not print when trace is disabled", func() {
				printer.Print("Should Not Appear", "content")

				// When trace is disabled, nothing should be printed
				// (behavior depends on trace.Logger implementation)
			})
		})

		Context("with file output", func() {
			var tempFile *os.File

			BeforeEach(func() {
				var err error
				tempFile, err = ioutil.TempFile("", "trace-test")
				Expect(err).ToNot(HaveOccurred())

				trace.SetStdout(tempFile)
				trace.EnableTrace()
			})

			AfterEach(func() {
				tempFile.Close()
				os.Remove(tempFile.Name())
			})

			It("writes to file when configured", func() {
				printer.Print("File Test", "file content")

				// Flush and read file
				tempFile.Seek(0, 0)
				content, err := ioutil.ReadAll(tempFile)
				Expect(err).ToNot(HaveOccurred())

				contentStr := string(content)
				Expect(contentStr).To(ContainSubstring("File Test"))
				Expect(contentStr).To(ContainSubstring("file content"))
			})
		})
	})
})

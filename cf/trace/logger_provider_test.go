package trace_test

import (
	"io/ioutil"
	"path"
	"runtime"

	. "code.cloudfoundry.org/cli/cf/trace"
	"code.cloudfoundry.org/gofileutils/fileutils"

	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("NewLogger", func() {
	var buffer *gbytes.Buffer
	BeforeEach(func() {
		buffer = gbytes.NewBuffer()
	})

	It("returns a logger that doesn't write anywhere when nothing is set", func() {
		logger := NewLogger(buffer, false, "", "")

		logger.Print("Hello World")

		Expect(buffer).NotTo(gbytes.Say("Hello World"))
	})

	It("returns a logger that only writes to STDOUT when verbose is set", func() {
		logger := NewLogger(buffer, true, "", "")

		logger.Print("Hello World")

		Expect(buffer).To(gbytes.Say("Hello World"))
	})

	It("returns a logger that only writes to STDOUT when CF_TRACE=true", func() {
		logger := NewLogger(buffer, false, "true", "")

		logger.Print("Hello World")

		Expect(buffer).To(gbytes.Say("Hello World"))

		_, err := os.Open("true")
		Expect(err).To(HaveOccurred())
	})

	It("returns a logger that only writes to STDOUT when config.trace=true", func() {
		logger := NewLogger(buffer, false, "", "true")

		logger.Print("Hello World")

		Expect(buffer).To(gbytes.Say("Hello World"))

		_, err := os.Open("true")
		Expect(err).To(HaveOccurred())
	})

	It("returns a logger that only writes to STDOUT when verbose is set and CF_TRACE=false", func() {
		logger := NewLogger(buffer, true, "false", "")

		logger.Print("Hello World")

		Expect(buffer).To(gbytes.Say("Hello World"))

		_, err := os.Open("false")
		Expect(err).To(HaveOccurred())
	})

	It("returns a logger that only writes to STDOUT when verbose is set and config.trace=false", func() {
		logger := NewLogger(buffer, true, "", "false")

		logger.Print("Hello World")

		Expect(buffer).To(gbytes.Say("Hello World"))

		_, err := os.Open("false")
		Expect(err).To(HaveOccurred())
	})

	It("returns a logger that writes to STDOUT and a file when verbose is set and CF_TRACE is a path", func() {
		fileutils.TempFile("trace_test", func(file *os.File, err error) {
			logger := NewLogger(buffer, true, file.Name(), "")

			logger.Print("Hello World")

			Expect(buffer).To(gbytes.Say("Hello World"))

			fileContents, _ := ioutil.ReadAll(file)
			Expect(fileContents).To(ContainSubstring("Hello World"))
		})
	})

	It("creates the file with 0600 permission", func() {
		// cannot use fileutils.TempFile because it sets the permissions to 0600
		// itself
		fileutils.TempDir("trace_test", func(tmpDir string, err error) {
			Expect(err).ToNot(HaveOccurred())

			fileName := path.Join(tmpDir, "trace_test")
			logger := NewLogger(buffer, true, fileName, "")
			logger.Print("Hello World")

			stat, err := os.Stat(fileName)
			Expect(err).ToNot(HaveOccurred())
			if runtime.GOOS == "windows" {
				Expect(stat.Mode().String()).To(Equal(os.FileMode(0666).String()))
			} else {
				Expect(stat.Mode().String()).To(Equal(os.FileMode(0600).String()))
			}
		})
	})

	It("returns a logger that writes to STDOUT and a file when verbose is set and config.trace is a path", func() {
		fileutils.TempFile("trace_test", func(file *os.File, err error) {
			logger := NewLogger(buffer, true, "", file.Name())

			logger.Print("Hello World")

			Expect(buffer).To(gbytes.Say("Hello World"))

			fileContents, _ := ioutil.ReadAll(file)
			Expect(fileContents).To(ContainSubstring("Hello World"))
		})
	})

	It("returns a logger that writes to a file when CF_TRACE is a path", func() {
		fileutils.TempFile("trace_test", func(file *os.File, err error) {
			logger := NewLogger(buffer, false, file.Name(), "")

			logger.Print("Hello World")

			Expect(buffer).NotTo(gbytes.Say("Hello World"))

			fileContents, _ := ioutil.ReadAll(file)
			Expect(fileContents).To(ContainSubstring("Hello World"))
		})
	})

	It("returns a logger that writes to a file when config.trace is a path", func() {
		fileutils.TempFile("trace_test", func(file *os.File, err error) {
			logger := NewLogger(buffer, false, "", file.Name())

			logger.Print("Hello World")

			Expect(buffer).NotTo(gbytes.Say("Hello World"))

			fileContents, _ := ioutil.ReadAll(file)
			Expect(fileContents).To(ContainSubstring("Hello World"))
		})
	})

	It("returns a logger that writes to multiple files when CF_TRACE and config.trace are both paths", func() {
		fileutils.TempFile("cf_trace_test", func(cfTraceFile *os.File, err error) {
			fileutils.TempFile("config_trace_test", func(configTraceFile *os.File, err error) {
				logger := NewLogger(buffer, false, cfTraceFile.Name(), configTraceFile.Name())

				logger.Print("Hello World")

				Expect(buffer).NotTo(gbytes.Say("Hello World"))

				cfTraceFileContents, _ := ioutil.ReadAll(cfTraceFile)
				Expect(cfTraceFileContents).To(ContainSubstring("Hello World"))

				configTraceFileContents, _ := ioutil.ReadAll(configTraceFile)
				Expect(configTraceFileContents).To(ContainSubstring("Hello World"))
			})
		})
	})

	It("returns a logger that writes to STDOUT when CF_TRACE is a path that cannot be opened", func() {
		if runtime.GOOS != "windows" {
			logger := NewLogger(buffer, false, "/dev/null/whoops", "")

			logger.Print("Hello World")

			Expect(buffer).To(gbytes.Say("Hello World"))
		}
	})

	It("returns a logger that writes to STDOUT when config.trace is a path that cannot be opened", func() {
		if runtime.GOOS != "windows" {
			logger := NewLogger(buffer, false, "", "/dev/null/whoops")

			logger.Print("Hello World")

			Expect(buffer).To(gbytes.Say("CF_TRACE ERROR CREATING LOG FILE /dev/null/whoops"))
			Expect(buffer).To(gbytes.Say("Hello World"))
		}
	})

	It("returns a logger that writes to STDOUT when verbose is set and CF_TRACE is a path that cannot be opened", func() {
		if runtime.GOOS != "windows" {
			logger := NewLogger(buffer, true, "", "/dev/null/whoops")

			logger.Print("Hello World")

			Expect(buffer).To(gbytes.Say("CF_TRACE ERROR CREATING LOG FILE /dev/null/whoops"))
			Expect(buffer).To(gbytes.Say("Hello World"))
		}
	})
})

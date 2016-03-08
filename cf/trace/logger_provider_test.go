package trace_test

import (
	"bytes"
	"io/ioutil"
	"runtime"

	. "github.com/cloudfoundry/cli/cf/trace"
	"github.com/cloudfoundry/gofileutils/fileutils"

	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NewLogger", func() {
	var stdout *bytes.Buffer

	BeforeEach(func() {
		stdout = bytes.NewBuffer([]byte{})
		SetStdout(stdout)
	})

	AfterEach(func() {
		SetStdout(os.Stdout)
	})

	It("returns a logger that doesn't write anywhere when nothing is set", func() {
		logger := NewLogger(false, "", "")

		logger.Print("Hello World")

		stdoutContents, _ := ioutil.ReadAll(stdout)
		Expect(stdoutContents).NotTo(ContainSubstring("Hello World"))
	})

	It("returns a logger that only writes to STDOUT when verbose is set", func() {
		logger := NewLogger(true, "", "")

		logger.Print("Hello World")

		stdoutContents, _ := ioutil.ReadAll(stdout)
		Expect(stdoutContents).To(ContainSubstring("Hello World"))
	})

	It("returns a logger that only writes to STDOUT when CF_TRACE=true", func() {
		logger := NewLogger(false, "true", "")

		logger.Print("Hello World")

		stdoutContents, _ := ioutil.ReadAll(stdout)
		Expect(stdoutContents).To(ContainSubstring("Hello World"))

		_, err := os.Open("true")
		Expect(err).To(HaveOccurred())
	})

	It("returns a logger that only writes to STDOUT when config.trace=true", func() {
		logger := NewLogger(false, "", "true")

		logger.Print("Hello World")

		stdoutContents, _ := ioutil.ReadAll(stdout)
		Expect(stdoutContents).To(ContainSubstring("Hello World"))

		_, err := os.Open("true")
		Expect(err).To(HaveOccurred())
	})

	It("returns a logger that only writes to STDOUT when verbose is set and CF_TRACE=false", func() {
		logger := NewLogger(true, "false", "")

		logger.Print("Hello World")

		stdoutContents, _ := ioutil.ReadAll(stdout)
		Expect(stdoutContents).To(ContainSubstring("Hello World"))

		_, err := os.Open("false")
		Expect(err).To(HaveOccurred())
	})

	It("returns a logger that only writes to STDOUT when verbose is set and config.trace=false", func() {
		logger := NewLogger(true, "", "false")

		logger.Print("Hello World")

		stdoutContents, _ := ioutil.ReadAll(stdout)
		Expect(stdoutContents).To(ContainSubstring("Hello World"))

		_, err := os.Open("false")
		Expect(err).To(HaveOccurred())
	})

	It("returns a logger that writes to STDOUT and a file when verbose is set and CF_TRACE is a path", func() {
		fileutils.TempFile("trace_test", func(file *os.File, err error) {
			logger := NewLogger(true, file.Name(), "")

			logger.Print("Hello World")

			stdoutContents, _ := ioutil.ReadAll(stdout)
			Expect(stdoutContents).To(ContainSubstring("Hello World"))

			fileContents, _ := ioutil.ReadAll(file)
			Expect(fileContents).To(ContainSubstring("Hello World"))
		})
	})

	It("returns a logger that writes to STDOUT and a file when verbose is set and config.trace is a path", func() {
		fileutils.TempFile("trace_test", func(file *os.File, err error) {
			logger := NewLogger(true, "", file.Name())

			logger.Print("Hello World")

			stdoutContents, _ := ioutil.ReadAll(stdout)
			Expect(stdoutContents).To(ContainSubstring("Hello World"))

			fileContents, _ := ioutil.ReadAll(file)
			Expect(fileContents).To(ContainSubstring("Hello World"))
		})
	})

	It("returns a logger that writes to a file when CF_TRACE is a path", func() {
		fileutils.TempFile("trace_test", func(file *os.File, err error) {
			logger := NewLogger(false, file.Name(), "")

			logger.Print("Hello World")

			stdoutContents, _ := ioutil.ReadAll(stdout)
			Expect(stdoutContents).NotTo(ContainSubstring("Hello World"))

			fileContents, _ := ioutil.ReadAll(file)
			Expect(fileContents).To(ContainSubstring("Hello World"))
		})
	})

	It("returns a logger that writes to a file when config.trace is a path", func() {
		fileutils.TempFile("trace_test", func(file *os.File, err error) {
			logger := NewLogger(false, "", file.Name())

			logger.Print("Hello World")

			stdoutContents, _ := ioutil.ReadAll(stdout)
			Expect(stdoutContents).NotTo(ContainSubstring("Hello World"))

			fileContents, _ := ioutil.ReadAll(file)
			Expect(fileContents).To(ContainSubstring("Hello World"))
		})
	})

	It("returns a logger that writes to multiple files when CF_TRACE and config.trace are both paths", func() {
		fileutils.TempFile("cf_trace_test", func(cfTraceFile *os.File, err error) {
			fileutils.TempFile("config_trace_test", func(configTraceFile *os.File, err error) {
				logger := NewLogger(false, cfTraceFile.Name(), configTraceFile.Name())

				logger.Print("Hello World")

				stdoutContents, _ := ioutil.ReadAll(stdout)
				Expect(stdoutContents).NotTo(ContainSubstring("Hello World"))

				cfTraceFileContents, _ := ioutil.ReadAll(cfTraceFile)
				Expect(cfTraceFileContents).To(ContainSubstring("Hello World"))

				configTraceFileContents, _ := ioutil.ReadAll(configTraceFile)
				Expect(configTraceFileContents).To(ContainSubstring("Hello World"))
			})
		})
	})

	It("returns a logger that writes to STDOUT when CF_TRACE is a path that cannot be opened", func() {
		if runtime.GOOS != "windows" {
			logger := NewLogger(false, "/dev/null/whoops", "")

			logger.Print("Hello World")

			stdoutContents, _ := ioutil.ReadAll(stdout)
			Expect(stdoutContents).To(ContainSubstring("Hello World"))
		}
	})

	It("returns a logger that writes to STDOUT when config.trace is a path that cannot be opened", func() {
		if runtime.GOOS != "windows" {
			logger := NewLogger(false, "", "/dev/null/whoops")

			logger.Print("Hello World")

			stdoutContents, _ := ioutil.ReadAll(stdout)
			Expect(stdoutContents).To(ContainSubstring("CF_TRACE ERROR CREATING LOG FILE /dev/null/whoops"))
			Expect(stdoutContents).To(ContainSubstring("Hello World"))
		}
	})

	It("returns a logger that writes to STDOUT when verbose is set and CF_TRACE is a path that cannot be opened", func() {
		if runtime.GOOS != "windows" {
			logger := NewLogger(true, "", "/dev/null/whoops")

			logger.Print("Hello World")

			stdoutContents, _ := ioutil.ReadAll(stdout)
			Expect(stdoutContents).To(ContainSubstring("CF_TRACE ERROR CREATING LOG FILE /dev/null/whoops"))
			Expect(stdoutContents).To(ContainSubstring("Hello World"))
		}
	})
})

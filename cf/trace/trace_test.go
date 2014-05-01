package trace_test

import (
	"bytes"
	"github.com/cloudfoundry/cli/cf/trace"
	"github.com/cloudfoundry/gofileutils/fileutils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
	"runtime"
)

var _ = Describe("trace logger", func() {
	It("does not emit logs when CF_TRACE not set", func() {
		stdOut := bytes.NewBuffer([]byte{})
		trace.SetStdout(stdOut)

		os.Setenv(trace.CF_TRACE, "false")

		logger := trace.NewLogger()
		logger.Print("hello world")

		result, _ := ioutil.ReadAll(stdOut)
		Expect(string(result)).To(Equal(""))
	})

	It("emits messages when CF_TRACE is set to 'true'", func() {
		stdOut := bytes.NewBuffer([]byte{})
		trace.SetStdout(stdOut)

		os.Setenv(trace.CF_TRACE, "true")

		logger := trace.NewLogger()
		logger.Print("hello world")

		result, _ := ioutil.ReadAll(stdOut)
		Expect(string(result)).To(ContainSubstring("hello world"))
	})

	Context("when CF_TRACE is set to a filename", func() {
		It("writes logs to that file", func() {
			stdOut := bytes.NewBuffer([]byte{})
			trace.SetStdout(stdOut)

			fileutils.TempFile("trace_test", func(file *os.File, err error) {
				Expect(err).NotTo(HaveOccurred())
				file.Write([]byte("pre-existing content"))

				os.Setenv(trace.CF_TRACE, file.Name())

				logger := trace.NewLogger()
				logger.Print("hello world")

				file.Seek(0, os.SEEK_SET)
				result, err := ioutil.ReadAll(file)
				Expect(err).NotTo(HaveOccurred())

				byteString := string(result)
				Expect(byteString).To(ContainSubstring("pre-existing content"))
				Expect(byteString).To(ContainSubstring("hello world"))

				result, _ = ioutil.ReadAll(stdOut)
				Expect(string(result)).To(Equal(""))
			})
		})
	})

	Context("when CF_TRACE is set to a file path that cannot be opened", func() {
		It("defaults to printing to its out pipe", func() {
			if runtime.GOOS != "windows" {
				stdOut := bytes.NewBuffer([]byte{})
				trace.SetStdout(stdOut)

				os.Setenv(trace.CF_TRACE, "/dev/null/whoops")

				logger := trace.NewLogger()
				logger.Print("hello world")

				result, _ := ioutil.ReadAll(stdOut)
				Expect(string(result)).To(ContainSubstring("hello world"))
			}
		})

	})
})

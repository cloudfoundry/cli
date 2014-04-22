/*
                       WARNING WARNING WARNING

                Attention all potential contributors

   This testfile is not in the best state. We've been slowly transitioning
   from the built in "testing" package to using Ginkgo. As you can see, we've
   changed the format, but a lot of the setup, test body, descriptions, etc
   are either hardcoded, completely lacking, or misleading.

   For example:

   Describe("Testing with ginkgo"...)      // This is not a great description
   It("TestDoesSoemthing"...)              // This is a horrible description

   Describe("create-user command"...       // Describe the actual object under test
   It("creates a user when provided ..."   // this is more descriptive

   For good examples of writing Ginkgo tests for the cli, refer to

   src/cf/commands/application/delete_app_test.go
   src/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

package trace_test

import (
	"bytes"
	"cf/trace"
	"github.com/cloudfoundry/gofileutils/fileutils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
	"runtime"
)

var _ = Describe("Testing with ginkgo", func() {
	It("TestTraceSetToFalse", func() {
		stdOut := bytes.NewBuffer([]byte{})
		trace.SetStdout(stdOut)

		os.Setenv(trace.CF_TRACE, "false")

		logger := trace.NewLogger()
		logger.Print("hello world")

		result, _ := ioutil.ReadAll(stdOut)
		Expect(string(result)).To(Equal(""))
	})

	It("TestTraceSetToTrue", func() {
		stdOut := bytes.NewBuffer([]byte{})
		trace.SetStdout(stdOut)

		os.Setenv(trace.CF_TRACE, "true")

		logger := trace.NewLogger()
		logger.Print("hello world")

		result, _ := ioutil.ReadAll(stdOut)
		Expect(string(result)).To(ContainSubstring("hello world"))
	})

	It("TestTraceSetToFile", func() {
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

	It("defaults to a StdoutLogger if the CF_TRACE file cannot be opened", func() {
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

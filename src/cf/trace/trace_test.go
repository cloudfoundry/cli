package trace_test

import (
	"bytes"
	"cf/trace"
	"fileutils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
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
		assert.Contains(mr.T(), string(result), "hello world")
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
			assert.Contains(mr.T(), byteString, "pre-existing content")
			assert.Contains(mr.T(), byteString, "hello world")

			result, _ = ioutil.ReadAll(stdOut)
			Expect(string(result)).To(Equal(""))
		})
	})
	It("TestTraceSetToInvalidFile", func() {

		if runtime.GOOS != "windows" {
			stdOut := bytes.NewBuffer([]byte{})
			trace.SetStdout(stdOut)

			fileutils.TempFile("trace_test", func(file *os.File, err error) {
				Expect(err).NotTo(HaveOccurred())

				file.Chmod(0000)

				os.Setenv(trace.CF_TRACE, file.Name())

				logger := trace.NewLogger()
				logger.Print("hello world")

				result, _ := ioutil.ReadAll(file)
				Expect(string(result)).To(Equal(""))

				result, _ = ioutil.ReadAll(stdOut)
				assert.Contains(mr.T(), string(result), "hello world")
			})
		}
	})
})

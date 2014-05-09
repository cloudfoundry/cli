package trace_test

import (
	"bytes"
	"github.com/cloudfoundry/gofileutils/fileutils"
	"io/ioutil"
	"os"
	"runtime"

	. "github.com/cloudfoundry/cli/cf/trace"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("trace logger", func() {
	Describe("a new, better API", func() {
		var (
			stdout *bytes.Buffer
		)

		BeforeEach(func() {
			stdout = bytes.NewBuffer([]byte{})
			SetStdout(stdout)
		})

		It("assumes it should write to stdout", func() {
			logger := NewLogger("true")
			logger.Print("hello whirled")

			result, err := ioutil.ReadAll(stdout)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(ContainSubstring("hello whirled"))
		})

		It("prints to nothing when given false", func() {
			logger := NewLogger("false")
			logger.Print("hello whirled")

			result, err := ioutil.ReadAll(stdout)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(BeEmpty())
		})

		It("prints to a file when given a string", func() {
			fileutils.TempFile("trace_test", func(file *os.File, err error) {
				Expect(err).NotTo(HaveOccurred())
				file.Write([]byte("pre-existing content"))

				logger := NewLogger(file.Name())
				logger.Print("hello world")

				file.Seek(0, os.SEEK_SET)
				result, err := ioutil.ReadAll(file)
				Expect(err).NotTo(HaveOccurred())

				byteString := string(result)
				Expect(byteString).To(ContainSubstring("pre-existing content"))
				Expect(byteString).To(ContainSubstring("hello world"))

				result, _ = ioutil.ReadAll(stdout)
				Expect(string(result)).To(BeEmpty())
			})
		})

		Context("when CF_TRACE is set to a file path that cannot be opened", func() {
			It("defaults to printing to its out pipe", func() {
				if runtime.GOOS != "windows" {
					stdOut := bytes.NewBuffer([]byte{})
					SetStdout(stdOut)

					logger := NewLogger("/dev/null/whoops")
					logger.Print("hello world")

					result, _ := ioutil.ReadAll(stdOut)
					Expect(string(result)).To(ContainSubstring("hello world"))
				}
			})
		})
	})
})

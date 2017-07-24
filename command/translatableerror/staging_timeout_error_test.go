package translatableerror_test

import (
	"bytes"
	"io"
	"text/template"
	"time"

	. "code.cloudfoundry.org/cli/command/translatableerror"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StagingTimeoutError", func() {
	Describe("Translate()", func() {
		var translateFunc func(string, ...interface{}) string

		BeforeEach(func() {
			translateFunc = func(templateStr string, subs ...interface{}) string {
				t := template.Must(template.New("some-text-template").Parse(templateStr))
				buffer := bytes.NewBuffer([]byte{})
				err := t.Execute(buffer, subs[0])
				Expect(err).NotTo(HaveOccurred())
				translatedStr, err := buffer.ReadString('\n')
				if err != io.EOF {
					Expect(err).NotTo(HaveOccurred())
				}
				return translatedStr
			}
		})

		Context("when called with a float", func() {
			It("prints the error without trailing zeros", func() {
				err := StagingTimeoutError{
					AppName: "sliders",
					Timeout: 150 * time.Second,
				}

				Expect(err.Translate(translateFunc)).To(Equal("Error staging application sliders: timed out after 2.5 minutes"))
			})
		})

		Context("when called with an integer", func() {
			It("prints the error with integer precision", func() {
				err := StagingTimeoutError{
					AppName: "sliders",
					Timeout: 120 * time.Second,
				}

				Expect(err.Translate(translateFunc)).To(Equal("Error staging application sliders: timed out after 2 minutes"))
			})
		})

		Context("when called with a timeout of less than one minute", func() {
			It("prints the error with 'minutes' instead of 'minute'", func() {
				err := StagingTimeoutError{
					AppName: "sliders",
					Timeout: 30 * time.Second,
				}

				Expect(err.Translate(translateFunc)).To(Equal("Error staging application sliders: timed out after 0.5 minutes"))
			})
		})

		Context("when called with a timeout of exactly one minute", func() {
			It("prints the error with 'minute' instead of 'minutes'", func() {
				err := StagingTimeoutError{
					AppName: "sliders",
					Timeout: 60 * time.Second,
				}

				Expect(err.Translate(translateFunc)).To(Equal("Error staging application sliders: timed out after 1 minute"))
			})
		})

		Context("when called with a timeout of more than one minute", func() {
			It("prints the error with 'minutes' instead of 'minute'", func() {
				err := StagingTimeoutError{
					AppName: "sliders",
					Timeout: 120 * time.Second,
				}

				Expect(err.Translate(translateFunc)).To(Equal("Error staging application sliders: timed out after 2 minutes"))
			})
		})
	})
})

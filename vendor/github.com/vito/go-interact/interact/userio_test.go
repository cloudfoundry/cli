package interact_test

import (
	"io/ioutil"
	"os"

	"github.com/kr/pty"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vito/go-interact/interact"
)

var _ = Describe("User IO", func() {
	Describe("fetching input from the user", func() {
		Context("when the terminal reports Ctrl-C was pressed", func() {
			It("returns ErrKeyboardInterrupt", func() {
				aPty, tty, err := pty.Open()
				Expect(err).NotTo(HaveOccurred())

				interaction := interact.NewInteraction("What is the air-speed of a Swallow?")
				interaction.Input = aPty
				interaction.Output = aPty

				go func() {
					defer GinkgoRecover()

					_, err = tty.Write([]byte{03})
					Expect(err).NotTo(HaveOccurred())
				}()

				var thing string
				err = interaction.Resolve(&thing)

				Expect(err).To(Equal(interact.ErrKeyboardInterrupt))
			})
		})
	})

	Describe("fetching input from a non-TTY user", func() {
		Context("when passed a CRLF", func() {
			var input, output *os.File

			BeforeEach(func() {
				var err error
				input, err = ioutil.TempFile("", "go-interact-input")
				Expect(err).ToNot(HaveOccurred())
				err = ioutil.WriteFile(input.Name(), []byte("What do you mean? An African or European swallow?\r\n"), 0644)
				Expect(err).ToNot(HaveOccurred())

				output, err = ioutil.TempFile("", "go-interact-output")
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				Expect(os.RemoveAll(input.Name())).ToNot(HaveOccurred())
				Expect(os.RemoveAll(output.Name())).ToNot(HaveOccurred())
			})

			It("ignores the CR", func() {
				interaction := interact.NewInteraction("What is the air-speed of a Swallow?")
				interaction.Input = input
				interaction.Output = output

				var thing string
				err := interaction.Resolve(&thing)
				Expect(err).ToNot(HaveOccurred())
				Expect(thing).To(Equal("What do you mean? An African or European swallow?"))
			})
		})
	})
})

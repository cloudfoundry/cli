package terminal_test

import (
	"bytes"
	"cf/configuration"
	"cf/models"
	. "cf/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io"
	"os"
	"strings"
	testassert "testhelpers/assert"
	testconfig "testhelpers/configuration"
)

var _ = Describe("UI", func() {
	Describe("Printing message to stdout with Say", func() {
		It("prints strings", func() {
			simulateStdin("", func(reader io.Reader) {
				output := captureOutput(func() {
					ui := NewUI(reader)
					ui.Say("Hello")
				})

				Expect("Hello").To(Equal(strings.Join(output, "")))
			})
		})

		It("prints formatted strings", func() {
			simulateStdin("", func(reader io.Reader) {
				output := captureOutput(func() {
					ui := NewUI(reader)
					ui.Say("Hello %s", "World!")
				})

				Expect("Hello World!").To(Equal(strings.Join(output, "")))
			})
		})
	})

	Describe("Confirming user input", func() {
		It("treats 'y' as an affirmative confirmation", func() {
			simulateStdin("y\n", func(reader io.Reader) {
				out := captureOutput(func() {
					ui := NewUI(reader)
					Expect(ui.Confirm("Hello %s", "World?")).To(BeTrue())
				})

				testassert.SliceContains(out, testassert.Lines{{"Hello World?"}})
			})
		})

		It("treats other input as a negative confirmation", func() {
			simulateStdin("wat\n", func(reader io.Reader) {
				out := captureOutput(func() {
					ui := NewUI(reader)
					Expect(ui.Confirm("Hello %s", "World?")).To(BeFalse())
				})

				testassert.SliceContains(out, testassert.Lines{{"Hello World?"}})
			})
		})
	})

	Context("when user is not logged in", func() {
		var config configuration.Reader

		BeforeEach(func() {
			config = testconfig.NewRepository()
		})

		It("prompts the user to login", func() {
			output := captureOutput(func() {
				ui := NewUI((os.Stdin))
				ui.ShowConfiguration(config)
			})

			testassert.SliceContains(output, testassert.Lines{
				{"Not logged in", "Use", "log in"},
			})
		})

	})

	Context("when an api endpoint is set and the user logged in", func() {
		var config configuration.ReadWriter

		BeforeEach(func() {
			config = testconfig.NewRepository()
			config.SetApiEndpoint("https://test.example.org")
			config.SetAccessToken("some-access-token")
		})

		It("prompts the user to target an org and space when no org or space is targeted", func() {
			output := captureOutput(func() {
				ui := NewUI(os.Stdin)
				ui.ShowConfiguration(config)
			})

			testassert.SliceContains(output, testassert.Lines{
				{"No", "org", "space", "targeted", "-o ORG", "-s SPACE"},
			})
		})

		It("prompts the user to target an org when no org is targeted", func() {
			sf := models.SpaceFields{}
			sf.Guid = "guid"
			sf.Name = "name"

			output := captureOutput(func() {
				ui := NewUI(os.Stdin)
				ui.ShowConfiguration(config)
			})

			testassert.SliceContains(output, testassert.Lines{
				{"No", "org", "targeted", "-o ORG"},
			})
		})

		It("prompts the user to target a space when no space is targeted", func() {
			of := models.OrganizationFields{}
			of.Guid = "of-guid"
			of.Name = "of-name"

			output := captureOutput(func() {
				ui := NewUI(os.Stdin)
				ui.ShowConfiguration(config)
			})

			testassert.SliceContains(output, testassert.Lines{
				{"No", "space", "targeted", "-s SPACE"},
			})
		})
	})

	Describe("failing", func() {
		It("panics with a specific string", func() {
			testassert.AssertPanic(FailedWasCalled, func() {
				NewUI(os.Stdin).Failed("uh oh")
			})
		})
	})
})

func simulateStdin(input string, block func(r io.Reader)) {
	reader, writer := io.Pipe()

	go func() {
		writer.Write([]byte(input))
		defer writer.Close()
	}()

	block(reader)
}

func captureOutput(block func()) []string {
	oldSTDOUT := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = oldSTDOUT
	}()

	block()
	w.Close()

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return strings.Split(buf.String(), "\n")
}

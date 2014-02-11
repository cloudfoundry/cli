package terminal_test

import (
	"bytes"
	"cf/models"
	. "cf/terminal"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	"io"
	"os"
	"strings"
	testassert "testhelpers/assert"
	testconfig "testhelpers/configuration"
)

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
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestSayWithStringOnly", func() {
			simulateStdin("", func(reader io.Reader) {
				output := captureOutput(func() {
					ui := NewUI(reader)
					ui.Say("Hello")
				})

				assert.Equal(mr.T(), "Hello", strings.Join(output, ""))
			})
		})
		It("TestSayWithStringWithFormat", func() {

			simulateStdin("", func(reader io.Reader) {
				output := captureOutput(func() {
					ui := NewUI(reader)
					ui.Say("Hello %s", "World!")
				})

				assert.Equal(mr.T(), "Hello World!", strings.Join(output, ""))
			})
		})
		It("TestConfirmYes", func() {

			simulateStdin("y\n", func(reader io.Reader) {
				out := captureOutput(func() {
					ui := NewUI(reader)
					assert.True(mr.T(), ui.Confirm("Hello %s", "World?"))
				})

				testassert.SliceContains(mr.T(), out, testassert.Lines{{"Hello World?"}})
			})
		})
		It("TestConfirmNo", func() {

			simulateStdin("wat\n", func(reader io.Reader) {
				_ = captureOutput(func() {
					ui := NewUI(reader)
					assert.False(mr.T(), ui.Confirm("Hello %s", "World?"))
				})
			})
		})
		It("TestShowConfigurationWhenNoOrgAndSpaceTargeted", func() {
			config := testconfig.NewRepository()
			output := captureOutput(func() {
				ui := NewUI(os.Stdin)
				ui.ShowConfiguration(config)
			})

			testassert.SliceContains(mr.T(), output, testassert.Lines{
				{"No", "org", "space", "targeted", "-o ORG", "-s SPACE"},
			})
		})
		It("TestShowConfigurationWhenNoOrgTargeted", func() {

			sf := models.SpaceFields{}
			sf.Guid = "guid"
			sf.Name = "name"

			config := testconfig.NewRepository()

			output := captureOutput(func() {
				ui := NewUI(os.Stdin)
				ui.ShowConfiguration(config)
			})

			testassert.SliceContains(mr.T(), output, testassert.Lines{
				{"No", "org", "targeted", "-o ORG"},
			})
		})
		It("TestShowConfigurationWhenNoSpaceTargeted", func() {

			of := models.OrganizationFields{}
			of.Guid = "of-guid"
			of.Name = "of-name"

			config := testconfig.NewRepository()

			output := captureOutput(func() {
				ui := NewUI(os.Stdin)
				ui.ShowConfiguration(config)
			})

			testassert.SliceContains(mr.T(), output, testassert.Lines{
				{"No", "space", "targeted", "-s SPACE"},
			})
		})
	})
}

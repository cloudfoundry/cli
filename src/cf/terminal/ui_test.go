package terminal_test

import (
	"bytes"
	"cf"
	"cf/configuration"
	. "cf/terminal"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"strings"
	testassert "testhelpers/assert"
	"testing"
)

func TestSayWithStringOnly(t *testing.T) {
	simulateStdin("", func(reader io.Reader) {
		output := captureOutput(func() {
			ui := NewUI(reader)
			ui.Say("Hello")
		})

		assert.Equal(t, "Hello", strings.Join(output, ""))
	})
}

func TestSayWithStringWithFormat(t *testing.T) {
	simulateStdin("", func(reader io.Reader) {
		output := captureOutput(func() {
			ui := NewUI(reader)
			ui.Say("Hello %s", "World!")
		})

		assert.Equal(t, "Hello World!", strings.Join(output, ""))
	})
}

func TestConfirmYes(t *testing.T) {
	simulateStdin("y\n", func(reader io.Reader) {
		out := captureOutput(func() {
			ui := NewUI(reader)
			assert.True(t, ui.Confirm("Hello %s", "World?"))
		})

		testassert.SliceContains(t, out, testassert.Lines{{"Hello World?"}})
	})
}

func TestConfirmNo(t *testing.T) {
	simulateStdin("wat\n", func(reader io.Reader) {
		_ = captureOutput(func() {
			ui := NewUI(reader)
			assert.False(t, ui.Confirm("Hello %s", "World?"))
		})
	})
}

func TestShowConfigurationWhenNoOrgAndSpaceTargeted(t *testing.T) {
	config := &configuration.Configuration{AccessToken: "speak, friend, and enter."}

	output := captureOutput(func() {
		ui := NewUI(os.Stdin)
		ui.ShowConfiguration(config)
	})

	testassert.SliceContains(t, output, testassert.Lines{
		{"No", "org", "space", "targeted", "-o ORG", "-s SPACE"},
	})
}

func TestShowConfigurationWhenNoOrgTargeted(t *testing.T) {
	sf := cf.SpaceFields{}
	sf.Guid = "guid"
	sf.Name = "name"

	config := &configuration.Configuration{
		AccessToken: "speak, friend, and enter.",
		SpaceFields: sf,
	}

	output := captureOutput(func() {
		ui := NewUI(os.Stdin)
		ui.ShowConfiguration(config)
	})

	testassert.SliceContains(t, output, testassert.Lines{
		{"No", "org", "targeted", "-o ORG"},
	})
}

func TestShowConfigurationWhenNoSpaceTargeted(t *testing.T) {
	of := cf.OrganizationFields{}
	of.Guid = "of-guid"
	of.Name = "of-name"

	config := &configuration.Configuration{
		AccessToken:        "speak, friend, and enter.",
		OrganizationFields: of,
	}

	output := captureOutput(func() {
		ui := NewUI(os.Stdin)
		ui.ShowConfiguration(config)
	})

	testassert.SliceContains(t, output, testassert.Lines{
		{"No", "space", "targeted", "-s SPACE"},
	})
}

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

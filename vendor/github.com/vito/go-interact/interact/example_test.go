package interact_test

import (
	"bytes"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/vito/go-interact/interact"
)

var (
	destination interface{}

	choices []interact.Choice
)

var _ = BeforeEach(func() {
	destination = nil
	choices = nil
})

type Example struct {
	Prompt  string
	Choices []interact.Choice

	Input string

	ExpectedAnswer interface{}
	ExpectedErr    error

	ExpectedOutput string
}

func (example Example) Run() {
	input := bytes.NewBufferString(example.Input)
	output := gbytes.NewBuffer()

	interaction := interact.NewInteraction(example.Prompt, choices...)
	interaction.Input = input
	interaction.Output = output

	resolveErr := interaction.Resolve(destination)

	if example.ExpectedErr != nil {
		Expect(resolveErr).To(Equal(example.ExpectedErr))
	} else {
		Expect(resolveErr).ToNot(HaveOccurred())
	}

	var finalDestination interface{}
	switch d := destination.(type) {
	case interact.RequiredDestination:
		finalDestination = d.Destination
	default:
		finalDestination = destination
	}

	Expect(reflect.Indirect(reflect.ValueOf(finalDestination)).Interface()).To(Equal(example.ExpectedAnswer))

	Expect(output.Contents()).To(Equal([]byte(example.ExpectedOutput)))
}

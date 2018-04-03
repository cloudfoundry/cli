package fakes

import (
	"fmt"
	"sync"

	. "github.com/cloudfoundry/bosh-cli/ui/table"
)

type FakeUI struct {
	Said   []string
	Errors []string

	Blocks []string // keep as string to make ginkgo err msgs easier

	Table  Table
	Tables []Table

	AskedTextLabels []string
	AskedText       []Answer

	AskedPasswordLabels []string
	AskedPasswords      []Answer

	AskedChoiceCalled  bool
	AskedChoiceLabel   string
	AskedChoiceOptions []string
	AskedChoiceChosens []int
	AskedChoiceErrs    []error

	AskedConfirmationCalled bool
	AskedConfirmationErr    error

	Interactive bool

	Flushed bool

	mutex sync.Mutex
}

type Answer struct {
	Text  string
	Error error
}

func (ui *FakeUI) ErrorLinef(pattern string, args ...interface{}) {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	ui.Errors = append(ui.Errors, fmt.Sprintf(pattern, args...))
}

func (ui *FakeUI) PrintLinef(pattern string, args ...interface{}) {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	ui.Said = append(ui.Said, fmt.Sprintf(pattern, args...))
}

func (ui *FakeUI) BeginLinef(pattern string, args ...interface{}) {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	ui.Said = append(ui.Said, fmt.Sprintf(pattern, args...))
}

func (ui *FakeUI) EndLinef(pattern string, args ...interface{}) {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	ui.Said = append(ui.Said, fmt.Sprintf(pattern, args...))
}

func (ui *FakeUI) PrintBlock(block []byte) {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	ui.Blocks = append(ui.Blocks, string(block))
}

func (ui *FakeUI) PrintErrorBlock(block string) {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	ui.Blocks = append(ui.Blocks, block)
}

func (ui *FakeUI) PrintTable(table Table) {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	ui.Table = table
	ui.Tables = append(ui.Tables, table)
}

func (ui *FakeUI) AskForText(label string) (string, error) {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	ui.AskedTextLabels = append(ui.AskedTextLabels, label)
	answer := ui.AskedText[0]
	ui.AskedText = ui.AskedText[1:]
	return answer.Text, answer.Error
}

func (ui *FakeUI) AskForChoice(label string, options []string) (int, error) {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	ui.AskedChoiceCalled = true

	ui.AskedChoiceLabel = label
	ui.AskedChoiceOptions = options

	chosen := ui.AskedChoiceChosens[0]
	ui.AskedChoiceChosens = ui.AskedChoiceChosens[1:]

	err := ui.AskedChoiceErrs[0]
	ui.AskedChoiceErrs = ui.AskedChoiceErrs[1:]

	return chosen, err
}

func (ui *FakeUI) AskForPassword(label string) (string, error) {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	ui.AskedPasswordLabels = append(ui.AskedPasswordLabels, label)
	answer := ui.AskedPasswords[0]
	ui.AskedPasswords = ui.AskedPasswords[1:]
	return answer.Text, answer.Error
}

func (ui *FakeUI) AskForConfirmation() error {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	ui.AskedConfirmationCalled = true
	return ui.AskedConfirmationErr
}

func (ui *FakeUI) IsInteractive() bool {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	return ui.Interactive
}

func (ui *FakeUI) Flush() {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	ui.Flushed = true
}

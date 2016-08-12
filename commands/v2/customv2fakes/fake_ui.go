package customv2fakes

import (
	"fmt"
	"text/template"

	"github.com/onsi/gomega/gbytes"
)

// FakeUI is a custom fake written to stub out the user interface. The reason
// this is a custom fake is because there are multiple output commands that can
// be intertwined and using straight counterfeiter would be difficult.
type FakeUI struct {
	// Out is the equivalent of STDOUT
	Out *gbytes.Buffer

	inTTY bool

	displayTextCallCount int

	displayTextWithKeyTranslationsCallCount int

	displayFlavorTextCallCount    int
	displayFlavorTextDisplayCount int

	displayFlavorTextWithKeyTranslationsCallCount    int
	displayFlavorTextWithKeyTranslationsDisplayCount int

	helpHeaderCallCount int
}

// NewFakeUI is the constructor for a FakeUI. If TTY is set to false, flavour
// text will not be captured.
func NewFakeUI(TTY bool) *FakeUI {
	return &FakeUI{
		Out:   gbytes.NewBuffer(),
		inTTY: TTY,
	}
}

// DisplayText captures the input text in the Fake UI buffer so that this can
// be asserted against in tests. If multiple maps are passed in, the merge will
// give precedence to the latter maps.
func (ui *FakeUI) DisplayText(template string, keys ...map[string]interface{}) {
	ui.displayTextCallCount = ui.displayTextCallCount + 1
	ui.outputToSTDOUT(template, keys...)
}

// DisplayTextCallCount returns the number of times DisplayText was called.
func (ui FakeUI) DisplayTextCallCount() int {
	return ui.displayTextCallCount
}

// DisplayTextWithKeyTranslations captures the input text in the Fake UI buffer
// so that this can be asserted against in tests. If multiple maps are passed
// in, the merge will give precedence to the latter maps.
func (ui *FakeUI) DisplayTextWithKeyTranslations(template string, _ []string, keys ...map[string]interface{}) {
	ui.displayTextWithKeyTranslationsCallCount = ui.displayTextWithKeyTranslationsCallCount + 1
	ui.outputToSTDOUT(template, keys...)
}

// DisplayTextWithKeyTranslationsCallCount returns the number of times
// DisplayTextWithKeyTranslations was called.
func (ui FakeUI) DisplayTextWithKeyTranslationsCallCount() int {
	return ui.displayTextWithKeyTranslationsCallCount
}

// DisplayFlavorText captures the input text in the Fake UI buffer so that this
// can be asserted against in tests. If TTY is false, flavour text will not be
// captured. If multiple maps are passed in, the merge will give precedence to
// the latter maps.
func (ui *FakeUI) DisplayFlavorText(template string, keys ...map[string]interface{}) {
	ui.displayFlavorTextCallCount = ui.displayFlavorTextCallCount + 1

	if ui.inTTY {
		ui.displayFlavorTextDisplayCount = ui.displayFlavorTextDisplayCount + 1
		ui.outputToSTDOUT(template, keys...)
	}
}

// DisplayFlavorTextCallCount returns the number of times DisplayFlavorText was
// called.
func (ui FakeUI) DisplayFlavorTextCallCount() int {
	return ui.displayFlavorTextCallCount
}

// DisplayFlavorTextDisplayCount returns the number of times DisplayFlavorText
// was displayed on the screen.
func (ui FakeUI) DisplayFlavorTextDisplayCount() int {
	return ui.displayFlavorTextDisplayCount
}

// DisplayFlavorTextWithKeyTranslations captures the input text in the Fake UI
// buffer so that this can be asserted against in tests. If TTY is false,
// flavour text will not be captured. If multiple maps are passed in, the merge
// will give precedence to the latter maps.
func (ui *FakeUI) DisplayFlavorTextWithKeyTranslations(template string, _ []string, keys ...map[string]interface{}) {
	ui.displayFlavorTextWithKeyTranslationsCallCount = ui.displayFlavorTextWithKeyTranslationsCallCount + 1

	if ui.inTTY {
		ui.displayFlavorTextWithKeyTranslationsDisplayCount = ui.displayFlavorTextWithKeyTranslationsDisplayCount + 1
		ui.outputToSTDOUT(template, keys...)
	}
}

// DisplayFlavorTextWithKeyTranslationsCallCount returns the number of times
// DisplayFlavorTextWithKeyTranslations was called.
func (ui FakeUI) DisplayFlavorTextWithKeyTranslationsCallCount() int {
	return ui.displayFlavorTextWithKeyTranslationsCallCount
}

// DisplayFlavorTextWithKeyTranslationsDisplayCount returns the number of times
// DisplayFlavorTextWithKeyTranslations was displayed on the screen.
func (ui FakeUI) DisplayFlavorTextWithKeyTranslationsDisplayCount() int {
	return ui.displayFlavorTextWithKeyTranslationsDisplayCount
}

// DisplayNewline adds a newline to the Out buffer.
func (ui FakeUI) DisplayNewline() {
	fmt.Fprintf(ui.Out, "\n")
}

// DisplayHelpHeader tracks the number of times DisplayHelpHeader is called.
func (ui FakeUI) DisplayHelpHeader(text string) {
	ui.helpHeaderCallCount += 1
	ui.outputToSTDOUT(text)
}

// DisplayHelpHeaderCount returns the number of times DisplayHelpHeader was called.
func (ui FakeUI) DisplayHelpHeaderCount() int {
	return ui.helpHeaderCallCount
}

func (ui FakeUI) mergeMap(maps []map[string]interface{}) map[string]interface{} {
	if len(maps) == 1 {
		return maps[0]
	}

	main := map[string]interface{}{}

	for _, minor := range maps {
		for key, value := range minor {
			main[key] = value
		}
	}

	return main
}

func (ui FakeUI) outputToSTDOUT(formattedString string, keys ...map[string]interface{}) {
	formattedTemplate := template.Must(template.New("Display Text").Parse(formattedString + "\n"))
	formattedTemplate.Execute(ui.Out, ui.mergeMap(keys))
}

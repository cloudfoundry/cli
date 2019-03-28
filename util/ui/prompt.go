package ui

import (
	"io"

	"github.com/vito/go-interact/interact"
	"github.com/vito/go-interact/interact/terminal"
)

const sigIntExitCode = 130

//go:generate counterfeiter . Resolver

type Resolver interface {
	Resolve(dst interface{}) error
	SetIn(io.Reader)
	SetOut(io.Writer)
}

// DisplayBoolPrompt outputs the prompt and waits for user input. It only
// allows for a boolean response. A default boolean response can be set with
// defaultResponse.
func (ui *UI) DisplayBoolPrompt(defaultResponse bool, template string, templateValues ...map[string]interface{}) (bool, error) {
	ui.terminalLock.Lock()
	defer ui.terminalLock.Unlock()

	response := defaultResponse
	interactivePrompt := ui.Interactor.NewInteraction(ui.TranslateText(template, templateValues...))
	interactivePrompt.SetIn(ui.In)
	interactivePrompt.SetOut(ui.OutForInteration)
	err := interactivePrompt.Resolve(&response)
	if isInterrupt(err) {
		ui.Exiter.Exit(sigIntExitCode)
	}
	return response, err
}

// DisplayPasswordPrompt outputs the prompt and waits for user input. Hides
// user's response from the screen.
func (ui *UI) DisplayPasswordPrompt(template string, templateValues ...map[string]interface{}) (string, error) {
	ui.terminalLock.Lock()
	defer ui.terminalLock.Unlock()

	var password interact.Password
	interactivePrompt := ui.Interactor.NewInteraction(ui.TranslateText(template, templateValues...))
	interactivePrompt.SetIn(ui.In)
	interactivePrompt.SetOut(ui.OutForInteration)
	err := interactivePrompt.Resolve(interact.Required(&password))
	if isInterrupt(err) {
		ui.Exiter.Exit(sigIntExitCode)
	}
	return string(password), err
}

// DisplayTextPrompt outputs the prompt and waits for user input.
func (ui *UI) DisplayTextPrompt(template string, templateValues ...map[string]interface{}) (string, error) {
	interactivePrompt := ui.Interactor.NewInteraction(ui.TranslateText(template, templateValues...))
	var value string
	interactivePrompt.SetIn(ui.In)
	interactivePrompt.SetOut(ui.OutForInteration)
	err := interactivePrompt.Resolve(interact.Required(&value))
	if isInterrupt(err) {
		ui.Exiter.Exit(sigIntExitCode)
	}
	return value, err
}

func isInterrupt(err error) bool {
	return err == interact.ErrKeyboardInterrupt || err == terminal.ErrKeyboardInterrupt
}

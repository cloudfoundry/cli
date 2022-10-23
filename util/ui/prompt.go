package ui

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/vito/go-interact/interact"
	"github.com/vito/go-interact/interact/terminal"
)

const sigIntExitCode = 130

var ErrInvalidIndex = errors.New("invalid list index")

type InvalidChoiceError struct {
	Choice string
}

func (InvalidChoiceError) Error() string {
	return "Some error"
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Resolver

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
	interactivePrompt.SetOut(ui.OutForInteraction)
	err := interactivePrompt.Resolve(&response)
	if isInterrupt(err) {
		ui.Exiter.Exit(sigIntExitCode)
	}
	return response, err
}

// DisplayOptionalTextPrompt outputs the prompt and waits for user input.
func (ui *UI) DisplayOptionalTextPrompt(defaultValue string, template string, templateValues ...map[string]interface{}) (string, error) {
	interactivePrompt := ui.Interactor.NewInteraction(ui.TranslateText(template, templateValues...))
	var value = defaultValue
	interactivePrompt.SetIn(ui.In)
	interactivePrompt.SetOut(ui.OutForInteraction)
	err := interactivePrompt.Resolve(&value)
	if isInterrupt(err) {
		ui.Exiter.Exit(sigIntExitCode)
	}
	return value, err
}

// DisplayPasswordPrompt outputs the prompt and waits for user input. Hides
// user's response from the screen.
func (ui *UI) DisplayPasswordPrompt(template string, templateValues ...map[string]interface{}) (string, error) {
	ui.terminalLock.Lock()
	defer ui.terminalLock.Unlock()

	var password interact.Password
	interactivePrompt := ui.Interactor.NewInteraction(ui.TranslateText(template, templateValues...))
	interactivePrompt.SetIn(ui.In)
	interactivePrompt.SetOut(ui.OutForInteraction)
	err := interactivePrompt.Resolve(interact.Required(&password))
	if isInterrupt(err) {
		ui.Exiter.Exit(sigIntExitCode)
	}
	return string(password), err
}

// DisplayTextMenu lets the user choose from a list of options, either by name
// or by number.
func (ui *UI) DisplayTextMenu(choices []string, promptTemplate string, templateValues ...map[string]interface{}) (string, error) {
	for i, c := range choices {
		t := fmt.Sprintf("%d. %s", i+1, c)
		ui.DisplayText(t)
	}
	ui.DisplayNewline()

	translatedPrompt := ui.TranslateText(promptTemplate, templateValues...)

	interactivePrompt := ui.Interactor.NewInteraction(translatedPrompt)

	interactivePrompt.SetIn(ui.In)
	interactivePrompt.SetOut(ui.OutForInteraction)

	var value string = "enter to skip"
	err := interactivePrompt.Resolve(&value)

	if isInterrupt(err) {
		ui.Exiter.Exit(sigIntExitCode)
	}

	if err != nil {
		return "", err
	}

	if value == "enter to skip" {
		return "", nil
	}

	i, err := strconv.Atoi(value)
	if err != nil {
		if contains(choices, value) {
			return value, nil
		}
		return "", InvalidChoiceError{Choice: value}
	}

	if i > len(choices) || i <= 0 {
		return "", ErrInvalidIndex
	}
	return choices[i-1], nil
}

// DisplayTextPrompt outputs the prompt and waits for user input.
func (ui *UI) DisplayTextPrompt(template string, templateValues ...map[string]interface{}) (string, error) {
	interactivePrompt := ui.Interactor.NewInteraction(ui.TranslateText(template, templateValues...))
	var value string
	interactivePrompt.SetIn(ui.In)
	interactivePrompt.SetOut(ui.OutForInteraction)
	err := interactivePrompt.Resolve(interact.Required(&value))
	if isInterrupt(err) {
		ui.Exiter.Exit(sigIntExitCode)
	}
	return value, err
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

func isInterrupt(err error) bool {
	return err == interact.ErrKeyboardInterrupt || err == terminal.ErrKeyboardInterrupt
}

package ui

import "github.com/vito/go-interact/interact"

// DisplayBoolPrompt outputs the prompt and waits for user input. It only
// allows for a boolean response. A default boolean response can be set with
// defaultResponse.
func (ui *UI) DisplayBoolPrompt(defaultResponse bool, template string, templateValues ...map[string]interface{}) (bool, error) {
	ui.terminalLock.Lock()
	defer ui.terminalLock.Unlock()

	response := defaultResponse
	interactivePrompt := interact.NewInteraction(ui.TranslateText(template, templateValues...))
	interactivePrompt.Input = ui.In
	interactivePrompt.Output = ui.OutForInteration
	err := interactivePrompt.Resolve(&response)
	return response, err
}

// DisplayPasswordPrompt outputs the prompt and waits for user input. Hides
// user's response from the screen.
func (ui *UI) DisplayPasswordPrompt(template string, templateValues ...map[string]interface{}) (string, error) {
	ui.terminalLock.Lock()
	defer ui.terminalLock.Unlock()

	var password interact.Password
	interactivePrompt := interact.NewInteraction(ui.TranslateText(template, templateValues...))
	interactivePrompt.Input = ui.In
	interactivePrompt.Output = ui.OutForInteration
	err := interactivePrompt.Resolve(interact.Required(&password))
	return string(password), err
}

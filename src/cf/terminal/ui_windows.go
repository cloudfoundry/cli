// +build windows

package terminal

func (ui TerminalUI) AskForPassword(prompt string, args ...interface{}) (passwd string) {
	return ui.Ask(prompt, args...)
}

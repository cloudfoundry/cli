package terminal

import "fmt"

type UI interface {
	Say(message string, args ...interface{})
	Ask(prompt string) (answer string)
}

type TerminalUI struct {
}

func (c TerminalUI) Say(message string, args ...interface{}) {
	fmt.Printf(message+"\n", args...)
	return
}

func (c TerminalUI) Ask(prompt string) (answer string) {
	fmt.Printf(prompt + " ")
	fmt.Scanln(&answer)
	return
}

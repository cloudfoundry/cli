package terminal

import "fmt"

type UI interface {
	Say(message string, args ...interface{})
	Ask(prompt string) (answer string)
}

type ConsoleUI struct {
}

func (c ConsoleUI) Say(message string, args ...interface{}) {
	fmt.Printf(message+"\n", args...)
	return
}

func (c ConsoleUI) Ask(prompt string) (answer string) {
	fmt.Printf(prompt + " ")
	fmt.Scanln(&answer)
	return
}

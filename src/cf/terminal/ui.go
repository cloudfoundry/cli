package terminal

import "fmt"

type UI interface {
	Say(message string, args ...interface{})
	Ask(prompt string) (answer string)
	Failed(message string, err error)
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

func (c TerminalUI) Failed(message string, err error) {
	c.Say(Red("FAILED"))
	c.Say(message)

	if err != nil {
		c.Say(err.Error())
	}
	return
}

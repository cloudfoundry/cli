package terminal

import "fmt"

func Say(message string, args ...interface{}) {
	fmt.Printf(message+"\n", args...)
	return
}

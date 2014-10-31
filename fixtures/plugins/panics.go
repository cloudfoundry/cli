/**
	* 1. Setup the server so cf can call it under main.
				e.g. `cf my-plugin` creates the callable server. now we can call the Run command
	* 2. Implement Run that is the actual code of the plugin!
	* 3. Return an error
**/

package main

import (
	"os"

	"github.com/cloudfoundry/cli/plugin"
)

type Panics struct {
}

func (c *Panics) Run(args []string) {
	if args[0] == "panic" {
		panic("OMG")
	} else if args[0] == "exit1" {
		os.Exit(1)
	}
}

func (c *Panics) GetCommands() []plugin.Command {
	return []plugin.Command{
		{
			Name:     "panic",
			HelpText: "omg panic",
		},
		{
			Name:     "exit1",
			HelpText: "omg exit1",
		},
	}
}

func (c *Panics) GetName() string {
	return "Panics"
}

func main() {
	plugin.Start(new(Panics))
}

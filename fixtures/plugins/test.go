/**
	* 1. Setup the server so cf can call it under main.
				e.g. `cf my-plugin` creates the callable server. now we can call the Run command
	* 2. Implement Run that is the actual code of the plugin!
	* 3. Return an error
**/

package main

import (
	"fmt"

	"github.com/cloudfoundry/cli/plugin"
)

type PushCommand struct{}

func (c *PushCommand) Run(args []string, cmds *string) error {
	fmt.Println("HaHaHaHa you called the push plugin")

	return nil
}

func main() {
	plugin.ServeCommand(new(PushCommand))
}

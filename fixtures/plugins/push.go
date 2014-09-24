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

type CliPlugin struct{}

func (c *CliPlugin) Run(args string, reply *bool) error {
	if args == "push1" {
		theFirstPush()
	} else if args == "push2" {
		theSecondPush()
	}
	return nil
}

func (c *CliPlugin) CmdExists(args string, exists *bool) error {
	var reply bool
	if args == "push1" || args == "push2" {
		reply = true
	} else {
		reply = false

	}
	exists = &reply
	return nil
}

func theSecondPush() {
	fmt.Println("HaHaHaHa you called THE SECOND PUSH")
}

func theFirstPush() {
	fmt.Println("HaHaHaHa you called THE FIRST PUSH")
}

func main() {
	plugin.ServeCommand(new(CliPlugin), "20001")
}

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
	if args == "test1" {
		theFirstPush()
	} else if args == "test2" {
		theSecondPush()
	}
	return nil
}

func (c *CliPlugin) CmdExists(args string, exists *bool) error {
	if args == "test1" || args == "test2" {
		*exists = true
	} else {
		*exists = false
	}
	return nil
}

func theSecondPush() {
	fmt.Println("HaHaHaHa you called THE SECOND TEST")
}

func theFirstPush() {
	fmt.Println("HaHaHaHa you called THE FIRST TEST")
}

func main() {
	plugin.ServeCommand(new(CliPlugin), "20001")
}

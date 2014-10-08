package plugin

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
)

type Command struct {
	Name     string
	HelpText string
}

/**
	Command interface needs to be implementd for a runnable sub-command of `cf`
**/
type RpcPlugin interface {
	//run is passed in all the command line parameter arguments and
	//an object containing all of the cli commands available to them
	Run(args string, reply *bool) error
	CmdExists(args string, exists *bool) error

	//We only care about the return value.
	//TODO: the first param could be used for obtaining help of a specific command.
	ListCmds(empty string, cmdList *[]Command) error
}

/**
	This function is called by the plugin to setup their server. This allows us to call Run on the plugin
**/
func ServeCommand(cmd RpcPlugin) {
	//register command
	rpc.Register(cmd)

	listener, err := net.Listen("tcp", ":"+os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//listen for the run command
	for {
		if conn, err := listener.Accept(); err != nil {
			log.Fatal("accept error: " + err.Error())
		} else {
			log.Printf("new connection established\n")
			go rpc.ServeConn(conn)
		}
	}
}

func CmdExists(cmd string, availableCmds []Command) bool {
	for _, availableCmd := range availableCmds {
		if cmd == availableCmd.Name {
			return true
		}
	}
	return false
}

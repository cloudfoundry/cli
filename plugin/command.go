package plugin

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
)

/**
	Command interface needs to be implementd for a runnable sub-command of `cf`
**/
type RpcPlugin interface {
	//run is passed in all the command line parameter arguments and
	//an object containing all of the cli commands available to them
	Run(args string, reply *bool) error
	CmdExists(args string, exists *bool) error

	//We only care about the return value.
	ListCmds(empty string, cmdList *[]string) error
}

/**
	This function is called by the plugin to setup their server. This allows us to call Run on the plugin
**/
func ServeCommand(cmd RpcPlugin, port string) {
	//register command
	rpc.Register(cmd)

	listener, err := net.Listen("tcp", ":"+port)
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

func CmdExists(cmd string, availableCmds []string) bool {
	for _, availableCmd := range availableCmds {
		if cmd == availableCmd {
			return true
		}
	}
	return false
}

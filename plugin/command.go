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
type Command interface {
	//run is passed in all the command line parameter arguments and
	//an object containing all of the cli commands available to them
	Run(args []string, cmds *string) error
}

/**
	This function is called by the plugin to setup their server. This allows us to call Run on the plugin
**/
func ServeCommand(cmd Command) {
	//register command
	rpc.Register(cmd)

	listener, err := net.Listen("tcp", ":20001")
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

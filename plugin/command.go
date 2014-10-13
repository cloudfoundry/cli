package plugin

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"reflect"
	"time"
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
	* This function is called by the plugin to setup their server. This allows us to call Run on the plugin
	* os.Args[1] port plugin rpc will be listening on
	* os.Args[2] port CF_CLI rpc server is running on
	* os.Args[3] **OPTIONAL**
		* install-plugin - used to fetch the command name
**/
func ServeCommand(cmd RpcPlugin) {
	//register command
	err := rpc.Register(cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	listener, err := net.Listen("tcp", ":"+os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	pingCLI()

	//Send CLI the plugin command name.
	if len(os.Args) == 4 {
		if os.Args[3] == "install-plugin" {
			pluginName := reflect.TypeOf(cmd).Elem().Name()
			fmt.Println("reflecting: ", reflect.Indirect(reflect.ValueOf(cmd)).Type().Name())
			client, err := rpc.Dial("tcp", "127.0.0.1:"+os.Args[2])
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			var success bool
			fmt.Println("In ServeCommand,name: ", pluginName)
			err = client.Call("CliRpcCmd.SetName", pluginName, &success)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			if !success {
				//fmt.Println(fmt.Sprintf("There was an error registering the plugin name: %s", pluginName))
				os.Exit(1)
			}

			os.Exit(0)
		}
	}

	//listen for the run command
	for {
		if conn, err := listener.Accept(); err != nil {
			fmt.Println("accept error: " + err.Error())
		} else {
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

func pingCLI() {
	//call back to cf saying we have been setup
	var connErr error
	var conn net.Conn
	for i := 0; i < 5; i++ {
		conn, connErr = net.Dial("tcp", "127.0.0.1:"+os.Args[2])
		if connErr != nil {
			time.Sleep(200 * time.Millisecond)
		} else {
			conn.Close()
			break
		}
	}
	if connErr != nil {
		fmt.Println(connErr)
		os.Exit(1)
	}
}

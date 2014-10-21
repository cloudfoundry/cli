package plugin

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"reflect"
	"time"
)

var CliServicePort string

type Command struct {
	Name     string
	HelpText string
}

type PluginMetadata struct {
	Name     string
	Commands []Command
}

/**
	Command interface needs to be implementd for a runnable sub-command of `cf`
**/
type RpcPlugin interface {
	//run is passed in all the command line parameter arguments and
	//an object containing all of the cli commands available to them
	Run(args []string, reply *bool) error
	GetCommands() []Command
}

/**
	* This function is called by the plugin to setup their server. This allows us to call Run on the plugin
	* os.Args[1] port plugin rpc will be listening on
	* os.Args[2] port CF_CLI rpc server is running on
	* os.Args[3] **OPTIONAL**
		* SendMetadata - used to fetch the plugin metadata
**/
func Start(cmd RpcPlugin) {
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
		if os.Args[3] == "SendMetadata" {
			pluginName := reflect.TypeOf(cmd).Elem().Name()
			client, err := rpc.Dial("tcp", "127.0.0.1:"+os.Args[2])
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			var success bool
			pluginMetadata := PluginMetadata{
				Name:     pluginName,
				Commands: cmd.GetCommands(),
			}
			err = client.Call("CliRpcCmd.SetPluginMetadata", pluginMetadata, &success)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			if !success {
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

func CliCommand(args ...string) error {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+CliServicePort)
	if err != nil {
		return err
	}

	var success bool
	err = client.Call("CliRpcCmd.CallCoreCommand", args, &success)
	if err != nil {
		return err
	} else if !success {
		return errors.New("Error executing cli core command")
	}

	return nil
}

func pingCLI() {
	//call back to cf saying we have been setup
	var connErr error
	var conn net.Conn
	for i := 0; i < 5; i++ {
		CliServicePort = os.Args[2]
		conn, connErr = net.Dial("tcp", "127.0.0.1:"+CliServicePort)
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

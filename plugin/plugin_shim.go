package plugin

import (
	"fmt"
	"os"
	"strconv"
)

/**
	* This function is called by the plugin to setup their server. This allows us to call Run on the plugin
	* os.Args[1] port CF_CLI rpc server is running on
	* os.Args[2] **OPTIONAL**
		* SendMetadata - used to fetch the plugin metadata
**/
func Start(cmd Plugin) {
	if len(os.Args) < 2 {
		fmt.Printf("This cf CLI plugin is not intended to be run on its own\n\n")
		os.Exit(1)
	}

	cliConnection := NewCliConnection(os.Args[1])
	cliConnection.pingCLI()
	if isMetadataRequest(os.Args) {
		cliConnection.sendPluginMetadataToCliServer(cmd.GetMetadata())
	} else {
		if version := MinCliVersionStr(cmd.GetMetadata().MinCliVersion); version != "" {
			ok := cliConnection.isMinCliVersion(version)
			if !ok {
				fmt.Printf("Minimum CLI version %s is required to run this plugin command\n\n", version)
				os.Exit(0)
			}
		}

		cmd.Run(cliConnection, os.Args[2:])
	}
}

func isMetadataRequest(args []string) bool {
	return len(args) == 3 && args[2] == "SendMetadata"
}

func MinCliVersionStr(version VersionType) string {
	if version.Major == 0 && version.Minor == 0 && version.Build == 0 {
		return ""
	}

	return strconv.Itoa(version.Major) + "." + strconv.Itoa(version.Minor) + "." + strconv.Itoa(version.Build)
}

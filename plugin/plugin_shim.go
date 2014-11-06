package plugin

import "os"

/**
	* This function is called by the plugin to setup their server. This allows us to call Run on the plugin
	* os.Args[1] port CF_CLI rpc server is running on
	* os.Args[2] **OPTIONAL**
		* SendMetadata - used to fetch the plugin metadata
**/
func Start(cmd Plugin) {
	cliConnection := NewCliConnection(os.Args[1])

	cliConnection.pingCLI()
	if isMetadataRequest(os.Args) {
		cliConnection.sendPluginMetadataToCliServer(cmd.GetMetadata())
	} else {
		cmd.Run(cliConnection, os.Args[2:])
	}
}

func isMetadataRequest(args []string) bool {
	return len(args) == 3 && args[2] == "SendMetadata"
}

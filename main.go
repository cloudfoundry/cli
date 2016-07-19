package main

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
)

func main() {
	traceEnv := os.Getenv("CF_TRACE")
	cmd.Main(traceEnv, os.Args)
}

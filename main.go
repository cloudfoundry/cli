package main

import (
	"os"

	"github.com/cloudfoundry/cli/cmd"
)

func main() {
	traceEnv := os.Getenv("CF_TRACE")
	cmd.Main(traceEnv, os.Args)
}

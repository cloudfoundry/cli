package cf

import (
	"os"
	"path/filepath"
)

const (
	Version = "BUILT_FROM_SOURCE"
	Usage   = "A command line tool to interact with Cloud Foundry"
)

func Name() string {
	return filepath.Base(os.Args[0])
}

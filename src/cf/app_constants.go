package cf

import (
	"os"
	"path/filepath"
)

const (
	Version = "6.0.2-SHA"
	Usage   = "A command line tool to interact with Cloud Foundry"
)

func Name() string {
	return filepath.Base(os.Args[0])
}

package cf

import "os"

const (
	Version = "1.0.0.rc1-SHA"
	Usage   = "A command line tool to interact with Cloud Foundry"
)

func Name() string {
	return os.Args[0]
}

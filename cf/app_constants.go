package cf

import (
	"os"
	"path/filepath"
)

var (
	Version     = "6.10.1"
	BuiltOnDate = "BUILT_AT_UNKNOWN_TIME"
)

func Name() string {
	return filepath.Base(os.Args[0])
}

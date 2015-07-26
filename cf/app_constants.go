package cf

import (
	"os"
	"path/filepath"
)

var (
	Version     = "BUILT_FROM_SOURCE"
	BuiltOnDate = "BUILT_AT_UNKNOWN_TIME"
)

func Name() string {
	return filepath.Base(os.Args[0])
}

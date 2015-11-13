package cf

import (
	"os"
	"path/filepath"
)

var (
	Version     string
	BuiltOnDate string
)

func Name() string {
	return filepath.Base(os.Args[0])
}

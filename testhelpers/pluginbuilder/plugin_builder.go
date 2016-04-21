package pluginbuilder

import (
	"os"
	"os/exec"
	"path/filepath"
)

func BuildTestBinary(relativePathToPluginDir, pluginFileName string) {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	binaryDestination := filepath.Join(dir, relativePathToPluginDir, pluginFileName+".exe")
	pluginSourceFile := filepath.Join(dir, relativePathToPluginDir, pluginFileName+".go")

	cmd := exec.Command("go", "build", "-o", binaryDestination, pluginSourceFile)
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
}

package plugin_builder

import (
	"os"
	"os/exec"
	"path/filepath"
)

func BuildTestBinary(relativePathToPluginFixturesDir, pluginFile string) {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	binaryDestination := filepath.Join(dir, relativePathToPluginFixturesDir, pluginFile+".exe")
	pluginSourceFile := filepath.Join(dir, relativePathToPluginFixturesDir, pluginFile+".go")

	cmd := exec.Command("go", "build", "-o", binaryDestination, pluginSourceFile)
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
}

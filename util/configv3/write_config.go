package configv3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/rogpeppe/go-internal/lockedfile"
)

// WriteConfig creates the .cf directory and then writes the config.json. The
// location of .cf directory is written in the same way LoadConfig reads .cf
// directory.
func (c *Config) WriteConfig() error {
	rawConfig, err := json.MarshalIndent(c.ConfigFile, "", "  ")
	if err != nil {
		return err
	}

	dir := configDirectory()
	err = os.MkdirAll(dir, 0700)
	if err != nil {
		return err
	}

	fmt.Println("WriteConfig():", ConfigFilePath())

	return lockedfile.Write(ConfigFilePath(), bytes.NewReader(rawConfig), 0600)
}

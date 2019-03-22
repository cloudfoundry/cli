package configv3

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
)

// WriteConfig creates the .cf directory and then writes the config.json. The
// location of .cf directory is written in the same way LoadConfig reads .cf
// directory.
func WriteConfig(c *Config) error {
	rawConfig, err := json.MarshalIndent(c.ConfigFile, "", "  ")
	if err != nil {
		return err
	}

	dir := configDirectory()
	err = os.MkdirAll(dir, 0700)
	if err != nil {
		return err
	}

	// Developer Note: The following is untested! Change at your own risk.
	// Setup notifications of termination signals to channel sig, create a process to
	// watch for these signals so we can remove transient config temp files.
	sig := make(chan os.Signal, 10)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, os.Interrupt)
	defer signal.Stop(sig)

	tempConfigFile, err := ioutil.TempFile(dir, "temp-config")
	if err != nil {
		return err
	}
	tempConfigFile.Close()
	tempConfigFileName := tempConfigFile.Name()

	go catchSignal(sig, tempConfigFileName)

	err = ioutil.WriteFile(tempConfigFileName, rawConfig, 0600)
	if err != nil {
		return err
	}

	return os.Rename(tempConfigFileName, ConfigFilePath())
}

// catchSignal tries to catch SIGHUP, SIGINT, SIGKILL, SIGQUIT and SIGTERM, and
// Interrupt for removing temporarily created config files before the program
// ends.  Note:  we cannot intercept a `kill -9`, so a well-timed `kill -9`
// will allow a temp config file to linger.
func catchSignal(sig chan os.Signal, tempConfigFileName string) {
	select {
	case <-sig:
		_ = os.Remove(tempConfigFileName)
		os.Exit(2)
	}
}

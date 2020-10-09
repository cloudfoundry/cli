package configv3

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"

	"github.com/gofrs/flock"
)

// WriteConfig creates the .cf directory and then writes the config.json. The
// location of .cf directory is written in the same way LoadConfig reads .cf
// directory.
func (c *Config) WriteConfig() error {
	fmt.Println("WriteConfig()")
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
	// sig := make(chan os.Signal, 10)
	// signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, os.Interrupt)
	// defer signal.Stop(sig)

	tempConfigFile, err := ioutil.TempFile(dir, randSeq(8))
	if err != nil {
		return err
	}
	tempConfigFile.Close()
	tempConfigFileName := tempConfigFile.Name()

	exists := fileExists(tempConfigFileName)
	if exists != true {
		fmt.Println("TEMP CONFIG FILE DOES NOT EXIST", tempConfigFileName)
	}

	// go catchSignal(sig, tempConfigFileName)

	fileLockConfig := flock.New(ConfigFilePath())
	locked, err := fileLockConfig.TryLock()
	defer handleUnlock(fileLockConfig)
	if err != nil {
		fmt.Println("flock: TryLock config failed")
		return err
	}

	if locked {
		fmt.Println("flock: lock obtained on config", fileLockConfig.Path())
		fileLock := flock.New(tempConfigFileName)
		// TODO: possibly poll, certain amount of retries, etc
		locked, err := fileLock.TryLock()
		defer handleUnlock(fileLock)
		if err != nil {
			fmt.Println("flock: TryLock temp failed")
			return err
		}

		if locked {
			fmt.Println("flock: lock obtained on temp config", fileLock.Path())
			err = ioutil.WriteFile(tempConfigFileName, rawConfig, 0600)
			if err != nil {
				return err
			}

			fmt.Println(fileExists(tempConfigFileName), ": temp config file", tempConfigFileName)
			fmt.Println(fileExists(ConfigFilePath()), ": real config file", ConfigFilePath())

			return os.Rename(tempConfigFileName, ConfigFilePath())
		} else {
			return errors.New("flock: could not acquire lock on temp")
		}
	} else {
		return errors.New("flock: could not acquire lock on config")
	}
}

func handleUnlock(f *flock.Flock) {
	err := f.Unlock()
	if err != nil {
		fmt.Println("flock: Unlock", err)
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	if info.IsDir() {
		fmt.Println(filename, "isDir")
	}
	return !info.IsDir()
}

// catchSignal tries to catch SIGHUP, SIGINT, SIGKILL, SIGQUIT and SIGTERM, and
// Interrupt for removing temporarily created config files before the program
// ends.  Note:  we cannot intercept a `kill -9`, so a well-timed `kill -9`
// will allow a temp config file to linger.
func catchSignal(sig chan os.Signal, tempConfigFileName string) {
	fmt.Println("catchSignal")
	<-sig
	_ = os.Remove(tempConfigFileName)
	os.Exit(2)
}

func randSeq(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

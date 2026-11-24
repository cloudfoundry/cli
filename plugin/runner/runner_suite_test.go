package runner_test

import (
	"encoding/json"
	"fmt"
	"net/rpc"
	"os"
	"testing"

	pluginmodels "code.cloudfoundry.org/cli/v8/plugin/models"
	"code.cloudfoundry.org/cli/v8/util/configv3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Run this test executable as a plugin subprocess without starting its suite.
func TestMain(m *testing.M) {
	mode := os.Getenv("CF_CLI_RUNNER_TEST_HELPER")
	if mode == "" {
		os.Exit(m.Run())
	}
	if err := runPluginTestHelper(mode); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}
	if len(os.Args) > 3 && os.Args[3] == "fail" {
		os.Exit(3)
	}
	os.Exit(0)
}

func runPluginTestHelper(mode string) error {
	if mode == "write-locale" {
		path := configv3.ConfigFilePath()
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var config configv3.JSONConfig
		if err := json.Unmarshal(data, &config); err != nil {
			return err
		}
		config.Locale = "fr-FR"
		data, err = json.Marshal(config)
		if err != nil {
			return err
		}
		return os.WriteFile(path, data, 0600)
	}
	if mode != "get-orgs" || len(os.Args) < 2 {
		return fmt.Errorf("invalid plugin test helper mode: %s", mode)
	}
	connection, err := rpc.Dial("tcp", "127.0.0.1:"+os.Args[1])
	if err != nil {
		return err
	}
	defer connection.Close()
	var orgs []pluginmodels.GetOrgs_Model
	return connection.Call("CliRpcCmd.GetOrgs", "", &orgs)
}

func TestRunner(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Plugin Runner Suite")
}

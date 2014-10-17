package rpc_test

import (
	"github.com/cloudfoundry/cli/plugin/rpc"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os"
	"os/exec"
	"path/filepath"

	"testing"
)

var rpcService *rpc.CliRpcService

func TestRpc(t *testing.T) {
	RegisterFailHandler(Fail)

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	cmd := exec.Command("go", "build", "-o", filepath.Join(dir, "..", "..", "fixtures", "plugins", "test_1.exe"), filepath.Join(dir, "..", "..", "fixtures", "plugins", "test_1.go"))
	err = cmd.Run()
	if err != nil {
		panic(err)
	}

	cmd = exec.Command("go", "build", "-o", filepath.Join(dir, "..", "..", "fixtures", "plugins", "test_2.exe"), filepath.Join(dir, "..", "..", "fixtures", "plugins", "test_2.go"))
	err = cmd.Run()
	if err != nil {
		panic(err)
	}

	RunSpecs(t, "Rpc Suite")
}

package commands_loader_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/commands_loader"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CommandsLoader", func() {
	It("references all command packages so all commands can be registered in command_registry", func() {

		commands_loader.Load()

		count := walkDirAndCountCommand("../cf/commands")
		count = count
		Ω(command_registry.Commands.TotalCommands()).To(Equal(count))
	})
})

func walkDirAndCountCommand(path string) int {
	cmdCount := 0

	filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("Error walking commands directories:", err)
			return err
		}

		if !info.IsDir() {
			if strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go") {
				cmdCount += 1
			}
		}
		return nil
	})

	return cmdCount
}

package commandsloader_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/commandsloader"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CommandsLoader", func() {
	It("references all command packages so all commands can be registered in commandregistry", func() {
		commandsloader.Load()

		count := walkDirAndCountCommand("../commands")
		Expect(commandregistry.Commands.TotalCommands()).To(Equal(count))
	})
})

func walkDirAndCountCommand(path string) int {
	cmdCount := 0

	filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("Error walking commands directories:", err)
			return err
		}

		dir := filepath.Dir(p)

		if !info.IsDir() && !strings.HasSuffix(dir, "fakes") {
			if strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go") {
				cmdCount += 1
			}
		}
		return nil
	})

	return cmdCount
}

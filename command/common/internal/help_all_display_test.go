// +build V7

package internal_test

import (
	"reflect"
	"sort"
	"strings"

	"code.cloudfoundry.org/cli/command/common"
	"code.cloudfoundry.org/cli/command/common/internal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("test help all display", func() {
	var fromCommandList []string
	var fromHelpAllDisplay []string

	BeforeEach(func() {
		commandList := common.Commands
		handler := reflect.TypeOf(commandList)
		for i := 0; i < handler.NumField(); i++ {
			fieldTag := handler.Field(i).Tag
			commandName := fieldTag.Get("command")
			if !(strings.HasPrefix(commandName, "v3-") || (commandName == "")) {
				fromCommandList = append(fromCommandList, commandName)
			}
		}

		for _, category := range internal.HelpCategoryList {
			for _, row := range category.CommandList {
				for _, command := range row {
					if !(strings.HasPrefix(command, "v3-") || (command == "")) {
						fromHelpAllDisplay = append(fromHelpAllDisplay, command)
					}
				}
			}
		}
	})

	It("matches up (not including experimental commands)", func() {
		sort.Strings(fromCommandList)
		sort.Strings(fromHelpAllDisplay)
		Expect(fromCommandList).To(ConsistOf(fromHelpAllDisplay))
	})
})

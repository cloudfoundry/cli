package command

import "fmt"

func DisplayNotLoggedInText(binaryName string, ui UI) {
	ui.DisplayText("Not logged in. Use '{{.CFLoginCommand}}' or '{{.CFLoginCommandSSO}}' to log in.",
		map[string]interface{}{
			"CFLoginCommand":    fmt.Sprintf("%s login", binaryName),
			"CFLoginCommandSSO": fmt.Sprintf("%s login --sso", binaryName),
		})
}

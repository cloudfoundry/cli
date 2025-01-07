package shared

import (
	"fmt"

	"code.cloudfoundry.org/cli/v8/actor/v7action"
	"code.cloudfoundry.org/cli/v8/command"
)

func WaitForResult(stream chan v7action.PollJobEvent, ui command.UI, waitForCompletion bool) (bool, error) {
	if stream == nil {
		return true, nil
	}

	if waitForCompletion {
		fmt.Fprint(ui.Writer(), "Waiting for the operation to complete")

		defer func() {
			ui.DisplayNewline()
			ui.DisplayNewline()
		}()
	}

	for event := range stream {
		ui.DisplayWarnings(event.Warnings)
		if waitForCompletion {
			fmt.Fprint(ui.Writer(), ".")
		}
		if event.Err != nil {
			return false, event.Err
		}
		if event.State == v7action.JobPolling && !waitForCompletion {
			return false, nil
		}
	}

	return true, nil
}

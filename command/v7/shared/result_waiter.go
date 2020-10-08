package shared

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
)

func WaitForResult(stream chan v7action.PollJobEvent, ui command.UI, waitForCompletion bool) (complete bool, err error) {
	if stream == nil {
		return true, nil
	}

	if waitForCompletion {
		fmt.Fprint(ui.Writer(), "Waiting for the operation to complete")
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
			break
		}
	}

	if waitForCompletion {
		ui.DisplayNewline()
		ui.DisplayNewline()
		return true, nil
	}
	return false, nil
}

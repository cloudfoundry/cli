package shared

import (
	"fmt"

	"code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/command"
)

func WaitForResult(stream chan v7action.PollJobEvent, ui command.UI, waitForCompletion bool) (bool, error) {
	if stream == nil {
		return true, nil
	}

	if waitForCompletion {
		fmt.Fprintln(ui.Writer(), "Waiting for the operation to complete")

		defer func() {
			ui.DisplayNewline()
			ui.DisplayNewline()
		}()
	}

	seen := map[string]bool{}
	for event := range stream {
		ui.DisplayWarnings(dedupeSeenWarnings(event.Warnings, seen))
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

// dedupeSeenWarnings prints each distinct warning at most once per operation:
// CC re-sends warnings like "still in progress" on every poll tick.
func dedupeSeenWarnings(warnings v7action.Warnings, seen map[string]bool) v7action.Warnings {
	var deduped v7action.Warnings
	for _, warning := range warnings {
		if seen[warning] {
			continue
		}
		seen[warning] = true
		deduped = append(deduped, warning)
	}

	return deduped
}

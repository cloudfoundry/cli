package flag

import flags "github.com/jessevdk/go-flags"

type SecurityGroupLifecycle string

func (SecurityGroupLifecycle) Complete(prefix string) []flags.Completion {
	return completions([]string{"staging", "running"}, prefix, false)
}

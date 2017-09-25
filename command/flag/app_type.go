package flag

import flags "github.com/jessevdk/go-flags"

type AppType string

func (AppType) Complete(prefix string) []flags.Completion {
	return completions([]string{"buildpack", "docker"}, prefix, false)
}

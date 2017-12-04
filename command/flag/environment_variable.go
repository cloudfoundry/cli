package flag

import (
	"fmt"
	"os"
	"strings"

	flags "github.com/jessevdk/go-flags"
)

// WorkAroundPrefix is the flag in hole emoji
const WorkAroundPrefix = "\U000026f3"

type EnvironmentVariable string

func (EnvironmentVariable) Complete(prefix string) []flags.Completion {
	if prefix == "" || prefix[0] != '$' {
		return nil
	}

	keyValPairs := os.Environ()
	envVars := make([]string, len(keyValPairs))
	for i, keyValPair := range keyValPairs {
		envVars[i] = fmt.Sprintf("$%s", strings.Split(keyValPair, "=")[0])
	}

	return completions(envVars, prefix, true)
}

func (e *EnvironmentVariable) UnmarshalFlag(val string) error {
	*e = EnvironmentVariable(strings.TrimLeft(val, WorkAroundPrefix))
	return nil
}

package helpers

import (
	"fmt"
	"strings"
)

func AddOrReplaceEnvironment(env []string, newEnvName string, newEnvVal string) []string {
	var found bool
	for i, envPair := range env {
		splitENV := strings.Split(envPair, "=")
		if splitENV[0] == newEnvName {
			env[i] = fmt.Sprintf("%s=%s", newEnvName, newEnvVal)
			found = true
		}
	}

	if !found {
		env = append(env, fmt.Sprintf("%s=%s", newEnvName, newEnvVal))
	}
	return env
}

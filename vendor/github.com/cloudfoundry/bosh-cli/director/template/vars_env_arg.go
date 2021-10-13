package template

import (
	"os"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"gopkg.in/yaml.v2"
)

type VarsEnvArg struct {
	Vars StaticVariables

	EnvironFunc func() []string
}

func (a *VarsEnvArg) UnmarshalFlag(prefix string) error {
	if len(prefix) == 0 {
		return bosherr.Errorf("Expected environment variable prefix to be non-empty")
	}

	if a.EnvironFunc == nil {
		a.EnvironFunc = os.Environ
	}

	vars := StaticVariables{}
	envVars := a.EnvironFunc()

	for _, envVar := range envVars {
		pieces := strings.SplitN(envVar, "=", 2)
		if len(pieces) != 2 {
			return bosherr.Error("Expected environment variable to be key-value pair")
		}

		if !strings.HasPrefix(pieces[0], prefix+"_") {
			continue
		}

		var val interface{}

		err := yaml.Unmarshal([]byte(pieces[1]), &val)
		if err != nil {
			return bosherr.WrapErrorf(err, "Deserializing YAML from environment variable '%s'", pieces[0])
		}

		vars[strings.TrimPrefix(pieces[0], prefix+"_")] = val
	}

	(*a).Vars = vars

	return nil
}

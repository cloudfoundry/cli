package cmd

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"github.com/cppforlife/go-patch/patch"
	"gopkg.in/yaml.v2"
)

type OpsFileArg struct {
	FS boshsys.FileSystem

	Ops patch.Ops
}

func (a *OpsFileArg) UnmarshalFlag(filePath string) error {
	if len(filePath) == 0 {
		return bosherr.Errorf("Expected file path to be non-empty")
	}

	bytes, err := a.FS.ReadFile(filePath)
	if err != nil {
		return bosherr.WrapErrorf(err, "Reading ops file '%s'", filePath)
	}

	var opDefs []patch.OpDefinition

	err = yaml.Unmarshal(bytes, &opDefs)
	if err != nil {
		return bosherr.WrapErrorf(err, "Deserializing ops file '%s'", filePath)
	}

	ops, err := patch.NewOpsFromDefinitions(opDefs)
	if err != nil {
		return bosherr.WrapErrorf(err, "Building ops")
	}

	(*a).Ops = ops

	return nil
}

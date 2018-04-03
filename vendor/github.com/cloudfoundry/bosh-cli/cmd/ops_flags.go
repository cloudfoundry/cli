package cmd

import (
	"github.com/cppforlife/go-patch/patch"
)

// Shared
type OpsFlags struct {
	OpsFiles []OpsFileArg `long:"ops-file" short:"o" value-name:"PATH" description:"Load manifest operations from a YAML file"`
}

func (f OpsFlags) AsOp() patch.Op {
	var ops patch.Ops

	for _, opsFile := range f.OpsFiles {
		ops = append(ops, opsFile.Ops...)
	}

	return ops
}

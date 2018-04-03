package pkg

import (
	birelpkg "github.com/cloudfoundry/bosh-cli/release/pkg"
)

type Compiler interface {
	Compile(birelpkg.Compilable) (CompiledPackageRecord, bool, error)
}

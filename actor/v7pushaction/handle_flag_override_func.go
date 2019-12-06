package v7pushaction

import "code.cloudfoundry.org/cli/util/manifestparser"

type HandleFlagOverrideFunc func(manifest manifestparser.Manifest, overrides FlagOverrides) (manifestparser.Manifest, error)

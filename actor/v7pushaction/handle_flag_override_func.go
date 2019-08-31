package v7pushaction

import "code.cloudfoundry.org/cli/util/pushmanifestparser"

type HandleFlagOverrideFunc func(manifest pushmanifestparser.Manifest, overrides FlagOverrides) (pushmanifestparser.Manifest, error)

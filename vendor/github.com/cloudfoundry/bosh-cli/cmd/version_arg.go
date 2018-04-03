package cmd

import (
	semver "github.com/cppforlife/go-semi-semantic/version"
)

type VersionArg semver.Version

func (a *VersionArg) UnmarshalFlag(data string) error {
	ver, err := semver.NewVersionFromString(data)
	if err != nil {
		return err
	}

	*a = VersionArg(ver)

	return nil
}

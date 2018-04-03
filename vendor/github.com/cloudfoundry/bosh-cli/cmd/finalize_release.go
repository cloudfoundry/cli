package cmd

import (
	semver "github.com/cppforlife/go-semi-semantic/version"

	boshrel "github.com/cloudfoundry/bosh-cli/release"
	boshreldir "github.com/cloudfoundry/bosh-cli/releasedir"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type FinalizeReleaseCmd struct {
	releaseReader boshrel.Reader
	releaseDir    boshreldir.ReleaseDir
	ui            boshui.UI
}

func NewFinalizeReleaseCmd(
	releaseReader boshrel.Reader,
	releaseDir boshreldir.ReleaseDir,
	ui boshui.UI,
) FinalizeReleaseCmd {
	return FinalizeReleaseCmd{
		releaseReader: releaseReader,
		releaseDir:    releaseDir,
		ui:            ui,
	}
}

func (c FinalizeReleaseCmd) Run(opts FinalizeReleaseOpts) error {
	release, err := c.releaseReader.Read(opts.Args.Path)
	if err != nil {
		return err
	}

	if len(opts.Name) > 0 {
		release.SetName(opts.Name)
	}

	version := semver.Version(opts.Version)

	if !version.Empty() {
		release.SetVersion(version.AsString())
	} else {
		version, err := c.releaseDir.NextFinalVersion(release.Name())
		if err != nil {
			return err
		}

		release.SetVersion(version.AsString())
	}

	err = c.releaseDir.FinalizeRelease(release, opts.Force)
	if err != nil {
		return err
	}

	ReleaseTables{Release: release}.Print(c.ui)

	return nil
}

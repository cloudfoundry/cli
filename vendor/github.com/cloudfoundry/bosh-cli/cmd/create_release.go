package cmd

import (
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshfu "github.com/cloudfoundry/bosh-utils/fileutil"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	semver "github.com/cppforlife/go-semi-semantic/version"

	boshrel "github.com/cloudfoundry/bosh-cli/release"
	boshreldir "github.com/cloudfoundry/bosh-cli/releasedir"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type CreateReleaseCmd struct {
	releaseDirFactory func(DirOrCWDArg) (boshrel.Reader, boshreldir.ReleaseDir)
	releaseWriter     boshrel.Writer
	fs                boshsys.FileSystem
	ui                boshui.UI
}

func NewCreateReleaseCmd(
	releaseDirFactory func(DirOrCWDArg) (boshrel.Reader, boshreldir.ReleaseDir),
	releaseWriter boshrel.Writer,
	fs boshsys.FileSystem,
	ui boshui.UI,
) CreateReleaseCmd {
	return CreateReleaseCmd{releaseDirFactory, releaseWriter, fs, ui}
}

func (c CreateReleaseCmd) Run(opts CreateReleaseOpts) (boshrel.Release, error) {
	releaseManifestReader, releaseDir := c.releaseDirFactory(opts.Directory)
	manifestGiven := len(opts.Args.Manifest.Path) > 0

	var release boshrel.Release
	var err error

	if manifestGiven {
		release, err = releaseManifestReader.Read(opts.Args.Manifest.Path)
		if err != nil {
			return nil, err
		}
	} else {
		release, err = c.buildRelease(releaseDir, opts)
		if err != nil {
			return nil, err
		}

		if opts.Final {
			err = c.finalizeRelease(releaseDir, release, opts)
			if err != nil {
				return nil, err
			}
		}
	}

	dstPath := opts.Tarball.ExpandedPath

	if dstPath != "" {
		path, err := c.releaseWriter.Write(release, nil)
		if err != nil {
			return nil, err
		}

		dstPath = strings.Replace(dstPath, "((name))", release.Name(), -1)
		dstPath = strings.Replace(dstPath, "((version))", release.Version(), -1)

		err = boshfu.NewFileMover(c.fs).Move(path, dstPath)
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Moving release archive to final destination")
		}
	}

	ReleaseTables{Release: release, ArchivePath: dstPath}.Print(c.ui)

	return release, nil
}

func (c CreateReleaseCmd) buildRelease(releaseDir boshreldir.ReleaseDir, opts CreateReleaseOpts) (boshrel.Release, error) {
	var err error

	name := opts.Name

	if len(name) == 0 {
		name, err = releaseDir.DefaultName()
		if err != nil {
			return nil, err
		}
	}

	version := semver.Version(opts.Version)

	if version.Empty() {
		version, err = releaseDir.NextDevVersion(name, opts.TimestampVersion)
		if err != nil {
			return nil, err
		}
	}

	return releaseDir.BuildRelease(name, version, opts.Force)
}

func (c CreateReleaseCmd) finalizeRelease(releaseDir boshreldir.ReleaseDir, release boshrel.Release, opts CreateReleaseOpts) error {
	version := semver.Version(opts.Version)

	if version.Empty() {
		version, err := releaseDir.NextFinalVersion(release.Name())
		if err != nil {
			return err
		}

		release.SetVersion(version.AsString())
	}

	return releaseDir.FinalizeRelease(release, opts.Force)
}

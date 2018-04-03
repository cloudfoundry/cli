package cmd

import (
	"fmt"

	semver "github.com/cppforlife/go-semi-semantic/version"

	boshreldir "github.com/cloudfoundry/bosh-cli/releasedir"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type VendorPackageCmd struct {
	releaseDirFactory func(DirOrCWDArg) boshreldir.ReleaseDir
	ui                boshui.UI
}

func NewVendorPackageCmd(
	releaseDirFactory func(DirOrCWDArg) boshreldir.ReleaseDir,
	ui boshui.UI,
) VendorPackageCmd {
	return VendorPackageCmd{releaseDirFactory, ui}
}

func (c VendorPackageCmd) Run(opts VendorPackageOpts) error {
	srcReleaseDir := c.releaseDirFactory(opts.Args.URL)
	dstReleaseDir := c.releaseDirFactory(opts.Directory)

	srcRelease, err := srcReleaseDir.FindRelease("", semver.Version{})
	if err != nil {
		return err
	}

	for _, pkg := range srcRelease.Packages() {
		if pkg.Name() == opts.Args.PackageName {
			return dstReleaseDir.VendorPackage(pkg)
		}
	}

	return fmt.Errorf("Expected to find package '%s'", opts.Args.PackageName)
}

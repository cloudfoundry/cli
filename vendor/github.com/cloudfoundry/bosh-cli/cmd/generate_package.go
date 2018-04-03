package cmd

import (
	boshreldir "github.com/cloudfoundry/bosh-cli/releasedir"
)

type GeneratePackageCmd struct {
	releaseDir boshreldir.ReleaseDir
}

func NewGeneratePackageCmd(releaseDir boshreldir.ReleaseDir) GeneratePackageCmd {
	return GeneratePackageCmd{releaseDir: releaseDir}
}

func (c GeneratePackageCmd) Run(opts GeneratePackageOpts) error {
	return c.releaseDir.GeneratePackage(opts.Args.Name)
}

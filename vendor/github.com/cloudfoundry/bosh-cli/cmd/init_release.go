package cmd

import (
	boshreldir "github.com/cloudfoundry/bosh-cli/releasedir"
)

type InitReleaseCmd struct {
	releaseDir boshreldir.ReleaseDir
}

func NewInitReleaseCmd(releaseDir boshreldir.ReleaseDir) InitReleaseCmd {
	return InitReleaseCmd{releaseDir: releaseDir}
}

func (c InitReleaseCmd) Run(opts InitReleaseOpts) error {
	return c.releaseDir.Init(opts.Git)
}

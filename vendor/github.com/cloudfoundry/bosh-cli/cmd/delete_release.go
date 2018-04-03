package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type DeleteReleaseCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewDeleteReleaseCmd(ui boshui.UI, director boshdir.Director) DeleteReleaseCmd {
	return DeleteReleaseCmd{ui: ui, director: director}
}

func (c DeleteReleaseCmd) Run(opts DeleteReleaseOpts) error {
	err := c.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	releaseSlug, ok := opts.Args.Slug.ReleaseSlug()
	if ok {
		release, err := c.director.FindRelease(releaseSlug)
		if err != nil {
			return err
		}

		return release.Delete(opts.Force)
	}

	releaseSeries, err := c.director.FindReleaseSeries(opts.Args.Slug.SeriesSlug())
	if err != nil {
		return err
	}

	return releaseSeries.Delete(opts.Force)
}

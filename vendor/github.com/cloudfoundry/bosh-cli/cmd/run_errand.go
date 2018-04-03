package cmd

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	"fmt"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
	biui "github.com/cloudfoundry/bosh-cli/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

type RunErrandCmd struct {
	deployment boshdir.Deployment
	downloader Downloader
	ui         biui.UI
}

func NewRunErrandCmd(
	deployment boshdir.Deployment,
	downloader Downloader,
	ui biui.UI,
) RunErrandCmd {
	return RunErrandCmd{deployment: deployment, downloader: downloader, ui: ui}
}

func (c RunErrandCmd) Run(opts RunErrandOpts) error {
	results, err := c.deployment.RunErrand(
		opts.Args.Name,
		opts.KeepAlive,
		opts.WhenChanged,
		opts.InstanceGroupOrInstanceSlugFlags.Slugs,
	)
	if err != nil {
		return err
	}

	errandErr := c.summarize(opts.Args.Name, results)
	for _, result := range results {

		if opts.DownloadLogs && len(result.LogsBlobstoreID) > 0 {
			err := c.downloader.Download(
				result.LogsBlobstoreID,
				result.LogsSHA1,
				opts.Args.Name,
				opts.LogsDirectory.Path,
			)
			if err != nil {
				return bosherr.WrapError(err, "Downloading errand logs")
			}
		}

	}

	return errandErr
}

func (c RunErrandCmd) summarize(errandName string, results []boshdir.ErrandResult) error {
	table := boshtbl.Table{
		Content: "errand(s)",

		Header: []boshtbl.Header{
			boshtbl.NewHeader("Instance"),
			boshtbl.NewHeader("Exit Code"),
			boshtbl.NewHeader("Stdout"),
			boshtbl.NewHeader("Stderr"),
		},

		SortBy: []boshtbl.ColumnSort{
			{Column: 0, Asc: true},
		},

		Notes: []string{},

		FillFirstColumn: true,

		Transpose: true,
	}

	var errandErr error
	for _, result := range results {
		instance := ""
		if result.InstanceGroup != "" {
			instance = boshdir.NewInstanceGroupOrInstanceSlug(result.InstanceGroup, result.InstanceID).String()
		}

		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.NewValueString(instance),
			boshtbl.NewValueInt(result.ExitCode),
			boshtbl.NewValueString(result.Stdout),
			boshtbl.NewValueString(result.Stderr),
		})

		prefix := fmt.Sprintf("Errand '%s'", errandName)
		suffix := fmt.Sprintf("(exit code %d)", result.ExitCode)

		switch {
		case result.ExitCode == 0:
		case result.ExitCode > 128:
			errandErr = bosherr.Errorf("%s was canceled %s", prefix, suffix)
		default:
			errandErr = bosherr.Errorf("%s completed with error %s", prefix, suffix)
		}
	}
	c.ui.PrintTable(table)

	return errandErr
}

func (c RunErrandCmd) printOutput(title, output string) {
	if len(output) > 0 {
		c.ui.PrintLinef("%s", title)
		c.ui.PrintLinef("%s", output)
	}
}

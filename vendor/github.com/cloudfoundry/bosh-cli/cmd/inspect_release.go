package cmd

import (
	"fmt"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

type InspectReleaseCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewInspectReleaseCmd(ui boshui.UI, director boshdir.Director) InspectReleaseCmd {
	return InspectReleaseCmd{ui: ui, director: director}
}

func (c InspectReleaseCmd) Run(opts InspectReleaseOpts) error {
	release, err := c.director.FindRelease(opts.Args.Slug)
	if err != nil {
		return err
	}

	jobsTable := boshtbl.Table{
		Content: "jobs",
		Header: []boshtbl.Header{
			boshtbl.NewHeader("Job"),
			boshtbl.NewHeader("Blobstore ID"),
			boshtbl.NewHeader("Digest"),
			boshtbl.NewHeader("Links Consumed"),
			boshtbl.NewHeader("Links Provided"),
		},
		SortBy: []boshtbl.ColumnSort{{Column: 0, Asc: true}},
	}

	jobs, err := release.Jobs()
	if err != nil {
		return err
	}

	for _, j := range jobs {
		jobsTable.Rows = append(jobsTable.Rows, []boshtbl.Value{
			boshtbl.NewValueString(fmt.Sprintf("%s/%s", j.Name, j.Fingerprint)),
			boshtbl.NewValueString(j.BlobstoreID),
			boshtbl.NewValueString(j.SHA1),
			boshtbl.NewValueInterface(j.LinksConsumed),
			boshtbl.NewValueInterface(j.LinksProvided),
		})
	}

	pkgsTable := boshtbl.Table{
		Content: "packages",
		Header: []boshtbl.Header{
			boshtbl.NewHeader("Package"),
			boshtbl.NewHeader("Compiled for"),
			boshtbl.NewHeader("Blobstore ID"),
			boshtbl.NewHeader("Digest"),
		},
		SortBy: []boshtbl.ColumnSort{{Column: 0, Asc: true}},
	}

	pkgs, err := release.Packages()
	if err != nil {
		return err
	}

	for _, p := range pkgs {
		section := boshtbl.Section{
			FirstColumn: boshtbl.NewValueString(fmt.Sprintf("%s/%s", p.Name, p.Fingerprint)),

			Rows: [][]boshtbl.Value{
				{
					boshtbl.NewValueString(""),
					boshtbl.NewValueString("(source)"),
					boshtbl.NewValueString(p.BlobstoreID),
					boshtbl.NewValueString(p.SHA1),
				},
			},
		}

		for _, cp := range p.CompiledPackages {
			section.Rows = append(section.Rows, []boshtbl.Value{
				boshtbl.NewValueString(""),
				boshtbl.NewValueString(cp.Stemcell.String()),
				boshtbl.NewValueString(cp.BlobstoreID),
				boshtbl.NewValueString(cp.SHA1),
			})
		}

		pkgsTable.Sections = append(pkgsTable.Sections, section)
	}

	c.ui.PrintTable(jobsTable)
	c.ui.PrintTable(pkgsTable)

	return nil
}

package cmd

import (
	"fmt"
	"sort"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

type InfoTable struct {
	Info boshdir.Info
	UI   boshui.UI
}

func (t InfoTable) Print() {
	table := boshtbl.Table{
		Header: []boshtbl.Header{
			boshtbl.NewHeader("Name"),
			boshtbl.NewHeader("UUID"),
			boshtbl.NewHeader("Version"),
		},
		Rows: [][]boshtbl.Value{
			{
				boshtbl.NewValueString(t.Info.Name),
				boshtbl.NewValueString(t.Info.UUID),
				boshtbl.NewValueString(t.Info.Version),
			},
		},
		Transpose: true,
	}

	if len(t.Info.CPI) > 0 {
		table = table.AddColumn("CPI", []boshtbl.Value{
			boshtbl.NewValueString(t.Info.CPI),
		})
	}

	if len(t.Info.Features) > 0 {
		desc := []string{}

		enabledText := map[bool]string{
			true:  "enabled",
			false: "disabled",
		}

		for name, enabled := range t.Info.Features {
			desc = append(desc, fmt.Sprintf("%s: %s", name, enabledText[enabled]))
		}

		sort.Sort(InfoFeatureSorting(desc))

		table = table.AddColumn("Features", []boshtbl.Value{
			boshtbl.NewValueStrings(desc),
		})
	}

	if len(t.Info.User) > 0 {
		table = table.AddColumn("User", []boshtbl.Value{
			boshtbl.NewValueString(t.Info.User),
		})
	} else {
		table = table.AddColumn("User", []boshtbl.Value{
			boshtbl.NewValueString("(not logged in)"),
		})
	}

	t.UI.PrintTable(table)
}

type InfoFeatureSorting []string

func (s InfoFeatureSorting) Len() int           { return len(s) }
func (s InfoFeatureSorting) Less(i, j int) bool { return s[i] < s[j] }
func (s InfoFeatureSorting) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

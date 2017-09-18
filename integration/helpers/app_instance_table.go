package helpers

import (
	"regexp"
	"strings"
)

type AppInstanceRow struct {
	Index  string
	State  string
	Since  string
	CPU    string
	Memory string
	Disk   string
}

type AppProcessTable struct {
	Title     string
	Instances []AppInstanceRow
}

type AppTable struct {
	Processes []AppProcessTable
}

func ParseV3AppProcessTable(input []byte) AppTable {
	appTable := AppTable{}

	rows := strings.Split(string(input), "\n")
	foundFirstProcess := false
	for _, row := range rows {
		if !foundFirstProcess {
			ok, err := regexp.MatchString(`\A([^:]+):\d/\d\z`, row)
			if err != nil {
				panic(err)
			}
			if ok {
				foundFirstProcess = true
			} else {
				continue
			}
		}

		if row == "" {
			continue
		}

		if strings.HasPrefix(row, "#") {
			// instance row
			columns := splitColumns(row)
			instanceRow := AppInstanceRow{
				Index:  columns[0],
				State:  columns[1],
				Since:  columns[2],
				CPU:    columns[3],
				Memory: columns[4],
				Disk:   columns[5],
			}
			lastProcessIndex := len(appTable.Processes) - 1
			appTable.Processes[lastProcessIndex].Instances = append(
				appTable.Processes[lastProcessIndex].Instances,
				instanceRow,
			)

		} else if !strings.HasPrefix(row, " ") {
			// process title
			appTable.Processes = append(appTable.Processes, AppProcessTable{
				Title: row,
			})
		} else {
			// column headers
			continue
		}

	}

	return appTable
}

func splitColumns(row string) []string {
	// uses 3 spaces between columns
	return regexp.MustCompile(`\s{3,}`).Split(strings.TrimSpace(row), -1)
}

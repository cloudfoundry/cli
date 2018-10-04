package helpers

import (
	"regexp"
	"strings"
)

type AppInstanceRow struct {
	Index   string
	State   string
	Since   string
	CPU     string
	Memory  string
	Disk    string
	Details string
}

type AppProcessTable struct {
	Type          string
	InstanceCount string
	MemUsage      string
	Instances     []AppInstanceRow
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
			ok, err := regexp.MatchString(`\Atype:([^:]+)\z`, row)
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
			details := ""
			if len(columns) >= 7 {
				details = columns[6]
			}

			instanceRow := AppInstanceRow{
				Index:   columns[0],
				State:   columns[1],
				Since:   columns[2],
				CPU:     columns[3],
				Memory:  columns[4],
				Disk:    columns[5],
				Details: details,
			}
			lastProcessIndex := len(appTable.Processes) - 1
			appTable.Processes[lastProcessIndex].Instances = append(
				appTable.Processes[lastProcessIndex].Instances,
				instanceRow,
			)

		} else if strings.HasPrefix(row, "type:") {
			appTable.Processes = append(appTable.Processes, AppProcessTable{
				Type: strings.TrimSpace(strings.TrimPrefix(row, "type:")),
			})
		} else if strings.HasPrefix(row, "instances:") {
			lpi := len(appTable.Processes) - 1
			iVal := strings.TrimSpace(strings.TrimPrefix(row, "instances:"))
			appTable.Processes[lpi].InstanceCount = iVal
		} else if strings.HasPrefix(row, "memory usage:") {
			lpi := len(appTable.Processes) - 1
			mVal := strings.TrimSpace(strings.TrimPrefix(row, "memory usage:"))
			appTable.Processes[lpi].MemUsage = mVal
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

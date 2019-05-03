package helpers

import (
	"regexp"
	"strings"
)

// AppInstanceRow represents an instance of a V3 app's process,
// as displayed in the 'cf app' output.
type AppInstanceRow struct {
	Index   string
	State   string
	Since   string
	CPU     string
	Memory  string
	Disk    string
	Details string
}

// AppProcessTable represents a process of a V3 app, as displayed in the 'cf
// app' output.
type AppProcessTable struct {
	Type          string
	InstanceCount string
	MemUsage      string
	Instances     []AppInstanceRow
}

// AppTable represents a V3 app as a collection of processes,
// as displayed in the 'cf app' output.
type AppTable struct {
	Processes []AppProcessTable
}

// ParseV3AppProcessTable parses bytes from 'cf app' stdout into an AppTable.
func ParseV3AppProcessTable(input []byte) AppTable {
	appTable := AppTable{}

	rows := strings.Split(string(input), "\n")
	foundFirstProcess := false
	for _, row := range rows {
		if !foundFirstProcess {
			ok := regexp.MustCompile(`\Atype:([^:]+)\z`).Match([]byte(row))
			if ok {
				foundFirstProcess = true
			} else {
				continue
			}
		}

		if row == "" {
			continue
		}

		switch {
		case strings.HasPrefix(row, "#"):
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

		case strings.HasPrefix(row, "type:"):
			appTable.Processes = append(appTable.Processes, AppProcessTable{
				Type: strings.TrimSpace(strings.TrimPrefix(row, "type:")),
			})

		case strings.HasPrefix(row, "instances:"):
			lpi := len(appTable.Processes) - 1
			iVal := strings.TrimSpace(strings.TrimPrefix(row, "instances:"))
			appTable.Processes[lpi].InstanceCount = iVal

		case strings.HasPrefix(row, "memory usage:"):
			lpi := len(appTable.Processes) - 1
			mVal := strings.TrimSpace(strings.TrimPrefix(row, "memory usage:"))
			appTable.Processes[lpi].MemUsage = mVal

		default:
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

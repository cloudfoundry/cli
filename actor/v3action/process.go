package v3action

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

// Process represents a V3 actor process.
type Process struct {
	Type       string
	Instances  []Instance
	MemoryInMB int
	DiskInMB   int
}

// Instance represents a V3 actor instance.
type Instance ccv3.Instance

// StartTime returns the time that the instance started.
func (instance *Instance) StartTime() time.Time {
	uptimeDuration := time.Duration(instance.Uptime) * time.Second

	return time.Now().Add(-uptimeDuration)
}

func (p Process) TotalInstanceCount() int {
	return len(p.Instances)
}

func (p Process) HealthyInstanceCount() int {
	count := 0
	for _, instance := range p.Instances {
		if instance.State == "RUNNING" {
			count++
		}
	}
	return count
}

type Processes []Process

func (ps Processes) Sort() {
	sort.Slice(ps, func(i int, j int) bool {
		var iScore int
		var jScore int

		switch ps[i].Type {
		case "web":
			iScore = 0
		default:
			iScore = 1
		}

		switch ps[j].Type {
		case "web":
			jScore = 0
		default:
			jScore = 1
		}

		if iScore == 1 && jScore == 1 {
			return ps[i].Type < ps[j].Type
		}
		return iScore < jScore
	})

}

func (ps Processes) Summary() string {
	ps.Sort()

	var summaries []string
	for _, p := range ps {
		summaries = append(summaries, fmt.Sprintf("%s:%d/%d", p.Type, p.HealthyInstanceCount(), p.TotalInstanceCount()))
	}

	return strings.Join(summaries, ", ")
}

func (actor Actor) GetProcessByApplication(appGUID string) (ccv3.Process, Warnings, error) {
	return ccv3.Process{}, nil, nil
}

func (actor Actor) ScaleProcessByApplication(appGUID string, process ccv3.Process) (ccv3.Process, Warnings, error) {
	return ccv3.Process{}, nil, nil
}

package director

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type VMInfo struct {
	AgentID string `json:"agent_id"`

	JobName      string `json:"job_name"`
	ID           string `json:"id"`
	Index        *int   `json:"index"`
	ProcessState string `json:"job_state"` // e.g. "running"
	Active       *bool  `json:"active"`
	Bootstrap    bool

	IPs []string `json:"ips"`
	DNS []string `json:"dns"`

	AZ              string      `json:"az"`
	State           string      `json:"state"`
	VMID            string      `json:"vm_cid"`
	VMType          string      `json:"vm_type"`
	ResourcePool    string      `json:"resource_pool"`
	DiskID          string      `json:"disk_cid"`
	Ignore          bool        `json:"ignore"`
	DiskIDs         []string    `json:"disk_cids"`
	VMCreatedAtRaw  string      `json:"vm_created_at"`
	VMCreatedAt     time.Time   `json:"-"`
	CloudProperties interface{} `json:"cloud_properties"`

	Processes []VMInfoProcess

	Vitals VMInfoVitals

	ResurrectionPaused bool `json:"resurrection_paused"`
}

type VMInfoProcess struct {
	Name  string
	State string // e.g. "running"

	CPU    VMInfoVitalsCPU `json:"cpu"`
	Mem    VMInfoVitalsMemIntSize
	Uptime VMInfoVitalsUptime
}

type VMInfoVitals struct {
	CPU    VMInfoVitalsCPU `json:"cpu"`
	Mem    VMInfoVitalsMemSize
	Swap   VMInfoVitalsMemSize
	Uptime VMInfoVitalsUptime

	Load []string
	Disk map[string]VMInfoVitalsDiskSize
}

func (v VMInfoVitals) SystemDisk() VMInfoVitalsDiskSize     { return v.Disk["system"] }
func (v VMInfoVitals) EphemeralDisk() VMInfoVitalsDiskSize  { return v.Disk["ephemeral"] }
func (v VMInfoVitals) PersistentDisk() VMInfoVitalsDiskSize { return v.Disk["persistent"] }

type VMInfoVitalsCPU struct {
	Total *float64 // used by VMInfoProcess
	Sys   string
	User  string
	Wait  string
}

type VMInfoVitalsDiskSize struct {
	InodePercent string `json:"inode_percent"`
	Percent      string
}

type VMInfoVitalsMemSize struct {
	KB      string `json:"kb"`
	Percent string
}

type VMInfoVitalsMemIntSize struct {
	KB      *uint64 `json:"kb"`
	Percent *float64
}

type VMInfoVitalsUptime struct {
	Seconds *uint64 `json:"secs"` // e.g. 48307
}

func (i VMInfo) IsRunning() bool {
	if i.ProcessState != "running" {
		return false
	}

	for _, p := range i.Processes {
		if !p.IsRunning() {
			return false
		}
	}

	return true
}

func (p VMInfoProcess) IsRunning() bool {
	return p.State == "running"
}

func (d DeploymentImpl) VMInfos() ([]VMInfo, error) {
	infos, err := d.client.DeploymentVMInfos(d.name)
	if err != nil {
		return nil, err
	}

	return infos, nil
}

func (c Client) DeploymentVMInfos(deploymentName string) ([]VMInfo, error) {
	return c.deploymentResourceInfos(deploymentName, "vms")
}

func (c Client) deploymentResourceInfos(deploymentName string, resourceType string) ([]VMInfo, error) {
	if len(deploymentName) == 0 {
		return nil, bosherr.Error("Expected non-empty deployment name")
	}

	path := fmt.Sprintf("/deployments/%s/%s?format=full", deploymentName, resourceType)

	_, resultBytes, err := c.taskClientRequest.GetResult(path)
	if err != nil {
		return nil, bosherr.WrapErrorf(
			err, "Listing deployment '%s' %s infos", deploymentName, resourceType)
	}

	var resps []VMInfo

	for _, piece := range strings.Split(string(resultBytes), "\n") {
		if len(piece) == 0 {
			continue
		}

		var resp VMInfo

		err := json.Unmarshal([]byte(piece), &resp)
		if err != nil {
			return nil, bosherr.WrapErrorf(
				err, "Unmarshaling %s info response: '%s'", strings.TrimSuffix(resourceType, "s"), string(piece))
		}

		if len(resp.DiskIDs) == 0 && resp.DiskID != "" {
			resp.DiskIDs = []string{resp.DiskID}
		}

		resp.VMCreatedAt, err = TimeParser{}.Parse(resp.VMCreatedAtRaw)
		if err != nil {
			return resps, bosherr.WrapErrorf(err, "Converting created_at '%s' to time", resp.VMCreatedAtRaw)
		}

		resps = append(resps, resp)
	}

	return resps, nil
}

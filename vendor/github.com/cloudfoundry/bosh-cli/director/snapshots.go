package director

import (
	"fmt"
	"strconv"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type Snapshot struct {
	Job   string
	Index *int

	CID       string
	CreatedAt time.Time

	Clean bool
}

type SnapshotResp struct {
	Job   string
	Index *int

	SnapshotCID string `json:"snapshot_cid"`
	CreatedAt   string `json:"created_at"`

	Clean bool
}

func (s Snapshot) InstanceDesc() string {
	jobDesc := "?"
	indexDesc := "?"

	if len(s.Job) > 0 {
		jobDesc = s.Job
	}

	if s.Index != nil {
		indexDesc = strconv.Itoa(*s.Index)
	}

	return fmt.Sprintf("%s/%s", jobDesc, indexDesc)
}

func (d DeploymentImpl) Snapshots() ([]Snapshot, error) {
	var snaps []Snapshot

	resps, err := d.client.Snapshots(d.name)
	if err != nil {
		return snaps, err
	}

	for _, r := range resps {
		createdAt, err := TimeParser{}.Parse(r.CreatedAt)
		if err != nil {
			return snaps, bosherr.WrapErrorf(err, "Converting created at '%s' to time", r.CreatedAt)
		}

		snap := Snapshot{
			Job:   r.Job,
			Index: r.Index,

			CID:       r.SnapshotCID,
			CreatedAt: createdAt.UTC(),

			Clean: r.Clean,
		}

		snaps = append(snaps, snap)
	}

	return snaps, nil
}

func (d DeploymentImpl) TakeSnapshot(slug InstanceSlug) error {
	return d.client.TakeSnapshot(d.name, slug.Name(), slug.IndexOrID())
}

func (d DeploymentImpl) DeleteSnapshot(cid string) error {
	err := d.client.DeleteSnapshot(d.name, cid)
	if err != nil {
		resps, listErr := d.client.Snapshots(d.name)
		if listErr != nil {
			return err
		}

		for _, resp := range resps {
			if resp.SnapshotCID == cid {
				return err
			}
		}
	}

	return nil
}

func (d DeploymentImpl) TakeSnapshots() error {
	return d.client.TakeSnapshots(d.name)
}

func (d DeploymentImpl) DeleteSnapshots() error {
	return d.client.DeleteSnapshots(d.name)
}

func (c Client) Snapshots(deploymentName string) ([]SnapshotResp, error) {
	var snaps []SnapshotResp

	if len(deploymentName) == 0 {
		return snaps, bosherr.Error("Expected non-empty deployment name")
	}

	path := fmt.Sprintf("/deployments/%s/snapshots", deploymentName)

	// [{"job":"first","index":0,"snapshot_cid":"snap-7b1f0b00",
	//   "created_at":"2015-10-03 18:02:09 +0000","clean":false}]
	err := c.clientRequest.Get(path, &snaps)
	if err != nil {
		return snaps, bosherr.WrapErrorf(err, "Finding snapshots")
	}

	return snaps, nil
}

func (c Client) TakeSnapshot(deploymentName, job, indexOrID string) error {
	if len(deploymentName) == 0 {
		return bosherr.Error("Expected non-empty deployment name")
	}

	if len(job) == 0 {
		return bosherr.Error("Expected non-empty job name")
	}

	if len(indexOrID) == 0 {
		return bosherr.Error("Expected non-empty index or ID")
	}

	path := fmt.Sprintf("/deployments/%s/jobs/%s/%s/snapshots",
		deploymentName, job, indexOrID)

	_, err := c.taskClientRequest.PostResult(path, nil, nil)
	if err != nil {
		return bosherr.WrapErrorf(
			err, "Taking snapshot for instance '%s/%s' in deployment '%s'",
			job, indexOrID, deploymentName)
	}

	return nil
}

func (c Client) DeleteSnapshot(deploymentName, cid string) error {
	if len(deploymentName) == 0 {
		return bosherr.Error("Expected non-empty deployment name")
	}

	if len(cid) == 0 {
		return bosherr.Error("Expected non-empty snapshot CID")
	}

	path := fmt.Sprintf("/deployments/%s/snapshots/%s", deploymentName, cid)

	_, err := c.taskClientRequest.DeleteResult(path)
	if err != nil {
		return bosherr.WrapErrorf(
			err, "Deleting snapshot '%s' from deployment '%s'", cid, deploymentName)
	}

	return nil
}

func (c Client) TakeSnapshots(deploymentName string) error {
	if len(deploymentName) == 0 {
		return bosherr.Error("Expected non-empty deployment name")
	}

	path := fmt.Sprintf("/deployments/%s/snapshots", deploymentName)

	_, err := c.taskClientRequest.PostResult(path, nil, nil)
	if err != nil {
		return bosherr.WrapErrorf(
			err, "Taking snapshots for deployment '%s'", deploymentName)
	}

	return nil
}

func (c Client) DeleteSnapshots(deploymentName string) error {
	if len(deploymentName) == 0 {
		return bosherr.Error("Expected non-empty deployment name")
	}

	path := fmt.Sprintf("/deployments/%s/snapshots", deploymentName)

	_, err := c.taskClientRequest.DeleteResult(path)
	if err != nil {
		return bosherr.WrapErrorf(
			err, "Deleting snapshots for deployment '%s'", deploymentName)
	}

	return nil
}

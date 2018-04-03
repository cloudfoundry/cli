package director

import (
	"fmt"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type OrphanDiskImpl struct {
	client Client

	cid  string
	size uint64

	deploymentName string
	instanceName   string
	azName         string

	orphanedAt time.Time
}

func (d OrphanDiskImpl) CID() string  { return d.cid }
func (d OrphanDiskImpl) Size() uint64 { return d.size }

func (d OrphanDiskImpl) Deployment() Deployment {
	return &DeploymentImpl{client: d.client, name: d.deploymentName}
}

func (d OrphanDiskImpl) InstanceName() string { return d.instanceName }
func (d OrphanDiskImpl) AZName() string       { return d.azName }

func (d OrphanDiskImpl) OrphanedAt() time.Time { return d.orphanedAt }

func (d OrphanDiskImpl) Delete() error {
	err := d.client.DeleteOrphanDisk(d.cid)
	if err != nil {
		resps, listErr := d.client.OrphanDisks()
		if listErr != nil {
			return err
		}

		for _, resp := range resps {
			if resp.CID == d.cid {
				return err
			}
		}
	}

	return nil
}

func (d DirectorImpl) OrphanDisk(cid string) error {
	return d.client.OrphanDisk(cid)
}

type OrphanDiskResp struct {
	CID  string `json:"disk_cid"`
	Size uint64

	DeploymentName string `json:"deployment_name"`
	InstanceName   string `json:"instance_name"`
	AZ             string `json:"az"`

	OrphanedAt string `json:"orphaned_at"` // e.g. "2016-01-09 06:23:25 +0000"
}

func (d DirectorImpl) FindOrphanDisk(cid string) (OrphanDisk, error) {
	return OrphanDiskImpl{client: d.client, cid: cid}, nil
}

func (d DirectorImpl) OrphanDisks() ([]OrphanDisk, error) {
	var disks []OrphanDisk

	resps, err := d.client.OrphanDisks()
	if err != nil {
		return disks, err
	}

	for _, r := range resps {
		orphanedAt, err := TimeParser{}.Parse(r.OrphanedAt)
		if err != nil {
			return disks, bosherr.WrapErrorf(err, "Converting orphaned at '%s' to time", r.OrphanedAt)
		}

		disk := OrphanDiskImpl{
			client: d.client,

			cid:  r.CID,
			size: r.Size,

			deploymentName: r.DeploymentName,
			instanceName:   r.InstanceName,
			azName:         r.AZ,

			orphanedAt: orphanedAt.UTC(),
		}

		disks = append(disks, disk)
	}

	return disks, nil
}

func (c Client) OrphanDisks() ([]OrphanDiskResp, error) {
	var disks []OrphanDiskResp

	err := c.clientRequest.Get("/disks", &disks)
	if err != nil {
		return disks, bosherr.WrapErrorf(err, "Finding orphaned disks")
	}

	return disks, nil
}

func (c Client) DeleteOrphanDisk(cid string) error {
	if len(cid) == 0 {
		return bosherr.Error("Expected non-empty orphaned disk CID")
	}

	path := fmt.Sprintf("/disks/%s", cid)

	_, err := c.taskClientRequest.DeleteResult(path)
	if err != nil {
		return bosherr.WrapErrorf(err, "Deleting orphaned disk '%s'", cid)
	}

	return nil
}

func (c Client) OrphanDisk(cid string) error {
	if len(cid) == 0 {
		return bosherr.Error("Expected non-empty disk CID")
	}

	path := fmt.Sprintf("/disks/%s?orphan=true", cid)

	_, err := c.taskClientRequest.DeleteResult(path)
	if err != nil {
		return bosherr.WrapErrorf(err, "Orphaning disk '%s'", cid)
	}

	return nil
}

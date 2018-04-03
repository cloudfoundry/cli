package disk

import (
	"reflect"

	bicloud "github.com/cloudfoundry/bosh-cli/cloud"
	biconfig "github.com/cloudfoundry/bosh-cli/config"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
)

type Disk interface {
	CID() string
	NeedsMigration(newSize int, newCloudProperties biproperty.Map) bool
	Delete() error
}

type disk struct {
	cid             string
	size            int
	cloudProperties biproperty.Map

	cloud bicloud.Cloud
	repo  biconfig.DiskRepo
}

func NewDisk(
	diskRecord biconfig.DiskRecord,
	cloud bicloud.Cloud,
	repo biconfig.DiskRepo,
) Disk {
	return &disk{
		cid:             diskRecord.CID,
		size:            diskRecord.Size,
		cloudProperties: diskRecord.CloudProperties,
		cloud:           cloud,
		repo:            repo,
	}
}

func (d *disk) CID() string {
	return d.cid
}

func (d *disk) NeedsMigration(newSize int, newCloudProperties biproperty.Map) bool {
	return d.size != newSize || !reflect.DeepEqual(d.cloudProperties, newCloudProperties)
}

func (d *disk) Delete() error {
	deleteErr := d.cloud.DeleteDisk(d.cid)
	if deleteErr != nil {
		// allow DiskNotFoundError for idempotency
		cloudErr, ok := deleteErr.(bicloud.Error)
		if !ok || cloudErr.Type() != bicloud.DiskNotFoundError {
			return bosherr.WrapError(deleteErr, "Deleting disk in the cloud")
		}
	}

	diskRecord, found, err := d.repo.Find(d.cid)
	if err != nil {
		return bosherr.WrapErrorf(err, "Finding disk record (cid=%s)", d.cid)
	}

	if !found {
		return nil
	}

	err = d.repo.Delete(diskRecord)
	if err != nil {
		return bosherr.WrapError(err, "Deleting disk record")
	}

	// returns bicloud.Error only if it is a DiskNotFoundError
	return deleteErr
}

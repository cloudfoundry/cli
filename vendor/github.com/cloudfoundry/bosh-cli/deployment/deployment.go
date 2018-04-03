package deployment

import (
	"fmt"
	"time"

	bicloud "github.com/cloudfoundry/bosh-cli/cloud"
	bidisk "github.com/cloudfoundry/bosh-cli/deployment/disk"
	biinstance "github.com/cloudfoundry/bosh-cli/deployment/instance"
	bistemcell "github.com/cloudfoundry/bosh-cli/stemcell"
	biui "github.com/cloudfoundry/bosh-cli/ui"
)

type Deployment interface {
	Delete(biui.Stage) error
}

type deployment struct {
	instances   []biinstance.Instance
	disks       []bidisk.Disk
	stemcells   []bistemcell.CloudStemcell
	pingTimeout time.Duration
	pingDelay   time.Duration
}

func NewDeployment(
	instances []biinstance.Instance,
	disks []bidisk.Disk,
	stemcells []bistemcell.CloudStemcell,
	pingTimeout time.Duration,
	pingDelay time.Duration,
) Deployment {
	return &deployment{
		instances:   instances,
		disks:       disks,
		stemcells:   stemcells,
		pingTimeout: pingTimeout,
		pingDelay:   pingDelay,
	}
}

func (d *deployment) Delete(deleteStage biui.Stage) error {
	// le sigh... consuming from an array sucks without generics
	for len(d.instances) > 0 {
		lastIdx := len(d.instances) - 1
		instance := d.instances[lastIdx]

		if err := instance.Delete(d.pingTimeout, d.pingDelay, deleteStage); err != nil {
			return err
		}

		d.instances = d.instances[:lastIdx]
	}

	for len(d.disks) > 0 {
		lastIdx := len(d.disks) - 1
		disk := d.disks[lastIdx]

		if err := d.deleteDisk(deleteStage, disk); err != nil {
			return err
		}

		d.disks = d.disks[:lastIdx]
	}

	for len(d.stemcells) > 0 {
		lastIdx := len(d.stemcells) - 1
		stemcell := d.stemcells[lastIdx]

		if err := d.deleteStemcell(deleteStage, stemcell); err != nil {
			return err
		}

		d.stemcells = d.stemcells[:lastIdx]
	}

	return nil
}

func (d *deployment) deleteDisk(deleteStage biui.Stage, disk bidisk.Disk) error {
	stepName := fmt.Sprintf("Deleting disk '%s'", disk.CID())
	return deleteStage.Perform(stepName, func() error {
		err := disk.Delete()
		cloudErr, ok := err.(bicloud.Error)
		if ok && cloudErr.Type() == bicloud.DiskNotFoundError {
			return biui.NewSkipStageError(cloudErr, "Disk not found")
		}
		return err
	})
}

func (d *deployment) deleteStemcell(deleteStage biui.Stage, stemcell bistemcell.CloudStemcell) error {
	stepName := fmt.Sprintf("Deleting stemcell '%s'", stemcell.CID())
	return deleteStage.Perform(stepName, func() error {
		err := stemcell.Delete()
		cloudErr, ok := err.(bicloud.Error)
		if ok && cloudErr.Type() == bicloud.StemcellNotFoundError {
			return biui.NewSkipStageError(cloudErr, "Stemcell not found")
		}
		return err
	})
}

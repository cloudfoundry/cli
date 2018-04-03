package vm

import (
	"fmt"

	bicloud "github.com/cloudfoundry/bosh-cli/cloud"
	biconfig "github.com/cloudfoundry/bosh-cli/config"
	bidisk "github.com/cloudfoundry/bosh-cli/deployment/disk"
	bideplmanifest "github.com/cloudfoundry/bosh-cli/deployment/manifest"
	biui "github.com/cloudfoundry/bosh-cli/ui"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

// DiskDeployer is in the vm package to avoid a [disk -> vm -> disk] dependency cycle
type DiskDeployer interface {
	Deploy(diskPool bideplmanifest.DiskPool, cloud bicloud.Cloud, vm VM, eventLoggerStage biui.Stage) ([]bidisk.Disk, error)
}

type diskDeployer struct {
	diskRepo           biconfig.DiskRepo
	diskManagerFactory bidisk.ManagerFactory
	diskManager        bidisk.Manager
	logger             boshlog.Logger
	logTag             string
}

func NewDiskDeployer(diskManagerFactory bidisk.ManagerFactory, diskRepo biconfig.DiskRepo, logger boshlog.Logger) DiskDeployer {
	return &diskDeployer{
		diskManagerFactory: diskManagerFactory,
		diskRepo:           diskRepo,
		logger:             logger,
		logTag:             "diskDeployer",
	}
}

func (d *diskDeployer) Deploy(diskPool bideplmanifest.DiskPool, cloud bicloud.Cloud, vm VM, stage biui.Stage) ([]bidisk.Disk, error) {
	if diskPool.DiskSize == 0 {
		return []bidisk.Disk{}, nil
	}

	d.diskManager = d.diskManagerFactory.NewManager(cloud)
	disks, err := d.diskManager.FindCurrent()
	if err != nil {
		return disks, bosherr.WrapError(err, "Finding existing disk")
	}

	if len(disks) > 1 {
		return disks, bosherr.WrapError(err, "Multiple current disks not supported")

	} else if len(disks) == 1 {
		disks, err = d.deployExistingDisk(disks[0], diskPool, vm, stage)
		if err != nil {
			return disks, err
		}

	} else {
		disks, err = d.deployNewDisk(diskPool, vm, stage)
		if err != nil {
			return disks, err
		}
	}

	err = d.diskManager.DeleteUnused(stage)
	if err != nil {
		return disks, err
	}

	return disks, nil
}

func (d *diskDeployer) deployExistingDisk(disk bidisk.Disk, diskPool bideplmanifest.DiskPool, vm VM, stage biui.Stage) ([]bidisk.Disk, error) {
	disks := []bidisk.Disk{}

	// the disk is already part of the deployment, and should already be attached
	disks = append(disks, disk)

	// attach is idempotent
	err := d.attachDisk(disk, vm, stage)
	if err != nil {
		return disks, err
	}

	if disk.NeedsMigration(diskPool.DiskSize, diskPool.CloudProperties) {
		disk, err = d.migrateDisk(disk, diskPool, vm, stage)
		if err != nil {
			return disks, err
		}

		// after migration, only the new disk is part of the deployment
		disks[0] = disk
	}

	return disks, nil
}

func (d *diskDeployer) deployNewDisk(diskPool bideplmanifest.DiskPool, vm VM, stage biui.Stage) ([]bidisk.Disk, error) {
	disks := []bidisk.Disk{}

	disk, err := d.createDisk(diskPool, vm, stage)
	if err != nil {
		return disks, err
	}

	err = d.attachDisk(disk, vm, stage)
	if err != nil {
		return disks, err
	}

	// once attached, the disk is part of the deployment
	disks = append(disks, disk)

	err = d.updateCurrentDiskRecord(disk)
	if err != nil {
		return disks, err
	}

	return disks, nil
}

func (d *diskDeployer) migrateDisk(
	originalDisk bidisk.Disk,
	diskPool bideplmanifest.DiskPool,
	vm VM,
	stage biui.Stage,
) (newDisk bidisk.Disk, err error) {
	d.logger.Debug(d.logTag, "Migrating disk '%s'", originalDisk.CID())

	err = stage.Perform("Creating disk", func() error {
		newDisk, err = d.diskManager.Create(diskPool, vm.CID())
		return err
	})
	if err != nil {
		return newDisk, err
	}

	stageName := fmt.Sprintf("Attaching disk '%s' to VM '%s'", newDisk.CID(), vm.CID())
	err = stage.Perform(stageName, func() error {
		return vm.AttachDisk(newDisk)
	})
	if err != nil {
		return newDisk, err
	}

	stageName = fmt.Sprintf("Migrating disk content from '%s' to '%s'", originalDisk.CID(), newDisk.CID())
	err = stage.Perform(stageName, func() error {
		return vm.MigrateDisk()
	})
	if err != nil {
		return newDisk, err
	}

	err = d.updateCurrentDiskRecord(newDisk)
	if err != nil {
		return newDisk, err
	}

	stageName = fmt.Sprintf("Detaching disk '%s'", originalDisk.CID())
	err = stage.Perform(stageName, func() error {
		return vm.DetachDisk(originalDisk)
	})
	if err != nil {
		return newDisk, err
	}

	stageName = fmt.Sprintf("Deleting disk '%s'", originalDisk.CID())
	err = stage.Perform(stageName, func() error {
		return originalDisk.Delete()
	})
	if err != nil {
		return newDisk, err
	}

	return newDisk, nil
}

func (d *diskDeployer) updateCurrentDiskRecord(disk bidisk.Disk) error {
	savedDiskRecord, found, err := d.diskRepo.Find(disk.CID())
	if err != nil {
		return bosherr.WrapError(err, "Finding disk record")
	}

	if !found {
		return bosherr.Error("Failed to find disk record for new disk")
	}

	err = d.diskRepo.UpdateCurrent(savedDiskRecord.ID)
	if err != nil {
		return bosherr.WrapError(err, "Updating current disk record")
	}

	return nil
}

func (d *diskDeployer) createDisk(diskPool bideplmanifest.DiskPool, vm VM, stage biui.Stage) (disk bidisk.Disk, err error) {
	err = stage.Perform("Creating disk", func() error {
		disk, err = d.diskManager.Create(diskPool, vm.CID())
		return err
	})

	return disk, err
}

func (d *diskDeployer) attachDisk(disk bidisk.Disk, vm VM, stage biui.Stage) error {
	stageName := fmt.Sprintf("Attaching disk '%s' to VM '%s'", disk.CID(), vm.CID())
	err := stage.Perform(stageName, func() error {
		return vm.AttachDisk(disk)
	})

	return err
}

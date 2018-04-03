package config

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
)

type DiskRepo interface {
	UpdateCurrent(diskID string) error
	FindCurrent() (DiskRecord, bool, error)
	ClearCurrent() error
	Save(cid string, size int, cloudProperties biproperty.Map) (DiskRecord, error)
	Find(cid string) (DiskRecord, bool, error)
	All() ([]DiskRecord, error)
	Delete(DiskRecord) error
}

type diskRepo struct {
	deploymentStateService DeploymentStateService
	uuidGenerator          boshuuid.Generator
}

func NewDiskRepo(deploymentStateService DeploymentStateService, uuidGenerator boshuuid.Generator) DiskRepo {
	return diskRepo{
		deploymentStateService: deploymentStateService,
		uuidGenerator:          uuidGenerator,
	}
}

func (r diskRepo) Save(cid string, size int, cloudProperties biproperty.Map) (DiskRecord, error) {
	config, records, err := r.load()
	if err != nil {
		return DiskRecord{}, err
	}

	oldRecord, found := r.find(records, cid)
	if found {
		return DiskRecord{}, bosherr.Errorf("Failed to save disk cid '%s', existing record found '%#v'", cid, oldRecord)
	}

	newRecord := DiskRecord{
		CID:             cid,
		Size:            size,
		CloudProperties: cloudProperties,
	}
	newRecord.ID, err = r.uuidGenerator.Generate()
	if err != nil {
		return newRecord, bosherr.WrapError(err, "Generating disk id")
	}

	records = append(records, newRecord)
	config.Disks = records

	err = r.deploymentStateService.Save(config)
	if err != nil {
		return newRecord, bosherr.WrapError(err, "Saving new config")
	}
	return newRecord, nil
}

func (r diskRepo) FindCurrent() (DiskRecord, bool, error) {
	deploymentState, err := r.deploymentStateService.Load()
	if err != nil {
		return DiskRecord{}, false, bosherr.WrapError(err, "Loading existing config")
	}

	currentDiskID := deploymentState.CurrentDiskID
	if currentDiskID == "" {
		return DiskRecord{}, false, nil
	}

	for _, oldRecord := range deploymentState.Disks {
		if oldRecord.ID == currentDiskID {
			return oldRecord, true, nil
		}
	}

	return DiskRecord{}, false, nil
}

func (r diskRepo) UpdateCurrent(diskID string) error {
	deploymentState, err := r.deploymentStateService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Loading existing config")
	}

	found := false
	for _, oldRecord := range deploymentState.Disks {
		if oldRecord.ID == diskID {
			found = true
		}
	}
	if !found {
		return bosherr.Errorf("Verifying disk record exists with id '%s'", diskID)
	}

	deploymentState.CurrentDiskID = diskID

	err = r.deploymentStateService.Save(deploymentState)
	if err != nil {
		return bosherr.WrapError(err, "Saving new config")
	}
	return nil
}

func (r diskRepo) Find(cid string) (DiskRecord, bool, error) {
	_, records, err := r.load()
	if err != nil {
		return DiskRecord{}, false, err
	}

	foundRecord, found := r.find(records, cid)
	return foundRecord, found, nil
}

func (r diskRepo) All() ([]DiskRecord, error) {
	deploymentState, err := r.deploymentStateService.Load()
	if err != nil {
		return []DiskRecord{}, bosherr.WrapError(err, "Loading existing config")
	}

	return deploymentState.Disks, nil
}

func (r diskRepo) Delete(diskRecord DiskRecord) error {
	config, records, err := r.load()
	if err != nil {
		return err
	}

	newRecords := []DiskRecord{}
	for _, record := range records {
		if record.ID != diskRecord.ID {
			newRecords = append(newRecords, record)
		}
	}

	config.Disks = newRecords

	if config.CurrentDiskID == diskRecord.ID {
		config.CurrentDiskID = ""
	}

	err = r.deploymentStateService.Save(config)
	if err != nil {
		return bosherr.WrapError(err, "Saving new config")
	}

	return nil
}

func (r diskRepo) ClearCurrent() error {
	deploymentState, err := r.deploymentStateService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Loading existing config")
	}

	deploymentState.CurrentDiskID = ""

	err = r.deploymentStateService.Save(deploymentState)
	if err != nil {
		return bosherr.WrapError(err, "Saving new config")
	}
	return nil
}

func (r diskRepo) load() (DeploymentState, []DiskRecord, error) {
	deploymentState, err := r.deploymentStateService.Load()
	if err != nil {
		return deploymentState, []DiskRecord{}, bosherr.WrapError(err, "Loading existing config")
	}

	records := deploymentState.Disks
	if records == nil {
		return deploymentState, []DiskRecord{}, nil
	}

	return deploymentState, records, nil
}

func (r diskRepo) find(records []DiskRecord, cid string) (DiskRecord, bool) {
	for _, existingRecord := range records {
		if existingRecord.CID == cid {
			return existingRecord, true
		}
	}
	return DiskRecord{}, false
}

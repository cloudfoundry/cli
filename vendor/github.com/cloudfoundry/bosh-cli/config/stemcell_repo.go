package config

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
)

type StemcellRepo interface // StemcellRepo persists stemcells metadata
{
	UpdateCurrent(recordID string) error
	FindCurrent() (StemcellRecord, bool, error)
	ClearCurrent() error
	Save(name, version, cid string) (StemcellRecord, error)
	Find(name, version string) (StemcellRecord, bool, error)
	All() ([]StemcellRecord, error)
	Delete(StemcellRecord) error
}

type stemcellRepo struct {
	deploymentStateService DeploymentStateService
	uuidGenerator          boshuuid.Generator
}

func NewStemcellRepo(deploymentStateService DeploymentStateService, uuidGenerator boshuuid.Generator) StemcellRepo {
	return stemcellRepo{
		deploymentStateService: deploymentStateService,
		uuidGenerator:          uuidGenerator,
	}
}

func (r stemcellRepo) Save(name, version, cid string) (StemcellRecord, error) {
	stemcellRecord := StemcellRecord{}

	err := r.updateConfig(func(config *DeploymentState) error {
		records := config.Stemcells
		if records == nil {
			records = []StemcellRecord{}
		}

		newRecord := StemcellRecord{
			Name:    name,
			Version: version,
			CID:     cid,
		}
		var err error
		newRecord.ID, err = r.uuidGenerator.Generate()
		if err != nil {
			return bosherr.WrapError(err, "Generating stemcell id")
		}

		for _, oldRecord := range records {
			if oldRecord.Name == newRecord.Name && oldRecord.Version == newRecord.Version {
				return bosherr.Errorf("Failed to save stemcell record '%s' (duplicate name/version), existing record found '%s'", newRecord, oldRecord)
			}
		}

		records = append(records, newRecord)
		config.Stemcells = records

		stemcellRecord = newRecord

		return nil
	})

	return stemcellRecord, err
}

func (r stemcellRepo) Find(name, version string) (StemcellRecord, bool, error) {
	_, records, err := r.load()
	if err != nil {
		return StemcellRecord{}, false, err
	}

	for _, oldRecord := range records {
		if oldRecord.Name == name && oldRecord.Version == version {
			return oldRecord, true, nil
		}
	}
	return StemcellRecord{}, false, nil
}

func (r stemcellRepo) FindCurrent() (StemcellRecord, bool, error) {
	deploymentState, err := r.deploymentStateService.Load()
	if err != nil {
		return StemcellRecord{}, false, bosherr.WrapError(err, "Loading existing config")
	}

	currentDiskID := deploymentState.CurrentStemcellID
	if currentDiskID == "" {
		return StemcellRecord{}, false, nil
	}

	for _, oldRecord := range deploymentState.Stemcells {
		if oldRecord.ID == currentDiskID {
			return oldRecord, true, nil
		}
	}

	return StemcellRecord{}, false, nil
}

func (r stemcellRepo) All() ([]StemcellRecord, error) {
	deploymentState, err := r.deploymentStateService.Load()
	if err != nil {
		return []StemcellRecord{}, bosherr.WrapError(err, "Loading existing config")
	}

	return deploymentState.Stemcells, nil
}

func (r stemcellRepo) Delete(stemcellRecord StemcellRecord) error {
	config, records, err := r.load()
	if err != nil {
		return err
	}

	newRecords := []StemcellRecord{}
	for _, record := range records {
		if stemcellRecord.ID != record.ID {
			newRecords = append(newRecords, record)
		}
	}

	config.Stemcells = newRecords

	if config.CurrentStemcellID == stemcellRecord.ID {
		config.CurrentStemcellID = ""
	}

	err = r.deploymentStateService.Save(config)
	if err != nil {
		return bosherr.WrapError(err, "Saving config")
	}

	return nil
}

func (r stemcellRepo) UpdateCurrent(recordID string) error {
	return r.updateConfig(func(config *DeploymentState) error {
		found := false
		for _, oldRecord := range config.Stemcells {
			if oldRecord.ID == recordID {
				found = true
			}
		}
		if !found {
			return bosherr.Errorf("Verifying stemcell record exists with id '%s'", recordID)
		}

		config.CurrentStemcellID = recordID

		return nil
	})
}

func (r stemcellRepo) ClearCurrent() error {
	return r.updateConfig(func(config *DeploymentState) error {
		config.CurrentStemcellID = ""

		return nil
	})
}

func (r stemcellRepo) updateConfig(updateFunc func(*DeploymentState) error) error {
	deploymentState, err := r.deploymentStateService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Loading existing config")
	}

	err = updateFunc(&deploymentState)
	if err != nil {
		return err
	}

	err = r.deploymentStateService.Save(deploymentState)
	if err != nil {
		return bosherr.WrapError(err, "Saving new config")
	}

	return nil
}

func (r stemcellRepo) load() (DeploymentState, []StemcellRecord, error) {
	deploymentState, err := r.deploymentStateService.Load()
	if err != nil {
		return deploymentState, []StemcellRecord{}, bosherr.WrapError(err, "Loading existing config")
	}

	records := deploymentState.Stemcells
	if records == nil {
		return deploymentState, []StemcellRecord{}, nil
	}

	return deploymentState, records, nil
}

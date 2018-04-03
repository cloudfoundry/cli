package fakes

import (
	biproperty "github.com/cloudfoundry/bosh-utils/property"
)

type FakeDisk struct {
	cid string

	NeedsMigrationInputs []NeedsMigrationInput
	needsMigrationOutput needsMigrationOutput

	DeleteCalledTimes int
	deleteErr         error
}

type NeedsMigrationInput struct {
	Size            int
	CloudProperties biproperty.Map
}

type needsMigrationOutput struct {
	needsMigration bool
}

func NewFakeDisk(cid string) *FakeDisk {
	return &FakeDisk{
		cid:                  cid,
		NeedsMigrationInputs: []NeedsMigrationInput{},
	}
}

func (d *FakeDisk) CID() string {
	return d.cid
}

func (d *FakeDisk) NeedsMigration(size int, cloudProperties biproperty.Map) bool {
	d.NeedsMigrationInputs = append(d.NeedsMigrationInputs, NeedsMigrationInput{
		Size:            size,
		CloudProperties: cloudProperties,
	})

	return d.needsMigrationOutput.needsMigration
}

func (d *FakeDisk) Delete() error {
	d.DeleteCalledTimes++
	return d.deleteErr
}

func (d *FakeDisk) SetNeedsMigrationBehavior(needsMigration bool) {
	d.needsMigrationOutput = needsMigrationOutput{
		needsMigration: needsMigration,
	}
}

func (d *FakeDisk) SetDeleteBehavior(err error) {
	d.deleteErr = err
}

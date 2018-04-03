package fakes

import (
	bidisk "github.com/cloudfoundry/bosh-cli/deployment/disk"
	bideplmanifest "github.com/cloudfoundry/bosh-cli/deployment/manifest"
	biui "github.com/cloudfoundry/bosh-cli/ui"
)

type FakeManager struct {
	CreateInputs []CreateInput
	CreateDisk   bidisk.Disk
	CreateErr    error

	findCurrentOutput findCurrentOutput

	DeleteUnusedCalledTimes int
	DeleteUnusedErr         error

	findUnusedOutput findUnusedOutput
}

type CreateInput struct {
	DiskPool   bideplmanifest.DiskPool
	InstanceID string
}

type findCurrentOutput struct {
	Disks []bidisk.Disk
	Err   error
}

type findUnusedOutput struct {
	disks []bidisk.Disk
	err   error
}

func NewFakeManager() *FakeManager {
	return &FakeManager{}
}

func (m *FakeManager) Create(diskPool bideplmanifest.DiskPool, instanceID string) (bidisk.Disk, error) {
	input := CreateInput{
		DiskPool:   diskPool,
		InstanceID: instanceID,
	}
	m.CreateInputs = append(m.CreateInputs, input)

	return m.CreateDisk, m.CreateErr
}

func (m *FakeManager) FindCurrent() ([]bidisk.Disk, error) {
	return m.findCurrentOutput.Disks, m.findCurrentOutput.Err
}

func (m *FakeManager) FindUnused() ([]bidisk.Disk, error) {
	return m.findUnusedOutput.disks, m.findUnusedOutput.err
}

func (m *FakeManager) DeleteUnused(eventLogStage biui.Stage) error {
	m.DeleteUnusedCalledTimes++
	return m.DeleteUnusedErr
}

func (m *FakeManager) SetFindCurrentBehavior(disks []bidisk.Disk, err error) {
	m.findCurrentOutput = findCurrentOutput{
		Disks: disks,
		Err:   err,
	}
}

func (m *FakeManager) SetFindUnusedBehavior(
	disks []bidisk.Disk,
	err error,
) {
	m.findUnusedOutput = findUnusedOutput{
		disks: disks,
		err:   err,
	}
}

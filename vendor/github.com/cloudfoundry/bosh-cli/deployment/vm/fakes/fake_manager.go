package fakes

import (
	bideplmanifest "github.com/cloudfoundry/bosh-cli/deployment/manifest"
	bivm "github.com/cloudfoundry/bosh-cli/deployment/vm"
	bistemcell "github.com/cloudfoundry/bosh-cli/stemcell"
)

type CreateInput struct {
	Stemcell bistemcell.CloudStemcell
	Manifest bideplmanifest.Manifest
}

type FakeManager struct {
	CreateInput CreateInput
	CreateVM    bivm.VM
	CreateErr   error

	findCurrentBehaviour findCurrentOutput
}

type findCurrentOutput struct {
	vm    bivm.VM
	found bool
	err   error
}

func NewFakeManager() *FakeManager {
	return &FakeManager{}
}

func (m *FakeManager) FindCurrent() (bivm.VM, bool, error) {
	return m.findCurrentBehaviour.vm, m.findCurrentBehaviour.found, m.findCurrentBehaviour.err
}

func (m *FakeManager) Create(stemcell bistemcell.CloudStemcell, deploymentManifest bideplmanifest.Manifest) (bivm.VM, error) {
	input := CreateInput{
		Stemcell: stemcell,
		Manifest: deploymentManifest,
	}
	m.CreateInput = input

	return m.CreateVM, m.CreateErr
}

func (m *FakeManager) SetFindCurrentBehavior(vm bivm.VM, found bool, err error) {
	m.findCurrentBehaviour = findCurrentOutput{
		vm:    vm,
		found: found,
		err:   err,
	}
}

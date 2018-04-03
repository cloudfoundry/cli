package fakes

import (
	bicloud "github.com/cloudfoundry/bosh-cli/cloud"
	bidisk "github.com/cloudfoundry/bosh-cli/deployment/disk"
)

type NewManagerInput struct {
	Cloud bicloud.Cloud
}

type FakeManagerFactory struct {
	NewManagerInputs  []NewManagerInput
	NewManagerManager bidisk.Manager
}

func NewFakeManagerFactory() *FakeManagerFactory {
	return &FakeManagerFactory{
		NewManagerInputs: []NewManagerInput{},
	}
}

func (f *FakeManagerFactory) NewManager(cloud bicloud.Cloud) bidisk.Manager {
	input := NewManagerInput{
		Cloud: cloud,
	}
	f.NewManagerInputs = append(f.NewManagerInputs, input)

	return f.NewManagerManager
}
